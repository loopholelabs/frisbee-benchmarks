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
	"github.com/loophole-labs/frisbee"
	"hash/crc32"
	"os"
	"os/signal"
)

const PUB = uint32(1)
const SUB = uint32(2)

var subscribers = make(map[uint32][]*frisbee.Conn)

func handleSub(c *frisbee.Conn, incomingMessage frisbee.Message, incomingContent []byte) (outgoingMessage *frisbee.Message, outgoingContent []byte, action frisbee.Action) {
	if incomingMessage.ContentLength > 0 {
		checksum := crc32.ChecksumIEEE(incomingContent)
		subscribers[checksum] = append(subscribers[checksum], c)
	}
	return
}

func handlePub(_ *frisbee.Conn, incomingMessage frisbee.Message, incomingContent []byte) (outgoingMessage *frisbee.Message, outgoingContent []byte, action frisbee.Action) {
	if connections := subscribers[incomingMessage.To]; connections != nil {
		for _, c := range connections {
			_ = c.Write(&incomingMessage, &incomingContent)
		}
	}

	return
}

func main() {
	router := make(frisbee.ServerRouter)
	router[SUB] = handleSub
	router[PUB] = handlePub
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	//emptyLogger := zerolog.New(ioutil.Discard)

	s := frisbee.NewServer(":8192", router)
	_ = s.Start()

	<-exit
	err := s.Shutdown()
	if err != nil {
		panic(err)
	}
}
