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
	"github.com/loov/hrtime"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"strconv"
)

var complete = make(chan struct{})

func main() {
	nc, err := nats.Connect(fmt.Sprintf("nats://%s", os.Args[1]))
	if err != nil {
		panic(err)
	}

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

	sendTopic := "0"
	if len(os.Args) > 5 {
		sendTopic = os.Args[5]
	}

	receiveTopic := "1"
	if len(os.Args) > 6 {
		receiveTopic = os.Args[6]
	}

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	_, err = nc.Subscribe(receiveTopic, func(m *nats.Msg) {
		complete <- struct{}{}
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Running benchmark with Message Size %d, Messages per Run %d, and Num Runs %d\n", messageSize, testSize, runs)

	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		for q := 0; q < testSize; q++ {
			err = nc.Publish(sendTopic, data)
			if err != nil {
				panic(err)
			}
		}
		err = nc.Publish(sendTopic, []byte("END"))
		if err != nil {
			panic(err)
		}
		<-complete
	}
	log.Println(bench.Histogram(10))
	nc.Close()
}
