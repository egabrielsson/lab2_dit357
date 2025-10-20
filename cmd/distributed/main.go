package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"Firetruck-sim/pkg/clock"
	"Firetruck-sim/pkg/message"
	"Firetruck-sim/pkg/simulation"
	"Firetruck-sim/pkg/transport"
)

func main() {
	// Command-line flags
	id := flag.String("id", "T1", "node identifier")
	natsURL := flag.String("nats", "nats://127.0.0.1:4222", "NATS server URL")
	role := flag.String("role", "truck", "role: truck, water-supply, observer")
	flag.Parse()

	// Connect to NATS
	t, err := transport.NewNATSTransport(*id, *natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer t.Close()

	// fmt.Printf("\n╔═══════════════════════════════════════════════════╗\n")
	// fmt.Printf("║  DECENTRALIZED Fire Truck System                 ║\n")
	// fmt.Printf("║  Node: %-42s ║\n", *id)
	// fmt.Printf("║  Role: %-42s ║\n", *role)
	// fmt.Printf("╚═══════════════════════════════════════════════════╝\n\n")

	// Launch appropriate role
	switch *role {
	case "truck":
		runFireTruck(t, *id)
	case "observer":
		runObserver(t, *id)
	default:
		log.Fatalf("Unknown role: %s. Valid roles: truck, observer", *role)
	}
}

// runFireTruck operates as an autonomous fire-fighting agent
func runFireTruck(t *transport.NATSTransport, truckID string) {
	// Initialize truck at starting position
	row, col := simulation.GetStartingPosition(truckID, simulation.GridSize)
	truck := simulation.NewFiretruck(truckID, row, col)
	truck.SetTransport(t)

	// Initialize Ricart-Agrawala for water
	truck.StartRA()

	// Lamport clock that is shared between truck and transport
	sharedClock := truck.Clock
	t.SetClock(sharedClock)

	// Initialize local grid simulation
	grid := simulation.NewGrid()

	// Track if currently assigned to a fire
	var assignedMu sync.Mutex
	var currentAssignment *simulation.FireLocation

	// Track last time we saw/announced a fire
	var fireMu sync.Mutex
	lastFireSeen := time.Now()

	// Bid collection and evaluation
	var mu sync.Mutex
	bidsByFire := make(map[string][]message.Message)
	timers := make(map[string]*time.Timer)

	log.Printf("Truck %s initialized at (%d,%d) with %d/%d water", truckID, row, col, truck.Water, truck.MaxWater)

	// Broadcast initial status
	truck.BroadcastStatus()

	// Subscribe to fire alerts and bid on fires
	t.Subscribe(transport.ChannelFireAlerts, func(msg message.Message) error {
		// Update Lamport clock on message receive
		sharedClock.Receive(msg.Lamport)

		// Check if already assigned
		assignedMu.Lock()
		if currentAssignment != nil {
			assignedMu.Unlock()
			return nil
		}
		assignedMu.Unlock()

		// Handle both old FireAlert and new FireAnnounce formats

		var fireRow, fireCol, intensity int
		if row, ok := msg.Payload["row"].(float64); ok {
			// Old format
			fireRow = int(row)
			fireCol = int(msg.Payload["col"].(float64))
			intensity = int(msg.Payload["intensity"].(float64))
		} else {
			// New FireAnnounce format
			fireRow = int(msg.Payload["id_x"].(float64))
			fireCol = int(msg.Payload["id_y"].(float64))
			intensity = int(msg.Payload["intensity"].(float64))
		}

		log.Printf("Truck %s: Fire alert received at (%d,%d), intensity %d", truckID, fireRow, fireCol, intensity)

		// Update local grid view
		grid.SetCell(fireRow, fireCol, simulation.Cell{
			State:     simulation.Fire,
			Intensity: intensity,
		})

		// Update last seen fire time
		fireMu.Lock()
		lastFireSeen = time.Now()
		fireMu.Unlock()

		// Bid if we have sufficient water to respond
		if truck.GetWater() >= intensity {
			distance := simulation.Abs(truck.Row-fireRow) + simulation.Abs(truck.Col-fireCol)
			// Lower score equals lower distance to fire
			score := distance

			// Broadcast bid using new typed message
			lamportTs := sharedClock.Tick()
			bid := message.Bid{
				Fire:    message.FireID{X: fireRow, Y: fireCol},
				Bidder:  truckID,
				Score:   score,
				Lamport: int(lamportTs),
			}

			bidMsg := message.Message{
				Type:    message.TypeBid,
				From:    truckID,
				Lamport: lamportTs,
				Payload: map[string]interface{}{
					"fire_x":  float64(fireRow),
					"fire_y":  float64(fireCol),
					"bidder":  bid.Bidder,
					"score":   float64(bid.Score),
					"lamport": float64(bid.Lamport),
				},
			}
			t.Publish(transport.ChannelFireBids, bidMsg)
			log.Printf("Truck %s: bid fire=(%d,%d) score=%d ts=%d", truckID, fireRow, fireCol, score, bid.Lamport)

			// Add own bid to local collection
			fireKey := fmt.Sprintf("%v,%v", fireRow, fireCol)
			mu.Lock()
			bidsByFire[fireKey] = append(bidsByFire[fireKey], bidMsg)

			// Start timer if not already running for this fire
			if timers[fireKey] == nil {
				timers[fireKey] = time.AfterFunc(1*time.Second, func() {
					mu.Lock()
					bids := bidsByFire[fireKey]
					delete(bidsByFire, fireKey)
					delete(timers, fireKey)
					mu.Unlock()

					if len(bids) == 0 {
						return
					}

					evaluateAndAnnounceDecentralized(t, truck, truckID, bids, sharedClock)
				})
			}
			mu.Unlock()
		} else {
			log.Printf("Truck %s: Low water (%d), requesting refill via RA", truckID, truck.GetWater())
			truck.RequestWaterRA()
		}

		return nil
	})

	// Collect bids from other trucks
	t.Subscribe(transport.ChannelFireBids, func(msg message.Message) error {
		// Update Lamport clock on message receive
		sharedClock.Receive(msg.Lamport)

		// Handle bid
		fireX := int(msg.Payload["fire_x"].(float64))
		fireY := int(msg.Payload["fire_y"].(float64))
		fireKey := fmt.Sprintf("%v,%v", fireX, fireY)

		mu.Lock()
		bidsByFire[fireKey] = append(bidsByFire[fireKey], msg)
		mu.Unlock()

		return nil
	})

	// Subscribe to bid decisions
	t.Subscribe(transport.ChannelFireDecision, func(msg message.Message) error {
		// Update Lamport clock on message receive
		sharedClock.Receive(msg.Lamport)

		winner := msg.Payload["winner"].(string)
		fireX := int(msg.Payload["fire_x"].(float64))
		fireY := int(msg.Payload["fire_y"].(float64))

		if winner == truckID {
			log.Printf("Truck %s: Assigned to fire at (%d,%d)", truckID, fireX, fireY)

			fire := &simulation.FireLocation{Row: fireX, Col: fireY}

			// Set assignment
			assignedMu.Lock()
			currentAssignment = fire
			assignedMu.Unlock()

			// Process assignment in goroutine
			go handleFireAssignment(t, truck, grid, fire, &assignedMu, &currentAssignment, sharedClock)
		} else {
			log.Printf("Truck %s: Assignment denied, winner is %s", truckID, winner)
		}

		return nil
	})

	// Subscribe to extinguish events to update local grid
	t.Subscribe(transport.ChannelCoordination, func(msg message.Message) error {
		// Update Lamport clock on message receive
		sharedClock.Receive(msg.Lamport)

		if action, ok := msg.Payload["action"].(string); ok && action == "extinguished" {
			row := int(msg.Payload["target_row"].(float64))
			col := int(msg.Payload["target_col"].(float64))
			grid.SetCell(row, col, simulation.Cell{State: simulation.Extinguished})
		}
		return nil
	})

	// Any truck may announce fires periodically
	go func() {
		randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))
		ticker := time.NewTicker(12 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// Check if we should generate a new fire
			fireMu.Lock()
			silent := time.Since(lastFireSeen) > 20*time.Second
			fireMu.Unlock()

			activeFires := len(grid.FindAllFires())

			// Generate fires when few fires exist OR long silence
			shouldGenerate := (activeFires < 2 && randSrc.Float32() < 0.4) ||
				(silent && randSrc.Float32() < 0.5)

			if shouldGenerate && activeFires < 5 {
				row := randSrc.Intn(simulation.GridSize)
				col := randSrc.Intn(simulation.GridSize)

				// Check if cell already has fire
				if grid.GetCell(row, col).State == simulation.Fire {
					continue
				}

				// fire intensity increases exponentially
				intensity := 2 + randSrc.Intn(3) // intensity 2-4
				msg := message.NewMessage(
					message.TypeFireAlert,
					truckID,
					message.FireAlertPayload(row, col, intensity),
				)
				msg.Lamport = sharedClock.Tick()
				t.Publish(transport.ChannelFireAlerts, msg)
				log.Printf("Truck %s: Generated fire at (%d,%d), intensity %d", truckID, row, col, intensity)
				fireMu.Lock()
				lastFireSeen = time.Now()
				fireMu.Unlock()
			}
		}
	}()

	// Periodic status broadcast to ensure trucks are always visible
	go func() {
		statusTicker := time.NewTicker(2 * time.Second)
		defer statusTicker.Stop()
		for range statusTicker.C {
			truck.BroadcastStatus()
		}
	}()
	// Keep running
	select {}
}

