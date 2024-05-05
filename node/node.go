package node

import (
	"context"
	"encoding/json"
	"simple-consensus-algorithm-impl/blockchain"
	"simple-consensus-algorithm-impl/domain"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/sirupsen/logrus"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "blockchain-channel"

// Create two channels for sending and receiving messages
var receiveChannel = make(chan []byte, 10)
var publishChannel = make(chan []byte, 10)

var NodeID int64

// We have created an Enum for the type of message
const (
	ACK   = iota // 0 : Acknowledgement
	BLOCK        // 1 : Block Proposed
)

// We have a custom Message type to send the message
type Message struct {
	FromNodeID int64
	Message    string
	Type       int
	Block      blockchain.Block
}

var ticker *time.Ticker

var isFirstBlockProposed = false

func StartNode(index int64, ctx context.Context, blockchainInstance *blockchain.Blockchain) {

	// Set the Node ID
	NodeID = index

	// create a new libp2p Host that listens on a random TCP port
	host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	// view host details and addresses
	logrus.Infof("node ID %s\n", host.ID().String())
	logrus.Infof("following are the assigned addresses\n")
	for _, addr := range host.Addrs() {
		logrus.Infof("%s\n", addr.String())
	}
	logrus.Infof("\n")

	// create a new PubSub service using the GossipSub router
	gossipSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(host); err != nil {
		panic(err)
	}

	// join the pubsub topic called blockchain-channel
	nodesRoom := domain.CHANNEL_NAME
	nodesTopic, err := gossipSub.Join(nodesRoom)
	if err != nil {
		panic(err)
	}

	// subscribe to topic
	subscriber, err := nodesTopic.Subscribe()
	if err != nil {
		panic(err)
	}

	defer subscriber.Cancel() // To avoid a dangling subscription in case we need to exit

	// Create a go routine to subscribe to the topic
	go subscribe(subscriber, ctx, host.ID())

	// Create a go routine to process the messages
	go processMessage(blockchainInstance)

	// create publisher
	go publishMessage(ctx, nodesTopic)

	// Now we want our node to propose a new block every genesis block time = 5*nodeID seconds
	// and propose a new block every 5*nodeID seconds from the genesisTimeStamp, so that we can
	// maintain the round-robin cycle for proposal of blocks.
	go proposeBlockLoop(ctx, blockchainInstance)

	go proposeBlockRoundRobin(ctx, blockchainInstance)

	// Wait for the main context to be canceled or times out
	<-ctx.Done()
}

// subscribe to messages from topic
func subscribe(subscriber *pubsub.Subscription, ctx context.Context, hostID peer.ID) {
	for {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			panic(err) // There can be a retry added for the subscription instead of panic
		}

		// only consider messages delivered by other peers
		if msg.ReceivedFrom == hostID {
			continue
		}
		// We will send the message in a reciever channel to another go routine.
		// There we will unmarshall the message and validate the message,
		// and then add it to the blockchain, and send the message
		// to another publisher channel to publish the message to the network.
		receiveChannel <- msg.Data
	}
}

// publish messages to topic
func publishMessage(ctx context.Context, topic *pubsub.Topic) {
	for {
		// switch for channel
		// Publish block and send ack for validated block
		msg := <-publishChannel
		if len(msg) != 0 {
			// publish message to topic
			bytes := []byte(msg)
			topic.Publish(ctx, bytes)
		}
	}
}

