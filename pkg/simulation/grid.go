package simulation

import (
	"fmt"
	"math/rand"
	"time"
)

// Initialize random seed when package is loaded
func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	GridSize      = 20
	FireChance    = 0.10  // Reasonable fire ignition rate
	SpreadChance  = 0.02  // Very low spread rate for clear demonstration
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
func (g *Grid) StepFires() {
	newCells := make([][]Cell, GridSize)
	for i := range newCells {
		newCells[i] = make([]Cell, GridSize)
		copy(newCells[i], g.cells[i])
	}

	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			switch g.cells[r][c].State {
			case Fire:
				newCells[r][c].Intensity += GrowthPerTick
				g.trySpread(newCells, r-1, c)
				g.trySpread(newCells, r+1, c)
				g.trySpread(newCells, r, c-1)
				g.trySpread(newCells, r, c+1)
			case Extinguished:
				// stays extinguished
			case Empty:
				// nothing
			}
		}
	}
	g.cells = newCells
}

// trySpread attempts to spread fire to adjacent cells
func (g *Grid) trySpread(newCells [][]Cell, r, c int) {
	if !g.InBounds(r, c) {
		return
	}
	if g.cells[r][c].State == Empty && rand.Float64() < SpreadChance {
		newCells[r][c] = Cell{State: Fire, Intensity: 1}
	}
}

// Extinguish applies up to `water` units to the cell at (r,c).
// Returns how much water was actually used.
func (g *Grid) Extinguish(r, c, water int) int {
	if !g.InBounds(r, c) || g.cells[r][c].State != Fire || water <= 0 {
		return 0
	}
	used := water
	if g.cells[r][c].Intensity < used {
		used = g.cells[r][c].Intensity
	}
	g.cells[r][c].Intensity -= used
	if g.cells[r][c].Intensity <= 0 {
		g.cells[r][c].Intensity = 0
		g.cells[r][c].State = Extinguished
	}
	return used
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

// Print shows a compact view: . empty, F fire, E extinguished, T truck
func (g *Grid) Print(trucks []*Firetruck) {
	overlay := make(map[[2]int]string)
	for _, t := range trucks {
		overlay[[2]int{t.Row, t.Col}] = "T"
	}
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if v, ok := overlay[[2]int{r, c}]; ok {
				fmt.Print(v)
				continue
			}
			switch g.cells[r][c].State {
			case Empty:
				fmt.Print(".")
			case Fire:
				fmt.Print("F")
			case Extinguished:
				fmt.Print("E")
			}
		}
		fmt.Println()
	}
}