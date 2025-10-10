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

## Architecture

**Components**

**Firetruck Agents:** Each firetruck is an independent agent that contains:
•  **A Lamport logical clock** for event ordering
•  **A transport layer** (NATS client) for communication
•  **Simulation logic** for movement, firefighting, and resource management
•  **Local state** (position, water level, current task)

**NATS Message Broker:** A central communication hub that manages:
•  **Multiple communication channels** (fires.alerts, trucks.status, water.requests, coordination, fires.bids, fires.assignment)
•  *Message routing* between firetrucks
•  **Publish-subscribe** patterns for broadcast messages

**Communication Channels:** Specialized message channels for different purposes:
•  **Fire-related channels** for alerts, bidding, and assignments
•  **Truck status channels** for sharing state information
•  **Water channels** for resource coordination
•  **Coordination channels** for movement planning

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