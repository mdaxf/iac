package messagebus

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
	"github.com/mdaxf/iac/logger"
)

/*
func Initialize(port int, channel string) {

	IACMB = mb.NewMessageBus(port, channel)

}
*/

func Initialize() *socketio.Server {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageBus"}
	iLog.Debug("Create new MessageBus Server ")
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		iLog.Debug(fmt.Sprintf("connected:", s.ID()))
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		iLog.Debug(fmt.Sprintf("notice:", msg))
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		iLog.Debug(fmt.Sprintf("receive chat msg:", msg))
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		iLog.Debug(fmt.Sprintf("bye:", s.ID()))
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		iLog.Error(fmt.Sprintf("meet error:", e))

	})

	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		iLog.Debug(fmt.Sprintf("closed:", msg))

	})

	if err := server.Serve(); err != nil {
		iLog.Error(fmt.Sprintf("socketio listen error: %s\n", err))
	}

	//defer server.Close()

	return server
}
