#!/bin/bash

# Log viewer script for DECENTRALIZED Fire Truck System
# Usage: ./logs.sh [observer|truck-t1|truck-t2|water-supply]

LOG_TYPE=${1:-observer}

case $LOG_TYPE in
    "observer")
        echo "Watching observer (DECENTRALIZED grid visualization)..."
        tail -f logs/observer.log
        ;;
    "truck-t1")
        echo "Watching Truck T1 logs (DECENTRALIZED bidding)..."
        tail -f logs/truck-t1.log
        ;;
    "truck-t2")
        echo "Watching Truck T2 logs (DECENTRALIZED bidding)..."
        tail -f logs/truck-t2.log
        ;;
    "water-supply")
        echo "Watching Water Supply logs (mutual exclusion)..."
        tail -f logs/water-supply.log
        ;;
    *)
        echo "Usage: $0 [observer|truck-t1|truck-t2|water-supply]"
        echo ""
        echo "Available log types:"
        echo "  observer     - DECENTRALIZED grid visualization and system state"
        echo "  truck-t1     - Truck T1 logs (DECENTRALIZED bidding, movement, etc.)"
        echo "  truck-t2     - Truck T2 logs (DECENTRALIZED bidding, movement, etc.)"
        echo "  water-supply - Water supply logs (mutual exclusion)"
        exit 1
        ;;
esac
