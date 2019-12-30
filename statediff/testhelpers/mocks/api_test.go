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

package mocks

import (
	"bytes"
	"math/big"
	"sort"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
)

var (
	config = statediff.Config{}
	block0, block1              *types.Block
	burnLeafKey                 = testhelpers.AddressToLeafKey(common.HexToAddress("0x0"))
	emptyAccountDiffEventualMap = make([]statediff.AccountDiff, 0)
	account1, _                 = rlp.EncodeToBytes(state.Account{
		Nonce:    uint64(0),
		Balance:  big.NewInt(10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	})
	burnAccount1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    uint64(0),
		Balance:  big.NewInt(2000000000000000000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	})
	bankAccount1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    uint64(1),
		Balance:  big.NewInt(testhelpers.TestBankFunds.Int64() - 10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	})
)

func TestAPI(t *testing.T) {
	_, blockMap, chain := testhelpers.MakeChain(3, testhelpers.Genesis)
	defer chain.Stop()
	block0Hash := common.HexToHash("0xd1721cfd0b29c36fd7a68f25c128e86413fb666a6e1d68e89b875bd299262661")
	block1Hash := common.HexToHash("0xbbe88de60ba33a3f18c0caa37d827bfb70252e19e40a07cd34041696c35ecb1a")
	block0 = blockMap[block0Hash]
	block1 = blockMap[block1Hash]
	blockChan := make(chan *types.Block)
	parentBlockChain := make(chan *types.Block)
	serviceQuitChan := make(chan bool)
	mockService := MockStateDiffService{
		Mutex:           sync.Mutex{},
		Builder:         statediff.NewBuilder(testhelpers.Testdb, chain, config),
		BlockChan:       blockChan,
		ParentBlockChan: parentBlockChain,
		QuitChan:        serviceQuitChan,
		Subscriptions:   make(map[rpc.ID]statediff.Subscription),
	}
	mockService.Start(nil)
	id := rpc.NewID()
	payloadChan := make(chan statediff.Payload)
	quitChan := make(chan bool)
	mockService.Subscribe(id, payloadChan, quitChan)
	blockChan <- block1
	parentBlockChain <- block0
	expectedStateDiff := statediff.StateDiff{
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
	}
	expectedStateDiffBytes, err := rlp.EncodeToBytes(expectedStateDiff)
	if err != nil {
		t.Error(err)
	}
	sort.Slice(expectedStateDiffBytes, func(i, j int) bool { return expectedStateDiffBytes[i] < expectedStateDiffBytes[j] })

	select {
	case payload := <-payloadChan:
		sort.Slice(payload.StateDiffRlp, func(i, j int) bool { return payload.StateDiffRlp[i] < payload.StateDiffRlp[j] })
		if !bytes.Equal(payload.StateDiffRlp, expectedStateDiffBytes) {
			t.Errorf("payload does not have expected state diff\r\actual state diff rlp: %v\r\nexpected state diff rlp: %v", payload.StateDiffRlp, expectedStateDiffBytes)
		}
	case <-quitChan:
		t.Errorf("channel quit before delivering payload")
	}
}
