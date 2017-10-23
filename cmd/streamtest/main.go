package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/cameronmaxwell/streamtest"
)

func handleInputError(message string) {
	fmt.Println(message)
	flag.PrintDefaults()
	os.Exit(1)
}

func handleClientError(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	hostPtr := flag.String("host", "", "name or IPv4 address of host")
	clientPortPtr := flag.Int("port", int(streamtest.DefaultPort), "host port number {0-65535}")
	clientProtocolPtr := flag.String("protocol", "tcp", "transport protocol {tcp|udp}")
	ackPtr := flag.String("ack", "streaming", "acknowledgment protocol {streaming|stopwait}")
	sizePtr := flag.Int("size", 512, "message payload size {1-65536}")

	flag.Parse()

	var ackProtocol streamtest.AckProtocol
	switch *ackPtr {
	case "streaming":
		ackProtocol = streamtest.Streaming
		break
	case "stopwait":
		ackProtocol = streamtest.StopWait
		break
	default:
		handleInputError("invalid acknowledgment protocol")
	}

	if *clientProtocolPtr != "tcp" && *clientProtocolPtr != "udp" {
		handleInputError("invalid transport protocol")
	}

	if *sizePtr <= 0 {
		handleInputError("message payload size out of range")
	}

	if *hostPtr == "" {
		handleInputError("host must be specified")
	}

	hostIPAddr, err := net.ResolveIPAddr("ip", *hostPtr)

	if err != nil {
		handleInputError(err.Error())
	}

	if *clientPortPtr == -1 {
		handleInputError("port must be specified")
	} else if *clientPortPtr < 0 || *clientPortPtr > 65535 {
		handleInputError("port number out of range")
	}

	optionsPtr := streamtest.ClientOptions{
		ServerIPAddress: *hostIPAddr,
		ServerPort:      uint16(*clientPortPtr),
		NetworkProtocol: *clientProtocolPtr,
		AckProtocol:     ackProtocol,
		MessageSize:     uint32(*sizePtr)}

	streamClient := streamtest.NewStreamClient(optionsPtr)

	connectErr := streamClient.Connect()

	if connectErr != nil {
		handleClientError(connectErr)
	}

	transferErr := streamClient.Start()

	if transferErr != nil {
		handleClientError(transferErr)
	}
}
