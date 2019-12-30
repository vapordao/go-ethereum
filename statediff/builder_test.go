// Copyright 2019 The go-ethereum Authors
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

package statediff_test

import (
	"bytes"
	"math/big"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
)

var (
	contractLeafKey                common.Hash
	emptyAccountDiffEventualMap    = make([]statediff.AccountDiff, 0)
	emptyAccountDiffIncrementalMap = make([]statediff.AccountDiff, 0)
	block0, block1, block2, block3 *types.Block
	builder                        statediff.Builder
	miningReward                   = int64(2000000000000000000)
	burnAddress                    = common.HexToAddress("0x0")
	burnLeafKey                    = testhelpers.AddressToLeafKey(burnAddress)
	block0Hash                     = common.HexToHash("0xd1721cfd0b29c36fd7a68f25c128e86413fb666a6e1d68e89b875bd299262661")
	block1Hash                     = common.HexToHash("0xbbe88de60ba33a3f18c0caa37d827bfb70252e19e40a07cd34041696c35ecb1a")
	block2Hash                     = common.HexToHash("0xbc57256a9d5ed6055d924da2b40118194c5a567025dc4d97c2267ab25afa72f9")
	block3Hash                     = common.HexToHash("0x93cf01a3e6a81b796f9b96509182d2a20e95ca48c990d2777628a5ba7e060312")
	balanceChange10000             = int64(10000)
	balanceChange1000              = int64(1000)
	block1BankBalance              = int64(99990000)
	block1Account1Balance          = int64(10000)
	block2Account2Balance          = int64(1000)
	nonce0                         = uint64(0)
	nonce1                         = uint64(1)
	nonce2                         = uint64(2)
	nonce3                         = uint64(3)
	originalContractRoot           = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
	contractContractRoot           = "0x821e2556a290c86405f8160a2d662042a431ba456b9db265c79bb837c04be5f0"
	newContractRoot                = "0x71e0d14b2b93e5c7f9748e69e1fe5f17498a1c3ac3cec29f96af13d7f8a4e070"
	originalStorageLocation        = common.HexToHash("0")
	originalStorageKey             = crypto.Keccak256Hash(originalStorageLocation[:]).Bytes()
	updatedStorageLocation         = common.HexToHash("2")
	updatedStorageKey              = crypto.Keccak256Hash(updatedStorageLocation[:]).Bytes()
	originalStorageValue           = common.Hex2Bytes("01")
	updatedStorageValue            = common.Hex2Bytes("03")
	config                         = statediff.Config{}

	account1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(balanceChange10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	burnAccount1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	bankAccount1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(testhelpers.TestBankFunds.Int64() - balanceChange10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	contractAccount, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0x753f98a8d4328b15636e46f66f2cb4bc860100aa17967cc145fcd17d1d4710ea").Bytes(),
		Root:     common.HexToHash(contractContractRoot),
	})
	bankAccount2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce2,
		Balance:  big.NewInt(block1BankBalance - balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account3, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce2,
		Balance:  big.NewInt(block1Account1Balance - balanceChange1000 + balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	burnAccount2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(miningReward + miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account4, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(block2Account2Balance + miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	contractAccount2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0x753f98a8d4328b15636e46f66f2cb4bc860100aa17967cc145fcd17d1d4710ea").Bytes(),
		Root:     common.HexToHash(newContractRoot),
	})
	bankAccount3, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce3,
		Balance:  big.NewInt(99989000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
)

type arguments struct {
	oldStateRoot common.Hash
	newStateRoot common.Hash
	blockNumber  *big.Int
	blockHash    common.Hash
}

func TestBuilder(t *testing.T) {
	_, blockMap, chain := testhelpers.MakeChain(3, testhelpers.Genesis)
	contractLeafKey = testhelpers.AddressToLeafKey(testhelpers.ContractAddr)
	defer chain.Stop()
	block0 = blockMap[block0Hash]
	block1 = blockMap[block1Hash]
	block2 = blockMap[block2Hash]
	block3 = blockMap[block3Hash]
	builder = statediff.NewBuilder(testhelpers.Testdb, chain, config)

	var tests = []struct {
		name              string
		startingArguments arguments
		expected          *statediff.StateDiff
	}{
		{
			"testEmptyDiff",
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block0.Root(),
				blockNumber:  block0.Number(),
				blockHash:    block0Hash,
			},
			&statediff.StateDiff{
				BlockNumber:     block0.Number(),
				BlockHash:       block0Hash,
				CreatedAccounts: emptyAccountDiffEventualMap,
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: emptyAccountDiffIncrementalMap,
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block1.Root(),
				blockNumber:  block1.Number(),
				blockHash:    block1Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   burnLeafKey.Bytes(),
						Value: burnAccount1,
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1,
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccount1,
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock2",
			//1000 transferred from testBankAddress to account1Addr
			//1000 transferred from account1Addr to account2Addr
			arguments{
				oldStateRoot: block1.Root(),
				newStateRoot: block2.Root(),
				blockNumber:  block2.Number(),
				blockHash:    block2Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccount,
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   originalStorageKey,
								Value: originalStorageValue,
							},
						},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account2LeafKey.Bytes(),
						Value: account2,
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccount2,
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   burnLeafKey.Bytes(),
						Value: burnAccount2,
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account3,
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			arguments{
				oldStateRoot: block2.Root(),
				newStateRoot: block3.Root(),
				blockNumber:  block3.Number(),
				blockHash:    block3.Hash(),
			},
			&statediff.StateDiff{
				BlockNumber:     block3.Number(),
				BlockHash:       block3.Hash(),
				CreatedAccounts: []statediff.AccountDiff{},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccount3,
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccount2,
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   updatedStorageKey,
								Value: updatedStorageValue,
							},
						},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account2LeafKey.Bytes(),
						Value: account4,
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		arguments := test.startingArguments
		diff, err := builder.BuildStateDiff(arguments.oldStateRoot, arguments.newStateRoot, arguments.blockNumber, arguments.blockHash)
		if err != nil {
			t.Error(err)
		}
		receivedStateDiffRlp, err := rlp.EncodeToBytes(diff)
		if err != nil {
			t.Error(err)
		}
		expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
		if err != nil {
			t.Error(err)
		}
		sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
		sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
		if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
			t.Logf("Test failed: %s", test.name)
			t.Errorf("actual state diff rlp: %+v\nexpected state diff rlp: %+v", receivedStateDiffRlp, expectedStateDiffRlp)
		}
	}
}

func TestBuilderWithWatchedAddressList(t *testing.T) {
	_, blockMap, chain := testhelpers.MakeChain(3, testhelpers.Genesis)
	contractLeafKey = testhelpers.AddressToLeafKey(testhelpers.ContractAddr)
	defer chain.Stop()
	block0 = blockMap[block0Hash]
	block1 = blockMap[block1Hash]
	block2 = blockMap[block2Hash]
	block3 = blockMap[block3Hash]
	config.WatchedAddresses = []string{testhelpers.Account1Addr.Hex(), testhelpers.ContractAddr.Hex()}
	builder = statediff.NewBuilder(testhelpers.Testdb, chain, config)

	var tests = []struct {
		name              string
		startingArguments arguments
		expected          *statediff.StateDiff
	}{
		{
			"testEmptyDiff",
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block0.Root(),
				blockNumber:  block0.Number(),
				blockHash:    block0Hash,
			},
			&statediff.StateDiff{
				BlockNumber:     block0.Number(),
				BlockHash:       block0Hash,
				CreatedAccounts: emptyAccountDiffEventualMap,
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: emptyAccountDiffIncrementalMap,
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block1.Root(),
				blockNumber:  block1.Number(),
				blockHash:    block1Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1,
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{},
			},
		},
		{
			"testBlock2",
			//1000 transferred from testBankAddress to account1Addr
			//1000 transferred from account1Addr to account2Addr
			arguments{
				oldStateRoot: block1.Root(),
				newStateRoot: block2.Root(),
				blockNumber:  block2.Number(),
				blockHash:    block2Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccount,
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   originalStorageKey,
								Value: originalStorageValue,
							},
						},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account3,
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			arguments{
				oldStateRoot: block2.Root(),
				newStateRoot: block3.Root(),
				blockNumber:  block3.Number(),
				blockHash:    block3.Hash(),
			},
			&statediff.StateDiff{
				BlockNumber:     block3.Number(),
				BlockHash:       block3.Hash(),
				CreatedAccounts: []statediff.AccountDiff{},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccount2,
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   updatedStorageKey,
								Value: updatedStorageValue,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		arguments := test.startingArguments
		diff, err := builder.BuildStateDiff(arguments.oldStateRoot, arguments.newStateRoot, arguments.blockNumber, arguments.blockHash)
		if err != nil {
			t.Error(err)
		}
		receivedStateDiffRlp, err := rlp.EncodeToBytes(diff)
		if err != nil {
			t.Error(err)
		}
		expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
		if err != nil {
			t.Error(err)
		}
		sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
		sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
		if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
			t.Logf("Test failed: %s", test.name)
			t.Errorf("actual state diff rlp: %+v\nexpected state diff rlp: %+v", receivedStateDiffRlp, expectedStateDiffRlp)
		}
	}
}

/*
contract test {

    uint256[100] data;

	constructor() public {
		data = [1];
	}

    function Put(uint256 addr, uint256 value) {
        data[addr] = value;
    }

    function Get(uint256 addr) constant returns (uint256 value) {
        return data[addr];
    }
}
*/
