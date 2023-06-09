package messagebus

import (
	"fmt"
	"net/http"

	"github.com/mdaxf/iac/integration/messagebus/glue"
	"github.com/mdaxf/iac/logger"
)

type MessageBus struct {
	// Set the http file server.
	Port      int
	ChannelID string
	Server    *glue.Server
	Channel   *glue.Channel
	iLog      logger.Log
}

func NewMessageBus(port int, channel string) *MessageBus {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageBus"}

	iLog.Debug(fmt.Sprintf(("Create new MessageBus with port %d and channel : %s"), port, channel))

	mb := &MessageBus{
		Port:      port,
		ChannelID: channel,
		iLog:      iLog,
	}
	mb.Server = mb.CreateServer()

	mb.iLog.Debug(fmt.Sprintf("MessageBus created: %v", mb))

	mb.Server.OnNewSocket(mb.CreateChannel)
	mb.iLog.Debug(fmt.Sprintf("MessageBus Server OnNewSocket: %v", mb.Server.OnNewSocket))

	// Run the glue server.

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../../client/dist"))))

	err := mb.Server.Run()
	if err != nil {
		mb.iLog.Error(fmt.Sprintf("Message Bus Server Run error: %v", err))
	}
	mb.iLog.Debug(fmt.Sprintf("Message Bus Server in port %d with channel%s Run", mb.Port, mb.ChannelID))

	return mb
}

func (mb *MessageBus) CreateServer() *glue.Server {

	// Create a new glue server.
	port := fmt.Sprintf(":%d", mb.Port)

	mb.iLog.Info(fmt.Sprintf("Starting MessageBus on port %s", port))

	server := glue.NewServer(glue.Options{
		HTTPListenAddress: port,
		HTTPSocketType:    glue.HTTPSocketTypeTCP,
	})
	return server
}

func (mb *MessageBus) CreateChannel(s *glue.Socket) {
	// We won't read any data from the socket itself.
	// Discard received data!
	s.DiscardRead()

	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		mb.iLog.Info(fmt.Sprintf("socket closed with remote address: %s", s.RemoteAddr()))
	})

	// Create a channel.
	mb.Channel = s.Channel(mb.ChannelID)

	/*
		// Set the channel on read event function.
		c.OnRead(func(data string) {
			// Echo the received data back to the client.
			c.Write("channel golang: " + data)
		})

		// Write to the channel.
		c.Write("Hello Gophers!") */

}
