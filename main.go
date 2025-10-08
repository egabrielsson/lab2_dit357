// main.go - Legacy simulation runner
// For new modular approach, use: go run cmd/simulate/main.go
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"Firetruck-sim/pkg/simulation"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	steps := flag.Int("steps", 50, "simulation steps")
	flag.Parse()

	fmt.Printf("Running fire simulation for %d steps\n", *steps)
	fmt.Println()

	// Create simulation with proper random seed
	sim := simulation.NewSimulator(2, 500)
	sim.Run(*steps)
}
