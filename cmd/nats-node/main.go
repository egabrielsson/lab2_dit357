package main

import (
	"flag"
	"fmt"
	"time"

	"Firetruck-sim/pkg/message"
	"Firetruck-sim/pkg/simulation"
	"Firetruck-sim/pkg/transport"
)

func main() {
	id := flag.String("id", "T1", "truck id")
	natsURL := flag.String("nats-url", "nats://127.0.0.1:4222", "NATS server URL")
	// amount := flag.Int("amount", 20, "amount to request/offer")
	// autoOffer := flag.Bool("auto-offer", false, "auto offer on incoming water_request")
	pubsub := flag.Bool("pubsub", false, "enable pub-sub testing")
	fireAlert := flag.Bool("fire-alert", false, "send a fire alert broadcast")
	status := flag.Bool("status", false, "send a status broadcast")
	bidding := flag.Bool("bidding", false, "enable fire bidding coordination")
	coordinator := flag.Bool("coordinator", false, "act as bid coordinator (evaluates all bids)")
	flag.Parse()

	natsTransport, err := transport.NewNATSTransport(*id, *natsURL)
	if err != nil {
		panic(err)
	}
	defer natsTransport.Close()

	// Enable pub-sub testing if requested
	if *pubsub {
		setupPubSubTesting(natsTransport)
	}

	// Enable bidding coordination if requested
	if *bidding {
		setupBiddingCoordination(natsTransport, *id, false)
	}

	// Enable coordinator mode if requested
	if *coordinator {
		setupBiddingCoordination(natsTransport, *id, true)
	}

	// Send fire alert broadcast
	if *fireAlert {
		msg := message.NewMessage(
			message.TypeFireAlert,
			natsTransport.GetID(),
			message.FireAlertPayload(5, 10, 3),
		)
		err := natsTransport.Publish(transport.ChannelFireAlerts, msg)
		if err != nil {
			fmt.Println("fire alert error:", err)
		} else {
			fmt.Printf("[%s] broadcast fire alert to channel %s\n", natsTransport.GetID(), transport.ChannelFireAlerts)
		}
	}

	// Send status broadcast
	if *status {
		msg := message.NewMessage(
			message.TypeTruckStatus,
			natsTransport.GetID(),
			message.TruckStatusPayload(2, 3, 45, 100, "moving_to_fire"),
		)
		err := natsTransport.Publish(transport.ChannelTruckStatus, msg)
		if err != nil {
			fmt.Println("status broadcast error:", err)
		} else {
			fmt.Printf("[%s] broadcast status to channel %s\n", natsTransport.GetID(), transport.ChannelTruckStatus)
		}
	}

	fmt.Printf("[%s] NATS node connected to %s\n", *id, *natsURL)
	select {} // keep process alive to handle incoming messages
}

