package main

import (
	"context"
	"crypto/rand"
	"github.com/loov/hrtime"
	benchmark "go.buf.build/loopholelabs/frisbee/loopholelabs/frisbee-benchmark"
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

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	req := new(benchmark.Request)
	req.Message = string(data)

	log.Printf("[CLIENT] Running benchmark with Message Size %d, Messages per Run %d, Num Runs %d, and Num Clients %d\n", messageSize, testSize, runs, clients)

	start := make(chan struct{}, clients)
	done := make(chan struct{}, clients)

	createClient := func(id int) {
		var c *benchmark.Client
		c, err = benchmark.NewClient(os.Args[1], nil, nil)
		if err != nil {
			panic(err)
		}

		err = c.Connect()
		if err != nil {
			panic(err)
		}

		var t time.Time
		for i := 0; i < runs; i++ {
			<-start
			t = time.Now()
			for q := 0; q < testSize; q++ {
				err = c.BenchmarkIgnore(context.Background(), req)
				if err != nil {
					panic(err)
				}
			}
			done <- struct{}{}
			log.Printf("Client %d Finish in %s\n", id, time.Since(t))
		}

		err = c.Close()
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < clients; i++ {
		go createClient(i)
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
