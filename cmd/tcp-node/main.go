package main

import (
	"flag"
	"fmt"
	"strings"

	"Firetruck-sim/pkg/message"
	"Firetruck-sim/pkg/transport"
)

func main() {
	id := flag.String("id", "T1", "truck id")
	addr := flag.String("addr", "127.0.0.1:9001", "listen addr")
	peersFlag := flag.String("peers", "", "comma list of host:port@ID for peers")
	request := flag.Bool("request", false, "send a water_request to first peer")
	amount := flag.Int("amount", 20, "amount to request/offer")
	autoOffer := flag.Bool("auto-offer", false, "auto offer on request")
	flag.Parse()

	peers := make(map[string]string)
	if *peersFlag != "" {
		for _, part := range strings.Split(*peersFlag, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			host, idp, _ := strings.Cut(part, "@")
			peers[idp] = host
		}
	}

	tcpTransport, err := transport.NewTCPTransport(*id, *addr, peers)
	if err != nil {
		panic(err)
	}
	defer tcpTransport.Close()

	// Start listening for messages
	err = tcpTransport.Listen(func(msg message.Message) (*message.Message, error) {
		fmt.Printf("[%s] recv %s from %s (lt=%d): %+v\n", 
			tcpTransport.GetID(), msg.Type, msg.From, msg.Lamport, msg.Payload)
		
		if *autoOffer && msg.Type == message.TypeWaterRequest {
			reply := message.NewMessage(
				message.TypeWaterOffer,
				tcpTransport.GetID(),
				message.WaterOfferPayload(*amount/2),
			)
			fmt.Printf("[%s] auto-reply water_offer: %+v\n", tcpTransport.GetID(), reply.Payload)
			return &reply, nil
		}
		return nil, nil
	})
	
	if err != nil {
		panic(err)
	}

	// Optionally initiate a request to the first peer
	if *request {
		for peerID := range peers {
			msg := message.NewMessage(
				message.TypeWaterRequest,
				tcpTransport.GetID(),
				message.WaterRequestPayload(*amount),
			)
			fmt.Printf("[%s] sending water_request to %s...\n", tcpTransport.GetID(), peerID)
			
			if err := tcpTransport.Send(peerID, msg); err != nil {
				fmt.Println("send error:", err)
			} else {
				fmt.Printf("[%s] sent water_request to %s\n", tcpTransport.GetID(), peerID)
			}
			break
		}
	}

	fmt.Printf("[%s] TCP node listening on %s\n", *id, *addr)
	select {} // keep running
}