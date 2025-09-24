package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Minimal console output that lists
	fmt.Println("Hejhejdå")

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
