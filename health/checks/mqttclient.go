package checks

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClientCheck struct {
	Ctx        context.Context
	MqttClient mqtt.Client
	Error      error
}

func NewMqttClientCheck(ctx context.Context, mqttClient mqtt.Client) MqttClientCheck {
	return MqttClientCheck{
		Ctx:        ctx,
		MqttClient: mqttClient,
		Error:      nil,
	}
}

func CheckMqttClientStatus(ctx context.Context, mqttClient mqtt.Client) error {
	check := NewMqttClientCheck(ctx, mqttClient)
	return check.CheckStatus()
}

func (check MqttClientCheck) CheckStatus() error {
	client := check.MqttClient
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
			if client.IsConnected() {
				checkErr = fmt.Errorf("MQTT Client is not connected")
				check.Error = checkErr
			}
		}
	}()
	return checkErr
}
