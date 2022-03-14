package main

import (
	"context"
	"github.com/loov/hrtime"
	benchmark "go.buf.build/grpc/go/loopholelabs/frisbee-benchmark"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	messageSize, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	testSize, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}

	runs, err := strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}

	clients, err := strconv.Atoi(os.Args[5])
	if err != nil {
		panic(err)
	}

	req := new(benchmark.Request)
	req.Message = "Test String"

	log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, and Num Clients %d\n", messageSize, testSize, runs, clients)

	start := make(chan struct{}, clients)
	done := make(chan struct{}, clients)

	createClient := func(id int, conn *grpc.ClientConn, c benchmark.BenchmarkServiceClient) {
		var t time.Time
		for i := 0; i < runs; i++ {
			<-start
			t = time.Now()
			for q := 0; q < testSize; q++ {
				_, err = c.Benchmark(context.Background(), req)
				if err != nil {
					panic(err)
				}
			}
			done <- struct{}{}
			log.Printf("Client %d finished run %d in %s\n", id, i, time.Since(t))
		}

		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < clients; i++ {
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(os.Args[1], grpc.WithInsecure())
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		client := benchmark.NewBenchmarkServiceClient(conn)

		go createClient(i, conn, client)
	}

	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		for i := 0; i < clients; i++ {
			start <- struct{}{}
		}
		for i := 0; i < clients; i++ {
			<-done
		}
	}
	log.Println(bench.Histogram(10))
}
