package simulation

import (
	"fmt"
	"sync"
)

// WaterTank represents the shared water supply for all firetrucks
type WaterTank struct {
	mu    sync.Mutex
	stock int
}

// NewWaterTank creates a new water tank with the given initial stock
func NewWaterTank(stock int) *WaterTank {
	return &WaterTank{stock: stock}
}

// Withdraw attempts to withdraw n units of water from the tank.
// Returns the actual amount withdrawn (may be less than requested).
func (w *WaterTank) Withdraw(n int, truckID string, lamportTime int64) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	if n <= 0 {
		return 0
	}
	requested := n
	if n > w.stock {
		n = w.stock
	}
	w.stock -= n
	
	// Log accepted or denied requests 
	if n == requested && n > 0 {
		fmt.Printf("\n Accepted water request: %d from %s (Lamport time %d)\n", n, truckID, lamportTime)
	} else if n < requested {
		fmt.Printf("-- Denied/Partial water request: got %d of %d from %s (insufficient supply)\n", n, requested, truckID)
	}
	
	return n
}

// Deposit adds n units of water to the tank
func (w *WaterTank) Deposit(n int) {
	if n <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stock += n
}

// GetStock returns the current water stock (for monitoring)
func (w *WaterTank) GetStock() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stock
}