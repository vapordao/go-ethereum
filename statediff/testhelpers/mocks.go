package testhelpers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/statediff"
	"errors"
)

var MockError = errors.New("mock error")

type MockBuilder struct {
	OldStateRoot common.Hash
	NewStateRoot common.Hash
	BlockNumber int64
	BlockHash common.Hash
	stateDiff *statediff.StateDiff
	builderError error
}

func (builder *MockBuilder) BuildStateDiff(oldStateRoot, newStateRoot common.Hash, blockNumber int64, blockHash common.Hash) (*statediff.StateDiff, error) {
	builder.OldStateRoot = oldStateRoot
	builder.NewStateRoot = newStateRoot
	builder.BlockNumber = blockNumber
	builder.BlockHash = blockHash

	return builder.stateDiff, builder.builderError
}

func (builder *MockBuilder) SetStateDiffToBuild(stateDiff *statediff.StateDiff) {
	builder.stateDiff = stateDiff
}

func (builder *MockBuilder) SetBuilderError(err error) {
	builder.builderError = err
}

type MockPublisher struct{
	StateDiff *statediff.StateDiff
	publisherError error
}

func (publisher *MockPublisher) PublishStateDiff(sd *statediff.StateDiff) (string, error) {
	publisher.StateDiff = sd
	return "", publisher.publisherError
}

func (publisher *MockPublisher) SetPublisherError(err error) {
	publisher.publisherError = err
}
