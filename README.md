# Firetruck Simulation

A distributed fire-fighting simulation system that demonstrates coordination, communication, and resource management between multiple firetrucks using the NATS messaging system and Lamport logical clocks for causal ordering.

### Installation

**Clone the repository**:
   ```bash
   git clone https://github.com/egabrielsson/lab2_dit357.git
   cd lab2_dit357
   ```

**Install Go dependencies**:
   ```bash
   go mod tidy
   ```
**Run program:**
```bash
# Build and run
go build -o distributed cmd/distributed/main.go
./start.sh

# View simulation
./logs.sh observer    # Grid and trucks
./logs.sh truck-t1    # Truck T1 activity
./logs.sh truck-t2    # Truck T2 activity
```

## Overview

This project simulates a distributed fire-fighting system where multiple firetrucks coordinate to extinguish fires on a grid. The system demonstrates:

- **Distributed coordination** between autonomous agents (firetrucks)
- **Publish-subscribe messaging** using NATS
- **Causal ordering** with Lamport logical clocks
- **Bidding protocols** for optimal fire assignment
- **Real-time communication** between distributed nodes

## Architecture

**Components**

**Each firetruck is an independent agent that contains:**
- **A Lamport logical clock** for event ordering
- **A transport layer** (NATS client) for communication
- **Simulation logic** for movement, firefighting, and resource management
- **Local state** (position, water level, current task)

- **NATS Message Broker:** A central communication hub that manages:
- **Multiple communication channels** (fires.alerts, trucks.status, water.requests...)
- **Message routing** between firetrucks
- **Publish-subscribe** pattern for broadcast messages

## Prerequisites & Installation

### Requirements
- **Go 1.24.2** or later
- **Docker** (for NATS server)
- **Git** (for cloning the repository)

## Key Files

- `cmd/distributed/main.go` - System orchestration
- `pkg/clock/lamport.go` - Lamport clock implementation
- `pkg/transport/nats.go` - Message transport layer
- `pkg/simulation/` - Fire grid, trucks, water supply
