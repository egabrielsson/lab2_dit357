package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSNode struct {
	ID    string
	URL   string
	Clock *LamportClock
	nc    *nats.Conn
	sub   *nats.Subscription
}

// Subject for directed messages to a specific truck.
func subjectFor(id string) string { return fmt.Sprintf("trucks.%s", id) }

func NewNATSNode(id, natsURL string) (*NATSNode, error) {
	n := &NATSNode{
		ID:    id,
		URL:   natsURL,
		Clock: &LamportClock{},
	}
	nc, err := nats.Connect(natsURL,
		nats.Name("truck-"+id),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}
	n.nc = nc
	return n, nil
}

// Listen starts a subscription to this node's directed inbox subject.
// handler is invoked for every inbound Message; use msg.Respond to reply.
func (n *NATSNode) Listen(handler func(Message) (reply *Message, err error)) error {
	subj := subjectFor(n.ID)
	sub, err := n.nc.Subscribe(subj, func(m *nats.Msg) {
		var msg Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			fmt.Println("bad msg:", err)
			return
		}
		n.Clock.Receive(msg.Lamport)
		if reply, err := handler(msg); err == nil && reply != nil {
			reply.From = n.ID
			reply.Lamport = n.Clock.Tick()
			b, _ := json.Marshal(reply)
			_ = m.Respond(b)
		}
	})
	if err != nil {
		return err
	}
	n.sub = sub
	return n.nc.Flush()
}

// Request sends a directed request to another truck ID and waits for a reply.
func (n *NATSNode) Request(toID string, msg Message, timeout time.Duration) (Message, error) {
	msg.From = n.ID
	msg.Lamport = n.Clock.Tick()
	b, _ := json.Marshal(msg)
	inbox := subjectFor(toID)
	resp, err := n.nc.Request(inbox, b, timeout)
	if err != nil {
		return Message{}, err
	}
	var out Message
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return Message{}, err
	}
	n.Clock.Receive(out.Lamport)
	return out, nil
}

// PublishFireAndForget is here if you later want broadcast-ish one-way sends.
func (n *NATSNode) Publish(toID string, msg Message) error {
	msg.From = n.ID
	msg.Lamport = n.Clock.Tick()
	b, _ := json.Marshal(msg)
	return n.nc.Publish(subjectFor(toID), b)
}

func (n *NATSNode) Close() {
	if n.sub != nil {
		_ = n.sub.Unsubscribe()
	}
	if n.nc != nil {
		n.nc.Close()
	}
}
