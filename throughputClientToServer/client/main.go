/*
	Copyright 2021 Loophole Labs

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		   http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package main

import (
	"crypto/rand"
	"fmt"
	"github.com/loophole-labs/frisbee"
	"github.com/loophole-labs/frisbee-benchmarks/internal/message"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"time"
)

const testSize = 100000
const messageSize = 512
const runs = 10
const port = 8192

var complete = make(chan struct{})

func handlePong(incomingMessage frisbee.Message, _ []byte) (outgoingMessage *frisbee.Message, outgoingContent []byte, action frisbee.Action) {
	if incomingMessage.Id == testSize {
		complete <- struct{}{}
	}
	return
}

func main() {
	router := make(frisbee.ClientRouter)
	router[message.PONG] = handlePong

	emptyLogger := zerolog.New(ioutil.Discard)

	c := frisbee.NewClient(fmt.Sprintf("127.0.0.1:%d", port), router, frisbee.WithLogger(&emptyLogger))
	_ = c.Connect()

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	duration := time.Nanosecond * 0
	for i := 1; i < runs+1; i++ {
		start := time.Now()
		for q := 0; q < testSize; q++ {
			err := c.Write(&frisbee.Message{
				To:            uint32(i),
				From:          uint32(i),
				Id:            uint32(q),
				Operation:     message.PING,
				ContentLength: messageSize,
			}, &data)
			if err != nil {
				panic(err)
			}
		}
		<-complete
		runTime := time.Since(start)
		log.Printf("Benchmark Time for test %d: %d ns", i, runTime.Nanoseconds())
		duration += runTime
	}
	log.Printf("Average Benchmark time for %d runs: %d ns, throughput: %f mb/s", runs, duration.Nanoseconds()/runs, (1/((duration.Seconds()/runs)/testSize)*messageSize)/(1024*1024))
	_ = c.Close()
}
