# Frisbee Benchmarks

This repository contains a series of benchmarks for Frisbee. You can learn more about these benchmarks at https://loopholelabs.io/docs/frisbee.

## NATS Benchmark

This benchmark is designed to be a 1:1 comparison with NATS for PUB/SUB. We have built simple broker, sender, and receiver examples in Frisbee, then
replicated those same examples using NATS (with the official NATS broker being used in place of the Frisbee broker).

### NATS Side

To start the NATS benchmark, you must have a NATS server accessible on the default NATS url (often `localhost:4222`). Then simply start the receivers and senders
with `go run receiver/main.go localhost:4222` and `go run sender/main.go localhost:4222 <bytes per message> <number of messages> <repetitions>`.

### Frisbee Side

To start the Frisbee benchmark, you need to start the broker, receiver, and sender in that order (the receiver and senders will error out if they are unable to
connect to a Frisbee server). You can start them like so:

```shell
go run broker/main.go # Will run on 0.0.0.0:8192 by default
go run reciever/main.go localhost:8192
go run sender/main.go localhost:8192 <bytes per message> <number of messages> <repetitions>
```

## GRPC Benchmark

This benchmark is designed to be a 1:1 comparison with GRPC for sending and receiving large numbers of messages. In order to facilitate this, both the GRPC and Frisbee
implementations make use of the same `benchmark.proto` file, and the RPC frameworks themselves are generated
via [Buf](https://buf.build) to guarantee clean builds. We've also kept the client and server implementations the exact same (with only the instantiation of the servers and clients being different).

### GRPC Side

To start the GRPC benchmark, you must start the server and then the client (in that order). You can start them like so:

```shell
go run server/main.go localhost:8192 
go run client/main.go localhost:8192 <bytes per message> <number of messages> <repetitions> <number of clients> <number of parallel senders per client>
```

### Frisbee Side

To start the Frisbee benchmark, you must start the server and then the client (in that order). You can start them like so:

```shell
go run server/main.go localhost:8192 
go run client/main.go localhost:8192 <bytes per message> <number of messages> <repetitions> <number of clients> <number of parallel senders per client>
```
