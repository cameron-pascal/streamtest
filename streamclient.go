package streamtest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	stopWaitTransferSize  uint32 = 10000   // 10KB
	streamingTransferSize uint32 = 1000000 // 10 MB
)

// ClientOptions Options when setting up a streamtest client
type ClientOptions struct {
	ServerIPAddress net.IPAddr
	ServerPort      uint16
	NetworkProtocol string
	AckProtocol     AckProtocol
	MessageSize     uint32
}

type ClientResult struct {
	PacketsSent uint32
	BytesSent   uint32
	Duration    int64
}

// StreamClient The streamtest client
type StreamClient struct {
	options ClientOptions
	conn    *net.Conn
}

// NewStreamClient Creates a new StreamClient
func NewStreamClient(options ClientOptions) *StreamClient {
	client := StreamClient{
		options: options,
		conn:    nil}

	return &client
}

// Connect Connects a StreamClient to its streamtest server
func (client *StreamClient) Connect() error {

	ipAddress := client.options.ServerIPAddress.IP.String()
	ipAndPort := fmt.Sprintf("%s:%d", ipAddress, client.options.ServerPort)
	conn, err := net.Dial(client.options.NetworkProtocol, ipAndPort)

	client.conn = &conn

	return err
}

// Start Begins a streamtest transfer
func (client *StreamClient) Start() (*ClientResult, error) {
	if client.conn == nil {
		return nil, errors.New("Client must be connected first")
	}

	dataTransferSize := streamingTransferSize

	if client.options.NetworkProtocol == "udp" {
		dataTransferSize = stopWaitTransferSize
	}

	preambleMessage := preambleMessage{
		AckProtocol:      client.options.AckProtocol,
		DataTransferSize: dataTransferSize,
		PayloadSize:      client.options.MessageSize}

	preambleJSON, marshalErr := json.Marshal(preambleMessage)

	// The preamble is a JSON serialized preambleMessage followed by a EOT byte
	preambleBuffer := bytes.NewBuffer(preambleJSON)
	preambleBuffer.WriteByte(eot)

	if marshalErr != nil {
		return nil, marshalErr
	}

	conn := *client.conn

	_, txErr := conn.Write(preambleBuffer.Bytes())

	if txErr != nil {
		return nil, txErr
	}

	reader := bufio.NewReader(conn)

	preambleResponse, rxErr := reader.ReadByte()

	if rxErr != nil {
		return nil, rxErr
	}

	if uint8(preambleResponse) != ackMessage {
		return nil, errors.New("server did not acknowledge preamble")
	}

	return client.transfer()
}

func (client *StreamClient) streamingUpload() (*ClientResult, error) {
	conn := *client.conn
	var byteCount uint32
	var packetCount uint32
	buf := make([]byte, client.options.MessageSize)

	start := time.Now()
	for byteCount < streamingTransferSize {
		conn.Write(buf)
		byteCount += uint32(len(buf))
		packetCount++
	}

	elapsed := time.Since(start)

	conn.Close()

	result := ClientResult{
		BytesSent:   byteCount,
		PacketsSent: packetCount,
		Duration:    elapsed.Nanoseconds()}

	return &result, nil
}

func (client *StreamClient) stopWaitUpload() (*ClientResult, error) {
	conn := *client.conn

	buf := make([]byte, client.options.MessageSize)
	ackBuf := make([]byte, 1)

	var byteCount uint32
	var packetCount uint32
	start := time.Now()

	for byteCount < stopWaitTransferSize {
		conn.Write(buf)
		byteCount += uint32(len(buf))
		packetCount++

		n, err := conn.Read(ackBuf)
		if n >= 1 && ackBuf[0] != ackMessage {
			return nil, errors.New("server did not acknowledge last message")
		}

		if err != nil {
			return nil, err
		}
	}

	elapsed := time.Since(start)

	conn.Close()

	result := ClientResult{
		BytesSent:   byteCount,
		PacketsSent: packetCount,
		Duration:    elapsed.Nanoseconds()}

	return &result, nil
}

func (client *StreamClient) transfer() (*ClientResult, error) {
	fmt.Println("server acknowledged preamble")

	if client.options.AckProtocol == Streaming {
		return client.streamingUpload()
	}

	if client.options.AckProtocol == StopWait {
		return client.stopWaitUpload()
	}

	return nil, nil
}
