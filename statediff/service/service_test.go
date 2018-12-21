package service_test

import (
	"testing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core"
	service2 "github.com/ethereum/go-ethereum/statediff/service"
	"reflect"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"math/rand"
)

type MockExtractor struct {
	ParentBlocks []types.Block
	CurrentBlocks  []types.Block
	extractError  error
}

func (me *MockExtractor) ExtractStateDiff(parent, current types.Block) (string, error) {
	me.ParentBlocks = append(me.ParentBlocks, parent)
	me.CurrentBlocks = append(me.CurrentBlocks, current)

	return "", me.extractError
}

func (me *MockExtractor) SetExtractError(err error) {
	me.extractError = err
}

type MockChain struct {
	ParentHashesLookedUp []common.Hash
	parentBlocksToReturn []*types.Block
	callCount int
}

func (mc *MockChain) SetParentBlockToReturn(blocks []*types.Block) {
	mc.parentBlocksToReturn = blocks
}

func (mc *MockChain) GetBlockByHash(hash common.Hash) *types.Block {
	mc.ParentHashesLookedUp = append(mc.ParentHashesLookedUp, hash)

	var parentBlock types.Block
	if len(mc.parentBlocksToReturn) > 0 {
		parentBlock = *mc.parentBlocksToReturn[mc.callCount]
	}

	mc.callCount++
	return &parentBlock
}

func (MockChain) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	panic("implement me")
}

func TestServiceLoop(t *testing.T) {
	testServiceLoop(t)
}

var (
	eventsChannel = make(chan core.ChainEvent, 10)

	parentHeader1 = types.Header{Number: big.NewInt(rand.Int63())}
	parentHeader2 = types.Header{Number: big.NewInt(rand.Int63())}

	parentBlock1 = types.NewBlock(&parentHeader1, nil, nil, nil)
	parentBlock2 = types.NewBlock(&parentHeader2, nil, nil, nil)

	parentHash1 = parentBlock1.Hash()
	parentHash2 = parentBlock2.Hash()

	header1 = types.Header{ ParentHash: parentHash1 }
	header2 = types.Header{ ParentHash: parentHash2 }

	block1 = types.NewBlock(&header1, nil, nil, nil)
	block2 = types.NewBlock(&header2, nil, nil, nil)

	event1 = core.ChainEvent{ Block: block1 }
	event2 = core.ChainEvent{ Block: block2 }
)

func testServiceLoop(t *testing.T) {
	eventsChannel <- event1
	eventsChannel <- event2

	extractor := MockExtractor{}
	close(eventsChannel)

	blockChain := MockChain{}
	service  := service2.StateDiffService{
		Builder:    nil,
		Extractor:  &extractor,
		BlockChain: &blockChain,
	}

	blockChain.SetParentBlockToReturn([]*types.Block{parentBlock1, parentBlock2})
	service.Loop(eventsChannel)

	//parent and current blocks are passed to the extractor
	expectedCurrentBlocks := []types.Block{*block1, *block2}
	if !reflect.DeepEqual(extractor.CurrentBlocks, expectedCurrentBlocks) {
		t.Errorf("Actual does not equal expected.\nactual:%+v\nexpected: %+v", extractor.CurrentBlocks, expectedCurrentBlocks)
	}
	expectedParentBlocks := []types.Block{*parentBlock1, *parentBlock2}
	if !reflect.DeepEqual(extractor.ParentBlocks, expectedParentBlocks) {
		t.Errorf("Actual does not equal expected.\nactual:%+v\nexpected: %+v", extractor.CurrentBlocks, expectedParentBlocks)
	}

	//look up the parent block from its hash
	expectedHashes := []common.Hash{block1.ParentHash(), block2.ParentHash()}
	if !reflect.DeepEqual(blockChain.ParentHashesLookedUp, expectedHashes) {
		t.Errorf("Actual does not equal expected.\nactual:%+v\nexpected: %+v", blockChain.ParentHashesLookedUp, expectedHashes)
	}
}
