package simulation

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"
	"Firetruck-sim/pkg/transport"
)

// Firetruck represents a fire-fighting truck agent
type Firetruck struct {
	ID           string
	Row, Col     int
	Water        int
	MaxWater     int
	Clock        *clock.LamportClock
	Transport    transport.Transport
	Task         string        // current task description
	AssignedFire *FireLocation // current fire assignment (nil if none)

	// Ricart-Agrawala state for water mutual exclusion
	ra            raState
	myReqTS       int
	replies       map[string]bool
	deferred      map[string]bool
	peers         map[string]bool
	lowWaterThresh int
}

type raState int
const (
	raIdle raState = iota
	raRequesting
	raHeld
)

// NewFiretruck creates a new firetruck at the given position
func NewFiretruck(id string, r, c int) *Firetruck {
	return &Firetruck{
		ID:             id,
		Row:            r,
		Col:            c,
		Water:          30, // Start with some water
		MaxWater:       50,
		Clock:          clock.NewLamportClock(),
		Task:           "idle",
		ra:             raIdle,
		replies:        make(map[string]bool),
		deferred:       make(map[string]bool),
		peers:          make(map[string]bool),
		lowWaterThresh: 10,
	}
}

// GetStartingPosition returns the starting position for a truck based on its ID
// This centralizes the spawn location logic used across simulation and NATS demo
func GetStartingPosition(truckID string, gridSize int) (row, col int) {
	switch truckID {
	case "T1":
		return 0, 0
	case "T2":
		return gridSize - 1, gridSize - 1
	case "T3":
		return 0, gridSize - 1
	case "T4":
		return gridSize - 1, 0
	default:
		// For T5+ or other IDs, place in center
		return gridSize / 2, gridSize / 2
	}
}

// SetTransport sets the communication transport for this firetruck
func (t *Firetruck) SetTransport(transport transport.Transport) {
	t.Transport = transport
}

// logf logs a simple message
func (t *Firetruck) logf(format string, a ...interface{}) {
	fmt.Printf("[%s] ", t.ID)
	fmt.Printf(format, a...)
	fmt.Println()
}

// MoveToward moves the firetruck one step toward the target coordinates
func (t *Firetruck) MoveToward(targetR, targetC int) {
	oldRow, oldCol := t.Row, t.Col

	dr := int(math.Copysign(1, float64(targetR-t.Row)))
	if targetR == t.Row {
		dr = 0
	}
	dc := int(math.Copysign(1, float64(targetC-t.Col)))
	if targetC == t.Col {
		dc = 0
	}
	// prefer vertical if far away vertically
	if abs(targetR-t.Row) >= abs(targetC-t.Col) && dr != 0 {
		t.Row += dr
	} else if dc != 0 {
		t.Col += dc
	}

	// Announce movement intention if transport is available and we actually moved
	if t.Transport != nil && (oldRow != t.Row || oldCol != t.Col) {
		t.AnnounceIntention("moving", targetR, targetC, map[string]interface{}{
			"from_row": oldRow,
			"from_col": oldCol,
		})
	}

	t.logf("moved to (%d,%d)", t.Row, t.Col)
}

// OnFireCell checks if the firetruck is currently on a cell with fire
func (t *Firetruck) OnFireCell(grid *Grid) bool {
	cell := grid.GetCell(t.Row, t.Col)
	return cell.State == Fire
}

// Extinguish attempts to extinguish fire at the firetruck's current position
func (t *Firetruck) Extinguish(grid *Grid) {
	if !t.OnFireCell(grid) {
		return
	}
	if t.Water <= 0 {
		t.logf("no water to extinguish fire")
		return
	}

	// Get fire intensity before extinguishing
	cell := grid.GetCell(t.Row, t.Col)
	fireIntensity := cell.Intensity

	used := grid.Extinguish(t.Row, t.Col, t.Water)
	t.Water -= used

	// Broadcast fire alert if we discovered a new fire
	if t.Transport != nil && fireIntensity > 0 {
		t.BroadcastFireAlert(t.Row, t.Col, fireIntensity)
	}

	t.SetTask("extinguishing")
	t.logf("extinguished fire, used %d water, remaining %d/%d", used, t.Water, t.MaxWater)
}

// GetPosition returns the current position of the firetruck
func (t *Firetruck) GetPosition() (int, int) {
	return t.Row, t.Col
}

// GetWater returns the current water level
func (t *Firetruck) GetWater() int {
	return t.Water
}

// GetLowWaterThresh returns the low water threshold
func (t *Firetruck) GetLowWaterThresh() int {
	return t.lowWaterThresh
}

