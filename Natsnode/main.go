package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	id := flag.String("id", "T1", "truck id")
	natsURL := flag.String("nats-url", "nats://127.0.0.1:4222", "NATS server URL")
	peer := flag.String("to", "", "target truck id to address (for request)")
	request := flag.Bool("request", false, "send a water_request to --to")
	amount := flag.Int("amount", 20, "amount to request/offer")
	autoOffer := flag.Bool("auto-offer", false, "auto offer on incoming water_request")
	timeout := flag.Duration("timeout", 3*time.Second, "request timeout")
	flag.Parse()

	node, err := NewNATSNode(*id, *natsURL)
	if err != nil {
		panic(err)
	}
	defer node.Close()

	// Start listener that optionally auto-replies to water_request
	if err := node.Listen(func(msg Message) (*Message, error) {
		fmt.Printf("[%s] recv %s from %s (lt=%d): %+v\n",
			node.ID, msg.Type, msg.From, msg.Lamport, msg.Payload)

		if *autoOffer && msg.Type == TypeWaterRequest {
			reply := &Message{
				Type:    TypeWaterOffer,
				Payload: map[string]any{"amount": *amount / 2},
			}
			fmt.Printf("[%s] auto-reply water_offer: %+v\n", node.ID, reply.Payload)
			return reply, nil
		}
		// No reply by default
		return nil, nil
	}); err != nil {
		panic(err)
	}

	// Optionally initiate a request to a peer
	if *request {
		if *peer == "" {
			panic("--request needs --to=<TruckID>")
		}
		msg := Message{
			Type:    TypeWaterRequest,
			Payload: map[string]any{"amount": *amount},
		}
		fmt.Printf("[%s] sending water_request to %s...\n", node.ID, *peer)
		resp, err := node.Request(*peer, msg, *timeout)
		if err != nil {
			fmt.Println("request error:", err)
		} else {
			fmt.Printf("[%s] got reply %s from %s (lt=%d): %+v\n",
				node.ID, resp.Type, resp.From, resp.Lamport, resp.Payload)
		}
	}

	select {} // keep process alive to handle incoming messages
}
