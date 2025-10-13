package transport

import (
	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

// NATSTransport implements the Transport interface using NATS messaging.
type NATSTransport struct {
	id      string
	url     string
	clock   *clock.LamportClock
	nc      *nats.Conn
	sub     *nats.Subscription
	pubSubs map[string]*nats.Subscription // track pub-sub subscriptions
}

// NewNATSTransport creates a new NATS transport instance.
func NewNATSTransport(id, natsURL string) (*NATSTransport, error) {
	nt := &NATSTransport{
		id:      id,
		url:     natsURL,
		clock:   clock.NewLamportClock(),
		pubSubs: make(map[string]*nats.Subscription),
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

// Publish broadcasts a message to all subscribers of a channel.
func (nt *NATSTransport) Publish(channel string, msg message.Message) error {
	msg.From = nt.id
	msg.Lamport = nt.clock.Tick()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %w", err)
	}

	return nt.nc.Publish(channel, data)
}

// Subscribe starts listening to broadcast messages on a channel.
func (nt *NATSTransport) Subscribe(channel string, handler SubscriptionHandler) error {
	sub, err := nt.nc.Subscribe(channel, func(m *nats.Msg) {
		var msg message.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			fmt.Printf("Error unmarshaling broadcast message: %v\n", err)
			return
		}

		// Skip messages from ourselves
		if msg.From == nt.id {
			return
		}

		// Update Lamport clock
		nt.clock.Receive(msg.Lamport)

		// Call handler (no reply expected for broadcasts)
		if err := handler(msg); err != nil {
			fmt.Printf("Error handling broadcast message: %v\n", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	// Store subscription for cleanup
	nt.pubSubs[channel] = sub
	return nt.nc.Flush()
}

// Close shuts down the NATS transport.
func (nt *NATSTransport) Close() error {
	// Unsubscribe from all pub-sub channels
	for _, sub := range nt.pubSubs {
		_ = sub.Unsubscribe()
	}

	if nt.sub != nil {
		_ = nt.sub.Unsubscribe()
	}
	if nt.nc != nil {
		nt.nc.Close()
	}
	return nil
}
