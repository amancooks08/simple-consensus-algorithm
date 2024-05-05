package main

import (
	"context"
	"os"
	"simple-consensus-algorithm-impl/blockchain"
	"simple-consensus-algorithm-impl/node"
	"strconv"
)

func main() {

	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	// defer cancel()
	ctx := context.Background()
	// Start the Blockchain
	blockchainInstance := blockchain.NewBlockchain()

	// Check for Node ID
	if len(os.Args) < 1 {
		panic("Node ID is required")
	}

	// Add the node to the network with the given ID
	nodeID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		panic(err)
	}
	node.StartNode(nodeID, ctx, blockchainInstance)

	// Wait for the main context to be canceled or times out
	<-ctx.Done()
}
