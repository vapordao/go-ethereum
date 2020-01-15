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
	testWhenThereAreNoStateDiffs(t)
	testWhenThereAreStateAndStorageDiffs(t)
	testWatchedAddresses(t)
}

var (
	stateChangeEventCh = make(chan core.StateChangeEvent, 1)

	testBlock1 = types.NewBlock(&types.Header{}, nil, nil, nil)
	testBlock2 = types.NewBlock(&types.Header{}, nil, nil, nil)
	testBlock3 = types.NewBlock(&types.Header{}, nil, nil, nil)

	testAccount1Address = common.HexToAddress("0x1")
	testAccount1 = state.Account{
		Nonce:   0,
		Balance: big.NewInt(100),
		Root:    common.HexToHash("0x01"),
	}
	modifiedAccount1 = state.ModifiedAccount{
		Account: testAccount1,
		Storage: nil,
	}
	testAccount2Address = common.HexToAddress("0x2")
	testAccount2 = state.Account{
		Nonce:   0,
		Balance: big.NewInt(200),
		Root:    common.HexToHash("0x02"),
	}
	account2StorageKey =common.HexToHash("0x0002")
	account2StorageValue = common.HexToHash("0x00002")
	account2Storage = state.Storage{account2StorageKey: account2StorageValue}
	modifiedAccount2 = state.ModifiedAccount{
		Account: testAccount2,
		Storage: account2Storage,
	}
	event1 = core.StateChangeEvent{
		Block: testBlock1,
		StateChanges: state.StateChanges{
			testAccount1Address: modifiedAccount1,
			testAccount2Address: modifiedAccount2,
		},
	}

	noStateChangeEvent1 = core.StateChangeEvent{Block: testBlock1, StateChanges: state.StateChanges{}}
	event2 = core.StateChangeEvent{Block: testBlock2, StateChanges: state.StateChanges{}}
	event3 = core.StateChangeEvent{Block: testBlock3, StateChanges: state.StateChanges{}}
)

func testWhenThereAreNoStateDiffs(t *testing.T) {
	blockChain := mocks.BlockChain{}
	service := statediff.Service{
		Mutex:         sync.Mutex{},
		BlockChain:    &blockChain,
		QuitChan:      make(chan bool),
		Subscriptions: make(map[rpc.ID]statediff.Subscription),
		WatchedAddresses: []common.Address{}, // when empty, return all diffs
	}
	payloadChan := make(chan statediff.Payload, 2)
	quitChan := make(chan bool)
	service.Subscribe(rpc.NewID(), payloadChan, quitChan)
	blockChain.SetStateChangeEvents([]core.StateChangeEvent{event2, event2, event3})

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
	expectedPayloads := []statediff.Payload{
		{StateDiffRlp: getEmptyStateDiffRlp(*testBlock1, t)},
		{StateDiffRlp: getEmptyStateDiffRlp(*testBlock1, t)},
	}

	// If there are no statediffs the payloads include empty statediffs
	if !reflect.DeepEqual(payloads, expectedPayloads) {
		t.Error("Test failure:", t.Name())
		t.Logf("Actual payload equal expected.\nactual:%+v\nexpected: %+v", payloads, expectedPayloads)
	}
}

func testWhenThereAreStateAndStorageDiffs(t *testing.T) {
	blockChain := mocks.BlockChain{}
	service := statediff.Service{
		Mutex:         sync.Mutex{},
		BlockChain:    &blockChain,
		QuitChan:      make(chan bool),
		Subscriptions: make(map[rpc.ID]statediff.Subscription),
		WatchedAddresses: []common.Address{}, // when empty, return all diffs
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

	stateDiff := statediff.StateDiff{
		BlockNumber:     testBlock1.Number(),
		BlockHash:       testBlock1.Hash(),
		UpdatedAccounts: []statediff.AccountDiff{getAccount1Diff(t), getAccount2Diff(t)},
	}
	expectedStateDiffRlp, err := rlp.EncodeToBytes(stateDiff)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode state diff to bytes")
	}
	expectedPayloads := []statediff.Payload{
		{StateDiffRlp: expectedStateDiffRlp},
		{StateDiffRlp: getEmptyStateDiffRlp(*testBlock1, t)},
	}
	if !reflect.DeepEqual(payloads, expectedPayloads) {
		t.Error("Test failure:", t.Name())
		t.Logf("Actual payload equal expected.\nactual:%+v\nexpected: %+v", payloads, expectedPayloads)
	}
}

func testWatchedAddresses(t *testing.T) {
	blockChain := mocks.BlockChain{}
	service := statediff.Service{
		Mutex:         sync.Mutex{},
		BlockChain:    &blockChain,
		QuitChan:      make(chan bool),
		Subscriptions: make(map[rpc.ID]statediff.Subscription),
		WatchedAddresses: []common.Address{testAccount2Address},
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

	stateDiff := statediff.StateDiff{
		BlockNumber:     testBlock1.Number(),
		BlockHash:       testBlock1.Hash(),
		UpdatedAccounts: []statediff.AccountDiff{getAccount2Diff(t)},
	}
	expectedStateDiffRlp, err := rlp.EncodeToBytes(stateDiff)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode state diff to bytes")
	}
	expectedPayloads := []statediff.Payload{
		{StateDiffRlp: expectedStateDiffRlp},
		{StateDiffRlp: getEmptyStateDiffRlp(*testBlock1, t)},
	}
	if !reflect.DeepEqual(payloads, expectedPayloads) {
		t.Error("Test failure:", t.Name())
		t.Logf("Actual payload equal expected.\nactual:%+v\nexpected: %+v", payloads, expectedPayloads)
	}

}

func getAccount1Diff(t *testing.T) statediff.AccountDiff{
	accountBlock1Bytes, err := rlp.EncodeToBytes(testAccount1)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode state diff to bytes")
	}

	return statediff.AccountDiff{
		Key:     testAccount1Address[:],
		Value:   accountBlock1Bytes,
		Storage: nil,
	}
}

func getAccount2Diff(t *testing.T) statediff.AccountDiff {
	account2Rlp, accountRlpErr := rlp.EncodeToBytes(testAccount2)
	if accountRlpErr != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode account diff")
	}

	storageValueRlp, storageRlpErr := rlp.EncodeToBytes(account2StorageValue)
	if storageRlpErr != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode storgae diff")
	}
	account2StorageDiff := statediff.StorageDiff{
		Key:   account2StorageKey[:],
		Value: storageValueRlp,
	}
	return statediff.AccountDiff{
		Key:     testAccount2Address[:],
		Value:   account2Rlp,
		Storage: []statediff.StorageDiff{account2StorageDiff},
	}
}

func getEmptyStateDiffRlp(block types.Block, t *testing.T) []byte {
	emptyStateDiff := statediff.StateDiff{
		BlockNumber:     block.Number(),
		BlockHash:       block.Hash(),
	}

	emptyStateDiffRlp, err := rlp.EncodeToBytes(emptyStateDiff)
	if err != nil {
		t.Error("Test failure:", t.Name())
		t.Logf("Failed to encode empty state diff to bytes")
	}

	return emptyStateDiffRlp
}
