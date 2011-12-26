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
	"net"
)

type msgReaderFunc func(peer *Peer, header *MsgHeader) error
type msgWriterFunc func(peer *Peer, data interface{}) error

type outputCall struct {
	f    msgWriterFunc
	data interface{}
}

type Peer struct {
	server        *Server
	conn          net.Conn
	incoming      bool
	pendingOutput chan *outputCall
	quit          chan bool
	versionKnown  bool
	version       int32
}

type supportedMsg struct {
	signature [12]byte
	handler   msgReaderFunc
}

var supportedMsgs = []supportedMsg{
	{
		versionCmdSig,
		handleVersionMsg,
	},
}

// Create new Peer from connection.
// Connection must be already opened.
func NewPeer(s *Server, conn net.Conn, incoming bool) (peer *Peer) {
	return &Peer{
		server:        s,
		conn:          conn,
		incoming:      incoming,
		pendingOutput: make(chan *outputCall),
		quit:          make(chan bool),
	}
}

// Start handling peer networking
func (peer *Peer) start() {
	go peer.inputHandler()
	go peer.outputHandler()

	if peer.incoming {
		peer.schedOutput(sendVersionMsg, nil)
	}
}

func (peer *Peer) schedOutput(f msgWriterFunc, data interface{}) {
	go func() {
		peer.pendingOutput <- &outputCall{
			f,
			data,
		}
	}()
}

// Provide serialized output agent, using pending functions channel
func (peer *Peer) outputHandler() {

mainLoop:
	for {
		select {
		case call := <-peer.pendingOutput:
			if err := call.f(peer, call.data); err != nil {
				peer.server.log.Printf("Error %s", err)
				break mainLoop
			}
		}
	}

}

func (peer *Peer) inputHandler() {
	var msgHeader MsgHeader
mainLoop:
	for {
		if err := readMsgHeader(peer.conn, &msgHeader); err != nil {
			break mainLoop
		}

		peer.server.log.Printf("%+v", &msgHeader)

		if peer.server.magic != msgHeader.Magic {
			// TODO: is this gentle enough?
			break mainLoop
		}

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
		peer.server.log.Printf("unsupported command: %s!", msgHeader.Command)
		break mainLoop
	}
	peer.server.quitingPeers <- peer
}

func handleVersionMsg(peer *Peer, msgHeader *MsgHeader) error {
	cmdHeader := new(VersionCmdHeader)

	err := readVersionMsg(peer.conn, cmdHeader)

	if err != nil {
		return err
	}

	peer.server.log.Printf("%+v", cmdHeader)

	peer.versionKnown = true
	peer.version = cmdHeader.Version

	peer.schedOutput(sendVerackMsg, nil)

	return err
}

func sendVersionMsg(peer *Peer, data interface{}) error {
	// TODO: implement
	return nil
}

func sendVerackMsg(peer *Peer, _ interface{}) error {
	peer.server.log.Printf("Sending verack")
	writeVerackMsg(peer.conn, peer.server.magic)
	return nil
}
