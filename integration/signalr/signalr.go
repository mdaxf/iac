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
	"strings"

	//	"strings"
	"time"

	//kitlog "github.com/go-kit/log"
	//"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/logger"
)

type SignalRLogAdapter struct {
	log logger.Log
}

func (l *SignalRLogAdapter) Log(keyVals ...interface{}) error {
	// Format the structured log like go-kit

	// Convert keyVals to a map for easy lookup
	m := make(map[string]interface{})
	for i := 0; i < len(keyVals); i += 2 {
		if i+1 < len(keyVals) {
			k := fmt.Sprintf("%v", keyVals[i])
			m[k] = keyVals[i+1]
		}
	}

	// Extract level if present
	level := ""
	if v, ok := m["level"]; ok {
		level = strings.ToLower(fmt.Sprintf("%v", v))
	}

	// Rebuild a clean structured message
	var sb strings.Builder
	for k, v := range m {
		sb.WriteString(fmt.Sprintf("%s=%v ", k, v))
	}
	msg := strings.TrimSpace(sb.String())

	// Route to correct log category
	switch level {

	case "warn", "warning":
		l.log.Warn(msg)
	case "error":
		l.log.Error(msg)
	default:
		l.log.Info(msg)
	}

	return nil
}

// Connect establishes a connection to a SignalR server using the provided configuration.
// It returns a signalr.Client and an error if any.
// The config parameter is a map containing the server and hub information.
// The server key should be a string representing the server address.
// The hub key should be a string representing the hub name.

func Connect(config map[string]interface{}) (signalr.Client, error) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalR Server connection"}

	//fmt.Println("Connecting", com.SingalRConfig)
	ilog.Info(fmt.Sprintf("Connecting with configuration: %v", config))
	address := fmt.Sprintf("%s/%s", config["server"].(string), config["hub"].(string))

	structuredLogger := &SignalRLogAdapter{log: ilog}

	/*	c, err := signalr.NewClient(context.Background(), nil,
		signalr.WithReceiver(&IACMessageBus{}),
		signalr.WithConnector(func() (signalr.Connection, error) {
			creationCtx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			//return signalr.NewHTTPConnection(creationCtx, address)
			return signalr.NewHTTPConnection(creationCtx, address,
				signalr.WithTransports(signalr.TransportWebSockets), // ← THIS LINE!
			)

		}),
		//signalr.Logger(kitlog.NewLogfmtLogger(os.Stdout), true))
		signalr.Logger(structuredLogger, true),
	)  */

	c, err := signalr.NewClient(
		context.Background(),
		signalr.WithReceiver(&IACMessageBus{}),
		signalr.WithConnector(func() (signalr.Connection, error) {
			// Create connection context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Connect with WebSocket transport only
			return signalr.NewHTTPConnection(ctx, address,
				signalr.WithTransports(signalr.TransportWebSockets),
			)
		}),
		signalr.Logger(structuredLogger, false),
		signalr.KeepAliveInterval(15*time.Second),
		signalr.TimeoutInterval(60*time.Second),
	)

	if err != nil {
		return nil, err
	}
	c.Start()
	//fmt.Println("Connected")
	ilog.Info("Connected to the signalR server!")
	return c, nil
}

type IACMessageBus struct {
	signalr.Hub
}

var groupname = "IAC_Internal_MessageBus"

// Receive receives a message from the SignalR client and logs it.
// It takes a string parameter 'message' which represents the received message.
func (c *IACMessageBus) Receive(message string) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalR Client Receive message"}
	//	fmt.Printf("Receive message: %s \n", message)
	ilog.Debug(fmt.Sprintf("Receive message: %s \n", message))
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
