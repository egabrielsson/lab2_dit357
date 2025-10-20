package message

// Message types for inter-truck communication
const (
	TypeMoveCommand    = "move_command"
	TypeFireAlert      = "fire_alert"
	TypeCoordination   = "coordination"
	TypeTruckStatus    = "truck_status"
	TypeWaterBroadcast = "water_broadcast"
	TypeFireBid        = "fire_bid"
	TypeFireAssignment = "fire_assignment"
	TypeWaterRequest   = "water_request"
	TypeWaterResponse  = "water_response"
	TypeFireAnnounce   = "fire_announce"
	TypeBid            = "bid"
	TypeBidDecision    = "bid_decision"
	TypeTick           = "tick"
	TypeWaterReq       = "water_req"
	TypeWaterReply     = "water_reply"
	TypeWaterRelease   = "water_release"
)

// Represents a communication message between fire trucks
type Message struct {
	Type    string                 `json:"type"`
	From    string                 `json:"from"`
	Lamport int64                  `json:"lamport"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Creates a new message with the specified type and payload
func NewMessage(msgType, from string, payload map[string]interface{}) Message {
	return Message{
		Type:    msgType,
		From:    from,
		Payload: payload,
	}
}

// Creates a payload for fire alert messages
func FireAlertPayload(row, col, intensity int) map[string]interface{} {
	return map[string]interface{}{
		"row":       row,
		"col":       col,
		"intensity": intensity,
	}
}

// Creates a payload for truck status broadcast messages
func TruckStatusPayload(row, col, water, maxWater int, task string) map[string]interface{} {
	return map[string]interface{}{
		"row":       row,
		"col":       col,
		"water":     water,
		"max_water": maxWater,
		"task":      task,
	}
}

// Creates a payload for coordination messages
func CoordinationPayload(action string, targetRow, targetCol int, details map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"action":     action,
		"target_row": targetRow,
		"target_col": targetCol,
	}
	for k, v := range details {
		payload[k] = v
	}
	return payload
}

// Creates a payload for fire bidding messages
func FireBidPayload(fireRow, fireCol, distance, water int, truckID string) map[string]interface{} {
	return map[string]interface{}{
		"fire_row": fireRow,
		"fire_col": fireCol,
		"distance": distance,
		"water":    water,
		"truck_id": truckID,
	}
}

// Creates a payload for fire assignment result messages
func FireAssignmentPayload(fireRow, fireCol int, assignedTruck string, reason string) map[string]interface{} {
	return map[string]interface{}{
		"fire_row":       fireRow,
		"fire_col":       fireCol,
		"assigned_truck": assignedTruck,
		"reason":         reason,
	}
}

// Typed payloads
type FireID struct{ X, Y int }

type FireAnnounce struct {
	ID        FireID
	Intensity int
	Tick      uint64
}

type Bid struct {
	Fire    FireID
	Bidder  string
	Score   int
	Lamport int
}

type BidDecision struct {
	Fire    FireID
	Winner  string
	Lamport int
}

type Tick struct {
	Tick uint64
	Seed int64
}

// RA messages
type WaterReq struct {
	From string
	TS   int
}
type WaterReply struct{ From string }
type WaterRelease struct{ From string }
