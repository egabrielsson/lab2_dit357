# Distributed Fire Truck System

**Decentralized coordination** with Lamport timestamps and mutual exclusion.

## Quick Start

```bash
# Build and run
go build -o distributed cmd/distributed/main.go
./start.sh

# View simulation
./logs.sh observer    # Grid and trucks
./logs.sh truck-t1    # Truck T1 activity
./logs.sh truck-t2    # Truck T2 activity
```

## Assignment Compliance

### Task 2: Lamport Timestamps & Mutual Exclusion
- ✅ **Lamport timestamps** on all truck actions (move, bid, extinguish, refill)
- ✅ **Mutual exclusion** for water supply using Lamport timestamps
- ✅ **Fairness** demonstrated - requests processed by timestamp order

### Task 3: Decentralized Strategy & Naming
- ✅ **Flat naming scheme** - trucks identified as T1, T2
- ✅ **Decentralized coordination** - trucks bid on fires and evaluate bids collectively
- ✅ **Robustness** - no single point of failure, handles message delays

### Task 4: Demo & Presentation
- ✅ **Concurrent trucks** - both T1 and T2 move simultaneously
- ✅ **Real-time visualization** - fires spawn, spread, trucks respond
- ✅ **Distributed consensus** - bidding and assignment without coordinator

## Architecture

- **Fire Trucks (T1, T2)**: Generate fires, bid on alerts, extinguish fires
- **Water Supply**: Shared resource with Lamport-based access control
- **Observer**: Real-time grid visualization
- **NATS**: Pub-sub messaging between all components

## Key Files

- `cmd/distributed/main.go` - System orchestration
- `pkg/clock/lamport.go` - Lamport clock implementation
- `pkg/transport/nats.go` - Message transport layer
- `pkg/simulation/` - Fire grid, trucks, water supply
