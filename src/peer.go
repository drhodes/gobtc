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
	"bytes"
	"github.com/kr/pretty.go"
	"net"
)

type msgHandlerFunc func(peer *Peer, header *MsgHeader) error

type Peer struct {
	server *Server
	conn   net.Conn
	quit   chan bool
}

type SupportedMsg struct {
	signature [12]byte
	handler   msgHandlerFunc
}

var supportedMsgs = []SupportedMsg{
	{
		versionCmdSig,
		handleVersionCmd,
	},
}

func handleVersionCmd(peer *Peer, header *MsgHeader) error {
	cmdHeader := new(VersionCmdHeader)

	err := parseVersionMsg(peer.conn, cmdHeader)

	if err == nil {
		peer.server.log.Printf("%+v", pretty.Formatter(cmdHeader))
	}

	return err
}

func (peer *Peer) handler() {
	var msgHeader MsgHeader
mainLoop:
	for {
		if err := parseMsgHeader(peer.conn, &msgHeader); err != nil {
			break
		}

		peer.server.log.Printf("%+v", pretty.Formatter(&msgHeader))
		// TODO: validate header

		for _, command := range supportedMsgs {
			if bytes.Compare(command.signature[:], msgHeader.Command[:]) == 0 {
				err := command.handler(peer, &msgHeader)
				if err != nil {
					peer.server.log.Printf("Error %s", err)
					break mainLoop
				}
				continue mainLoop
			}
		}
		peer.server.log.Printf("unknown command: %s!", msgHeader.Command)
		break
	}
	peer.server.quitingPeers <- peer
}
