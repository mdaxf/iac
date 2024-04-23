package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac-signalr/signalr"
)

type SignalClientCheck struct {
	Ctx          context.Context
	SignalClient signalr.Client
	Error        error
}

func NewSignalClientCheck(ctx context.Context, client signalr.Client) SignalClientCheck {
	return SignalClientCheck{
		Ctx:          ctx,
		SignalClient: client,
		Error:        nil,
	}
}

func CheckSignalClientStatus(ctx context.Context, client signalr.Client) error {
	check := NewSignalClientCheck(ctx, client)
	return check.CheckStatus()
}

func (check SignalClientCheck) CheckStatus() error {
	client := check.SignalClient
	var checkErr error
	checkErr = nil
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if client.State() != signalr.ClientConnected {
				checkErr = fmt.Errorf("SignalR Client is not connected")
				check.Error = checkErr
			}
		}
	}()
	return checkErr
}
