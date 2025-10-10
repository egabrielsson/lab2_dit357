package transport

import (
	"Firetruck-sim/pkg/message"
)

type Transport interface {
	// GetID returns the unique identifier of this transport node
	GetID() string

	// Publish broadcasts a message to all subscribers of a channel
	Publish(channel string, msg message.Message) error

	// Subscribe starts listening to broadcast messages on a channel
	Subscribe(channel string, handler func(message.Message) error) error

	// Close shuts down the transport
	Close() error
}

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
