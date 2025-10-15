package transport

import (
	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"
)

type Transport interface {
	// GetID returns the unique identifier of this transport node
	GetID() string

	// Publish broadcasts a message to all subscribers of a channel
	Publish(channel string, msg message.Message) error

	// Subscribe starts listening to broadcast messages on a channel
	Subscribe(channel string, handler SubscriptionHandler) error

	// SetClock sets the shared Lamport clock for this transport
	SetClock(clock *clock.LamportClock)

	// Close shuts down the transport
	Close() error
}

// SubscriptionHandler is a function that processes broadcast messages.
// No reply is expected for broadcast messages.
type SubscriptionHandler func(message.Message) error

// Common broadcast channels for coordination
const (
	ChannelFireAlerts     = "fires.alerts"      // FireAnnounce
	ChannelFireBids       = "fires.bids"        // Bid
	ChannelFireDecision   = "fires.decision"    // BidDecision
	ChannelTruckStatus    = "trucks.status"     // discovery/heartbeats
	ChannelWorldTick      = "world.tick"        // optional deterministic ticks

	// Ricart–Agrawala for water (NEW)
	ChannelWaterReq       = "water.req"
	ChannelWaterReply     = "water.reply"
	ChannelWaterRelease   = "water.release"

	// Legacy (keep for compatibility)
	ChannelCoordination   = "coordination"
)
