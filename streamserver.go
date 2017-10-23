package streamtest

import (
	"encoding/json"
	"fmt"
	"net"
)

type StreamServer struct {
}

func udpStreamingDownload(conn net.PacketConn, connAddr net.Addr, preamble *preambleMessage) error {
	buf := make([]byte, preamble.PayloadSize)

	var totalBytesRead uint32
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)

		if connAddr.String() != remoteAddr.String() {
			fmt.Println("streaming: ignored udp packet")
			continue
		}

		if err != nil {
			return err
		}

		totalBytesRead += uint32(n)

		if totalBytesRead >= preamble.DataTransferSize {
			break
		}
	}

	conn.Close()

	return nil
}

func udpStopWaitDownload(conn net.PacketConn, connAddr net.Addr, preamble *preambleMessage) error {
	buf := make([]byte, preamble.PayloadSize)

	var totalBytesRead uint32
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if connAddr.String() != remoteAddr.String() {
			fmt.Println("stopwait: ignored udp packet")
			continue
		}

		if err != nil {
			return err
		}

		totalBytesRead += uint32(n)
		conn.WriteTo([]byte{ackMessage}, connAddr)

		if totalBytesRead >= preamble.DataTransferSize {
			break
		}
	}

	conn.Close()

	return nil
}

func listenUDP(portStr string) error {
	conn, err := net.ListenPacket("udp", portStr)

	if err != nil {
		return err
	}

	buf := make([]byte, 128)

	for {
		n, remoteAddr, err := conn.ReadFrom(buf)

		if err != nil {
			fmt.Printf("could not read udp: %s\n", err.Error())
			continue
		}

		preamble, preambleErr := deserailizePreamble(buf[:n])
		if preambleErr == nil {
			fmt.Printf("acknowledging preamble over udp from %s\n", remoteAddr.String())
			conn.WriteTo([]byte{ackMessage}, remoteAddr)

			var downloadErr error
			if preamble.AckProtocol == Streaming {
				downloadErr = udpStreamingDownload(conn, remoteAddr, preamble)
			} else if preamble.AckProtocol == StopWait {
				downloadErr = udpStopWaitDownload(conn, remoteAddr, preamble)
			}

			if downloadErr != nil {
				fmt.Printf("could not perform udp download: %s\n", downloadErr)
			}

		} else {
			fmt.Printf("could not udp preamble: %s\n", preambleErr.Error())
		}
	}
}

func tcpStreamingDownload(conn net.Conn, preamble *preambleMessage) error {
	buf := make([]byte, preamble.PayloadSize)

	var totalBytesRead uint32
	for {
		n, err := conn.Read(buf)

		if err != nil {
			return err
		}

		totalBytesRead += uint32(n)

		if totalBytesRead >= preamble.DataTransferSize {
			break
		}
	}

	conn.Close()

	return nil
}

func tcpStopWaitDownload(conn net.Conn, preamble *preambleMessage) error {
	buf := make([]byte, preamble.PayloadSize)

	var totalBytesRead uint32
	for {
		n, err := conn.Read(buf)

		if err != nil {
			return err
		}

		totalBytesRead += uint32(n)
		conn.Write([]byte{ackMessage})

		if totalBytesRead >= preamble.DataTransferSize {
			break
		}
	}

	conn.Close()

	return nil
}

func listenTCP(portStr string) error {
	listener, listenErr := net.Listen("tcp", portStr)

	if listenErr != nil {
		return listenErr
	}

	buf := make([]byte, 128)
	for {
		conn, acceptErr := listener.Accept()

		if acceptErr != nil {
			fmt.Printf("could not accept tcp: %s\n", acceptErr.Error())
			continue
		}

		n, readErr := conn.Read(buf)

		if readErr != nil {
			fmt.Printf("could not read tcp: %s\n", readErr.Error())
			continue
		}

		preamble, preambleErr := deserailizePreamble(buf[:n])
		if preambleErr == nil {
			fmt.Printf("acknowledging preamble over tcp from %s\n", conn.LocalAddr().String())
			conn.Write([]byte{ackMessage})

			var downloadErr error
			if preamble.AckProtocol == Streaming {
				downloadErr = tcpStreamingDownload(conn, preamble)
			} else if preamble.AckProtocol == StopWait {
				downloadErr = tcpStopWaitDownload(conn, preamble)
			}

			if downloadErr != nil {
				fmt.Printf("could not perform tcp download: %s\n", downloadErr)
			}

		} else {
			fmt.Printf("could not read tcp preamble: %s\n", preambleErr.Error())
		}
	}
}

func deserailizePreamble(preambleBuf []byte) (*preambleMessage, error) {
	currByte := preambleBuf[0]
	idx := 1

	for currByte != eot {
		currByte = preambleBuf[idx]
		idx++
	}

	preambleJSONBytes := preambleBuf[:idx-1]

	preamble := new(preambleMessage)
	parseErr := json.Unmarshal(preambleJSONBytes, preamble)

	if parseErr != nil {
		return nil, parseErr
	}

	return preamble, nil
}

func (server *StreamServer) Start(port uint16, protocol string) error {
	portStr := fmt.Sprintf(":%d", int(port))

	if protocol == "tcp" {
		listenTCP(portStr)
	} else if protocol == "udp" {
		listenUDP(portStr)
	}

	return nil
}
