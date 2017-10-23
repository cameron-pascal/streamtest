package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cameronmaxwell/streamtest"
)

func main() {
	portPtr := flag.Int("port", int(streamtest.DefaultPort), "port to listen on {0-65535}")

	flag.Parse()

	if *portPtr < 0 || *portPtr > 65535 {
		fmt.Println("port out of range")
		flag.PrintDefaults()
		os.Exit(1)
	}

	server := new(streamtest.StreamServer)

	serverErr := server.Start(uint16(*portPtr), "tcp")

	if serverErr != nil {
		fmt.Println(serverErr.Error())
		os.Exit(1)
	}
}
