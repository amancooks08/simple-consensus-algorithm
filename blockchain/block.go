package blockchain

import (
	"math/rand"
	"simple-consensus-algorithm-impl/domain"
	"sync"

	"github.com/sirupsen/logrus"
)

type Block struct {
	Data int64
}

type Votes struct {
	For     int
	Against int
}

type VotesMap struct {
	Count map[int64]Votes
	sync.Mutex
}

// We'll create a map of Blocks and the Votes they received
var BlockVotes = &VotesMap{
	Count: make(map[int64]Votes),
}

func NewBlock(nodeID int64, instance *Blockchain) Block {

	var data int64
	var newBlock Block
	for {
		// Create a new block with some random data and the nodeID.
		// We are generating a random number between 1 and the last block's data * nodeID
		// We are making an exception for the first node, where we are multiplying the
		// last block's data by 5, since the first node's ID is 1.
		if nodeID == 1 {
			data = int64(rand.Int63n(instance.getLastBlockData()+1) * 5)
		} else {
			data = int64(rand.Int63n(instance.getLastBlockData())+1) * nodeID
		}

		// Also add a check if the data is smaller than the last block's data
		// If it is, we will continue to generate a new block
		if data < instance.getLastBlockData() {
			continue
		}
		// We check if there is an entry for the data in the map
		// If there is it means the block has already been proposed
		// We will not propose the same block again
		// We will only propose a block if it is not in the map
		if _, ok := BlockVotes.getVotesFromMap(data); !ok {
			newBlock.Data = data
			break
		}
	}

	// Update the votes for the new block in the map
	BlockVotes.updateVotesInMap(newBlock, Votes{
		For:     1,
		Against: 0,
	})

	return newBlock
}


// This function will validate the block
func (block Block) IsBlockValid(bc *Blockchain, data int64) string {

	// Check if the data is divisible by the last block's data
	if data%bc.getLastBlockData() == 0 {
		return domain.BLOCK_IS_VALID
	} else {
		return domain.BLOCK_IS_INVALID
	}
}

func (block Block) VoteForBlock(instance *Blockchain, msg string) {

	// Get the votes for the data from the map
	votes, ok := BlockVotes.getVotesFromMap(block.Data)
	if !ok {
		// If the block is not in the map, then we initialize the votes
		votes = Votes{}

		// We will initialize the votes with 1 for the block
		votes.For = 1
	}

	// Update the vote count based on the message
	if msg == domain.BLOCK_IS_VALID {
		votes.For++
	} else {
		votes.Against++
	}

	// Update the BlockVotes map with the updated vote count
	BlockVotes.updateVotesInMap(block, votes)

	// Check if the block has received 3 votes for approval
	var voteCount Votes
	if vote, ok := BlockVotes.getVotesFromMap(block.Data); ok {
		voteCount = vote
	}
	if voteCount.For == 4 {

		logrus.Infof("Block %d with the data: %d has been approved", instance.GetHeight()+1, block.Data)

		// Add the block to the blockchain
		instance.addBlock(block)
	} else if voteCount.Against > 0 {

		// The block has been rejected
		logrus.Infof("Block with the data: %d has been rejected", block.Data)
	}
}

// This function will update the votes for a block in the map
func (BlockVotes *VotesMap) updateVotesInMap(block Block, vote Votes) {
	BlockVotes.Lock()
	defer BlockVotes.Unlock()
	BlockVotes.Count[block.Data] = vote
}

// This function will get us the votes for a block from the map
func (BlockVotes *VotesMap) getVotesFromMap(data int64) (Votes, bool) {
	BlockVotes.Lock()
	defer BlockVotes.Unlock()
	if vote, ok := BlockVotes.Count[data]; ok {
		return vote, true
	}
	return Votes{}, false
}
