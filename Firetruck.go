package main

import (
	"fmt"
	"math"
)

type Firetruck struct {
	ID       string
	Row, Col int
	Water    int
	MaxWater int
	Clock    *LamportClock
}

func NewFiretruck(id string, r, c int) Firetruck {
	return Firetruck{
		ID:       id,
		Row:      r,
		Col:      c,
		Water:    50,
		MaxWater: 100,
		Clock:    &LamportClock{},
	}
}

func (t *Firetruck) logf(format string, a ...any) {
	lt := t.Clock.Tick()
	prefix := fmt.Sprintf("[%s lt=%d] ", t.ID, lt)
	fmt.Println(prefix + fmt.Sprintf(format, a...))
}

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

func (t *Firetruck) OnFireCell(g [][]Cell) bool {
	return inBounds(t.Row, t.Col) && g[t.Row][t.Col].State == Fire
}

func (t *Firetruck) Extinguish(g [][]Cell) {
	if !t.OnFireCell(g) {
		return
	}
	if t.Water <= 0 {
		t.logf("no water to extinguish")
		return
	}
	used := extinguish(g, t.Row, t.Col, t.Water)
	t.Water -= used
	t.logf("extinguish used=%d remaining=%d", used, t.Water)
}

func (t *Firetruck) Refill(tank *WaterTank) {
	if t.Water >= t.MaxWater {
		return
	}
	need := t.MaxWater - t.Water
	got := tank.Withdraw(need)
	t.Water += got
	t.logf("refill got=%d new=%d", got, t.Water)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
