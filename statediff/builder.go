// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains a batch of utility type declarations used by the tests. As the node
// operates on unique types, a lot of them are needed to check various features.

package statediff

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

type Builder interface {
	BuildStateDiff(oldStateRoot, newStateRoot common.Hash, blockNumber int64, blockHash common.Hash) (*StateDiff, error)
}

type builder struct {
	chainDB    ethdb.Database
	trieDB     *trie.Database
	cachedTrie *trie.Trie
}

func NewBuilder(db ethdb.Database) *builder {
	return &builder{
		chainDB: db,
		trieDB:  trie.NewDatabase(db),
	}
}

func (sdb *builder) BuildStateDiff(oldStateRoot, newStateRoot common.Hash, blockNumber int64, blockHash common.Hash) (*StateDiff, error) {
	// Generate tries for old and new states
	oldTrie, err := trie.New(oldStateRoot, sdb.trieDB)
	if err != nil {
		return nil, err
	}
	newTrie, err := trie.New(newStateRoot, sdb.trieDB)
	if err != nil {
		return nil, err
	}

	// Find created accounts
	oldIt := oldTrie.NodeIterator([]byte{})
	newIt := newTrie.NodeIterator([]byte{})
	creations, err := sdb.collectDiffNodes(oldIt, newIt)
	if err != nil {
		return nil, err
	}

	// Find deleted accounts
	oldIt = oldTrie.NodeIterator(make([]byte, 0))
	newIt = newTrie.NodeIterator(make([]byte, 0))
	deletions, err := sdb.collectDiffNodes(newIt, oldIt)
	if err != nil {
		return nil, err
	}

	// Find all the diffed keys
	createKeys := sortKeys(creations)
	deleteKeys := sortKeys(deletions)
	updatedKeys := findIntersection(createKeys, deleteKeys)

	// Build and return the statediff
	updatedAccounts, err := sdb.buildDiffIncremental(creations, deletions, &updatedKeys)
	if err != nil {
		return nil, err
	}
	createdAccounts, err := sdb.buildDiffEventual(creations, true)
	if err != nil {
		return nil, err
	}
	deletedAccounts, err := sdb.buildDiffEventual(deletions, false)
	if err != nil {
		return nil, err
	}

	return &StateDiff{
		BlockNumber:     blockNumber,
		BlockHash:       blockHash,
		CreatedAccounts: createdAccounts,
		DeletedAccounts: deletedAccounts,
		UpdatedAccounts: updatedAccounts,
	}, nil
}

func (sdb *builder) collectDiffNodes(a, b trie.NodeIterator) (map[common.Address]*state.Account, error) {
	var diffAccounts map[common.Address]*state.Account
	it, _ := trie.NewDifferenceIterator(a, b)

	for {
		log.Debug("Current Path and Hash", "path", pathToStr(it), "hashold", common.Hash(it.Hash()))
		if it.Leaf() {

			// lookup address
			path := make([]byte, len(it.Path())-1)
			copy(path, it.Path())
			addr, err := sdb.addressByPath(path)
			if err != nil {
				log.Error("Error looking up address via path", "path", path, "error", err)
				return nil, err
			}

			// lookup account state
			var account state.Account
			if err := rlp.DecodeBytes(it.LeafBlob(), &account); err != nil {
				log.Error("Error looking up account via address", "address", addr, "error", err)
				return nil, err
			}

			// record account to diffs (creation if we are looking at new - old; deletion if old - new)
			log.Debug("Account lookup successful", "address", addr, "account", account)
			diffAccounts[*addr] = &account
		}
		cont := it.Next(true)
		if !cont {
			break
		}
	}
	return diffAccounts, nil
}

func (sdb *builder) buildDiffEventual(accounts map[common.Address]*state.Account, created bool) (map[common.Address]AccountDiffEventual, error) {
	accountDiffs := make(map[common.Address]AccountDiffEventual)
	for addr, val := range accounts {
		sr := val.Root
		if storageDiffs, err := sdb.buildStorageDiffsEventual(sr, created); err != nil {
			log.Error("Failed building eventual storage diffs", "Address", val, "error", err)
			return nil, err
		} else {
			code := ""
			codeBytes, err := sdb.chainDB.Get(val.CodeHash)
			if err == nil && len(codeBytes) != 0 {
				code = common.ToHex(codeBytes)
			} else {
				log.Debug("No code field.", "codehash", val.CodeHash, "Address", val, "error", err)
			}
			codeHash := common.ToHex(val.CodeHash)
			if created {
				nonce := diffUint64{
					NewValue: &val.Nonce,
				}

				balance := diffBigInt{
					NewValue: val.Balance,
				}

				hexRoot := val.Root.Hex()
				contractRoot := diffString{
					NewValue: &hexRoot,
				}
				accountDiffs[addr] = AccountDiffEventual{
					Nonce:        nonce,
					Balance:      balance,
					CodeHash:     codeHash,
					Code:         code,
					ContractRoot: contractRoot,
					Storage:      storageDiffs,
				}
			} else {
				nonce := diffUint64{
					OldValue: &val.Nonce,
				}
				balance := diffBigInt{
					OldValue: val.Balance,
				}
				hexRoot := val.Root.Hex()
				contractRoot := diffString{
					OldValue: &hexRoot,
				}
				accountDiffs[addr] = AccountDiffEventual{
					Nonce:        nonce,
					Balance:      balance,
					CodeHash:     codeHash,
					ContractRoot: contractRoot,
					Storage:      storageDiffs,
				}
			}
		}
	}
	return accountDiffs, nil
}

