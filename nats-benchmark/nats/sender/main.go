/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"crypto/rand"
	"github.com/loov/hrtime"
	"github.com/nats-io/nats.go"
	"log"
)

const testSize = 100000
const messageSize = 2048
const runs = 100

var complete = make(chan struct{})

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)

	data := make([]byte, messageSize)
	_, _ = rand.Read(data)

	_, _ = nc.Subscribe("DONE", func(m *nats.Msg) {
		complete <- struct{}{}
	})

	i := 0
	bench := hrtime.NewBenchmark(runs)
	for bench.Next() {
		for q := 0; q < testSize; q++ {
			err := nc.Publish("BENCH", data)
			if err != nil {
				panic(err)
			}
		}
		err := nc.Publish("BENCH", []byte("END"))
		if err != nil {
			panic(err)
		}
		<-complete
		i++
	}
	log.Println(bench.Histogram(10))
	nc.Close()
}
