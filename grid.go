package main

import (
	"fmt"
	"math/rand"
)

const (
	GridSize      = 20
	FireChance    = 0.02
	SpreadChance  = 0.10
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

func createGrid() [][]Cell {
	g := make([][]Cell, GridSize)
	for i := range g {
		g[i] = make([]Cell, GridSize)
	}
	return g
}

// igniteRandom may ignite a random empty cell with a new fire.
func igniteRandom(g [][]Cell, chance float64) {
	if rand.Float64() < chance {
		r := rand.Intn(GridSize)
		c := rand.Intn(GridSize)
		if g[r][c].State == Empty {
			g[r][c] = Cell{State: Fire, Intensity: 1}
		}
	}
}

// stepFires advances the fire dynamics by one tick: fires grow and may spread.
func stepFires(g [][]Cell) {
	ng := createGrid()
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			ng[r][c] = g[r][c]
			switch g[r][c].State {
			case Fire:
				ng[r][c].Intensity += GrowthPerTick
				trySpread(g, ng, r-1, c)
				trySpread(g, ng, r+1, c)
				trySpread(g, ng, r, c-1)
				trySpread(g, ng, r, c+1)
			case Extinguished:
				// stays extinguished
			case Empty:
				// nothing
			}
		}
	}
	for r := range g {
		copy(g[r], ng[r])
	}
}

func inBounds(r, c int) bool { return r >= 0 && r < GridSize && c >= 0 && c < GridSize }

func trySpread(g, ng [][]Cell, r, c int) {
	if !inBounds(r, c) {
		return
	}
	if g[r][c].State == Empty && rand.Float64() < SpreadChance {
		ng[r][c] = Cell{State: Fire, Intensity: 1}
	}
}

// extinguish applies up to `water` units to the cell at (r,c).
// Returns how much water was actually used.
func extinguish(g [][]Cell, r, c, water int) int {
	if !inBounds(r, c) || g[r][c].State != Fire || water <= 0 {
		return 0
	}
	used := water
	if g[r][c].Intensity < used {
		used = g[r][c].Intensity
	}
	g[r][c].Intensity -= used
	if g[r][c].Intensity <= 0 {
		g[r][c].Intensity = 0
		g[r][c].State = Extinguished
	}
	return used
}

// printGrid shows a compact view: . empty, F fire, E extinguished, T truck.
func printGrid(g [][]Cell, trucks []Firetruck) {
	overlay := map[[2]int]string{}
	for _, t := range trucks {
		overlay[[2]int{t.Row, t.Col}] = "T"
	}
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if v, ok := overlay[[2]int{r, c}]; ok {
				fmt.Print(v)
				continue
			}
			switch g[r][c].State {
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
