package simulation

import (
	"fmt"
	"math"

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

// logf logs a message with Lamport timestamp
func (t *Firetruck) logf(format string, a ...interface{}) {
	lt := t.Clock.Tick()
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
	got := tank.Withdraw(need)
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

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}