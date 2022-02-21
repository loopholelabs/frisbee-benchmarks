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
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
)

func main() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	nc, err := nats.Connect(fmt.Sprintf("nats://%s", os.Args[1]))
	if err != nil {
		panic(err)
	}

	sendTopic := "1"
	if len(os.Args) > 2 {
		sendTopic = os.Args[2]
	}

	receiveTopic := "0"
	if len(os.Args) > 3 {
		receiveTopic = os.Args[3]
	}

	_, err = nc.Subscribe(receiveTopic, func(m *nats.Msg) {
		if string(m.Data) == "END" {
			err := nc.Publish(sendTopic, []byte("END"))
			if err != nil {
				panic(err)
			}
		}
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Ready to Receive\n")

	<-exit
	nc.Close()
}
