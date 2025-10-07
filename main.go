// main.go
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	steps := flag.Int("steps", 50, "simulation steps")
	flag.Parse()

	// Grid & trucks
	grid := createGrid()
	trucks := []Firetruck{
		NewFiretruck("T1", 0, 0),
		NewFiretruck("T2", GridSize-1, GridSize-1),
	}
	tank := NewWaterTank(500)

	for step := 1; step <= *steps; step++ {
		fmt.Printf("\nStep %d\n", step)

		igniteRandom(grid, FireChance)
		stepFires(grid)

		// simple policy: each truck moves toward the first fire found; if on fire, extinguish; if water low, refill.
		fireR, fireC, found := findFirstFire(grid)
		for i := range trucks {
			t := &trucks[i]
			if t.Water < 10 {
				t.Refill(tank)
			}
			if found {
				if t.Row == fireR && t.Col == fireC {
					t.Extinguish(grid)
				} else {
					t.MoveToward(fireR, fireC)
					if t.OnFireCell(grid) {
						t.Extinguish(grid)
					}
				}
			}
		}

		printGrid(grid, trucks)
	}
}

func findFirstFire(g [][]Cell) (int, int, bool) {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if g[r][c].State == Fire {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}
