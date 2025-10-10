package transport

import (
	"time"

	"Firetruck-sim/pkg/message"
)

// Transport defines the interface for communication between fire trucks.
// Both TCP and NATS implementations must satisfy this interface.
type Transport interface {
	// GetID returns the unique identifier of this transport node
	GetID() string

	// Send sends a one-way message to another node
	Send(toID string, msg message.Message) error

	// Request sends a message and waits for a reply
	Request(toID string, msg message.Message, timeout time.Duration) (message.Message, error)

	// Listen starts listening for incoming messages with the given handler
	// The handler should return a reply message if one is needed
	Listen(handler func(message.Message) (*message.Message, error)) error

	// Publish broadcasts a message to all subscribers of a channel
	Publish(channel string, msg message.Message) error

	// Subscribe starts listening to broadcast messages on a channel
	Subscribe(channel string, handler func(message.Message) error) error

	// Close shuts down the transport
	Close() error
}

// MessageHandler is a function that processes incoming messages.
// It should return a reply message if one is needed, or nil for no reply.
type MessageHandler func(message.Message) (*message.Message, error)

// SubscriptionHandler is a function that processes broadcast messages.
// No reply is expected for broadcast messages.
type SubscriptionHandler func(message.Message) error

// Common broadcast channels for coordination
const (
	ChannelFireAlerts     = "fires.alerts"
	ChannelTruckStatus    = "trucks.status"
	ChannelWaterRequests  = "water.requests"
	ChannelCoordination   = "coordination"
	ChannelFireBids       = "fires.bids"
	ChannelFireAssignment = "fires.assignment"
)