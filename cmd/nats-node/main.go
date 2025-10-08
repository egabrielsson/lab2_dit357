package main

import (
	"flag"
	"fmt"
	"time"

	"Firetruck-sim/pkg/message"
	"Firetruck-sim/pkg/transport"
)

func main() {
	id := flag.String("id", "T1", "truck id")
	natsURL := flag.String("nats-url", "nats://127.0.0.1:4222", "NATS server URL")
	peer := flag.String("to", "", "target truck id to address (for request)")
	request := flag.Bool("request", false, "send a water_request to --to")
	amount := flag.Int("amount", 20, "amount to request/offer")
	autoOffer := flag.Bool("auto-offer", false, "auto offer on incoming water_request")
	timeout := flag.Duration("timeout", 3*time.Second, "request timeout")
	pubsub := flag.Bool("pubsub", false, "enable pub-sub testing")
	fireAlert := flag.Bool("fire-alert", false, "send a fire alert broadcast")
	status := flag.Bool("status", false, "send a status broadcast")
	flag.Parse()

	natsTransport, err := transport.NewNATSTransport(*id, *natsURL)
	if err != nil {
		panic(err)
	}
	defer natsTransport.Close()

	// Start listener that optionally auto-replies to water_request
	err = natsTransport.Listen(func(msg message.Message) (*message.Message, error) {
		fmt.Printf("[%s] recv %s from %s (lt=%d): %+v\n",
			natsTransport.GetID(), msg.Type, msg.From, msg.Lamport, msg.Payload)

		if *autoOffer && msg.Type == message.TypeWaterRequest {
			reply := message.NewMessage(
				message.TypeWaterOffer,
				natsTransport.GetID(),
				message.WaterOfferPayload(*amount/2),
			)
			fmt.Printf("[%s] auto-reply water_offer: %+v\n", natsTransport.GetID(), reply.Payload)
			return &reply, nil
		}
		// No reply by default
		return nil, nil
	})
	
	if err != nil {
		panic(err)
	}

	// Enable pub-sub testing if requested
	if *pubsub {
		setupPubSubTesting(natsTransport)
	}

	// Optionally initiate a request to a peer
	if *request {
		if *peer == "" {
			panic("--request needs --to=<TruckID>")
		}
		msg := message.NewMessage(
			message.TypeWaterRequest,
			natsTransport.GetID(),
			message.WaterRequestPayload(*amount),
		)
		fmt.Printf("[%s] sending water_request to %s...\n", natsTransport.GetID(), *peer)
		resp, err := natsTransport.Request(*peer, msg, *timeout)
		if err != nil {
			fmt.Println("request error:", err)
		} else {
			fmt.Printf("[%s] got reply %s from %s (lt=%d): %+v\n",
				natsTransport.GetID(), resp.Type, resp.From, resp.Lamport, resp.Payload)
		}
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