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
)

const PUB = uint16(10)
const SUB = uint16(11)
const ConnKey = "conn"

var subscribers = make(map[uint16][]frisbee.Conn)

func handleSub(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	subscribers[incoming.Metadata.Id] = append(subscribers[incoming.Metadata.Id], ctx.Value(ConnKey).(*frisbee.Async))
	return
}

func handlePub(_ context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	if connections := subscribers[incoming.Metadata.Id]; connections != nil {
		for _, c := range connections {
			p := packet.Get()
			p.Metadata.Operation = incoming.Metadata.Operation
			p.Metadata.ContentLength = incoming.Metadata.ContentLength
			p.Metadata.Id = incoming.Metadata.Id
			p.Write(incoming.Content)
			_ = c.WritePacket(p)
			packet.Put(p)
		}
	}

	return
}

func main() {
	router := make(frisbee.HandlerTable)
	router[SUB] = handleSub
	router[PUB] = handlePub
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	emptyLogger := zerolog.New(ioutil.Discard)

	s, err := frisbee.NewServer(":8192", router, frisbee.WithLogger(&emptyLogger))
	if err != nil {
		panic(err)
	}
	s.ConnContext = func(ctx context.Context, conn *frisbee.Async) context.Context {
		return context.WithValue(ctx, ConnKey, conn)
	}
	err = s.Start()
	if err != nil {
		panic(err)
	}

	<-exit
	err = s.Shutdown()
	if err != nil {
		panic(err)
	}
}
