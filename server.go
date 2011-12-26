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
	"container/list"
	"log"
	"net"
	"os"
)

type Server struct {
	waitPeerHandler     chan bool
	waitListenerHandler chan bool
	newPeers            chan *Peer
	quitingPeers        chan *Peer
	listener            net.Listener
	log                 *log.Logger
	maxPeers            int
	magic               uint32
}

func NewServer(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		waitPeerHandler:     make(chan bool),
		waitListenerHandler: make(chan bool),
		newPeers:            make(chan *Peer),
		quitingPeers:        make(chan *Peer),
		listener:            listener,
		log:                 log.New(os.Stderr, "gobtc: ", log.LstdFlags),
		maxPeers:            10,
		magic:               MAGIC_MAIN,
	}
	return s, nil
}

func (s *Server) Start() {
	go s.peerHandler()
	go s.listenerHandler()
	// TODO: implement peerSeeker
}

func (s *Server) SetLogger(log *log.Logger) {
	s.log = log
}

func (s *Server) Wait() {
	<-s.waitPeerHandler
	<-s.waitListenerHandler
}

func (s *Server) AddPeer(peer *Peer) {
	go func() { s.newPeers <- peer }()
}

// Manage server peer list.
// Handle new and quiting peers.
func (s *Server) peerHandler() {
	var peers *list.List = list.New()
	defer func() {
		s.waitListenerHandler <- true
		recover()
	}()

	func() {
		for {
			select {
			case peer := <-s.newPeers:
				if peers.Len() >= s.maxPeers {
					peer.conn.Close()
				}

				peers.PushBack(peer)
				s.log.Printf("Added peer %s", peer.conn.RemoteAddr())
				peer.start()

			case peer := <-s.quitingPeers:
				// TODO: remove peer
				found := false
				for e := peers.Front(); e != nil; e = e.Next() {
					tpeer := e.Value.(*Peer)
					if tpeer == peer {
						peers.Remove(e)
						s.log.Printf("Removed peer %s", peer.conn.RemoteAddr())
						found = true
					}
				}
				if !found {
					s.log.Printf("assert error: quiting peer not found on the list")
				}
			}
		}
	}()
}

func (s *Server) listenerHandler() {
	var err error
	s.log.Printf("Listening on %s", s.listener.Addr())
	defer func() {
		s.waitListenerHandler <- true
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		s.AddPeer(NewPeer(s, conn, true))
	}
	s.log.Printf("Stopped listening on %s. Error: %s", s.listener.Addr(), err)
}
