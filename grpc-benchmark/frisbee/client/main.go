package main

import (
	"context"
	benchmark "github.com/loopholelabs/frisbee-benchmarks/grpc-benchmark/frisbee/proto"
	"github.com/loov/hrtime"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type client struct {
	*benchmark.Client
}

func (c *client) run(wg *sync.WaitGroup, id int, concurrency int, size int, req *benchmark.Request, shouldLog bool) {
	var res *benchmark.Response
	var err error
	t := time.Now()
	for q := 0; q < size; q++ {
		res, err = c.Benchmark(context.Background(), req)
		if err != nil {
			panic(err)
		}
		if res.Message != req.Message {
			panic("invalid response")
		}
	}
	if shouldLog {
		log.Printf("Client with ID %d and concurrency %d completed in %s\n", id, concurrency, time.Since(t))
	}
	wg.Done()
}

func (c *client) start(wg *sync.WaitGroup, id int, concurrent int, size int, req *benchmark.Request, shouldLog bool) {
	var runWg sync.WaitGroup
	runWg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go c.run(&runWg, id, i, size, req, shouldLog)
	}
	runWg.Wait()
	wg.Done()
}

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

	numClients, err := strconv.Atoi(os.Args[5])
	if err != nil {
		panic(err)
	}

	numConcurrent, err := strconv.Atoi(os.Args[6])
	if err != nil {
		panic(err)
	}

	shouldLog := len(os.Args) > 7

	req := new(benchmark.Request)
	req.Message = RandomString(messageSize)

	log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, Num Clients %d, an Num Concurrent %d\n", messageSize, testSize, runs, numClients, numConcurrent)

	clients := make([]*client, 0, numClients)

	for i := 0; i < numClients; i++ {
		var c *benchmark.Client
		c, err = benchmark.NewClient(os.Args[1], nil, nil)
		if err != nil {
			panic(err)
		}

		err = c.Connect()
		if err != nil {
			panic(err)
		}
		clients = append(clients, &client{Client: c})
	}

	var wg sync.WaitGroup
	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		wg.Add(numClients)
		for id, c := range clients {
			go c.start(&wg, id, numConcurrent, testSize, req, shouldLog)
		}
		wg.Wait()
	}

	for _, c := range clients {
		err = c.Close()
		if err != nil {
			panic(err)
		}
	}

	log.Println(bench.Histogram(10))
}
