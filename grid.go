package main

import (
	"math/rand"
)

const (
	GridSize    = 20   // The set gridsize
	FireChance  = 0.01 // Chance of a fire starting
	SpeadChance = 0.1  // Chance of the fire spreading
)

// Defines Cellstate as int 0-2
type CellState int

// Creates constants for Cellstate with iota
// Which sets Empty to 0, Fire 1 and Ext 2
const (
	NoFire CellState = iota
	Fire
)

// Defines Cell as a Struct that holds
// State of the Cell
type Cell struct {
	State CellState
}

// Creates the grid without taking in parameters
// Which always creates the GridSize at the set value of
// 20. It then returns a grid that has all the values
// assigned to 0.
func createGrid() [][]int {
	rows := GridSize
	cols := GridSize

	grid := make([][]int, rows)

	for i := range grid {
		grid[i] = make([]int, cols)
	}
	return grid
}

// This function takes a grid and loops trough it
// At each place in the grid it
func igniteRandom(grid [][]Cell) {
	for i := range grid {
		for j := range grid[i] {
			if grid[ı][j].State == NoFire && rand.ExpFloat64() < FireChance {
				grid[i][j].State = Fire
			}

		}
	}

}

func SpreadFires(grid [][]Cell) {

}
