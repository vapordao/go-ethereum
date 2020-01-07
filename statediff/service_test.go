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
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/testhelpers/mocks"
)

func TestServiceLoop(t *testing.T) {
	testErrorInStateChangeEventLoop(t)
}

var (
	stateChangeEventCh = make(chan core.StateChangeEvent, 1)

	testBlock1 = types.NewBlock(&types.Header{}, nil, nil, nil)
	testBlock2 = types.NewBlock(&types.Header{}, nil, nil, nil)
	testBlock3 = types.NewBlock(&types.Header{}, nil, nil, nil)

	account1Address = common.HexToAddress("0x1")
	accountBlock1 = state.Account{
		Nonce:    0,
		Balance:  big.NewInt(100),
		Root:     common.HexToHash("0x01"),
	}
	modifiedAccount1 = state.ModifiedAccount{
		Account: accountBlock1,
		Storage: nil,
	}

	event1 = core.StateChangeEvent{
		Block:            testBlock1,
		StateChanges: state.StateChanges{
			ModifiedAccounts: map[common.Address]state.ModifiedAccount{account1Address: modifiedAccount1},
		},
	}
	event2 = core.StateChangeEvent{
		Block:            testBlock2,
		StateChanges: state.StateChanges{
			ModifiedAccounts: nil,
		},
	}
	event3 = core.StateChangeEvent{
		Block:            testBlock3,
		StateChanges: state.StateChanges{
			ModifiedAccounts: nil,
		},
	}
)

func testErrorInStateChangeEventLoop(t *testing.T) {
	blockChain := mocks.BlockChain{}
	service := statediff.Service{
		Mutex:         sync.Mutex{},
		BlockChain:    &blockChain,
		QuitChan:      make(chan bool),
		Subscriptions: make(map[rpc.ID]statediff.Subscription),
	}
	payloadChan := make(chan statediff.Payload, 2)
	quitChan := make(chan bool)
	service.Subscribe(rpc.NewID(), payloadChan, quitChan)
	blockChain.SetStateChangeEvents([]core.StateChangeEvent{event1, event2, event3})

	payloads := make([]statediff.Payload, 0, 2)
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		for i := 0; i < 2; i++ {
			select {
			case payload := <-payloadChan:
				payloads = append(payloads, payload)
			case <-quitChan:
			}
		}
		wg.Done()
	}()

	service.Loop(stateChangeEventCh)
	wg.Wait()
	if len(payloads) != 2 {
		t.Error("Test failure:", t.Name())
		t.Logf("Actual number of payloads does not equal expected.\nactual: %+v\nexpected: 2", len(payloads))
	}

	accountBlock1Bytes, err  := rlp.EncodeToBytes(accountBlock1)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode state diff to bytes")
	}

	accountDiff := statediff.AccountDiff{
		Key:     account1Address[:],
		Value:   accountBlock1Bytes,
		Storage: nil,
	}
	stateDiff := statediff.StateDiff{
		BlockNumber:     testBlock1.Number(),
		BlockHash:       testBlock1.Hash(),
		UpdatedAccounts: []statediff.AccountDiff{accountDiff},
	}

	expectedStateDiffRlp, err := rlp.EncodeToBytes(stateDiff)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode state diff to bytes")
	}

	emptyStateDiffRlp := []byte{ 229, 128, 160, 177, 89, 160, 119, 252, 42, 247, 155, 154, 156, 116, 140, 156, 14, 80, 255, 149, 183, 76, 50, 148, 110, 213, 36, 24, 252, 192, 147, 208, 149, 63, 38, 192, 192, 192 }
	expectedPayloads := []statediff.Payload{{
		StateDiffRlp: expectedStateDiffRlp,
	}, {StateDiffRlp: emptyStateDiffRlp}}
	if !reflect.DeepEqual(payloads, expectedPayloads) {
		t.Error("Test failure:", t.Name())
		t.Logf("Actual payload equal expected.\nactual:%+v\nexpected: %+v", payloads, expectedPayloads)
	}
}
