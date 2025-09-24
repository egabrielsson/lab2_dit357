package main

import (
    "fmt"
    "time"
)

func printGrid(grid [][]Cell) {
    for i := range grid {
        for j := range grid[i] {
            if grid[i][j].State == Fire {
                fmt.Print("F")
            } else {
                fmt.Print(".")
            }
        }
        fmt.Println()
    }
}

func main() {
    grid := createGrid()

    steps := 50
    for s := 0; s < steps; s++ {
        igniteRandom(grid)
        SpreadFires(grid)

        fmt.Printf("Step %d\n", s)
        printGrid(grid)
        fmt.Println()

        time.Sleep(150 * time.Millisecond)
    }
}

/*
a) Implement a grid (e.g., 20×20) where fires can randomly ignite at different locations
and spread over time if not extinguished.
b) Implement basic firetruck agents that can move around the grid (north, south, east,
west), request water, and extinguish fire.
c) Define simple rules for fire dynamics: e.g., each timestep, a fire grows in intensity if
not attended; extinguishing requires water proportional to intensity.
d) Provide a minimal console output that lists / shows grid state (fires, firetrucks,
extinguished cells).

*/