// Processes collected bids and announces winner
func evaluateAndAnnounceDecentralized(t *transport.NATSTransport, truck *simulation.Firetruck, truckID string, bids []message.Message, clock *clock.LamportClock) {
	if len(bids) == 0 {
		return
	}

	// Extract fire location
	fireX := int(bids[0].Payload["fire_x"].(float64))
	fireY := int(bids[0].Payload["fire_y"].(float64))

	log.Printf("Truck %s: Evaluating %d bids for fire=(%d,%d)", truckID, len(bids), fireX, fireY)

	// Convert to typed bids for sorting
	var typedBids []message.Bid
	for _, b := range bids {
		typedBids = append(typedBids, message.Bid{
			Fire:    message.FireID{X: fireX, Y: fireY},
			Bidder:  b.Payload["bidder"].(string),
			Score:   int(b.Payload["score"].(float64)),
			Lamport: int(b.Payload["lamport"].(float64)),
		})
	}

	// Sort bids with proper tie-breaking: Score ASC, Lamport ASC, Bidder ASC
	sort.Slice(typedBids, func(i, j int) bool {
		a, b := typedBids[i], typedBids[j]
		if a.Score != b.Score {
			return a.Score < b.Score
		}
		if a.Lamport != b.Lamport {
			return a.Lamport < b.Lamport
		}
		return a.Bidder < b.Bidder
	})

	winner := typedBids[0].Bidder
	reason := "lowest score"
	log.Printf("Truck %s: Winner=%s (%s)", truckID, winner, reason)

	// Find lowest truck ID among all bidders
	announcer := winner
	for _, b := range typedBids {
		if b.Bidder < announcer {
			announcer = b.Bidder
		}
	}

	// Only the lowest truck ID announces to prevent duplicates
	if truckID == announcer {
		decision := message.Message{
			Type:    message.TypeBidDecision,
			From:    truckID,
			Lamport: clock.Tick(),
			Payload: map[string]interface{}{
				"fire_x":  fireX,
				"fire_y":  fireY,
				"winner":  winner,
				"lamport": clock.Now(),
			},
		}
		t.Publish(transport.ChannelFireDecision, decision)
		log.Printf("Truck %s: DECISION fire=(%d,%d) winner=%s by (score,ts,id)", truckID, fireX, fireY, winner)
	} else {
		log.Printf("Truck %s: Assignment deferred, announcer is %s", truckID, announcer)
	}
}

