package simulation

import "sync"

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
func (w *WaterTank) Withdraw(n int) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	if n <= 0 {
		return 0
	}
	if n > w.stock {
		n = w.stock
	}
	w.stock -= n
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