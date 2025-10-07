// waterlock.go
package main

import "sync"

type WaterTank struct {
	mu    sync.Mutex
	stock int
}

func NewWaterTank(stock int) *WaterTank {
	return &WaterTank{stock: stock}
}

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

func (w *WaterTank) Deposit(n int) {
	if n <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stock += n
}
