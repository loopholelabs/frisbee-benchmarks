#!/bin/bash

echo "Running Using $1 as Broker"

nats-server &
BROKER_PID=$!

sleep 1

go run ../receiver/main.go "$1"
RECEIVER_PID=$!

sleep 1

go run main.go "$1" 32 100000 100
go run main.go "$1" 128 100000 100
go run main.go "$1" 512 100000 100
go run main.go "$1" 4096 50000 100
go run main.go "$1" 8192 50000 100
go run main.go "$1" 131072 1000 100
go run main.go "$1" 1048576 100 100

kill $BROKER_PID
kill $RECEIVER_PID