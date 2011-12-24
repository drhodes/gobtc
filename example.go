package main

import (
	"flag"
	"log"
	"net"
	"github.com/dpc/gobtc"
)

var listener net.Listener

var listenAddr = flag.String("addr", ":8883", "listen address")


func main() {
	flag.Parse()
	server, err := gobtc.New(*listenAddr);
	if err != nil {
		log.Fatalf("Couldn't listen on: %s; err: %s", *listenAddr, err)
	}

	server.Start()
	server.Wait()
}

