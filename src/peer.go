package gobtc

import (
	"net"
)



type Peer struct {
	server *Server
	conn net.Conn
	quit chan bool
}

func (peer *Peer) handler() {
	for {
		var buf [64]byte
		count, err := peer.conn.Read(buf[:])

		// TODO: handle bytes here
		count++

		if err != nil {
			break;
		}
	}
	peer.server.quitingPeers <- peer
}


