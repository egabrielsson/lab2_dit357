package main

import "sync/atomic"

// Lamport clock for causal ordering.
type LamportClock struct {
	t int64
}

func (lc *LamportClock) Now() int64 {
	return atomic.LoadInt64(&lc.t)
}

func (lc *LamportClock) Tick() int64 {
	return atomic.AddInt64(&lc.t, 1)
}

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
