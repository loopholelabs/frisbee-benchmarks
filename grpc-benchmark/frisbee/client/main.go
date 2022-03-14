package main

import (
	"context"
	benchmark "github.com/loopholelabs/frisbee-benchmarks/grpc-benchmark/frisbee/proto"
	"github.com/loov/hrtime"
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

	runBenchmark := func(id int, c *benchmark.Client) {
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
			}
			if shouldLog {
				log.Printf("Client with ID %d completed run %d in %s\n", id, i, time.Since(t))
			}
			done <- struct{}{}
		}
	}

	var c *benchmark.Client
	c, err = benchmark.NewClient(os.Args[1], nil, nil)
	if err != nil {
		panic(err)
	}

	err = c.Connect()
	if err != nil {
		panic(err)
	}

	for i := 0; i < clients; i++ {
		go runBenchmark(i, c)
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
	err = c.Close()
	if err != nil {
		panic(err)
	}
	log.Println(bench.Histogram(10))
}