// Moves truck to fire and extinguishes it
func handleFireAssignment(t *transport.NATSTransport, truck *simulation.Firetruck,
	grid *simulation.Grid, fire *simulation.FireLocation, assignedMu *sync.Mutex, currentAssignment **simulation.FireLocation, clock *clock.LamportClock) {

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		row, col := truck.GetPosition()

		// Check if at fire location
		if row == fire.Row && col == fire.Col {
			cell := grid.GetCell(row, col)
			if cell.State == simulation.Fire && truck.GetWater() > 0 {
				used := grid.Extinguish(row, col, truck.GetWater())
				truck.Water -= used

				log.Printf("\nFIRE EXTINGUISHED SUCCESSFULLY")
				log.Printf("   Location: (%d,%d)", row, col)
				log.Printf("   Water used: %d units", used)
				log.Printf("   Water remaining: %d/%d units", truck.GetWater(), truck.MaxWater)
				log.Printf("   Extinguished by: Truck %s", truck.ID)
				log.Printf("   Lamport timestamp: %d", clock.Tick())
				log.Printf("   Broadcasting extinguish event to all trucks...")

				// Broadcast extinguish event
				msg := message.NewMessage(
					message.TypeCoordination,
					truck.ID,
					message.CoordinationPayload("extinguished", row, col, map[string]interface{}{
						"water_used": used,
					}),
				)
				msg.Lamport = clock.Tick()
				t.Publish(transport.ChannelCoordination, msg)
				truck.BroadcastStatus()
			} else if cell.State != simulation.Fire {
				log.Printf("[%s] DECENTRALIZED: Fire at (%d,%d) already extinguished", truck.ID, row, col)
			}

			// Clear assignment
			assignedMu.Lock()
			*currentAssignment = nil
			assignedMu.Unlock()

			// Request water refill if below threshold
			if truck.GetWater() <= truck.GetLowWaterThresh() {
				truck.RequestWaterRA()
			}
			return
		}

		// Move toward fire
		oldRow, oldCol := truck.GetPosition()
		truck.MoveToward(fire.Row, fire.Col)
		newRow, newCol := truck.GetPosition()

		// Broadcast position update if moved
		if oldRow != newRow || oldCol != newCol {
			log.Printf("\nTRUCK MOVEMENT")
			log.Printf("   From: (%d,%d) -> To: (%d,%d)", oldRow, oldCol, newRow, newCol)
			log.Printf("   Target fire: (%d,%d)", fire.Row, fire.Col)
			log.Printf("   Truck: %s", truck.ID)
			log.Printf("   Water level: %d/%d units", truck.GetWater(), truck.MaxWater)
			log.Printf("   Lamport timestamp: %d", clock.Now())
			truck.BroadcastStatus()
		}
	}
}

