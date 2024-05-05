package domain


const (
	ACK   = iota // 0 : Acknowledgement
	BLOCK        // 1 : Block Proposed
)

const (
	BLOCK_IS_VALID   = "VALID"
	BLOCK_IS_INVALID = "INVALID"

	CHANNEL_NAME = "blockchain-channel"

	MESSAGE_BLOCK_PROPOSED = "New Block Proposed"

	GENESIS_TIMESTAMP = int64(1710845056)
)