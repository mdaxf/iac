package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/com"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/queue"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type MqttConfig struct {
	Mqtts []Mqtt `json:"mqtts"`
}

type Mqtt struct {
	Type       string      `json:"type"` // tcp, ws, wss
	Broker     string      `json:"broker"`
	Port       string      `json:"port"`
	CertFile   string      `json:"certFile"`
	KeyFile    string      `json:"keyFile"`
	CaCertFile string      `json:"caFile"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	Topics     []MqttTopic `json:"topics"`
}

type MqttTopic struct {
	Topic   string `json:"topic"`
	Qos     byte   `json:"qos"`
	Handler string `json:"handler"`
}

type MqttClient struct {
	mqttBrokertype string
	mqttBroker     string
	mqttPort       string
	certFile       string
	keyFile        string
	caCertFile     string
	username       string
	password       string
	mqttClientID   string
	mqttTopics     []MqttTopic
	iLog           logger.Log
	client         mqtt.Client
	Queue          *queue.MessageQueue
	DocDBconn      *documents.DocDB
	DB             *sql.DB
	SignalRClient  signalr.Client
}

func NewMqttClient(configurations Mqtt) *MqttClient {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MqttClient"}

	iLog.Debug(fmt.Sprintf(("Create MqttClient with configuration : %s"), logger.ConvertJson(configurations)))

	mqttclient := &MqttClient{
		mqttBrokertype: configurations.Type, // tcp, ws, wss
		mqttBroker:     configurations.Broker,
		mqttPort:       configurations.Port,
		certFile:       configurations.CertFile,
		keyFile:        configurations.KeyFile,
		caCertFile:     configurations.CaCertFile,
		mqttClientID:   (uuid.New()).String(),
		mqttTopics:     configurations.Topics,
		iLog:           iLog,
	}
	iLog.Debug(fmt.Sprintf(("Create MqttClient: %s"), logger.ConvertJson(mqttclient)))
	uuid := uuid.New().String()

	mqttclient.DocDBconn = documents.DocDBCon
	mqttclient.DB = dbconn.DB
	mqttclient.SignalRClient = com.IACMessageBusClient
	mqttclient.Queue = queue.NewMessageQueue(uuid, "mqttclient")
	mqttclient.Queue.DocDBconn = documents.DocDBCon
	mqttclient.Queue.DB = dbconn.DB
	mqttclient.Queue.SignalRClient = com.IACMessageBusClient

	mqttclient.Initialize_mqttClient()
	return mqttclient
}

func NewMqttClientbyExternal(configurations Mqtt, DB *sql.DB, DocDBconn *documents.DocDB, SignalRClient signalr.Client) *MqttClient {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MqttClient"}

	iLog.Debug(fmt.Sprintf(("Create MqttClient with configuration : %s"), logger.ConvertJson(configurations)))

	mqttclient := &MqttClient{
		mqttBrokertype: configurations.Type, // tcp, ws, wss
		mqttBroker:     configurations.Broker,
		mqttPort:       configurations.Port,
		certFile:       configurations.CertFile,
		keyFile:        configurations.KeyFile,
		caCertFile:     configurations.CaCertFile,
		mqttClientID:   (uuid.New()).String(),
		mqttTopics:     configurations.Topics,
		iLog:           iLog,
		DocDBconn:      DocDBconn,
		DB:             DB,
		SignalRClient:  SignalRClient,
	}
	iLog.Debug(fmt.Sprintf(("Create MqttClient: %s"), logger.ConvertJson(mqttclient)))
	uuid := uuid.New().String()

	mqttclient.Queue = queue.NewMessageQueuebyExternal(uuid, "mqttclient", DB, DocDBconn, SignalRClient)

	mqttclient.Initialize_mqttClient()
	return mqttclient
}

func (mqttClient *MqttClient) Initialize_mqttClient() {
	// Create an MQTT client options
	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetClientID(mqttClient.mqttClientID)
	opts.SetUsername(mqttClient.username)
	opts.SetPassword(mqttClient.password)
	mqttClient.iLog.Debug(fmt.Sprintf("%s://%s:%s", mqttClient.mqttBrokertype, mqttClient.mqttBroker, mqttClient.mqttPort))

	if mqttClient.mqttBrokertype == "" {
		opts.AddBroker(fmt.Sprintf("%s:%s", mqttClient.mqttBroker, mqttClient.mqttPort)) // Replace with your MQTT broker address
	} else {
		opts.AddBroker(fmt.Sprintf("%s://%s:%s", mqttClient.mqttBrokertype, mqttClient.mqttBroker, mqttClient.mqttPort)) // Replace with your MQTT broker address
	}

	if mqttClient.mqttBrokertype == "ssl" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(mqttClient.certFile, mqttClient.keyFile)
		if err != nil {
			mqttClient.iLog.Critical(fmt.Sprintf("Failed to load client certificates: %v", err))

		}
		opts.SetTLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      mqttClient.loadCACert(mqttClient.caCertFile),
		})
	}
	// Create an MQTT client
	client := mqtt.NewClient(opts)
	mqttClient.client = client
	// Create a channel to receive MQTT messages
	messageChannel := make(chan mqtt.Message)

	// Define the MQTT message handler
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		messageChannel <- msg
	}

	// Set the message handler
	//client.SetDefaultPublishHandler(messageHandler)

	// Connect to the MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		mqttClient.iLog.Critical(fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error()))

	}

	// Subscribe to the desired MQTT topic(s)
	for _, data := range mqttClient.mqttTopics {
		//topic := data["topic"].(string)

		topic := data.Topic
		qos := data.Qos
		if token := client.Subscribe(topic, qos, messageHandler); token.Wait() && token.Error() != nil {
			mqttClient.iLog.Error(fmt.Sprintf("Failed to subscribe to MQTT topic: %v", token.Error()))
		}
		mqttClient.iLog.Debug(fmt.Sprintf("Subscribed to topic: %s", topic))
	}

	// Start a goroutine to handle MQTT messages
	go func() {
		for {
			select {
			case msg := <-messageChannel:
				// Process the received MQTT message
				mqttClient.iLog.Debug(fmt.Sprintf("Received message: %s from topic: %s , %d ,%d", msg.Payload(), msg.Topic(), msg.MessageID(), msg.Qos()))
				handler := ""
				for _, data := range mqttClient.mqttTopics {
					if data.Topic == msg.Topic() {
						handler = data.Handler
						break
					}
				}
				/*			data := map[string]interface{}{"Topic": msg.Topic(), "Payload": msg.Payload()}
							jsonData, err := json.Marshal(data)
							if err != nil {
								mqttClient.iLog.Error(fmt.Sprintf("Failed to marshal data: %v", err))
							}
							rawMessage := json.RawMessage(jsonData)  */
				/*var jsonData interface{}
				err := json.Unmarshal(msg.Payload(), &jsonData)
				if err != nil {
					mqttClient.iLog.Error(fmt.Sprintf("Failed to unmarshal data: %v", err))
					return
				} */
				message := queue.Message{
					Id:        strconv.FormatUint(uint64(msg.MessageID()), 10),
					UUID:      uuid.New().String(),
					Retry:     3,
					Execute:   0,
					Topic:     msg.Topic(),
					PayLoad:   msg.Payload(),
					Handler:   handler,
					CreatedOn: time.Now().UTC(),
				}
				mqttClient.iLog.Debug(fmt.Sprintf("Push message %s to queue: %s", message, mqttClient.Queue.QueueID))
				mqttClient.Queue.Push(message)

			}
		}
	}()

	// Wait for termination signal to gracefully shutdown
	mqttClient.waitForTerminationSignal()
}
func (mqttClient *MqttClient) Publish(topic string, payload string) {

	token := mqttClient.client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		mqttClient.iLog.Debug(fmt.Sprintf("Failed to publish message: topic %s, payload %s\n %s\n with error:", topic, payload, token.Error()))
	} else {
		mqttClient.iLog.Debug(fmt.Sprintf("Message published successfully: topic %s, payload %s\n", topic, payload))
	}
}

func (mqttClient *MqttClient) loadCACert(caCertFile string) *x509.CertPool {
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		mqttClient.iLog.Error(fmt.Sprintf("Failed to read CA certificate: %v", err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}

func (mqttClient *MqttClient) waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")
	mqttClient.client.Disconnect(250)
	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}
