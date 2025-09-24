package main

import (
	"math/rand"
	"time"
	"fmt"
)

const (
	GridSize    = 20   // The set gridsize
	FireChance  = 0.01 // Chance of a fire starting
	SpreadChance = 0.05  // Chance of the fire spreading
)

// Defines Cellstate as int
type CellState int

// Creates constants for Cellstate with iota
// Which sets Empty to 0, Fire 1
const (
	NoFire CellState = iota
	Fire
)

// Defines Cell as a Struct that holds
// State of the Cell
type Cell struct {
	State CellState
}
// This is a function that randomizes the seed so that the outcome
// Of the fire starting and spreading is different each time you run
// the program.
func init(){
	rand.Seed(time.Now().UnixNano())
}

// Creates the grid without taking in parameters
// Which always creates the GridSize at the set value of
// 20. It then returns a grid that has all the values
// assigned to 0.
func createGrid() [][]Cell {
	rows := GridSize
	cols := GridSize

	grid := make([][]Cell, rows)

	for i := range grid {
		grid[i] = make([]Cell, cols)
	}
	return grid
}

// This function takes a grid and loops trough it
// At each place in the grid it
func igniteRandom(grid [][]Cell) {
	for i := range grid {
		for j := range grid[i] {
			if grid[i][j].State == NoFire && rand.Float64() < FireChance {
				grid[i][j].State = Fire
			}

		}
	}

}

func SpreadFires(grid [][]Cell) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		fmt.Println("No active fires")
		return
	}
	rows, cols := len(grid), len(grid[0])

	// Makes a copy of the currect state so that newly ignited
	// Cells won't spread in the same step (exponentially)
	newGrid := make([][]Cell, rows)
	for i := range newGrid{
		newGrid[i] = make([]Cell, cols)
		copy(newGrid[i], grid[i])
	}

	// Declares directions to hold up, down, left and right
	direction := [][2]int{{-1, 0}, {+1, 0}, {0, -1}, {0, +1}}


	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if grid[i][j].State == Fire {
				for _, dirValue := range direction{
					ni, nj := i+dirValue[0], j+dirValue[1] // ni is neighbours of i and nj is neighbours of j (up, down, left, right)
					if ni >= 0 && ni < rows && nj >= 0 && nj < cols && grid[ni][nj].State == NoFire {
						if rand.Float64() < SpreadChance {
							newGrid[ni][nj].State = Fire
						}
					}
				}
			}
		}
	}

	for i := range grid {
		copy(grid[i], newGrid[i])
	}

}
