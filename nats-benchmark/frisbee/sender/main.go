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
	"github.com/loophole-labs/frisbee"
	"github.com/loov/hrtime"
	"github.com/rs/zerolog"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const PUB = uint32(1)
const SUB = uint32(2)

var complete = make(chan struct{})

var topic = []byte("SENDING")
var topicHash = crc32.ChecksumIEEE(topic)

var receiveTopic = []byte("RECEIVING")
var receiveTopicHash = crc32.ChecksumIEEE(receiveTopic)

func handlePub(incomingMessage frisbee.Message, _ []byte) (outgoingMessage *frisbee.Message, outgoingContent []byte, action frisbee.Action) {
	if incomingMessage.To == receiveTopicHash {
		complete <- struct{}{}
	}
	return
}

func main() {
	router := make(frisbee.ClientRouter)

	router[PUB] = handlePub

	emptyLogger := zerolog.New(ioutil.Discard)

	c := frisbee.NewClient(os.Args[1], router, frisbee.WithLogger(&emptyLogger))
	err := c.Connect()
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

	if os.Args[5] != "" {
		topic = []byte(os.Args[5])
		topicHash = crc32.ChecksumIEEE(topic)
	}

	if os.Args[6] != "" {
		receiveTopic = []byte(os.Args[6])
		receiveTopicHash = crc32.ChecksumIEEE(topic)
	}

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	END := []byte("END")

	err = c.Write(&frisbee.Message{
		From:          0,
		To:            0,
		Id:            0,
		Operation:     SUB,
		ContentLength: uint64(len(receiveTopic)),
	}, &receiveTopic)
	if err != nil {
		panic(err)
	}

	i := 0
	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		for q := 0; q < testSize; q++ {
			err := c.Write(&frisbee.Message{
				From:          topicHash,
				To:            topicHash,
				Id:            uint32(i),
				Operation:     PUB,
				ContentLength: uint64(len(data)),
			}, &data)
			if err != nil {
				panic(err)
			}
		}
		err := c.Write(&frisbee.Message{
			From:          topicHash,
			To:            topicHash,
			Id:            uint32(i),
			Operation:     PUB,
			ContentLength: uint64(len(END)),
		}, &END)
		if err != nil {
			panic(err)
		}
		i++
		<-complete
	}
	log.Println(bench.Histogram(10))

	err = c.Close()
	if err != nil {
		panic(err)
	}
}
