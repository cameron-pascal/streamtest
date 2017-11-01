package streamtest

// AckProtocol Acknowledgement protocol used during streaming tests
type AckProtocol uint8

const (
	// Streaming No acknowledgement between packets
	Streaming AckProtocol = iota

	// StopWait Wait for packet acknowledgement before sending the next packet
	StopWait
)

// DefaultPort The default IP port that streamtest communicates on
const DefaultPort uint16 = 5991

type preambleMessage struct {
	AckProtocol      AckProtocol
	DataTransferSize uint32
	PayloadSize      uint32
}

const (
	ackMessage uint8 = 6
	eot        uint8 = 4
)
