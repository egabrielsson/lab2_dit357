# Firetruck Simulation

A distributed fire-fighting simulation system that demonstrates coordination, communication, and resource management between multiple firetrucks using the NATS messaging system and Lamport logical clocks for causal ordering.

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites & Installation](#prerequisites--installation)
- [Usage](#usage)
  - [1. Standalone Simulation](#1-standalone-simulation)
  - [2. Distributed NATS Communication](#2-distributed-nats-communication)
  - [3. Fire Bidding Coordination](#3-fire-bidding-coordination)

## Overview

This project simulates a distributed fire-fighting system where multiple firetrucks coordinate to extinguish fires on a grid. The system demonstrates:

- **Distributed coordination** between autonomous agents (firetrucks)
- **Publish-subscribe messaging** using NATS
- **Causal ordering** with Lamport logical clocks
- **Resource management** (water supply coordination)
- **Bidding protocols** for optimal fire assignment
- **Real-time communication** between distributed nodes

### Core Simulation
- 20x20 grid-based fire simulation
- Multiple firetrucks with autonomous behavior
- Dynamic fire ignition, spreading, and extinguishing
- Water tank management and refilling mechanics
- Real-time visualization of simulation state

### Distributed Communication
- **NATS messaging** for inter-truck communication
- **Lamport logical clocks** for event ordering
- **Publish-subscribe** for broadcast messages

### Coordination Protocols
- **Fire bidding system** - trucks compete for fire assignments
- **Distance-based optimization** - closest truck typically wins
- **Water level consideration** - trucks with more water are preferred
- **Tie-breaking mechanisms** - uses truck IDs for deterministic results
- **Distributed consensus** - no single point of failure


## Prerequisites & Installation

### Requirements
- **Go 1.24.2** or later
- **Docker** (for NATS server)
- **Git** (for cloning the repository)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/egabrielsson/lab2_dit357.git
   cd lab2_dit357
   ```

2. **Install Go dependencies**:
   ```bash
   go mod tidy
   ```

3. **Start NATS server** (required for distributed features):
   ```bash
   docker run -p 4222:4222 nats:latest
   ```

## Usage

### 1. Standalone Simulation

Run a local simulation without distributed communication:

```bash
cd cmd/simulate
go run .
```

### 2. Distributed NATS Communication

Demonstrates publish-subscribe messaging between trucks:

**Terminal 1** - NATS Server:
```bash
docker run -p 4222:4222 nats:latest
```

**Terminal 2** - Truck T1 (subscriber):
```bash
go run cmd/nats-node/main.go --id T1 --pubsub
```

**Terminal 3** - Truck T2 (subscriber):
```bash
go run cmd/nats-node/main.go --id T2 --pubsub
```

**Terminal 4** - Fire Station (publisher):
```bash
go run cmd/nats-node/main.go --id T3 --fire-alert
```

### 3. Fire Bidding Coordination

Demonstrates distributed decision-making where trucks bid for fires:

**Terminal 1** - NATS Server:
```bash
docker run -p 4222:4222 nats:latest
```

**Terminal 2** - Coordinator (evaluates bids):
```bash
go run cmd/nats-node/main.go --id FireCentral --coordinator
```

**Terminal 3** - Truck T1 (bidder):
```bash
go run cmd/nats-node/main.go --id T1 --bidding
```

**Terminal 4** - Truck T2 (bidder):
```bash
go run cmd/nats-node/main.go --id T2 --bidding
```

**Terminal 5** - Trigger fire alert:
```bash
go run cmd/nats-node/main.go --id FireStation --fire-alert
```

**Bidding Process**:
1. Fire station broadcasts fire alert
2. All trucks calculate distance and submit bids
3. Coordinator evaluates bids using criteria:
   - **Primary**: Closest distance wins
   - **Secondary**: If tied, most water wins  
   - **Tertiary**: If still tied, lowest ID wins
4. Winner announcement broadcasted to all trucks