// SetTask sets the current task and broadcasts status if transport is available
func (t *Firetruck) SetTask(task string) {
	t.Task = task
	t.BroadcastStatus()
}

// BroadcastFireAlert sends a fire alert to all trucks
func (t *Firetruck) BroadcastFireAlert(row, col, intensity int) {
	if t.Transport == nil {
		return
	}

	msg := message.NewMessage(
		message.TypeFireAlert,
		t.ID,
		message.FireAlertPayload(row, col, intensity),
	)

	if err := t.Transport.Publish(transport.ChannelFireAlerts, msg); err != nil {
		t.logf("failed to broadcast fire alert: %v", err)
	} else {
		t.logf("broadcast fire alert at (%d,%d) intensity=%d", row, col, intensity)
	}
}

// BroadcastStatus sends current status to all trucks
func (t *Firetruck) BroadcastStatus() {
	if t.Transport == nil {
		return
	}

	msg := message.NewMessage(
		message.TypeTruckStatus,
		t.ID,
		message.TruckStatusPayload(t.Row, t.Col, t.Water, t.MaxWater, t.Task),
	)
	msg.Lamport = t.Clock.Tick()

	if err := t.Transport.Publish(transport.ChannelTruckStatus, msg); err != nil {
		t.logf("failed to broadcast status: %v", err)
	}
}

// AddWater adds water to the firetruck's tank (used when receiving from water supply)
func (t *Firetruck) AddWater(amount int) {
	t.Water += amount
	if t.Water > t.MaxWater {
		t.Water = t.MaxWater
	}
	t.logf("received water=%d new_total=%d", amount, t.Water)
}

// AnnounceIntention broadcasts coordination message about planned action
func (t *Firetruck) AnnounceIntention(action string, targetRow, targetCol int, details map[string]interface{}) {
	if t.Transport == nil {
		return
	}

	msg := message.NewMessage(
		message.TypeCoordination,
		t.ID,
		message.CoordinationPayload(action, targetRow, targetCol, details),
	)

	if err := t.Transport.Publish(transport.ChannelCoordination, msg); err != nil {
		t.logf("failed to announce intention: %v", err)
	} else {
		t.logf("announced intention: %s to (%d,%d)", action, targetRow, targetCol)
	}
}

// BidForFire broadcasts a bid to respond to a fire at the given location
func (t *Firetruck) BidForFire(fireRow, fireCol int) {
	if t.Transport == nil {
		return
	}

	distance := abs(t.Row-fireRow) + abs(t.Col-fireCol) // Manhattan distance

	msg := message.NewMessage(
		message.TypeFireBid,
		t.ID,
		message.FireBidPayload(fireRow, fireCol, distance, t.Water, t.ID),
	)
	msg.Lamport = t.Clock.Tick()

	if err := t.Transport.Publish(transport.ChannelFireBids, msg); err != nil {
		t.logf("failed to broadcast fire bid: %v", err)
	} else {
		t.logf("bidding for fire at (%d,%d), distance %d, water %d", fireRow, fireCol, distance, t.Water)
	}
}

// CalculateDistance returns Manhattan distance to target
func (t *Firetruck) CalculateDistance(targetRow, targetCol int) int {
	return abs(t.Row-targetRow) + abs(t.Col-targetCol)
}

// FireBid represents a truck's bid to respond to a fire
type FireBid struct {
	TruckID  string
	Distance int
	Water    int
	Lamport  int64
}

// EvaluateFireBids determines which truck should respond to a fire
// Rules: 1) Closest, 2) Most water if tied, 3) Lowest ID if still tied
func EvaluateFireBids(bids []FireBid) (winner string, reason string) {
	if len(bids) == 0 {
		return "", "no bids"
	}

	// Sort by: distance (asc), water (desc), ID (asc)
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].Distance != bids[j].Distance {
			return bids[i].Distance < bids[j].Distance
		}
		if bids[i].Water != bids[j].Water {
			return bids[i].Water > bids[j].Water
		}
		return strings.Compare(bids[i].TruckID, bids[j].TruckID) < 0
	})

	winner = bids[0].TruckID

	// Build reason string
	if len(bids) == 1 {
		reason = "only bidder"
	} else if bids[0].Distance < bids[1].Distance {
		reason = fmt.Sprintf("closest (distance=%d)", bids[0].Distance)
	} else if bids[0].Water > bids[1].Water {
		reason = fmt.Sprintf("most water (water=%d, tied distance=%d)", bids[0].Water, bids[0].Distance)
	} else {
		reason = fmt.Sprintf("lowest ID (tied distance=%d, water=%d)", bids[0].Distance, bids[0].Water)
	}

	return winner, reason
}

