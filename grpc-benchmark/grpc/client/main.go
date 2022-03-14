package main

import (
	"context"
	"github.com/loov/hrtime"
	benchmark "go.buf.build/grpc/go/loopholelabs/frisbee-benchmark"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}

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
	shouldLog := len(os.Args) > 6

	req := new(benchmark.Request)
	req.Message = RandomString(messageSize)

	log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, and Num Clients %d\n", messageSize, testSize, runs, clients)

	start := make(chan struct{}, clients)
	done := make(chan struct{}, clients)

	runBenchmark := func(id int, c benchmark.BenchmarkServiceClient) {
		var t time.Time
		for i := 0; i < runs; i++ {
			<-start
			t = time.Now()
			var res *benchmark.Response
			for q := 0; q < testSize; q++ {
				res, err = c.Benchmark(context.Background(), req)
				if err != nil {
					panic(err)
				}
				if res.Message != req.Message {
					panic("invalid response")
				}
				if shouldLog {
					log.Printf("Client with ID %d completed run %d in %s\n", id, i, time.Since(t))
				}
			}
			done <- struct{}{}
		}
	}

	var conn *grpc.ClientConn
	conn, err = grpc.Dial(os.Args[1], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	for i := 0; i < clients; i++ {
		client := benchmark.NewBenchmarkServiceClient(conn)
		go runBenchmark(i, client)
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

	err = conn.Close()
	if err != nil {
		panic(err)
	}

	log.Println(bench.Histogram(10))
}
