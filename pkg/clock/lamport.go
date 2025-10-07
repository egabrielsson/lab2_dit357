package clock

import "sync/atomic"

// LamportClock implements Lamport's logical clock for causal ordering.
type LamportClock struct {
	t int64
}

// NewLamportClock creates a new Lamport clock starting from 0.
func NewLamportClock() *LamportClock {
	return &LamportClock{t: 0}
}

// Now returns the current clock value without incrementing it.
func (lc *LamportClock) Now() int64 {
	return atomic.LoadInt64(&lc.t)
}

// Tick increments the clock and returns the new value.
// Use this when a local event occurs.
func (lc *LamportClock) Tick() int64 {
	return atomic.AddInt64(&lc.t, 1)
}

// Receive updates the clock when receiving a message from another process.
// It sets the clock to max(local_time, received_time) + 1.
func (lc *LamportClock) Receive(other int64) int64 {
	for {
		cur := atomic.LoadInt64(&lc.t)
		next := cur
		if other > cur {
			next = other
		}
		next++
		if atomic.CompareAndSwapInt64(&lc.t, cur, next) {
			return next
		}
	}
}