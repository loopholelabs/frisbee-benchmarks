#!/bin/bash

echo "Running Broker, binding to $1"
go run broker/main.go "$1" &
BROKER_PID=$!
echo "Broker started with PID $BROKER_PID"

sleep 1

echo "Running Receiver, connecting to $1"
go run receiver/main.go "$1" &
RECEIVER_PID=$!
echo "Receiver started with PID $RECEIVER_PID"

sleep 1

echo "Running Sender. connecting to $1"
go run sender/main.go "$1" 32 100000 100
go run sender/main.go "$1" 128 100000 100
go run sender/main.go "$1" 512 100000 100
go run sender/main.go "$1" 4096 50000 100
go run sender/main.go "$1" 8192 50000 100
go run sender/main.go "$1" 131072 1000 100
go run sender/main.go "$1" 1048576 100 100

kill -9 $BROKER_PID
kill -9 $RECEIVER_PID
pkill main