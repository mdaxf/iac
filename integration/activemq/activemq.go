package activemq

import (
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-stomp/stomp"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/com"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/queue"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type ActiveMQconfigs struct {
	ActiveMQs []ActiveMQconfig `json:"activemqs"`
	ApiKey    string           `json:"apikey"`
}

// ActiveMQ struct
type ActiveMQconfig struct {
	Host      string          `json:"host"`
	Port      string          `json:"port"`
	Username  string          `json:"username"`
	Password  string          `json:"passwrod"`
	TLS       string          `json:"tls"`
	TLSVerify bool            `json:"tlsverify"`
	CAPath    string          `json:"CAPath"`
	CertPath  string          `json:"CertPath"`
	KeyPath   string          `json:"keypath"`
	Topics    []ActiveMQtopic `json:"topics"`
}

type ActiveMQtopic struct {
	Topic    string `json:"topic"`
	Handler  string `json:"handler"`
	SQLQuery string `json:"sqlquery"`
	Mode     string `json:"mode"`
	Type     string `json:"type"`
}

// ActiveMQ struct
type ActiveMQ struct {
	Config        ActiveMQconfig
	Conn          *stomp.Conn
	Subs          []*stomp.Subscription
	ConnectionID  string
	QueueID       string
	iLog          logger.Log
	Queue         *queue.MessageQueue
	DocDBconn     *documents.DocDB
	DB            *sql.DB
	SignalRClient signalr.Client
	AppServer     string
	ApiKey        string
}

/*
	type msghandler struct {
		Topic   string
		Handler string
		Message stomp.Message
	}
*/
func NewActiveMQConnection(config ActiveMQconfig) *ActiveMQ {

	activeMQ := connectActiveMQ(config)

	if activeMQ == nil {
		iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ActiveMQConnection"}

		iLog.Critical(fmt.Sprintf(("Fail to create activeMQ connection with configuration : %s"), logger.ConvertJson(config)))

		return nil
	}
	activeMQ.iLog.Debug(fmt.Sprintf("Create ActiveMQ connection successful!"))

	uuid_ := uuid.New().String()

	activeMQ.DocDBconn = documents.DocDBCon
	activeMQ.DB = dbconn.DB
	activeMQ.SignalRClient = com.IACMessageBusClient
	activeMQ.Queue = queue.NewMessageQueue(uuid_, "ActiveMQ")
	activeMQ.Queue.DocDBconn = documents.DocDBCon
	activeMQ.Queue.DB = dbconn.DB
	activeMQ.Queue.SignalRClient = com.IACMessageBusClient

	activeMQ.Subscribes()

	return activeMQ
}

