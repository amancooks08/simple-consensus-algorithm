# Consensus Algorithm
This is a simple Consensus Algorithm and is defined as follows: blocks are produced in a constant frequency of 1 block every 5 seconds. Nodes propose new blocks one after another in a round-robin cycle. If one node skips its turn, the block is not proposed in the given time window. Every new block contains a single information: an integer. The block is valid if the integer proposed is divisible by the integer from the previous block. All nodes should validate this rule and reject invalid blocks.

For the purpose of communication for this task, I have used p2p communication. The nodes are communicating over the network layer. I have used the [`libp2p`](https://docs.libp2p.io/) library for the same.

I have used the libp2p's gossipsub for the communication between the nodes.
You can learn more about Gossipsub [here](https://github.com/libp2p/specs/tree/master/pubsub/gossipsub).

Also for the purpose of peer discovery I have used `mDNS`. `mDNS` is a protocol that allows you to resolve hostnames to IP addresses within small networks that do not include a local name server. It is a zero-configuration service, using essentially the same programming interfaces, packet formats and operating semantics as the unicast Domain Name System (DNS). You can learn more about mDNS [here](https://en.wikipedia.org/wiki/Multicast_DNS).

## Installation and Running
1. Clone the repository.
2. Enter the repository directory.
3. Make sure you have `go` installed on your system.
4. Install the dependencies using the command `go mod tidy`.
5. Run the following command in to start a node: `go run main.go 1`, we need to repeat this process 4 times in order to run 4 nodes, with the node IDs as 1, 2, 3 and 4 respectively.
6. The nodes will start running and will start proposing blocks in a round-robin fashion.


## Assumptions
1. The nodes are running on the same machine and have system time synchronized.
2. The nodes are communicating over the network layer.
3. There are always 4 nodes in the network.
4. The state is always kept in the memory.

## Notes
1. You'll see the logs for a particular node in the terminal where you started it.
2. The logs will show the proposed block, the validation result, ans the state of 
the blockchain for that node.
3. The logs will also show the messages received by the node from the other nodes.