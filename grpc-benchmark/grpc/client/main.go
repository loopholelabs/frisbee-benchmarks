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
	"sync"
	"time"
)

type client struct {
	benchmark.BenchmarkServiceClient
	*grpc.ClientConn
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
		var conn *grpc.ClientConn
		conn, err = grpc.Dial(os.Args[1], grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		clients = append(clients, &client{ClientConn: conn, BenchmarkServiceClient: benchmark.NewBenchmarkServiceClient(conn)})
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