// Abs returns the absolute value of an integer (exported for use in other packages)
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// abs is kept for internal backwards compatibility
func abs(x int) int {
	return Abs(x)
}

// Ricart-Agrawala mutual exclusion methods

// StartRA initializes RA subscriptions and peer discovery
func (t *Firetruck) StartRA() {
	// Subscribe to RA channels
	t.Transport.Subscribe(transport.ChannelWaterReq, t.handleWaterReq)
	t.Transport.Subscribe(transport.ChannelWaterReply, t.handleWaterReply)
	t.Transport.Subscribe(transport.ChannelWaterRelease, t.handleWaterRelease)
	t.Transport.Subscribe(transport.ChannelTruckStatus, t.handleTruckStatus)
}

// handleTruckStatus discovers peers
func (t *Firetruck) handleTruckStatus(msg message.Message) error {
	if msg.From != t.ID {
		t.peers[msg.From] = true
	}
	return nil
}

// RequestWaterRA initiates Ricart-Agrawala protocol for water refill
func (t *Firetruck) RequestWaterRA() {
	if t.ra != raIdle || t.Water > t.lowWaterThresh {
		return
	}

	t.ra = raRequesting
	t.myReqTS = int(t.Clock.Tick())
	t.replies = make(map[string]bool)

	t.logf("[ME] REQUEST ts=%d", t.myReqTS)

	// Send request to all peers
	req := message.Message{
		Type:    message.TypeWaterReq,
		From:    t.ID,
		Lamport: int64(t.myReqTS),
		Payload: map[string]interface{}{"ts": t.myReqTS},
	}
	t.Transport.Publish(transport.ChannelWaterReq, req)
}

// handleWaterReq processes incoming water requests
func (t *Firetruck) handleWaterReq(msg message.Message) error {
	ts := int(msg.Payload["ts"].(float64))

	if t.ra == raHeld || (t.ra == raRequesting && (ts > t.myReqTS || (ts == t.myReqTS && msg.From > t.ID))) {
		// Defer reply
		t.deferred[msg.From] = true
		t.logf("[ME] DEFER %s", msg.From)
	} else {
		// Reply immediately
		reply := message.Message{
			Type:    message.TypeWaterReply,
			From:    t.ID,
			Lamport: t.Clock.Tick(),
		}
		t.Transport.Publish(transport.ChannelWaterReply, reply)
		t.logf("[ME] REPLY-> %s", msg.From)
	}
	return nil
}

// handleWaterReply processes replies
func (t *Firetruck) handleWaterReply(msg message.Message) error {
	if t.ra == raRequesting {
		t.replies[msg.From] = true
		// Check if we have all replies
		allReplied := true
		for peer := range t.peers {
			if peer != t.ID && !t.replies[peer] {
				allReplied = false
				break
			}
		}
		if allReplied {
			t.enterCS()
		}
	}
	return nil
}

// handleWaterRelease processes releases
func (t *Firetruck) handleWaterRelease(msg message.Message) error {
	if t.deferred[msg.From] {
		delete(t.deferred, msg.From)
		reply := message.Message{
			Type:    message.TypeWaterReply,
			From:    t.ID,
			Lamport: t.Clock.Tick(),
		}
		t.Transport.Publish(transport.ChannelWaterReply, reply)
		t.logf("[ME] REPLY-> %s (deferred)", msg.From)
	}
	return nil
}

// enterCS enters the critical section (water refill)
func (t *Firetruck) enterCS() {
	t.ra = raHeld
	t.Water = t.MaxWater
	t.logf("[ME] ENTER CS (refill)")

	// Exit CS immediately after refill
	t.exitCS()
}

// exitCS exits the critical section
func (t *Firetruck) exitCS() {
	t.ra = raIdle

	// Send release to all peers
	release := message.Message{
		Type:    message.TypeWaterRelease,
		From:    t.ID,
		Lamport: t.Clock.Tick(),
	}
	t.Transport.Publish(transport.ChannelWaterRelease, release)
	t.logf("[ME] RELEASE")

	// Reply to all deferred requests
	for peer := range t.deferred {
		delete(t.deferred, peer)
		reply := message.Message{
			Type:    message.TypeWaterReply,
			From:    t.ID,
			Lamport: t.Clock.Tick(),
		}
		t.Transport.Publish(transport.ChannelWaterReply, reply)
		t.logf("[ME] REPLY-> %s (deferred)", peer)
	}
}
