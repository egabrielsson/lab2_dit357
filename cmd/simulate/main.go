package main

import (
	"Firetruck-sim/pkg/simulation"
	"flag"
	"fmt"
)

func main() {
	// Random seed is initialized in pkg/simulation/grid.go init()

	steps := flag.Int("steps", 200, "simulation steps")
	trucks := flag.Int("trucks", 5, "number of fire trucks")
	water := flag.Int("water", 1000, "initial water supply")
	flag.Parse()

	fmt.Printf("Starting fire simulation with %d trucks, %d water units, %d steps\n",
		*trucks, *water, *steps)

	// Create and run simulation
	sim := simulation.NewSimulator(*trucks, *water)
	sim.Run(*steps)

	fmt.Println("\nSimulation completed!")
}
