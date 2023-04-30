// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nats

import (
	"log"

	//nats "github.com/nats-io/nats-server"
	natc "github.com/nats-io/nats.go"
)

var NATS_Server_Port = "4222"
var NATS_Server_Host = "localhost"
var MB_NATS_CONN *natc.Conn

func ConnectNATSServer() (*natc.Conn, error) {
	nc, err := natc.Connect(natc.DefaultURL)
	if err != nil {
		log.Println(err)
	}
	return nc, err
}

/*
	func createnatsserverinstance()(*nats.Server, error){
		    // Create a new NATS server instance
			ns, err := nats.StartServer(&nats.Options{
				Port: 4222,
			})
			if err != nil {
				log.Fatal(err)
			}
			defer ns.Shutdown()

			// Print server information
			fmt.Println("NATS server running on port 4222")

			// Wait for server to shut down
			select {
			case <-ns.Done():
				fmt.Println("NATS server shut down")
			}
		return ns, err
	}
*/
