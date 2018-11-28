package statediff

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"fmt"
)

type StateDiffService struct {
	builder    *builder
	extractor  *extractor
	blockchain *core.BlockChain
}

func NewStateDiffService(db ethdb.Database, blockChain *core.BlockChain) (*StateDiffService, error) {
	config := Config{}
	extractor, _ := NewExtractor(db, config)
	return &StateDiffService{
		blockchain: blockChain,
		extractor: extractor,
	}, nil
}

func (StateDiffService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}

}

func (StateDiffService) APIs() []rpc.API {
	return []rpc.API{}
}

func (sds *StateDiffService) Start(server *p2p.Server) error {
	fmt.Println("starting the state diff service")
	blockChannel := make(chan core.ChainHeadEvent)
	sds.blockchain.SubscribeChainHeadEvent(blockChannel)
	for {
		select {
		case <-blockChannel:
			headOfChainEvent := <-blockChannel
			previousBlock := headOfChainEvent.Block
			//TODO: figure out the best way to get the previous block
			currentBlock := headOfChainEvent.Block
			sds.extractor.ExtractStateDiff(*previousBlock, *currentBlock)
		}
	}
	return nil
}
func (StateDiffService) Stop() error {
	fmt.Println("stopping the state diff service")
	return nil
}