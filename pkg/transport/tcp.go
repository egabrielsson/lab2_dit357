package transport

import (
	"fmt"
	"net"
	"sync"
	"time"

	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"
)

// TCPTransport implements the Transport interface using TCP sockets.
type TCPTransport struct {
	id       string
	addr     string
	peers    map[string]string // peerID -> address
	clock    *clock.LamportClock
	listener net.Listener
	conns    map[string]net.Conn
	connsMu  sync.RWMutex
	handler  MessageHandler
}

// NewTCPTransport creates a new TCP transport instance.
func NewTCPTransport(id, addr string, peers map[string]string) (*TCPTransport, error) {
	tt := &TCPTransport{
		id:    id,
		addr:  addr,
		peers: peers,
		clock: clock.NewLamportClock(),
		conns: make(map[string]net.Conn),
	}
	return tt, nil
}

// GetID returns the transport's unique identifier.
func (tt *TCPTransport) GetID() string {
	return tt.id
}

// Listen starts the TCP server and listens for incoming connections.
func (tt *TCPTransport) Listen(handler MessageHandler) error {
	tt.handler = handler

	listener, err := net.Listen("tcp", tt.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", tt.addr, err)
	}
	tt.listener = listener

	go tt.acceptLoop()
	return nil
}

// acceptLoop continuously accepts incoming connections.
func (tt *TCPTransport) acceptLoop() {
	for {
		conn, err := tt.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go tt.handleConnection(conn)
	}
}

// handleConnection handles messages from a single connection.
func (tt *TCPTransport) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		var msg message.Message
		if err := message.WireRead(conn, &msg); err != nil {
			return // connection closed or error
		}

		// Update Lamport clock
		tt.clock.Receive(msg.Lamport)

		// Call handler
		if tt.handler != nil {
			if reply, err := tt.handler(msg); err == nil && reply != nil {
				reply.From = tt.id
				reply.Lamport = tt.clock.Tick()
				_ = message.WireWrite(conn, *reply)
			}
		}
	}
}

// getConnection gets or creates a connection to a peer.
func (tt *TCPTransport) getConnection(peerID string) (net.Conn, error) {
	tt.connsMu.RLock()
	if conn, exists := tt.conns[peerID]; exists {
		tt.connsMu.RUnlock()
		return conn, nil
	}
	tt.connsMu.RUnlock()

	// Need to create connection
	peerAddr, exists := tt.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("unknown peer: %s", peerID)
	}

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", peerAddr, err)
	}

	tt.connsMu.Lock()
	tt.conns[peerID] = conn
	tt.connsMu.Unlock()

	return conn, nil
}

// Send sends a one-way message to another truck.
func (tt *TCPTransport) Send(toID string, msg message.Message) error {
	conn, err := tt.getConnection(toID)
	if err != nil {
		return err
	}

	msg.From = tt.id
	msg.To = toID
	msg.Lamport = tt.clock.Tick()

	return message.WireWrite(conn, msg)
}

// Request sends a request message and waits for a reply (simplified for TCP).
// Note: This is a basic implementation - in a real system you'd want proper correlation IDs.
func (tt *TCPTransport) Request(toID string, msg message.Message, timeout time.Duration) (message.Message, error) {
	conn, err := tt.getConnection(toID)
	if err != nil {
		return message.Message{}, err
	}

	msg.From = tt.id
	msg.To = toID
	msg.Lamport = tt.clock.Tick()

	// Send request
	if err := message.WireWrite(conn, msg); err != nil {
		return message.Message{}, err
	}

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})

	// Read reply
	var reply message.Message
	if err := message.WireRead(conn, &reply); err != nil {
		return message.Message{}, err
	}

	tt.clock.Receive(reply.Lamport)
	return reply, nil
}

// Publish is not implemented for TCP transport (use NATS for pub-sub).
func (tt *TCPTransport) Publish(channel string, msg message.Message) error {
	return fmt.Errorf("publish not supported by TCP transport - use NATS for pub-sub")
}

// Subscribe is not implemented for TCP transport (use NATS for pub-sub).
func (tt *TCPTransport) Subscribe(channel string, handler func(message.Message) error) error {
	return fmt.Errorf("subscribe not supported by TCP transport - use NATS for pub-sub")
}

// Close shuts down the TCP transport.
func (tt *TCPTransport) Close() error {
	if tt.listener != nil {
		tt.listener.Close()
	}

	tt.connsMu.Lock()
	for _, conn := range tt.conns {
		conn.Close()
	}
	tt.connsMu.Unlock()

	return nil
}