func setupPubSubTesting(natsTransport *transport.NATSTransport) {
	// Subscribe to fire alerts
	err := natsTransport.Subscribe(transport.ChannelFireAlerts, func(msg message.Message) error {
		fmt.Printf("[%s] FIRE ALERT from %s: %+v\n", natsTransport.GetID(), msg.From, msg.Payload)
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to fire alerts: %v\n", err)
	}

	// Subscribe to truck status updates
	err = natsTransport.Subscribe(transport.ChannelTruckStatus, func(msg message.Message) error {
		fmt.Printf("[%s] TRUCK STATUS from %s: %+v\n", natsTransport.GetID(), msg.From, msg.Payload)
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to truck status: %v\n", err)
	}

	// Subscribe to coordination messages
	err = natsTransport.Subscribe(transport.ChannelCoordination, func(msg message.Message) error {
		fmt.Printf("[%s] COORDINATION from %s: %+v\n", natsTransport.GetID(), msg.From, msg.Payload)
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to coordination: %v\n", err)
	}

	fmt.Printf("[%s] Subscribed to all broadcast channels\n", natsTransport.GetID())
}

func setupBiddingCoordination(natsTransport *transport.NATSTransport, truckID string, isCoordinator bool) {
	bids := make(map[string][]message.Message)
	bidTimeout := 1500 * time.Millisecond
	truckRow, truckCol := 0, 0
	truckWater := 50

	// Assign different starting positions based on ID
	if truckID == "T1" {
		truckRow, truckCol = 0, 0
	} else if truckID == "T2" {
		truckRow, truckCol = 19, 19
	} else {
		truckRow, truckCol = 10, 10
	}

	// Trucks (not coordinator) subscribe to fire alerts and auto-bid
	if !isCoordinator {
		err := natsTransport.Subscribe(transport.ChannelFireAlerts, func(msg message.Message) error {
			fireRow := int(msg.Payload["row"].(float64))
			fireCol := int(msg.Payload["col"].(float64))
			intensity := int(msg.Payload["intensity"].(float64))

			fmt.Printf("[%s] Fire detected at (%d,%d) intensity=%d\n", truckID, fireRow, fireCol, intensity)

			// Calculate distance and submit bid (Manhattan distance)
			distance := simulation.Abs(truckRow-fireRow) + simulation.Abs(truckCol-fireCol)

			bidMsg := message.NewMessage(
				message.TypeFireBid,
				truckID,
				message.FireBidPayload(fireRow, fireCol, distance, truckWater, truckID),
			)

			fmt.Printf("[%s] Submitting bid: distance=%d water=%d\n", truckID, distance, truckWater)
			natsTransport.Publish(transport.ChannelFireBids, bidMsg)

			return nil
		})
		if err != nil {
			fmt.Printf("Failed to subscribe to fire alerts: %v\n", err)
		}
	}

	// Coordinator or trucks subscribe to fire bids
	err := natsTransport.Subscribe(transport.ChannelFireBids, func(msg message.Message) error {
		fireKey := fmt.Sprintf("%v,%v", msg.Payload["fire_row"], msg.Payload["fire_col"])

		// Store the bid
		if bids[fireKey] == nil {
			bids[fireKey] = []message.Message{}

			// Start timer - only coordinator or lowest ID evaluates
			go func(key string) {
				time.Sleep(bidTimeout)

				if isCoordinator {
					// Coordinator always evaluates
					evaluateBidsForFire(natsTransport, truckID, key, bids[key])
				} else {
					// Check if this truck should evaluate (lowest ID among bidders)
					lowestID := truckID
					for _, bid := range bids[key] {
						if bid.From < lowestID {
							lowestID = bid.From
						}
					}

					// Only evaluate if this truck has the lowest ID
					if lowestID == truckID {
						evaluateBidsForFire(natsTransport, truckID, key, bids[key])
					}
				}
				delete(bids, key)
			}(fireKey)
		}
		bids[fireKey] = append(bids[fireKey], msg)
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to fire bids: %v\n", err)
	}

	// Subscribe to fire assignments (only if not coordinator)
	if !isCoordinator {
		err = natsTransport.Subscribe(transport.ChannelFireAssignment, func(msg message.Message) error {
			assignedTruck := msg.Payload["assigned_truck"].(string)
			fireRow := msg.Payload["fire_row"]
			fireCol := msg.Payload["fire_col"]
			reason := msg.Payload["reason"].(string)

			if assignedTruck == truckID {
				fmt.Printf("\n[%s] ✓ ACCEPTED for fire at (%v,%v): %s (Lamport time: %d)\n\n",
					truckID, fireRow, fireCol, reason, msg.Lamport)
			} else {
				fmt.Printf("\n[%s] ✗ DENIED for fire at (%v,%v): %s won (%s) (Lamport time: %d)\n\n",
					truckID, fireRow, fireCol, assignedTruck, reason, msg.Lamport)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Failed to subscribe to fire assignments: %v\n", err)
		}
	}

	if isCoordinator {
		fmt.Printf("[%s] Coordinator mode enabled\n", truckID)
	} else {
		fmt.Printf("[%s] Bidding coordination enabled\n", truckID)
	}
}

func evaluateBidsForFire(natsTransport *transport.NATSTransport, evaluatorID string, _ string, bidMsgs []message.Message) {
	if len(bidMsgs) == 0 {
		return
	}

	// Convert messages to FireBid structs
	var bids []struct {
		TruckID  string
		Distance int
		Water    int
		Lamport  int64
	}

	for _, msg := range bidMsgs {
		bid := struct {
			TruckID  string
			Distance int
			Water    int
			Lamport  int64
		}{
			TruckID:  msg.From,
			Distance: int(msg.Payload["distance"].(float64)),
			Water:    int(msg.Payload["water"].(float64)),
			Lamport:  msg.Lamport,
		}
		bids = append(bids, bid)
	}

	// Evaluate using the same logic as EvaluateFireBids
	fireBids := make([]simulation.FireBid, len(bids))
	for i, b := range bids {
		fireBids[i] = simulation.FireBid{
			TruckID:  b.TruckID,
			Distance: b.Distance,
			Water:    b.Water,
			Lamport:  b.Lamport,
		}
	}

	winner, reason := simulation.EvaluateFireBids(fireBids)

	// Broadcast the assignment
	fireRow := bidMsgs[0].Payload["fire_row"]
	fireCol := bidMsgs[0].Payload["fire_col"]

	assignment := message.NewMessage(
		message.TypeFireAssignment,
		evaluatorID,
		message.FireAssignmentPayload(int(fireRow.(float64)), int(fireCol.(float64)), winner, reason),
	)

	natsTransport.Publish(transport.ChannelFireAssignment, assignment)
}
