// trucknode.go
package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

func mainTruckNode() {
	id := flag.String("id", "T1", "truck id")
	addr := flag.String("addr", "127.0.0.1:9001", "listen addr")
	peersFlag := flag.String("peers", "", "comma list of host:port@ID for peers")
	request := flag.Bool("request", false, "send a water_request to first peer")
	amount := flag.Int("amount", 20, "amount to request/offer")
	autoOffer := flag.Bool("auto-offer", false, "auto offer on request")
	flag.Parse()

	peers := map[string]string{}
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

	node := NewTCPNode(*id, *addr, peers)
	node.Listen(func(msg Message, conn net.Conn) {
		fmt.Printf("[%s] recv %s from %s (lt=%d): %+v\n", node.ID, msg.Type, msg.From, msg.Lamport, msg.Payload)
		if *autoOffer && msg.Type == TypeWaterRequest {
			reply := Message{
				Type:    TypeWaterOffer,
				Payload: map[string]any{"amount": *amount / 2},
			}
			// reply on same conn
			reply.Lamport = node.Clock.Tick()
			reply.From = node.ID
			wireWrite(conn, reply)
			fmt.Printf("[%s] sent water_offer back (lt=%d)\n", node.ID, reply.Lamport)
		}
	})

	// optionally initiate a request to the first peer we know
	if *request {
		for pid := range peers {
			msg := Message{
				Type:    TypeWaterRequest,
				Payload: map[string]any{"amount": *amount},
			}
			if err := node.Send(pid, msg); err != nil {
				fmt.Println("send error:", err)
			} else {
				fmt.Printf("[%s] sent water_request to %s\n", node.ID, pid)
			}
			break
		}
	}

	select {} // keep running
}

// We keep the original 'main' in main.go for the sim. If you want this file to be the entry,
// run it directly: `go run trucknode.go --id T1 --addr ...`
