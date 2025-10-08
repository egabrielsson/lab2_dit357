package simulation

import (
	"fmt"
	"time"
)

// Simulator manages the overall fire simulation
type Simulator struct {
	Grid      *Grid
	Trucks    []*Firetruck
	WaterTank *WaterTank
	Steps     int
}

// NewSimulator creates a new simulation with the given parameters
func NewSimulator(numTrucks int, waterSupply int) *Simulator {
	grid := NewGrid()
	tank := NewWaterTank(waterSupply)
	
	trucks := make([]*Firetruck, numTrucks)
	for i := 0; i < numTrucks; i++ {
		// Place trucks at different corners/positions
		var row, col int
		switch i % 4 {
		case 0:
			row, col = 0, 0
		case 1:
			row, col = GridSize-1, GridSize-1
		case 2:
			row, col = 0, GridSize-1
		case 3:
			row, col = GridSize-1, 0
		}
		
		truck := NewFiretruck(fmt.Sprintf("T%d", i+1), row, col)
		trucks[i] = truck
	}
	
	return &Simulator{
		Grid:      grid,
		Trucks:    trucks,
		WaterTank: tank,
		Steps:     0,
	}
}

// Step advances the simulation by one timestep
func (s *Simulator) Step() {
	s.Steps++
	
	// 1. Random fire ignition
	s.Grid.IgniteRandom(FireChance)
	
	// 2. Fire spreading and growth
	s.Grid.StepFires()
	
	// 3. Truck actions (this is where distributed coordination would go)
	s.simpleFirefightingPolicy()
}

// simpleFirefightingPolicy implements a basic centralized policy for testing
// In the full solution, this would be replaced by distributed coordination
func (s *Simulator) simpleFirefightingPolicy() {
	fireR, fireC, found := s.Grid.FindFirstFire()
	
	for _, truck := range s.Trucks {
		// Refill if water is low
		if truck.GetWater() < 10 {
			truck.Refill(s.WaterTank)
		}
		
		// Move toward fire if one exists
		if found {
			row, col := truck.GetPosition()
			if row == fireR && col == fireC {
				truck.Extinguish(s.Grid)
			} else {
				truck.MoveToward(fireR, fireC)
				// Check if moved onto a fire cell
				if truck.OnFireCell(s.Grid) {
					truck.Extinguish(s.Grid)
				}
			}
		}
	}
}

// Run runs the simulation for the specified number of steps
func (s *Simulator) Run(maxSteps int) {
	for step := 1; step <= maxSteps; step++ {
		fmt.Printf("\nStep %d\n", step)
		s.Step()
		s.Print()
		
		// Small delay for visibility
		time.Sleep(100 * time.Millisecond)
	}
}

// Print displays the current state of the simulation
func (s *Simulator) Print() {
	s.Grid.Print(s.Trucks)
	fmt.Printf("Water tank: %d units\n", s.WaterTank.GetStock())
}

// GetTruck returns a truck by its ID
func (s *Simulator) GetTruck(id string) *Firetruck {
	for _, truck := range s.Trucks {
		if truck.ID == id {
			return truck
		}
	}
	return nil
}

// AddTruck adds a new truck to the simulation
func (s *Simulator) AddTruck(truck *Firetruck) {
	s.Trucks = append(s.Trucks, truck)
}