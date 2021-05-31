# Frisbee Benchmarks

This repository contains a series of benchmarks for Frisbee. You can learn more about these benchmarks at https://loopholelabs.io/docs/frisbee.

## Nats Benchmark

This benchmark is designed to be a 1:1 comparison with NATS for PUB/SUB. We have built simple broker, sender, and receiver examples in Frisbee, then
replicated those same examples using NATS (with the official NATS broker being used in place of the Frisbee broker).

### NATS Side

To start the NATS benchmark, you must have a NATS server accessible on the default NATS url (often `localhost:4222`). Then simply start the receivers and senders
with `go run receiver/main.go` and `go run sender/main.go`.

### Frisbee Side

To start the Frisbee benchmark, you need to start the broker, receiver, and sender in that order (the receiver and senders will error out if they are unable to 
connect to a Frisbee server). You can start them like so:


```shell
go run broker/main.go
go run reciever/main.go
go run sender/main.go
```

## Throughput Client to Server Benchmark

This benchmark is designed to test the Frisbee throughput when pushing data from the Frisbee Client to the Frisbee Server.

To start this benchmark, you mush start the server and then the client (in that order).

```shell
go run server/main.go
go run client/main.go
```

## Throughput Server to Client Benchmark

This benchmark is designed to test the Frisbee throughput when pushing data from the Frisbee Server to the Frisbee Client.

To start this benchmark, you mush start the server and then the client (in that order).

```shell
go run server/main.go
go run client/main.go
```