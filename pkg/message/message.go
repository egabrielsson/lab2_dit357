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
)

// Message represents a communication message between fire trucks.
type Message struct {
	Type    string                 `json:"type"`
	From    string                 `json:"from"`
	Lamport int64                  `json:"lamport"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// NewMessage creates a new message with the specified type and payload.
func NewMessage(msgType, from string, payload map[string]interface{}) Message {
	return Message{
		Type:    msgType,
		From:    from,
		Payload: payload,
	}
}

// FireAlertPayload creates a payload for fire alert messages.
func FireAlertPayload(row, col, intensity int) map[string]interface{} {
	return map[string]interface{}{
		"row":       row,
		"col":       col,
		"intensity": intensity,
	}
}

// TruckStatusPayload creates a payload for truck status broadcast messages.
func TruckStatusPayload(row, col, water, maxWater int, task string) map[string]interface{} {
	return map[string]interface{}{
		"row":       row,
		"col":       col,
		"water":     water,
		"max_water": maxWater,
		"task":      task,
	}
}

// CoordinationPayload creates a payload for coordination messages.
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

// FireBidPayload creates a payload for fire bidding messages.
func FireBidPayload(fireRow, fireCol, distance, water int, truckID string) map[string]interface{} {
	return map[string]interface{}{
		"fire_row": fireRow,
		"fire_col": fireCol,
		"distance": distance,
		"water":    water,
		"truck_id": truckID,
	}
}

// FireAssignmentPayload creates a payload for fire assignment result messages.
func FireAssignmentPayload(fireRow, fireCol int, assignedTruck string, reason string) map[string]interface{} {
	return map[string]interface{}{
		"fire_row":       fireRow,
		"fire_col":       fireCol,
		"assigned_truck": assignedTruck,
		"reason":         reason,
	}
}
