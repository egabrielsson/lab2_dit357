package transport

import (
	"encoding/json"
	"fmt"
	"time"

	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"

	"github.com/nats-io/nats.go"
)

// NATSTransport implements the Transport interface using NATS messaging.
type NATSTransport struct {
	id    string
	url   string
	clock *clock.LamportClock
	nc    *nats.Conn
	sub   *nats.Subscription
}

// NewNATSTransport creates a new NATS transport instance.
func NewNATSTransport(id, natsURL string) (*NATSTransport, error) {
	nt := &NATSTransport{
		id:    id,
		url:   natsURL,
		clock: clock.NewLamportClock(),
	}

	nc, err := nats.Connect(natsURL,
		nats.Name("truck-"+id),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	nt.nc = nc
	return nt, nil
}

// GetID returns the transport's unique identifier.
func (nt *NATSTransport) GetID() string {
	return nt.id
}

// subjectFor generates the NATS subject for a specific truck ID.
func (nt *NATSTransport) subjectFor(id string) string {
	return fmt.Sprintf("trucks.%s", id)
}

// Listen starts listening for incoming messages on this transport's subject.
func (nt *NATSTransport) Listen(handler MessageHandler) error {
	subj := nt.subjectFor(nt.id)
	sub, err := nt.nc.Subscribe(subj, func(m *nats.Msg) {
		var msg message.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			return
		}

		// Update Lamport clock
		nt.clock.Receive(msg.Lamport)

		// Call handler
		if reply, err := handler(msg); err == nil && reply != nil {
			reply.From = nt.id
			reply.Lamport = nt.clock.Tick()
			
			if replyData, err := json.Marshal(reply); err == nil {
				_ = m.Respond(replyData)
			}
		}
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	
	nt.sub = sub
	return nt.nc.Flush()
}

// Send sends a one-way message to another truck.
func (nt *NATSTransport) Send(toID string, msg message.Message) error {
	msg.From = nt.id
	msg.To = toID
	msg.Lamport = nt.clock.Tick()
	
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	return nt.nc.Publish(nt.subjectFor(toID), data)
}

// Request sends a request message and waits for a reply.
func (nt *NATSTransport) Request(toID string, msg message.Message, timeout time.Duration) (message.Message, error) {
	msg.From = nt.id
	msg.To = toID
	msg.Lamport = nt.clock.Tick()
	
	data, err := json.Marshal(msg)
	if err != nil {
		return message.Message{}, fmt.Errorf("failed to marshal message: %w", err)
	}
	
	resp, err := nt.nc.Request(nt.subjectFor(toID), data, timeout)
	if err != nil {
		return message.Message{}, fmt.Errorf("request failed: %w", err)
	}
	
	var reply message.Message
	if err := json.Unmarshal(resp.Data, &reply); err != nil {
		return message.Message{}, fmt.Errorf("failed to unmarshal reply: %w", err)
	}
	
	nt.clock.Receive(reply.Lamport)
	return reply, nil
}

// Close shuts down the NATS transport.
func (nt *NATSTransport) Close() error {
	if nt.sub != nil {
		_ = nt.sub.Unsubscribe()
	}
	if nt.nc != nil {
		nt.nc.Close()
	}
	return nil
}