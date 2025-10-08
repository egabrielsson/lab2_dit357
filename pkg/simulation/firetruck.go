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
	ID        string
	Row, Col  int
	Water     int
	MaxWater  int
	Clock     *clock.LamportClock
	Transport transport.Transport
	Task      string // current task description
}

// NewFiretruck creates a new firetruck at the given position
func NewFiretruck(id string, r, c int) *Firetruck {
	return &Firetruck{
		ID:       id,
		Row:      r,
		Col:      c,
		Water:    50,
		MaxWater: 100,
		Clock:    clock.NewLamportClock(),
		Task:     "idle",
	}
}

// SetTransport sets the communication transport for this firetruck
func (t *Firetruck) SetTransport(transport transport.Transport) {
	t.Transport = transport
}

// logf logs a message with Lamport timestamp (disabled for clean simulation output)
func (t *Firetruck) logf(format string, a ...interface{}) {
	// Uncomment for debugging
	lt := t.Clock.Tick()
	_ = lt // prevent unused variable warning
	prefix := fmt.Sprintf("[%s lt=%d] ", t.ID, lt)
	fmt.Println(prefix + fmt.Sprintf(format, a...))
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
	
	t.logf("move to (%d,%d)", t.Row, t.Col)
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
		t.logf("no water to extinguish")
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
	t.logf("extinguish used=%d remaining=%d", used, t.Water)
}

// Refill refills the firetruck's water tank from the central water supply
func (t *Firetruck) Refill(tank *WaterTank) {
	if t.Water >= t.MaxWater {
		return
	}
	need := t.MaxWater - t.Water
	lamportTime := t.Clock.Tick()
	got := tank.Withdraw(need, t.ID, lamportTime)
	t.Water += got
	t.logf("refill got=%d new=%d", got, t.Water)
}

// GetPosition returns the current position of the firetruck
func (t *Firetruck) GetPosition() (int, int) {
	return t.Row, t.Col
}

// GetWater returns the current water level
func (t *Firetruck) GetWater() int {
	return t.Water
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
	
	if err := t.Transport.Publish(transport.ChannelTruckStatus, msg); err != nil {
		t.logf("failed to broadcast status: %v", err)
	}
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
		t.logf("bid for fire at (%d,%d): distance=%d water=%d", fireRow, fireCol, distance, t.Water)
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

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}