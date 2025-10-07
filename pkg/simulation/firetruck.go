package simulation

import (
	"fmt"
	"math"

	"Firetruck-sim/pkg/clock"
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
	used := grid.Extinguish(t.Row, t.Col, t.Water)
	t.Water -= used
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

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}