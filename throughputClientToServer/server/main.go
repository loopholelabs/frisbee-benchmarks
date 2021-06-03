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
	"fmt"
	"github.com/loophole-labs/frisbee"
	"github.com/loophole-labs/frisbee-benchmarks/internal/message"
	"github.com/rs/zerolog"
	"io/ioutil"
	"os"
	"os/signal"
)

const testSize = 100000
const port = 8192

func handlePing(_ *frisbee.Conn, incomingMessage frisbee.Message, _ []byte) (outgoingMessage *frisbee.Message, outgoingContent []byte, action frisbee.Action) {
	if incomingMessage.Id == testSize-1 {
		outgoingMessage = &frisbee.Message{
			To:            0,
			From:          0,
			Id:            testSize,
			Operation:     message.PONG,
			ContentLength: 0,
		}
	}
	return
}

func main() {
	router := make(frisbee.ServerRouter)
	router[message.PING] = handlePing
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	emptyLogger := zerolog.New(ioutil.Discard)

	s := frisbee.NewServer(fmt.Sprintf(":%d", port), router, frisbee.WithLogger(&emptyLogger))
	s.OnOpened = func(server *frisbee.Server, c *frisbee.Conn) frisbee.Action {
		server.Logger().Debug().Msgf("Client connected: %s", c.RemoteAddr())
		return frisbee.NONE
	}

	_ = s.Start()

	<-exit
	err := s.Shutdown()
	if err != nil {
		panic(err)
	}
}
