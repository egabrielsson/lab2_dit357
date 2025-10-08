# lab2_dit357

## Simulation

```
go run cmd/simulate/main.go --steps 20
```

## NATS request-reply

T3. If still tied, lowest ID wins:

```
docker run -p 4222:4222 nats:latest
```

Terminal 2:

```
go run cmd/nats-node/main.go --id T1 --auto-offer
```

Terminal 3:

```
go run cmd/nats-node/main.go --id T2 --to T1 --request --amount 30
```

## NATS publish-subscribe

Terminal 1:

```
docker run -p 4222:4222 nats:latest
```

Terminal 2:

```
go run cmd/nats-node/main.go --id T1 --pubsub
```

Terminal 3:

```
go run cmd/nats-node/main.go --id T2 --pubsub
```

Terminal 4:

```
go run cmd/nats-node/main.go --id T3 --fire-alert
```

## Fire Bidding Coordination (Task 2)

Demonstrates distributed decision-making where trucks bid for fires based on distance, water, and ID.

Terminal 1 - NATS server:

```

docker run -p 4222:4222 nats:latest

```

Terminal 2 - Coordinator (evaluates bids):

```

go run cmd/nats-node/main.go --id FireCentral --coordinator

```

Terminal 3 - Truck T1:

```

go run cmd/nats-node/main.go --id T1 --bidding

```

Terminal 4 - Truck T2:

```

go run cmd/nats-node/main.go --id T2 --bidding

```

Terminal 5 - Trigger fire alert:

```

go run cmd/nats-node/main.go --id FireStation --fire-alert

```

Both T1 and T2 bid. FireCentral evaluates and announces winner:

1. Closest distance wins
2. If tied, most water wins
3. If still tied, lowest ID wins

```

```

```

```
