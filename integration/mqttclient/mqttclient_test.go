package mqttclient

import (
	"crypto/x509"
	"database/sql"
	"reflect"
	"testing"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac-signalr/signalr"
)

func TestNewMqttClientbyExternal(t *testing.T) {
	type args struct {
		configurations Mqtt
		DB             *sql.DB
		DocDBconn      *documents.DocDB
		SignalRClient  signalr.Client
	}
	tests := []struct {
		name string
		args args
		want *MqttClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMqttClientbyExternal(tt.args.configurations, tt.args.DB, tt.args.DocDBconn, tt.args.SignalRClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMqttClientbyExternal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMqttClient_Initialize_mqttClient(t *testing.T) {
	tests := []struct {
		name       string
		mqttClient *MqttClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mqttClient.Initialize_mqttClient()
		})
	}
}

func TestMqttClient_Publish(t *testing.T) {
	type args struct {
		topic   string
		payload string
	}
	tests := []struct {
		name       string
		mqttClient *MqttClient
		args       args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mqttClient.Publish(tt.args.topic, tt.args.payload)
		})
	}
}

func TestMqttClient_loadCACert(t *testing.T) {
	type args struct {
		caCertFile string
	}
	tests := []struct {
		name       string
		mqttClient *MqttClient
		args       args
		want       *x509.CertPool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mqttClient.loadCACert(tt.args.caCertFile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MqttClient.loadCACert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMqttClient_waitForTerminationSignal(t *testing.T) {
	tests := []struct {
		name       string
		mqttClient *MqttClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mqttClient.waitForTerminationSignal()
		})
	}
}
