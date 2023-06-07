package messagebus

import (
	mb "github.com/mdaxf/iac/integration/messagebus"
)

var IACMB *mb.MessageBus

func Initialize(port int, channel string) {

	IACMB = mb.NewMessageBus(port, channel)

}
