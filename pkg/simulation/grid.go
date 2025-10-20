package simulation

import (
	//"fmt"
	"math/rand"
	"time"
)

// Initialize random seed when package is loaded
func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	GridSize      = 20
	FireChance    = 0.03 // Reasonable fire ignition rate
	SpreadChance  = 0.02 // Low spread rate for demonstration
	GrowthPerTick = 1
)

type CellState int

const (
	Empty CellState = iota
	Fire
	Extinguished
)

type Cell struct {
	State     CellState
	Intensity int
}

// Grid represents the 2D simulation grid
type Grid struct {
	cells [][]Cell
}

// NewGrid creates a new empty grid
func NewGrid() *Grid {
	g := &Grid{
		cells: make([][]Cell, GridSize),
	}
	for i := range g.cells {
		g.cells[i] = make([]Cell, GridSize)
	}
	return g
}

// GetCell returns the cell at the given coordinates
func (g *Grid) GetCell(row, col int) Cell {
	if !g.InBounds(row, col) {
		return Cell{State: Empty, Intensity: 0}
	}
	return g.cells[row][col]
}

// SetCell sets the cell at the given coordinates
func (g *Grid) SetCell(row, col int, cell Cell) {
	if g.InBounds(row, col) {
		g.cells[row][col] = cell
	}
}

// InBounds checks if the coordinates are within the grid bounds
func (g *Grid) InBounds(row, col int) bool {
	return row >= 0 && row < GridSize && col >= 0 && col < GridSize
}

// GetCells returns the raw cell array (for compatibility)
func (g *Grid) GetCells() [][]Cell {
	return g.cells
}

// IgniteRandom may ignite a random empty cell with a new fire
func (g *Grid) IgniteRandom(chance float64) {
	if rand.Float64() < chance {
		r := rand.Intn(GridSize)
		c := rand.Intn(GridSize)
		if g.cells[r][c].State == Empty {
			g.cells[r][c] = Cell{State: Fire, Intensity: 1}
		}
	}
}

// StepFires advances the fire dynamics by one tick: fires grow and may spread
// Returns a list of new fire locations that were created by spreading
func (g *Grid) StepFires() []FireLocation {
	newCells := make([][]Cell, GridSize)
	for i := range newCells {
		newCells[i] = make([]Cell, GridSize)
		copy(newCells[i], g.cells[i])
	}

	var newFires []FireLocation

	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			switch g.cells[r][c].State {
			case Fire:
				newCells[r][c].Intensity += GrowthPerTick
				if g.trySpread(newCells, r-1, c) {
					newFires = append(newFires, FireLocation{Row: r - 1, Col: c})
				}
				if g.trySpread(newCells, r+1, c) {
					newFires = append(newFires, FireLocation{Row: r + 1, Col: c})
				}
				if g.trySpread(newCells, r, c-1) {
					newFires = append(newFires, FireLocation{Row: r, Col: c - 1})
				}
				if g.trySpread(newCells, r, c+1) {
					newFires = append(newFires, FireLocation{Row: r, Col: c + 1})
				}
			case Extinguished:
				// stays extinguished
			case Empty:
				// nothing
			}
		}
	}
	g.cells = newCells
	return newFires
}

// trySpread attempts to spread fire to adjacent cells
// Returns true if a new fire was created
func (g *Grid) trySpread(newCells [][]Cell, r, c int) bool {
	if !g.InBounds(r, c) {
		return false
	}
	if g.cells[r][c].State == Empty && rand.Float64() < SpreadChance {
		newCells[r][c] = Cell{State: Fire, Intensity: 1}
		return true
	}
	return false
}

// WaterCostForStep returns exponential cost for extinguishing one intensity step
func WaterCostForStep(intensity int) int {
	if intensity <= 0 {
		return 0
	}
	if intensity > 10 {
		intensity = 10
	}
	return 1 << intensity
}

// Extinguish applies up to `water` units to the cell at (r,c) using exponential cost.
// Returns how much water was actually used.
func (g *Grid) Extinguish(r, c, water int) int {
	if !g.InBounds(r, c) || g.cells[r][c].State != Fire || water <= 0 {
		return 0
	}

	totalUsed := 0
	for water > 0 && g.cells[r][c].Intensity > 0 {
		cost := WaterCostForStep(g.cells[r][c].Intensity)
		if water >= cost {
			g.cells[r][c].Intensity--
			water -= cost
			totalUsed += cost
		} else {
			break // not enough water for this step
		}
	}

	if g.cells[r][c].Intensity <= 0 {
		g.cells[r][c].State = Extinguished
	}
	return totalUsed
}

// FireLocation represents a fire location with its intensity
type FireLocation struct {
	Row       int
	Col       int
	Intensity int
}

// FindAllFires returns all fire locations on the grid
func (g *Grid) FindAllFires() []FireLocation {
	var fires []FireLocation
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if g.cells[r][c].State == Fire {
				fires = append(fires, FireLocation{
					Row:       r,
					Col:       c,
					Intensity: g.cells[r][c].Intensity,
				})
			}
		}
	}
	return fires
}
