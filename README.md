# lab2_dit357

## Simulation

```
go run cmd/simulate/main.go --steps 20
```

## TCP sockets

Terminal 1:

```
go run cmd/tcp-node/main.go --id T1 --addr :9001 --auto-offer
```

Terminal 2:

```
go run cmd/tcp-node/main.go --id T2 --addr :9002 --peers :9001@T1 --request --amount 30
```

## NATS request-reply

Terminal 1:

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
