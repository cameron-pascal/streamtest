package streamtest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

const dataTransferSize uint32 = 1000000000 // 1 GB

// ClientOptions Options when setting up a streamtest client
type ClientOptions struct {
	ServerIPAddress net.IPAddr
	ServerPort      uint16
	NetworkProtocol string
	AckProtocol     AckProtocol
	MessageSize     uint32
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
func (client *StreamClient) Start() error {

	if client.conn == nil {
		return errors.New("Client must be connected first")
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
		return marshalErr
	}

	conn := *client.conn

	_, txErr := conn.Write(preambleBuffer.Bytes())

	if txErr != nil {
		return txErr
	}

	reader := bufio.NewReader(conn)

	preambleResponse, rxErr := reader.ReadByte()

	if rxErr != nil {
		return rxErr
	}

	if uint8(preambleResponse) != ackMessage {
		return errors.New("server did not acknowledge preamble")
	}

	return client.transfer()
}

func (client *StreamClient) transfer() error {
	fmt.Println("server acknowledged preamble")

	return nil
}
