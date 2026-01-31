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

package signalrserver

import (
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/logger"
)

// IACMessageBus is the SignalR hub for IAC message bus functionality
type IACMessageBus struct {
	signalr.Hub
	ilog logger.Log
}

// NewIACMessageBus creates a new IACMessageBus hub instance
func NewIACMessageBus() *IACMessageBus {
	return &IACMessageBus{
		ilog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalRHub"},
	}
}

var groupname = "IAC_Internal_MessageBus"

// Subscribe subscribes a client to a topic
func (c *IACMessageBus) Subscribe(topic string, connectionID string) {
	c.ilog.Debug(fmt.Sprintf("Subscribe: topic: %s, sender: %s\n", topic, connectionID))
}

// Send sends a message to all clients in the group
func (c *IACMessageBus) Send(topic string, message string, connectionID string) {
	c.ilog.Debug(fmt.Sprintf("Send: topic: %s, message: %s, sender: %s\n", topic, message, connectionID))
	c.Clients().Group(groupname).Send(topic, message)
}

// SendToBackEnd sends a message to backend clients
func (c *IACMessageBus) SendToBackEnd(topic string, message string, connectionID string) {
	c.ilog.Debug(fmt.Sprintf("SendToBackEnd: topic: %s, message: %s, sender: %s\n", topic, message, connectionID))
	JsonMsg := make(map[string]interface{})
	JsonMsg["topic"] = topic
	JsonMsg["message"] = message
	JsonMsg["sender"] = connectionID

	c.ilog.Debug(fmt.Sprintf("SendToBackEnd: JsonMsg: %s\n", JsonMsg))
	c.Clients().Group(groupname).Send("sendtobackend", JsonMsg)
}

// AddMessage adds a message to the specified topic
func (c *IACMessageBus) AddMessage(message string, topic string, sender string) {
	c.ilog.Debug(fmt.Sprintf("AddMessage: topic: %s, message: %s, sender: %s\n", topic, message, sender))
	c.Clients().Group(groupname).Send(topic, message)
}

// OnConnected is called when a client connects
func (c *IACMessageBus) OnConnected(connectionID string) {
	c.ilog.Info(fmt.Sprintf("Client %s connected and joining group %s", connectionID, groupname))
	c.Groups().AddToGroup(groupname, connectionID)
}

// OnDisconnected is called when a client disconnects
func (c *IACMessageBus) OnDisconnected(connectionID string) {
	c.ilog.Info(fmt.Sprintf("Client %s disconnected from group %s", connectionID, groupname))
	c.Groups().RemoveFromGroup(groupname, connectionID)
}

// Broadcast sends a message to all clients
func (c *IACMessageBus) Broadcast(message string) {
	c.ilog.Debug(fmt.Sprintf("broadcast message: %s\n", message))
	c.Clients().Group(groupname).Send("broadcast", message)
	c.Clients().Group(groupname).Send("receive", message)
}

// Echo sends a message back to the caller
func (c *IACMessageBus) Echo(message string) {
	c.Clients().Caller().Send("echo", message)
}

// Panic is a test method that triggers a panic
func (c *IACMessageBus) Panic() {
	panic("Don't panic!")
}

// RequestAsync processes a message asynchronously
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

// RequestTuple returns multiple values for a message
func (c *IACMessageBus) RequestTuple(message string) (string, string, int) {
	return strings.ToUpper(message), strings.ToLower(message), len(message)
}

// DateStream streams date/time information
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

// UploadStream handles streaming uploads
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

// Abort aborts the current connection
func (c *IACMessageBus) Abort() {
	c.ilog.Info("Abort called")
	c.Hub.Abort()
}
