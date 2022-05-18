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
	"log"
	"os"
	"os/signal"
	"sync"
)

const PUB = uint16(10)
const SUB = uint16(11)
const ConnKey = "conn"

var mu sync.RWMutex
var subscribers = make(map[uint16][]*frisbee.Async)
var subscriptions = make(map[*frisbee.Async]map[uint16]bool)

func handleSub(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	conn := ctx.Value(ConnKey).(*frisbee.Async)
	log.Printf("[BROKER] Adding subscriber for ID %d, with Remote IP %s\n", incoming.Metadata.Id, conn.RemoteAddr())
	mu.Lock()
	subscribers[incoming.Metadata.Id] = append(subscribers[incoming.Metadata.Id], conn)
	if m, ok := subscriptions[conn]; !ok {
		m = make(map[uint16]bool)
		subscriptions[conn] = m
	} else {
		m[incoming.Metadata.Id] = true
	}
	mu.Unlock()
	return
}

func handlePub(_ context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
	mu.RLock()
	if connections := subscribers[incoming.Metadata.Id]; connections != nil {
		p := packet.Get()
		p.Metadata.Operation = incoming.Metadata.Operation
		p.Metadata.ContentLength = incoming.Metadata.ContentLength
		p.Metadata.Id = incoming.Metadata.Id
		p.Content.Write(incoming.Content.B)
		for _, c := range connections {
			_ = c.WritePacket(p)
		}
		packet.Put(p)
	}
	mu.RUnlock()

	return
}

func main() {
	router := make(frisbee.HandlerTable)
	router[SUB] = handleSub
	router[PUB] = handlePub
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	emptyLogger := zerolog.New(ioutil.Discard)

	s, err := frisbee.NewServer(router, frisbee.WithLogger(&emptyLogger))
	if err != nil {
		panic(err)
	}
	s.ConnContext = func(ctx context.Context, conn *frisbee.Async) context.Context {
		return context.WithValue(ctx, ConnKey, conn)
	}
	err = s.SetOnClosed(func(conn *frisbee.Async, err error) {
		log.Printf("[BROKER] Removing subscriber with Remote IP %s\n", conn.RemoteAddr())
		mu.Lock()
		if m, ok := subscriptions[conn]; ok {
			for k, v := range m {
				if v {
					for i, c := range subscribers[k] {
						if c == conn {
							subscribers[k][i] = subscribers[k][len(subscribers[k])-1]
							subscribers[k] = subscribers[k][:len(subscribers[k])-1]
						}
					}
				}
			}
		}
		delete(subscriptions, conn)
		mu.Unlock()
	})
	if err != nil {
		panic(err)
	}

	err = s.Start(os.Args[1])
	if err != nil {
		panic(err)
	}

	<-exit
	err = s.Shutdown()
	if err != nil {
		panic(err)
	}
}
