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

// simpleFirefightingPolicy implements distributed fire assignment
// Multiple trucks can work on different fires simultaneously
func (s *Simulator) simpleFirefightingPolicy() {
	// Find ALL fires on the grid
	fires := s.Grid.FindAllFires()

	if len(fires) == 0 {
		// No fires - clear all assignments and handle refills
		for _, truck := range s.Trucks {
			truck.AssignedFire = nil
			if truck.GetWater() < 10 {
				truck.Refill(s.WaterTank)
			}
		}
		return
	}

	// Track which trucks get assigned
	assigned := make(map[string]*FireLocation)

	// For each fire, find the best available truck
	for i := range fires {
		fire := &fires[i]
		var bids []FireBid

		for _, truck := range s.Trucks {
			// Skip if already assigned to another fire or low water
			if assigned[truck.ID] != nil || truck.GetWater() < 10 {
				continue
			}

			distance := truck.CalculateDistance(fire.Row, fire.Col)
			bids = append(bids, FireBid{
				TruckID:  truck.ID,
				Distance: distance,
				Water:    truck.GetWater(),
				Lamport:  truck.Clock.Tick(),
			})
		}

		if len(bids) == 0 {
			continue // No available trucks for this fire
		}

		// Evaluate bids to determine winner for this fire
		winnerID, reason := EvaluateFireBids(bids)
		assigned[winnerID] = fire

		fmt.Printf("Fire at (%d,%d): %s assigned to %s\n", fire.Row, fire.Col, reason, winnerID)
	}

	// Execute actions for ALL trucks simultaneously
	for _, truck := range s.Trucks {
		if assignedFire := assigned[truck.ID]; assignedFire != nil {
			// This truck won a bid - assign and move toward fire
			truck.AssignedFire = assignedFire

			row, col := truck.GetPosition()
			if row == assignedFire.Row && col == assignedFire.Col {
				// At the fire location - extinguish it
				truck.Extinguish(s.Grid)
				truck.AssignedFire = nil // Clear assignment after extinguishing
			} else {
				// Move toward assigned fire
				truck.MoveToward(assignedFire.Row, assignedFire.Col)
				// Check if we moved onto another fire
				if truck.OnFireCell(s.Grid) {
					truck.Extinguish(s.Grid)
				}
			}
		} else {
			// Not assigned to any fire
			truck.AssignedFire = nil

			if truck.GetWater() < 10 {
				// Low water - refill
				truck.Refill(s.WaterTank)
			} else {
				// Has water but no assignment - stay idle
				fmt.Printf("[%s] Idle - no fires available\n", truck.ID)
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
