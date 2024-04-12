package mqttclient

import (
	"bytes"
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

	"net/http"

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
	Type          string      `json:"type"` // tcp, ws, wss
	Broker        string      `json:"broker"`
	Port          string      `json:"port"`
	CertFile      string      `json:"certFile"`
	KeyFile       string      `json:"keyFile"`
	CaCertFile    string      `json:"caFile"`
	Username      string      `json:"username"`
	Password      string      `json:"password"`
	Topics        []MqttTopic `json:"topics"`
	AutoReconnect bool        `json:"reconnect"`
}

type MqttTopic struct {
	Topic   string `json:"topic"`
	Qos     byte   `json:"qos"`
	Handler string `json:"handler"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
}

type MqttClient struct {
	Config         Mqtt
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
	Client         mqtt.Client
	Queue          *queue.MessageQueue
	DocDBconn      *documents.DocDB
	DB             *sql.DB
	SignalRClient  signalr.Client
	monitoring     bool
	AppServer      string
}

// NewMqttClient creates a new instance of MqttClient with the given configurations.
// It initializes the MqttClient with the provided configurations and returns a pointer to the created MqttClient.
// The MqttClient is initialized with the following parameters:
// - mqttBrokertype: a string representing the type of the MQTT broker (tcp, ws, wss).
// - mqttBroker: a string representing the MQTT broker address.
// - mqttPort: a string representing the MQTT broker port.
// - certFile: a string representing the certificate file path.
// - keyFile: a string representing the key file path.
// - caCertFile: a string representing the CA certificate file path.
// - mqttClientID: a string representing the MQTT client ID.
// - mqttTopics: a slice of MqttTopic structs representing the MQTT topics to subscribe to.
// - iLog: a logger.Log struct representing the logger.
// - Client: a mqtt.Client struct representing the MQTT client.
func NewMqttClient(configurations Mqtt) *MqttClient {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MqttClient"}

	iLog.Debug(fmt.Sprintf(("Create MqttClient with configuration : %s"), logger.ConvertJson(configurations)))

	mqttclient := &MqttClient{
		Config:         configurations,
		mqttBrokertype: configurations.Type, // tcp, ws, wss
		mqttBroker:     configurations.Broker,
		mqttPort:       configurations.Port,
		certFile:       configurations.CertFile,
		keyFile:        configurations.KeyFile,
		caCertFile:     configurations.CaCertFile,
		mqttClientID:   (uuid.New()).String(),
		mqttTopics:     configurations.Topics,
		iLog:           iLog,
		monitoring:     false,
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

	//	mqttclient.Initialize_mqttClient()
	return mqttclient
}

// NewMqttClientbyExternal creates a new instance of MqttClient with the provided configurations, database connection, document database connection, and SignalR client.
// It initializes the MqttClient with the given configurations and returns the created MqttClient instance.
// The MqttClient is initialized with the following parameters:
// - mqttBrokertype: a string representing the type of the MQTT broker (tcp, ws, wss).
// - mqttBroker: a string representing the MQTT broker address.
// - mqttPort: a string representing the MQTT broker port.
// - certFile: a string representing the certificate file path.
// - keyFile: a string representing the key file path.
// - caCertFile: a string representing the CA certificate file path.
// - mqttClientID: a string representing the MQTT client ID.
// - mqttTopics: a slice of MqttTopic structs representing the MQTT topics to subscribe to.
// - iLog: a logger.Log struct representing the logger.
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
		monitoring:     false,
	}
	iLog.Debug(fmt.Sprintf(("Create MqttClient: %s"), logger.ConvertJson(mqttclient)))
	uuid := uuid.New().String()

	mqttclient.Queue = queue.NewMessageQueuebyExternal(uuid, "mqttclient", DB, DocDBconn, SignalRClient)

	mqttclient.Initialize_mqttClient()
	return mqttclient
}

// Initialize_mqttClient initializes the MQTT client by setting up the client options,
// connecting to the MQTT broker, subscribing to the desired MQTT topics,
// and starting a goroutine to handle incoming MQTT messages.
// It takes no parameters and returns nothing.

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
	mqttClient.Client = client

	// Set the message handler
	//client.SetDefaultPublishHandler(messageHandler)
	// Connect to the MQTT broker

	err := mqttClient.Connect()

	if err != nil {
		return
	}

	if !mqttClient.monitoring {
		go func() {
			mqttClient.MonitorAndReconnect()
		}()
	}

}

func (mqttClient *MqttClient) Connect() error {
	if token := mqttClient.Client.Connect(); token.Wait() && token.Error() != nil {
		mqttClient.iLog.Critical(fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error()))
		return token.Error()
	}

	mqttClient.SubscribeTopics()

	return nil
}

func (mqttClient *MqttClient) SubscribeTopics() {
	// Create a channel to receive MQTT messages
	messageChannel := make(chan mqtt.Message)

	// Define the MQTT message handler
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		messageChannel <- msg
	}

	// Subscribe to the desired MQTT topic(s)
	for _, data := range mqttClient.mqttTopics {
		//topic := data["topic"].(string)

		topic := data.Topic
		qos := data.Qos

		if token := mqttClient.Client.Subscribe(topic, qos, messageHandler); token.Wait() && token.Error() != nil {
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
				mode := ""
				executetype := ""
				for _, data := range mqttClient.mqttTopics {
					if data.Topic == msg.Topic() {
						handler = data.Handler
						mode = data.Mode
						executetype = data.Type

						mqttClient.iLog.Debug(fmt.Sprintf("topic: %s handler: %s mode: %s type: %s", data.Topic, handler, mode, executetype))

						break
					}
				}
				if executetype == "local" {
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
					mqttClient.iLog.Debug(fmt.Sprintf("Push message %v to queue: %s", message, mqttClient.Queue.QueueID))
					mqttClient.Queue.Push(message)
				} else {
					method := "POST"
					url := mqttClient.AppServer + "/trancode/execute"

					client := &http.Client{}
					req, err := http.NewRequest(method, url, bytes.NewBuffer(msg.Payload()))

					if err != nil {
						mqttClient.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
						break
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(req)
					if err != nil {
						mqttClient.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
						break
					}
					respBody, err := ioutil.ReadAll(resp.Body)

					mqttClient.iLog.Debug(fmt.Sprintf("Response data: %v", respBody))

					resp.Body.Close()

				}

			}
		}
	}()

	// Wait for termination signal to gracefully shutdown
	mqttClient.waitForTerminationSignal()
}

func (mqttClient *MqttClient) MonitorAndReconnect() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mqttClient.iLog.PerformanceWithDuration("MQTTClient.monitorAndReconnect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			mqttClient.iLog.Error(fmt.Sprintf("MQTTClient.monitorAndReconnect defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	mqttClient.iLog.Debug(fmt.Sprintf("Start the mqttclient connection monitoring for %s", mqttClient.mqttBroker))
	mqttClient.monitoring = true

	for {
		isconnected := mqttClient.Client.IsConnected()
		if !isconnected {
			mqttClient.iLog.Error(fmt.Sprintf("Mqttclient connection lost, %v %v", mqttClient.mqttBroker, mqttClient.mqttPort))

			err := mqttClient.Connect()

			if err != nil {
				mqttClient.iLog.Error("Reconnect to mqtt broker fail!")
				time.Sleep(5 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
			}

		} else {
			time.Sleep(10 * time.Second)
		}
	}

}

// Publish publishes a message to the MQTT broker with the specified topic and payload.
// It waits for the operation to complete and logs the result.
// It takes two parameters:
// - topic: a string representing the MQTT topic to publish to.
// - payload: a string representing the MQTT message payload.
// It returns nothing.

func (mqttClient *MqttClient) Publish(topic string, payload string) {

	token := mqttClient.Client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		mqttClient.iLog.Debug(fmt.Sprintf("Failed to publish message: topic %s, payload %s\n %s\n with error:", topic, payload, token.Error()))
	} else {
		mqttClient.iLog.Debug(fmt.Sprintf("Message published successfully: topic %s, payload %s\n", topic, payload))
	}
}

// loadCACert loads the CA certificate from the specified file and returns a *x509.CertPool.
// It reads the contents of the file using ioutil.ReadFile and appends the certificate to a new CertPool.
// If there is an error reading the file, it logs an error message and returns nil.
// It takes one parameter:
// - caCertFile: a string representing the CA certificate file path.
// It returns a *x509.CertPool.

func (mqttClient *MqttClient) loadCACert(caCertFile string) *x509.CertPool {
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		mqttClient.iLog.Error(fmt.Sprintf("Failed to read CA certificate: %v", err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}

// waitForTerminationSignal waits for an interrupt signal or a termination signal and performs a graceful shutdown of the MQTT client.
// It listens for the os.Interrupt and syscall.SIGTERM signals and upon receiving the signal, it disconnects the MQTT client, performs any necessary cleanup or graceful shutdown logic, and exits the program.
func (mqttClient *MqttClient) waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")
	mqttClient.Client.Disconnect(250)
	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}