// Monitors and visualizes the system state
func runObserver(t *transport.NATSTransport, observerID string) {
	grid := simulation.NewGrid()
	trucks := make(map[string]*simulation.Firetruck)

	log.Printf("\n==================================================================================")
	log.Printf("OBSERVER - SYSTEM MONITOR")
	log.Printf("==================================================================================")
	log.Printf("Role: Passive observer of distributed system")
	log.Printf("Monitoring: Fire generation, truck movement, extinguishing")
	log.Printf("Tracking: Water supply and refill operations")
	log.Printf("Visualizing: Decentralized coordination in real-time")
	log.Printf("Lamport clocks: Synchronized across all processes")
	log.Printf("==================================================================================\n")

	// Subscribe to all events
	t.Subscribe(transport.ChannelFireAlerts, func(msg message.Message) error {
		// Handle both old and new fire alert formats
		var row, col, intensity int
		if r, ok := msg.Payload["row"].(float64); ok {
			// Old format
			row = int(r)
			col = int(msg.Payload["col"].(float64))
			intensity = int(msg.Payload["intensity"].(float64))
		} else {
			// New FireAnnounce format
			row = int(msg.Payload["id_x"].(float64))
			col = int(msg.Payload["id_y"].(float64))
			intensity = int(msg.Payload["intensity"].(float64))
		}

		grid.SetCell(row, col, simulation.Cell{
			State:     simulation.Fire,
			Intensity: intensity,
		})

		fmt.Printf("\nNEW FIRE DETECTED: (%d,%d) | Intensity: %d | Lamport: %d\n", row, col, intensity, msg.Lamport)
		return nil
	})

	t.Subscribe(transport.ChannelTruckStatus, func(msg message.Message) error {
		truckID := msg.From
		row := int(msg.Payload["row"].(float64))
		col := int(msg.Payload["col"].(float64))
		water := int(msg.Payload["water"].(float64))
		maxWater := int(msg.Payload["max_water"].(float64))

		if trucks[truckID] == nil {
			trucks[truckID] = simulation.NewFiretruck(truckID, row, col)
		}
		trucks[truckID].Row = row
		trucks[truckID].Col = col
		trucks[truckID].Water = water
		trucks[truckID].MaxWater = maxWater

		return nil
	})

	t.Subscribe(transport.ChannelCoordination, func(msg message.Message) error {
		if action, ok := msg.Payload["action"].(string); ok && action == "extinguished" {
			row := int(msg.Payload["target_row"].(float64))
			col := int(msg.Payload["target_col"].(float64))

			grid.SetCell(row, col, simulation.Cell{State: simulation.Extinguished})
			fmt.Printf("\nFIRE EXTINGUISHED: (%d,%d) | By: Truck %s | Lamport: %d\n", row, col, msg.From, msg.Lamport)
		}
		return nil
	})

	// Observer advances fire simulation and publishes alerts
	go func() {
		growthTicker := time.NewTicker(5 * time.Second)
		defer growthTicker.Stop()
		for range growthTicker.C {
			newFires := grid.StepFires()

			// Publish alerts for newly spread fires
			for _, fire := range newFires {
				cell := grid.GetCell(fire.Row, fire.Col)
				alert := message.Message{
					Type:    message.TypeFireAnnounce,
					From:    observerID,
					Lamport: 1, // Observer timestamp
					Payload: map[string]interface{}{
						"id_x":      float64(fire.Row),
						"id_y":      float64(fire.Col),
						"intensity": float64(cell.Intensity),
						"tick":      float64(0),
					},
				}
				t.Publish(transport.ChannelFireAlerts, alert)
			}
		}
	}()

	// Periodic status display
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("\n" + "═══════════════════════════════════════════════════")
		printSystemState(grid, trucks)
	}
}

