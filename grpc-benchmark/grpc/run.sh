#!/bin/bash

echo "Running Server, binding to $1"
go run server/main.go "$1" &
SERVER_PID=$!
echo "Server started with PID $SERVER_PID"

sleep 1

echo "Running Client connecting to $1"
go run client/main.go "$1" 32 1000 100 1
go run client/main.go "$1" 32 1000 100 2
go run client/main.go "$1" 32 1000 100 5
go run client/main.go "$1" 32 1000 100 10
go run client/main.go "$1" 32 1000 100 25
go run client/main.go "$1" 32 1000 100 50
go run client/main.go "$1" 32 1000 100 100

go run client/main.go "$1" 512 1000 100 1
go run client/main.go "$1" 512 1000 100 2
go run client/main.go "$1" 512 1000 100 5
go run client/main.go "$1" 512 1000 100 10
go run client/main.go "$1" 512 1000 100 25
go run client/main.go "$1" 512 1000 100 50
go run client/main.go "$1" 512 1000 100 100

go run client/main.go "$1" 131072 100 100 1
go run client/main.go "$1" 131072 100 100 2
go run client/main.go "$1" 131072 100 100 5
go run client/main.go "$1" 131072 100 100 10
go run client/main.go "$1" 131072 100 100 25
go run client/main.go "$1" 131072 100 100 50
go run client/main.go "$1" 131072 100 100 100

go run client/main.go "$1" 1048576 10 100 1
go run client/main.go "$1" 1048576 10 100 2
go run client/main.go "$1" 1048576 10 100 5
go run client/main.go "$1" 1048576 10 100 10
go run client/main.go "$1" 1048576 10 100 25
go run client/main.go "$1" 1048576 10 100 50
go run client/main.go "$1" 1048576 10 100 100

kill -9 $SERVER_PID
pkill main