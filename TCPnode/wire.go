package main

import (
	"bufio"
	"encoding/json"
	"net"
)

const (
	TypeWaterRequest = "water_request"
	TypeWaterOffer   = "water_offer"
)

type Message struct {
	Type    string         `json:"type"`
	From    string         `json:"from"`
	Lamport int64          `json:"lamport"`
	Payload map[string]any `json:"payload,omitempty"`
}

func wireWrite(conn net.Conn, m Message) error {
	data, _ := json.Marshal(m)
	data = append(data, '\n')
	_, err := conn.Write(data)
	return err
}

func wireRead(conn net.Conn, m *Message) error {
	sc := bufio.NewScanner(conn)
	if sc.Scan() {
		return json.Unmarshal(sc.Bytes(), m)
	}
	return sc.Err()
}
