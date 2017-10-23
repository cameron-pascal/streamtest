package streamtest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

type StreamServer struct {
	listener *net.Listener
}

func (server *StreamServer) Start(port uint16, protocol string) error {
	portStr := fmt.Sprintf(":%d", int(port))
	listener, listenErr := net.Listen(protocol, portStr)

	if listenErr != nil {
		return listenErr
	}

	for {
		conn, acceptErr := listener.Accept()

		if acceptErr != nil {
			return acceptErr
		}

		return beginConnection(conn)
	}
}

func beginConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)

	currByte, rxErr := reader.ReadByte()
	if rxErr != nil {
		return rxErr
	}

	preambleCount := 1
	preambleBuffer := new(bytes.Buffer)

	for currByte != eot {
		currByte, rxErr = reader.ReadByte()
		if rxErr != nil {
			return rxErr
		}
		preambleBuffer.WriteByte(currByte)
		preambleCount++
	}

	preambleJSONBytes := preambleBuffer.Bytes()[:preambleCount-1]

	preamble := new(preambleMessage)
	parseErr := json.Unmarshal(preambleJSONBytes, preamble)

	if parseErr != nil {
		return parseErr
	}

	_, txErr := conn.Write([]byte{ackMessage})

	if txErr != nil {
		return txErr
	}

	fmt.Println(preamble.AckProtocol)
	fmt.Println(preamble.DataTransferSize)

	return nil
}