func proposeBlockLoop(ctx context.Context, blockchainInstance *blockchain.Blockchain) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			currentTime := time.Now().Unix()

			genesisTimStamp := domain.GENESIS_TIMESTAMP // This is a random timestamp

			// Here we are calculating the remainder of the current time and the genesisTimeStamp
			// For the first execution of the proposeBlockInternal function for each node, we are checking if the
			// remainder is 0, for the local NodeID, and is not 0 for the other NodeIDs.
			var shouldExecute bool
			if NodeID == 4 && (currentTime-genesisTimStamp)%(5*NodeID) == 0 {
				shouldExecute = true
			} else if NodeID == 3 && (currentTime-genesisTimStamp)%(5*NodeID) == 0 && (currentTime-genesisTimStamp)%(5*4) != 0 {
				shouldExecute = true
			} else if NodeID == 2 && (currentTime-genesisTimStamp)%(5*NodeID) == 0 && (currentTime-genesisTimStamp)%(5*3) != 0 && (currentTime-genesisTimStamp)%(5*4) != 0 {
				shouldExecute = true
			} else if NodeID == 1 && (currentTime-genesisTimStamp)%(5*NodeID) == 0 && (currentTime-genesisTimStamp)%(5*3) != 0 && (currentTime-genesisTimStamp)%(5*4) != 0 && (currentTime-genesisTimStamp)%(5*2) != 0 {
				shouldExecute = true
			}

			// We are executing the proposeBlockInternal function once
			// and then we are starting a ticker to propose a new block every 20 seconds.
			if shouldExecute {
				proposeBlockInternal(blockchainInstance)
				ticker = time.NewTicker(20 * time.Second)
				isFirstBlockProposed = true
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func proposeBlockRoundRobin(ctx context.Context, blockchainInstance *blockchain.Blockchain) {
	for {
		if isFirstBlockProposed {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Propose a new block
				proposeBlockInternal(blockchainInstance)
			}
		}
	}
}
func proposeBlockInternal(blockchainInstance *blockchain.Blockchain) {
	// Create a new block
	newBlock := blockchain.NewBlock(NodeID, blockchainInstance)

	// Propose the Block to the peers
	// Create a new message to send to the peers
	message := Message{
		FromNodeID: NodeID,
		Block:      newBlock,
		Message:    domain.MESSAGE_BLOCK_PROPOSED,
		Type:       BLOCK,
	}

	// Convert the message to bytes
	msgBytes, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Error marshalling block: %s", err)
	}

	// Send the block to the publisher channel
	publishChannel <- msgBytes

	logrus.Infof("Block proposed by the node: %d, with the data: %d", NodeID, newBlock.Data)
}

func processMessage(blockchainInstance *blockchain.Blockchain) {
	for {
		// Receive the message from the receive channel
		receivedMessage := <-receiveChannel

		// Create a new message object
		var message Message

		// Check if the message is a block or an ACK
		err := json.Unmarshal(receivedMessage, &message)
		if err != nil {

			// This is an errorneous message
			// revert an error
			logrus.Errorf("Error unmarshalling message: %s", err)
		} else {

			// It's a Message
			// Now we look if the message received is an acknowledgement or a block proposed.
			// If it's a block, then we validate the block and send an ACK or NACK
			// to the publisher channel.
			if message.Type == BLOCK {
				// Validate the block
				// Check if the block is valid
				logrus.Infof("Block received from the node: %d, with the data: %d", message.FromNodeID, message.Block.Data)

				// Check if the block is valid
				isValid := message.Block.IsBlockValid(blockchainInstance, message.Block.Data)

				// Create a new message to send to the peers
				responseMessage := Message{
					FromNodeID: NodeID,
					Block:      message.Block,
					Message:    isValid,
					Type:       ACK,
				}

				// Convert the message to bytes
				msgBytes, err := json.Marshal(responseMessage)
				if err != nil {
					logrus.Errorf("Error marshalling NACK: %s", err)
				}

				// Send the NACK to the publisher channel
				publishChannel <- msgBytes

				// Vote for the block
				message.Block.VoteForBlock(blockchainInstance, responseMessage.Message)
			} else if message.Type == ACK {

				// It's an ACK
				logrus.Infof("ACK received from node: %d, for the block: %d", message.FromNodeID, blockchainInstance.GetHeight()+1)
				// Vote for the block
				message.Block.VoteForBlock(blockchainInstance, message.Message)
			}
		}
	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	logrus.Infof("discovered new peer %s\n", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		logrus.Infof("error connecting to peer %s: %s\n", pi.ID.String(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
