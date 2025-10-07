// tcpnode.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type TCPNode struct {
	ID    string
	Addr  string
	Clock *LamportClock

	PeersMu sync.RWMutex
	Peers   map[string]string // id -> addr
}

func NewTCPNode(id, addr string, peers map[string]string) *TCPNode {
	return &TCPNode{
		ID:    id,
		Addr:  addr,
		Clock: &LamportClock{},
		Peers: peers,
	}
}

func (n *TCPNode) Listen(handler func(Message, net.Conn)) error {
	ln, err := net.Listen("tcp", n.Addr)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] listening on %s\n", n.ID, n.Addr)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go n.handleConn(conn, handler)
		}
	}()
	return nil
}

func (n *TCPNode) handleConn(conn net.Conn, handler func(Message, net.Conn)) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		line := sc.Bytes()
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Println("bad msg:", err)
			continue
		}
		n.Clock.Receive(msg.Lamport)
		handler(msg, conn)
	}
}

func (n *TCPNode) Send(peerID string, msg Message) error {
	n.PeersMu.RLock()
	addr, ok := n.Peers[peerID]
	n.PeersMu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown peer %s", peerID)
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg.Lamport = n.Clock.Tick()
	msg.From = n.ID

	enc, _ := json.Marshal(msg)
	enc = append(enc, '\n')
	_, err = conn.Write(enc)
	return err
}
