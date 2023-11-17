package checks

import (
	"context"
	"fmt"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
)

type SignalRSrvCheck struct {
	Ctx           context.Context
	WcAddress     string
	ServerAddress string
	Error         error
}

func NewSignalRSrvCheck(ctx context.Context, address string, wcaddress string) SignalRSrvCheck {
	return SignalRSrvCheck{
		Ctx:           ctx,
		WcAddress:     wcaddress,
		ServerAddress: address,
		Error:         nil,
	}
}

func CheckSignalRSrvStatus(ctx context.Context, address string, wcaddress string) error {
	check := NewSignalRSrvCheck(ctx, address, wcaddress)
	return check.CheckStatus()
}

func CheckSignalRSrvHttpStatus(ctx context.Context, address string, wcaddress string) error {
	check := NewSignalRSrvCheck(ctx, address, wcaddress)
	return check.CheckhttpStatus()
}

func (check SignalRSrvCheck) CheckStatus() error {

	var checkErr error
	checkErr = nil
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {

			conn, _, err := websocket.DefaultDialer.Dial(check.WcAddress, nil)
			if err != nil {
				checkErr = fmt.Errorf("websocket connection failed: %w", err)
				check.Error = checkErr
			}
			defer conn.Close()
		}
	}()
	return checkErr
}

func (check SignalRSrvCheck) CheckhttpStatus() error {

	var checkErr error
	checkErr = nil
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			resp, err := http.Get(check.ServerAddress)
			if err != nil {
				checkErr = fmt.Errorf("http connection failed: %w", err)
				check.Error = checkErr
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				checkErr = nil
			} else {
				checkErr = fmt.Errorf("SignalR server returned a non-OK status:", resp.Status)
				check.Error = checkErr
			}
		}
	}()
	return checkErr
}