// Displays current grid and truck status
func printSystemState(grid *simulation.Grid, trucks map[string]*simulation.Firetruck) {
	fmt.Println("GRID STATE:")

	// Create truck position map
	truckPos := make(map[[2]int]string)
	for id, t := range trucks {
		truckPos[[2]int{t.Row, t.Col}] = id
	}

	// Print grid
	for r := 0; r < simulation.GridSize; r++ {
		for c := 0; c < simulation.GridSize; c++ {
			if tid, ok := truckPos[[2]int{r, c}]; ok {
				fmt.Printf("%3s", tid)
			} else {
				cell := grid.GetCell(r, c)
				switch cell.State {
				case simulation.Empty:
					fmt.Print("  .")
				case simulation.Fire:
					fmt.Print("  F")
				case simulation.Extinguished:
					fmt.Print("  E")
				}
			}
		}
		fmt.Println()
	}

	// Print truck and water supply info
	fmt.Println("\nTRUCK STATUS:")
	for id, t := range trucks {
		fmt.Printf("  %s: position=(%d,%d) water=%d/%d\n",
			id, t.Row, t.Col, t.Water, t.MaxWater)
	}

	// Fire count
	fires := grid.FindAllFires()
	fmt.Printf("\nActive fires: %d\n", len(fires))
}
