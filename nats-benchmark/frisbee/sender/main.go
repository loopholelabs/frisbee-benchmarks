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
	"context"
	"crypto/rand"
	"github.com/loopholelabs/frisbee"
	"github.com/loopholelabs/frisbee/pkg/packet"
	"github.com/loov/hrtime"
	"github.com/rs/zerolog"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const PUB = uint16(10)
const SUB = uint16(11)

var complete = make(chan struct{})

var sendTopic = uint16(0)
var receiveTopic = uint16(1)

func handlePub(_ context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	if incoming.Metadata.Id == receiveTopic {
		complete <- struct{}{}
	}
	return
}

func main() {
	router := make(frisbee.HandlerTable)
	router[PUB] = handlePub

	emptyLogger := zerolog.New(ioutil.Discard)
	c, err := frisbee.NewClient(os.Args[1], router, context.Background(), frisbee.WithLogger(&emptyLogger))
	if err != nil {
		panic(err)
	}

	err = c.Connect()
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

	if len(os.Args) > 5 {
		topic, err := strconv.Atoi(os.Args[5])
		if err != nil {
			panic(err)
		}
		sendTopic = uint16(topic)
	}

	if len(os.Args) > 6 {
		topic, err := strconv.Atoi(os.Args[6])
		if err != nil {
			panic(err)
		}
		receiveTopic = uint16(topic)
	}

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	END := []byte("END")

	p := packet.Get()
	p.Metadata.Operation = SUB
	p.Metadata.Id = receiveTopic
	err = c.WritePacket(p)
	if err != nil {
		panic(err)
	}

	p.Reset()
	p.Metadata.Id = sendTopic
	p.Metadata.Operation = PUB
	p.Metadata.ContentLength = uint32(len(data))
	p.Content.Write(data)

	endPacket := packet.Get()
	endPacket.Metadata.Id = sendTopic
	endPacket.Metadata.Operation = PUB
	endPacket.Metadata.ContentLength = uint32(len(END))
	endPacket.Content.Write(END)

	log.Printf("[SENDER] Running benchmark with Message Size %d, Messages per Run %d, and Num Runs %d\n", messageSize, testSize, runs)

	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		for q := 0; q < testSize; q++ {
			err = c.WritePacket(p)
			if err != nil {
				panic(err)
			}
		}
		err = c.WritePacket(endPacket)
		if err != nil {
			panic(err)
		}
		<-complete
	}
	log.Println(bench.Histogram(10))

	packet.Put(p)
	packet.Put(endPacket)

	err = c.Close()
	if err != nil {
		panic(err)
	}
}
