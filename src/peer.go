/*
 * Copyright (c) 2011, Dawid Ciężarkiewicz. All rights reserved.
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
 * MA 02110-1301  USA
 */


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


