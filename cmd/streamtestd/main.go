package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/cameronmaxwell/streamtest"
)

func startServer(port uint16, protocol string) {
	server := new(streamtest.StreamServer)
	fmt.Printf("starting %s on %d\n", protocol, port)
	err := server.Start(port, protocol)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	portPtr := flag.Int("port", int(streamtest.DefaultPort), "port to listen on {0-65535}")

	flag.Parse()

	if *portPtr < 0 || *portPtr > 65535 {
		fmt.Println("port out of range")
		flag.PrintDefaults()
		os.Exit(1)
	}

	port := uint16(*portPtr)

	var wg sync.WaitGroup

	wg.Add(2)
	go startServer(port, "tcp")
	go startServer(port, "udp")

	wg.Wait()
}