func (sdb *builder) buildDiffIncremental(creations map[common.Address]*state.Account, deletions map[common.Address]*state.Account, updatedKeys *[]string) (map[common.Address]AccountDiffIncremental, error) {
	updatedAccounts := make(map[common.Address]AccountDiffIncremental)
	for _, val := range *updatedKeys {
		createdAcc := creations[common.HexToAddress(val)]
		deletedAcc := deletions[common.HexToAddress(val)]
		oldSR := deletedAcc.Root
		newSR := createdAcc.Root
		if storageDiffs, err := sdb.buildStorageDiffsIncremental(oldSR, newSR); err != nil {
			log.Error("Failed building storage diffs", "Address", val, "error", err)
			return nil, err
		} else {
			nonce := diffUint64{
				NewValue: &createdAcc.Nonce,
				OldValue: &deletedAcc.Nonce,
			}

			balance := diffBigInt{
				NewValue: createdAcc.Balance,
				OldValue: deletedAcc.Balance,
			}
			codeHash := common.ToHex(createdAcc.CodeHash)

			nHexRoot := createdAcc.Root.Hex()
			oHexRoot := deletedAcc.Root.Hex()
			contractRoot := diffString{
				NewValue: &nHexRoot,
				OldValue: &oHexRoot,
			}

			updatedAccounts[common.HexToAddress(val)] = AccountDiffIncremental{
				Nonce:        nonce,
				Balance:      balance,
				CodeHash:     codeHash,
				ContractRoot: contractRoot,
				Storage:      storageDiffs,
			}
			delete(creations, common.HexToAddress(val))
			delete(deletions, common.HexToAddress(val))
		}
	}
	return updatedAccounts, nil
}

func (sdb *builder) buildStorageDiffsEventual(sr common.Hash, creation bool) (map[string]diffString, error) {
	log.Debug("Storage Root For Eventual Diff", "root", sr.Hex())
	sTrie, err := trie.New(sr, sdb.trieDB)
	if err != nil {
		return nil, err
	}
	it := sTrie.NodeIterator(make([]byte, 0))
	storageDiffs := make(map[string]diffString)
	for {
		log.Debug("Iterating over state at path ", "path", pathToStr(it))
		if it.Leaf() {
			log.Debug("Found leaf in storage", "path", pathToStr(it))
			path := pathToStr(it)
			value := common.ToHex(it.LeafBlob())
			if creation {
				storageDiffs[path] = diffString{NewValue: &value}
			} else {
				storageDiffs[path] = diffString{OldValue: &value}
			}
		}
		cont := it.Next(true)
		if !cont {
			break
		}
	}
	return storageDiffs, nil
}

func (sdb *builder) buildStorageDiffsIncremental(oldSR common.Hash, newSR common.Hash) (map[string]diffString, error) {
	log.Debug("Storage Roots for Incremental Diff", "old", oldSR.Hex(), "new", newSR.Hex())
	oldTrie, err := trie.New(oldSR, sdb.trieDB)
	if err != nil {
		return nil, err
	}
	newTrie, err := trie.New(newSR, sdb.trieDB)
	if err != nil {
		return nil, err
	}

	oldIt := oldTrie.NodeIterator(make([]byte, 0))
	newIt := newTrie.NodeIterator(make([]byte, 0))
	it, _ := trie.NewDifferenceIterator(oldIt, newIt)
	storageDiffs := make(map[string]diffString)
	for {
		if it.Leaf() {
			log.Debug("Found leaf in storage", "path", pathToStr(it))
			path := pathToStr(it)
			value := common.ToHex(it.LeafBlob())
			if oldVal, err := oldTrie.TryGet(it.LeafKey()); err != nil {
				log.Error("Failed to look up value in oldTrie", "path", path, "error", err)
			} else {
				hexOldVal := common.ToHex(oldVal)
				storageDiffs[path] = diffString{OldValue: &hexOldVal, NewValue: &value}
			}
		}

		cont := it.Next(true)
		if !cont {
			break
		}
	}
	return storageDiffs, nil
}

func (sdb *builder) addressByPath(path []byte) (*common.Address, error) {
	// db := core.PreimageTable(sdb.chainDb)
	log.Debug("Looking up address from path", "path", common.ToHex(append([]byte("secure-key-"), path...)))
	// if addrBytes,err := db.Get(path); err != nil {
	if addrBytes, err := sdb.chainDB.Get(append([]byte("secure-key-"), hexToKeybytes(path)...)); err != nil {
		log.Error("Error looking up address via path", "path", common.ToHex(append([]byte("secure-key-"), path...)), "error", err)
		return nil, err
	} else {
		addr := common.BytesToAddress(addrBytes)
		log.Debug("Address found", "Address", addr)
		return &addr, nil
	}

}
