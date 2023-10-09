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

package signalr

import (
	"context"
	"fmt"

	//	"strings"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/signalrsrv/signalr"
)

func Connect() (signalr.Client, error) {
	fmt.Println("Connecting", com.SingalRConfig)
	address := fmt.Sprintf("%s/%s", com.SingalRConfig["server"].(string), com.SingalRConfig["hub"].(string))

	c, err := signalr.NewClient(context.Background(), nil,
		signalr.WithReceiver(&IACMessageBus{}),
		signalr.WithConnector(func() (signalr.Connection, error) {
			creationCtx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			return signalr.NewHTTPConnection(creationCtx, address)
		}))
	if err != nil {
		return nil, err
	}
	c.Start()
	fmt.Println("Connected")

	return c, nil
}

type IACMessageBus struct {
	signalr.Hub
}

var groupname = "IAC_Internal_MessageBus"

func (c *IACMessageBus) Receive(message string) {
	fmt.Printf("Receive message: %s \n", message)
}

/*
func (c *IACMessageBus) Subscribe(topic string, connectionID string) {
	fmt.Printf("Subscribe: topic: %s, sender: %s\n", topic, connectionID)
}


func (c *IACMessageBus) AddMessage(message string, topic string, sender string) {
	fmt.Printf("AddMessage: topic: %s, message: %s, sender: %s\n", topic, message, sender)
	c.Clients().Group(groupname).Send(topic, message)
}

// add the client to the connection
func (c *IACMessageBus) OnConnected(connectionID string) {
	fmt.Printf("%s connected\n", connectionID)
	c.Groups().AddToGroup(groupname, connectionID)
	fmt.Printf("%s connected and added to group %s\n", connectionID, groupname)
}

func (c *IACMessageBus) OnDisconnected(connectionID string) {
	fmt.Printf("%s disconnected\n", connectionID)
	c.Groups().RemoveFromGroup(groupname, connectionID)
	fmt.Printf("%s disconnected and removed from group %s\n", connectionID, groupname)
}

func (c *IACMessageBus) Broadcast(message string) {
	// Broadcast to all clients
	fmt.Printf("broadcast message: %s\n", message)
	c.Clients().Group(groupname).Send("broadcast", message)
	//	c.Clients().Group(groupname).Send("receive", message)
}

func (c *IACMessageBus) Echo(message string) {
	c.Clients().Caller().Send("echo", message)
	//	c.Clients().Caller().Send("receive", message)
}

func (c *IACMessageBus) Panic() {
	panic("Don't panic!")
}

func (c *IACMessageBus) RequestAsync(message string) <-chan map[string]string {
	r := make(chan map[string]string)
	go func() {
		defer close(r)
		time.Sleep(4 * time.Second)
		m := make(map[string]string)
		m["ToUpper"] = strings.ToUpper(message)
		m["ToLower"] = strings.ToLower(message)
		m["len"] = fmt.Sprint(len(message))
		r <- m
	}()
	return r
}

func (c *IACMessageBus) RequestTuple(message string) (string, string, int) {
	return strings.ToUpper(message), strings.ToLower(message), len(message)
}

func (c *IACMessageBus) DateStream() <-chan string {
	r := make(chan string)
	go func() {
		defer close(r)
		for i := 0; i < 50; i++ {
			r <- fmt.Sprint(time.Now().Clock())
			time.Sleep(time.Second)
		}
	}()
	return r
}

func (c *IACMessageBus) UploadStream(upload1 <-chan int, factor float64, upload2 <-chan float64) {
	ok1 := true
	ok2 := true
	u1 := 0
	u2 := 0.0
	c.Echo(fmt.Sprintf("f: %v", factor))
	for {
		select {
		case u1, ok1 = <-upload1:
			if ok1 {
				c.Echo(fmt.Sprintf("u1: %v", u1))
			} else if !ok2 {
				c.Echo("Finished")
				return
			}
		case u2, ok2 = <-upload2:
			if ok2 {
				c.Echo(fmt.Sprintf("u2: %v", u2))
			} else if !ok1 {
				c.Echo("Finished")
				return
			}
		}
	}
}

func (c *IACMessageBus) Abort() {
	fmt.Println("Abort")
	c.Hub.Abort()
}
*/
