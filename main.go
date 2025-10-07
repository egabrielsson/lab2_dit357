// main.go - Legacy simulation runner
// For new modular approach, use: go run cmd/simulate/main.go
package main

import (
	"flag"
	"fmt"

	"Firetruck-sim/pkg/simulation"
)

func main() {
	steps := flag.Int("steps", 50, "simulation steps")
	flag.Parse()

	fmt.Printf("Running legacy simulation for %d steps\n", *steps)
	fmt.Println("For the new modular version, use: go run cmd/simulate/main.go")
	fmt.Println()

	// Create simple simulation
	sim := simulation.NewSimulator(2, 500)
	sim.Run(*steps)
}
