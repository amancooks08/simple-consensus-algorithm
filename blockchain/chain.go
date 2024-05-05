package blockchain

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	Blocks []Block
	sync.Mutex
}

// NewBlockchain initializes the blockchain with the genesis block
func NewBlockchain() *Blockchain{

	// Add the Entry for the Genesis Block
	BlockVotes.Count[4] = Votes{
		For:     4,
		Against: 0,
	}
	return &Blockchain{
		Blocks: []Block{{
			Data: 4,
		}},
	}
}

// A private function to add a block to the blockchain
func (instance *Blockchain) addBlock(block Block) {
	instance.Lock()
	defer instance.Unlock()
	// Add the block to the blockchain
	instance.Blocks = append(instance.Blocks, block)
	logrus.Infof("Block %d added to the blockchain with the data: %d", len(instance.Blocks), block.Data)

	// We print the blockchain every time a block is added, so that we can see/verify the added block.
	printBlockchain(instance)
}

func printBlockchain(instance *Blockchain) {

	for index, block := range instance.Blocks {

		// Print the block ID and the data
		logrus.Infof("Block ID: %d with Data: %d", index+1, block.Data)
	}
}

// getLastBlockData returns the data of the last block in the blockchain
func (instance *Blockchain) getLastBlockData() int64 {
	instance.Lock()
	defer instance.Unlock()
	return instance.Blocks[len(instance.Blocks)-1].Data
}

// GetHeight returns the height of the blockchain, i.e. the number of blocks in the blockchain
func (instance *Blockchain) GetHeight() int {
	instance.Lock()
	defer instance.Unlock()
	return len(instance.Blocks)
}
