package main

import (
	"context"
	"github.com/loov/hrtime"
	benchmark "go.buf.build/loopholelabs/frisbee/loopholelabs/frisbee-benchmark"
	"log"
	"math/rand"
	"os"
	"strconv"
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

	req := new(benchmark.Request)
	req.Message = RandomString(messageSize)

	log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, and Num Clients %d\n", messageSize, testSize, runs, clients)

	start := make(chan struct{}, clients)
	done := make(chan struct{}, clients)

	createClient := func(id int, c *benchmark.Client) {
		for i := 0; i < runs; i++ {
			<-start
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
			done <- struct{}{}
		}

		err = c.Close()
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < clients; i++ {
		var c *benchmark.Client
		c, err = benchmark.NewClient(os.Args[1], nil, nil)
		if err != nil {
			panic(err)
		}

		err = c.Connect()
		if err != nil {
			panic(err)
		}

		go createClient(i, c)
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
