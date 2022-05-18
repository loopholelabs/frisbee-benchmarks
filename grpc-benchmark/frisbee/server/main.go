package main

import (
	"context"
	"github.com/loopholelabs/frisbee"
	"github.com/rs/zerolog"
	benchmark "go.buf.build/loopholelabs/frpc/loopholelabs/frisbee-benchmark"
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
	shouldLog := len(os.Args) > 2
	var logger *zerolog.Logger
	if shouldLog {
		l := zerolog.New(os.Stdout)
		logger = &l
	}
	frisbeeServer, err := benchmark.NewServer(new(svc), nil, logger)
	if err != nil {
		panic(err)
	}

	if shouldLog {
		err = frisbeeServer.SetOnClosed(func(async *frisbee.Async, err error) {
			logger.Error().Err(err).Msg("Error caused connection to close")
		})
		if err != nil {
			panic(err)
		}
	}

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
