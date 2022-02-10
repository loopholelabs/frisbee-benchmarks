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
	"github.com/loopholelabs/frisbee"
	"github.com/loopholelabs/frisbee/pkg/packet"
	"github.com/rs/zerolog"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
)

const PUB = uint16(10)
const SUB = uint16(11)

var receiveTopic = uint16(0)
var sendTopic = uint16(1)

const END = "END"

// Handle the PUB message type
func handlePub(_ context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	if incoming.Metadata.Id == receiveTopic {
		if string(incoming.Content) == END {
			incoming.Reset()
			incoming.Metadata.Id = sendTopic
			incoming.Metadata.Operation = PUB
			outgoing = incoming
		}
	}
	return
}

func main() {
	router := make(frisbee.HandlerTable)
	router[PUB] = handlePub
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	emptyLogger := zerolog.New(ioutil.Discard)
	c, err := frisbee.NewClient(os.Args[1], router, context.Background(), frisbee.WithLogger(&emptyLogger))
	if err != nil {
		panic(err)
	}

	err = c.Connect()
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

	p := packet.Get()
	p.Metadata.Id = receiveTopic
	p.Metadata.Operation = SUB

	err = c.WritePacket(p)
	if err != nil {
		panic(err)
	}

	// Now the handle pub function will be called
	// automatically whenever a message that matches the topic arrives

	<-exit
	err = c.Close()
	if err != nil {
		panic(err)
	}
}