func NewActiveMQConnectionExternal(config ActiveMQconfig, docDBconn *documents.DocDB, db *sql.DB, signalRClient signalr.Client) *ActiveMQ {

	activeMQ := connectActiveMQ(config)

	if activeMQ == nil {
		iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ActiveMQConnection"}

		iLog.Critical(fmt.Sprintf(("Fail to create activeMQ connection with configuration : %s"), logger.ConvertJson(config)))

		return nil
	}
	activeMQ.iLog.Debug(fmt.Sprintf("Create ActiveMQ connection successful!"))

	uuid_ := uuid.New().String()
	activeMQ.Queue = queue.NewMessageQueue(uuid_, "ActiveMQ")
	activeMQ.Queue.DocDBconn = docDBconn
	activeMQ.Queue.DB = db
	activeMQ.Queue.SignalRClient = signalRClient

	activeMQ.DocDBconn = docDBconn
	activeMQ.DB = db
	activeMQ.SignalRClient = signalRClient

	//	activeMQ.Subscribes()

	return activeMQ

}
func CheckConnection(activemq ActiveMQ) bool {
	if activemq.Conn == nil {
		return false
	}

	return true
}
func connectActiveMQ(config ActiveMQconfig) *ActiveMQ {

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "ActiveMQConnection"}

	iLog.Debug(fmt.Sprintf(("Create activeMQ connection with configuration : %s"), logger.ConvertJson(config)))

	if config.Host == "" {
		iLog.Error("Host is required")
		return nil
	}

	if config.Port == "" {
		iLog.Error("Port is required")
		return nil
	}

	var conn *stomp.Conn
	var err error

	if config.Username == "" && config.TLS == "" {
		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.AcceptVersion(stomp.V12),
		)
	} else if config.Username != "" && config.TLS == "" {
		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.Login(config.Username, config.Password),
		)
	} else if config.Username == "" && config.TLS != "" && config.TLSVerify == false {
		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.AcceptVersion(stomp.V12),
		//	stomp.ConnOpt.TLSConfig(&tls.Config{InsecureSkipVerify: true}), // Set to true for testing purposes only
		)
	} else if config.Username != "" && config.TLS != "" && config.TLSVerify == false {
		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.Login(config.Username, config.Password),
			stomp.ConnOpt.AcceptVersion(stomp.V12),
		//	stomp.ConnOpt.TLSConfig(&tls.Config{InsecureSkipVerify: true}), // Set to true for testing purposes only
		)
	} else if config.Username == "" && config.TLSVerify == true && config.CAPath != "" && config.CertPath != "" {
		//	cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
		if err != nil {
			iLog.Critical(fmt.Sprintf("Failed to load client certificates: %v", err))
			return nil
		}

		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.AcceptVersion(stomp.V12),
		/*	stomp.ConnOpt.TLSConfig(&tls.Config{
			InsecureSkipVerify: false, // Set to true for testing purposes only
			RootCAs:            loadCACert(config.CAPath, iLog),
			Certificates:       []tls.Certificate{cert},
		}), */
		)
	} else if config.Username != "" && config.Password != "" && config.TLSVerify == true && config.CAPath != "" && config.CertPath != "" {
		//	cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
		if err != nil {
			iLog.Critical(fmt.Sprintf("Failed to load client certificates: %v", err))
			return nil
		}

		conn, err = stomp.Dial(
			"tcp",
			config.Host+":"+config.Port,
			stomp.ConnOpt.Login(config.Username, config.Password),
			stomp.ConnOpt.AcceptVersion(stomp.V12),
		/*	stomp.ConnOpt.TLSConfig(&tls.Config{
			InsecureSkipVerify: false, // Set to true for testing purposes only
			RootCAs:            loadCACert(config.CAPath, iLog),
			Certificates:       []tls.Certificate{cert},
		}),  */
		)
	} else {
		iLog.Error("Invalid configuration")
		return nil
	}

	if err != nil {
		iLog.Error(fmt.Sprintf("Error while connecting to ActiveMQ: %s", err.Error()))
		return nil
	}
	activeMQ := &ActiveMQ{
		Config: config,
		Conn:   conn,
		iLog:   iLog,
	}

	return activeMQ

}

func loadCACert(caCertFile string, iLog logger.Log) *x509.CertPool {
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read CA certificate: %v", err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}

func (a *ActiveMQ) Subscribes() {
	/*
		messageChannel := make(chan msghandler)

		messageHandler := func(msg stomp.Message) {
			messageChannel <- msg
		}  */

	for _, item := range a.Config.Topics {
		topic := item.Topic
		handler := item.Handler
		//	mode := item.Mode
		executetype := item.Type

		sub, err := a.Conn.Subscribe(topic, stomp.AckAuto)
		if err != nil {
			a.iLog.Error(fmt.Sprintf("Error while subscribing to topic %s: %s", topic, err.Error()))
		}
		a.Subs = append(a.Subs, sub)
		go func() {
			if executetype == "local" {
				msgID := 0
				for {
					msg := <-sub.C
					msgID += 1
					fmt.Println("Received message", string(msg.Body))
					message := queue.Message{
						Id:        strconv.FormatUint(uint64(msgID), 10),
						UUID:      uuid.New().String(),
						Retry:     3,
						Execute:   0,
						Topic:     topic,
						PayLoad:   msg.Body,
						Handler:   handler,
						CreatedOn: time.Now().UTC(),
					}

					a.iLog.Debug(fmt.Sprintf("Push message %v to queue: %s", message, a.Queue.QueueID))
					a.Queue.Push(message)
				}
			} else {
				for {
					msg := <-sub.C
					fmt.Println("Received message", string(msg.Body))
					a.CallWebService(msg, topic, handler)
				}
			}
		}()
	}

	a.waitForTerminationSignal()

}

func (a *ActiveMQ) CallWebService(msg *stomp.Message, topic string, handler string) {

	method := "POST"
	url := a.AppServer + "/trancode/execute"

	var result map[string]interface{}
	err := json.Unmarshal(msg.Body, &result)
	if err != nil {
		a.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}
	var inputs map[string]interface{}

	inputs["Payload"] = result
	inputs["Topic"] = topic

	msgdata := make(map[string]interface{})
	msgdata["TranCode"] = handler
	msgdata["Inputs"] = inputs

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "apikey " + a.ApiKey

	result, err = com.CallWebService(url, method, msgdata, headers)

	if err != nil {
		a.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		return
	}

	a.iLog.Debug(fmt.Sprintf("Response data: %v", result))

}

func (a *ActiveMQ) waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	for _, sub := range a.Subs {
		sub.Unsubscribe()
	}

	a.Conn.Disconnect()
	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}
