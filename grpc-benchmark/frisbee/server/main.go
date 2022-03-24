package main

import (
	"context"
	benchmark "go.buf.build/loopholelabs/frisbee/loopholelabs/frisbee-benchmark"
	"log"
	"os"
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
	frisbeeServer, err := benchmark.NewServer(new(svc), nil, nil)
	if err != nil {
		panic(err)
	}

	shouldLog := len(os.Args) > 2

	if shouldLog {
		go func() {
			err = frisbeeServer.Start(os.Args[1])
			if err != nil {
				panic(err)
			}
		}()

		for {
			log.Printf("Num goroutines: %d\n", runtime.NumGoroutine())
			time.Sleep(time.Millisecond * 500)
		}
	} else {
		err = frisbeeServer.Start(os.Args[1])
		if err != nil {
			panic(err)
		}
	}
}
