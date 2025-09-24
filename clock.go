package main

import (
	"sync"
)

// Boiler plate Lamport clock
// Standard implementation of a Classic Lamport Clock
// used to send and receive event.
//With this we can make the firetruck agent use time.

// LamportClock struct
type LamportClock struct {
	mu   sync.Mutex // Mutex to ensure safe and concurrent access
	time int        // Current time
}

// Constructor function for new lamport clock
func NewLamportClock() *LamportClock {
	return &LamportClock{time: 0} // Starts at time 0
}

// Tick represents an internal event or sending a message
func (lc *LamportClock) Tick() int {
	lc.mu.Lock()         // Lock to prevent race conditions
	defer lc.mu.Unlock() // Unlock when function exits
	lc.time++            // Increment logical value
	return lc.time       // Return updated time
}

// Receive event
func (lc *LamportClock) Receive(received int) int {
	lc.mu.Lock()         // Lock for safe concurrent access
	defer lc.mu.Unlock() // Unlock after update
	// Set the local time to max(local, received) + 1
	if received > lc.time {
		lc.time = received
	}
	lc.time++      // Always increment after receive
	return lc.time // Return the updated time
}

// Time simply returns the current logical time
func (lc *LamportClock) Time() int {
	lc.mu.Lock()         // Lock so that it is thread-safe to read
	defer lc.mu.Unlock() // Unlock immediately after
	return lc.time       // Return current logical time
}
