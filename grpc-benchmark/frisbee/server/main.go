package main

import (
	"context"
	benchmark "github.com/loopholelabs/frisbee-benchmarks/grpc-benchmark/frisbee/proto"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type svc struct{}

func (s *svc) Benchmark(_ context.Context, req *benchmark.Request) (*benchmark.Response, error) {
	res := new(benchmark.Response)
	res.Message = req.Message
	return res, nil
}

func main() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	server, err := benchmark.NewServer(new(svc), os.Args[1], nil, nil)
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-exit:
			err = server.Shutdown()
			if err != nil {
				panic(err)
			}
			return
		default:
			log.Printf("Num goroutines: %d\n", runtime.NumGoroutine())
			time.Sleep(time.Millisecond * 500)
		}

	}

	return
}
