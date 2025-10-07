package main

import (
	"flag"
	"fmt"

	"Firetruck-sim/pkg/simulation"
)

func main() {
	steps := flag.Int("steps", 50, "simulation steps")
	trucks := flag.Int("trucks", 2, "number of fire trucks")
	water := flag.Int("water", 500, "initial water supply")
	flag.Parse()

	fmt.Printf("Starting fire simulation with %d trucks, %d water units, %d steps\n", 
		*trucks, *water, *steps)

	// Create and run simulation
	sim := simulation.NewSimulator(*trucks, *water)
	sim.Run(*steps)

	fmt.Println("\nSimulation completed!")
}