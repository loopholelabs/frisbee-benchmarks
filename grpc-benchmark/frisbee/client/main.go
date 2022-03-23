package main

import (
	"context"
	"fmt"
	"github.com/loov/hrtime"
	benchmark "go.buf.build/loopholelabs/frisbee/loopholelabs/frisbee-benchmark"
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

func (c *client) run(wg *sync.WaitGroup, size int, req *benchmark.Request) {
	var res *benchmark.Response
	var err error
	for q := 0; q < size; q++ {
		res, err = c.Benchmark(context.Background(), req)
		if err != nil {
			panic(err)
		}
		if res.Message != req.Message {
			panic("invalid response")
		}
	}
	wg.Done()
}

func (c *client) start(wg *sync.WaitGroup, id int, concurrent int, run int, size int, req *benchmark.Request, shouldLog bool) {
	var runWg sync.WaitGroup
	t := time.Now()
	runWg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go c.run(&runWg, size, req)
	}
	runWg.Wait()
	if shouldLog {
		log.Printf("Clients (%d concurrent) with ID %d completed run %d in %s\n", concurrent, id, run, time.Since(t))
	}
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

	if shouldLog {
		log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, Num Clients %d, and Num Concurrent %d\n", messageSize, testSize, runs, numClients, numConcurrent)
	}

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
	i := 0
	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		wg.Add(numClients)
		for id, c := range clients {
			go c.start(&wg, id, numConcurrent, i, testSize, req, shouldLog)
		}
		wg.Wait()
		i++
	}

	for _, c := range clients {
		err = c.Close()
		if err != nil {
			panic(err)
		}
	}

	if shouldLog {
		log.Println(bench.Histogram(10))
	} else {
		m := bench.Float64s()
		sum := float64(0)
		for i := 0; i < len(m); i++ {
			sum += m[i]
		}
		sum /= float64(len(m))
		fmt.Printf("%s\n", time.Duration(sum))
	}
}
