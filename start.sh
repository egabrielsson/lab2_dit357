#!/bin/bash

# Start script for Fire Truck System
# This script starts all components: NATS server, coordinator, trucks, water supply, and observer

echo "Starting Fire Truck System..."

# Check if NATS server is running
if ! nc -z 127.0.0.1 4222 2>/dev/null; then
    echo "Starting NATS server..."
    nats-server &
    NATS_PID=$!
    sleep 2
    echo "NATS server started (PID: $NATS_PID)"
else
    echo "NATS server already running"
fi

# Build the system
echo "Building system..."
go build -o distributed cmd/distributed/main.go
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

# Start water supply
echo "Starting water supply..."
./distributed -id=WATER-SUPPLY -role=water-supply > logs/water-supply.log 2>&1 &
WATER_PID=$!
sleep 1

# Start trucks (DECENTRALIZED - NO COORDINATOR!)
echo "Starting fire trucks (DECENTRALIZED)..."
./distributed -id=T1 -role=truck > logs/truck-t1.log 2>&1 &
T1_PID=$!
sleep 0.5

./distributed -id=T2 -role=truck > logs/truck-t2.log 2>&1 &
T2_PID=$!
sleep 0.5

# Start observer
echo "Starting observer..."
./distributed -id=OBSERVER -role=observer > logs/observer.log 2>&1 &
OBS_PID=$!

echo ""
echo "╔═══════════════════════════════════════════════════╗"
echo "║  System Started Successfully!                   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""
echo "Process IDs:"
echo "  Water Supply: $WATER_PID"
echo "  Truck T1:     $T1_PID"
echo "  Truck T2:     $T2_PID"
echo "  Observer:     $OBS_PID"
echo ""
echo "To watch the simulation:"
echo "  ./logs.sh observer"
echo ""
echo "To watch truck logs:"
echo "  ./logs.sh truck-t1"
echo "  ./logs.sh truck-t2"
echo ""
echo "To stop the system:"
echo "  pkill -f distributed"
echo "  pkill -f nats-server"
echo ""

# Keep script running
wait
