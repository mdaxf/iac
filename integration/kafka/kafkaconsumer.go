package kafka

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	//	cluster "github.com/bsm/sarama-cluster"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/queue"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type KafkasConfig struct {
	Kafkas []KafkaConfig `json:"kafkas"`
	ApiKey string        `json:"apikey"`
}

type KafkaConfig struct {
	Server string       `json:"server"`
	Topics []KafkaTopic `json:"topics"`
}

type KafkaTopic struct {
	Topic   string `json:"topic"`
	Handler string `json:"handler"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
}

type KafkaConsumer struct {
	Config        KafkaConfig
	Queue         *queue.MessageQueue
	iLog          logger.Log
	Consumer      sarama.Consumer
	DocDBconn     *documents.DocDB
	DB            *sql.DB
	SignalRClient signalr.Client
	AppServer     string
	ApiKey        string
}

func NewKafkaConsumer(config KafkaConfig) *KafkaConsumer {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "KafkaConsumer"}

	iLog.Debug(fmt.Sprintf(("Create Kafkaconsumer with configuration : %s"), logger.ConvertJson(config)))

	uuid := uuid.New().String()
	q := queue.NewMessageQueue(uuid, "Kafkaconsumer")

	Kafkaconsumer := &KafkaConsumer{
		Config: config,
		Queue:  q,
		iLog:   iLog,
	}

	iLog.Debug(fmt.Sprintf(("Create Kafkaconsumer: %s"), logger.ConvertJson(Kafkaconsumer)))
	Kafkaconsumer.BuildKafkaConsumer()
	return Kafkaconsumer
}

func NewKafkaConsumerExternal(config KafkaConfig, docDBconn *documents.DocDB, db *sql.DB, signalRClient signalr.Client) *KafkaConsumer {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "KafkaConsumer"}

	iLog.Debug(fmt.Sprintf(("Create Kafkaconsumer with configuration : %s"), logger.ConvertJson(config)))
	uuid := uuid.New().String()
	q := queue.NewMessageQueue(uuid, "Kafkaconsumer")

	Kafkaconsumer := &KafkaConsumer{
		Config: config,
		Queue:  q,
		iLog:   iLog,
	}

	Kafkaconsumer.Queue.DocDBconn = docDBconn
	Kafkaconsumer.Queue.DB = db
	Kafkaconsumer.Queue.SignalRClient = signalRClient

	iLog.Debug(fmt.Sprintf(("Create Kafkaconsumer: %s"), logger.ConvertJson(Kafkaconsumer)))
	Kafkaconsumer.BuildKafkaConsumer()
	return Kafkaconsumer
}

func (KafkaConsumer *KafkaConsumer) BuildKafkaConsumer() {

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	consumer, err := sarama.NewConsumer([]string{KafkaConsumer.Config.Server}, config)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error creating consumer: %v", err))
		return
	}

	KafkaConsumer.Consumer = consumer

	KafkaConsumer.PartitionTopics()

}

/*
func (KafkaConsumer *KafkaConsumer) ClusterGroup() {

		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true
		config.Consumer.Offsets.AutoCommit.Enable = true
		config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

		group := uuid.New().String()
		topics := []string{}
		for _, data := range KafkaConsumer.Config.Topics {
			topics.append(data.Topic)
		}

		consumerGroup, err := cluster.NewConsumer(
			[]string{KafkaConsumer.Config.Server},
			group,
			topics,
			config,
		)

		KafkaConsumer.Consumer = consumerGroup

		if err != nil {
			KafkaConsumer.iLog.Error(fmt.Sprintf("Failed to create consumer group: %v", err))
		}
		defer consumerGroup.Close()

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT)

		q := KafkaConsumer.Queue
		go func() {
			for message := range consumerGroup.Messages() {
				KafkaConsumer.iLog.Debug(fmt.Sprintf("Received message: %s", message.Value))
				for _, data := range KafkaConsumer.Config.Topics {
					if message.Topic == data.Topic {
						handler = data.Handler
						if Handler != "" {
							ID := uuid.New().String()
							msg := queue.Message{
								Id:        ID,
								UUID:      ID,
								Retry:     3,
								Execute:   0,
								Topic:     message.Topic,
								PayLoad:   []byte(message.Value),
								Handler:   handler,
								CreatedOn: time.Now(),
							}
							iLog.Debug(fmt.Sprintf("Push message %s to queue: %s", msg, q.QueueID))
							q.Push(msg)
						}
						break
					}
				}

				consumerGroup.MarkMessage(message, "") // Mark the message as processed
			}
		}()
		KafkaConsumer.waitForTerminationSignal()
	}
*/
func (KafkaConsumer *KafkaConsumer) PartitionTopics() {

	for _, data := range KafkaConsumer.Config.Topics {
		topic := data.Topic
		handler := data.Handler
		KafkaConsumer.initKafkaConsumerbyTopic(topic, handler, data)
	}
}

func (KafkaConsumer *KafkaConsumer) initKafkaConsumerbyTopic(topic string, handler string, data KafkaTopic) {

	consumer := KafkaConsumer.Consumer
	iLog := KafkaConsumer.iLog
	q := KafkaConsumer.Queue

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating partition consumer: %v", err))
		return
	}
	defer partitionConsumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
	ConsumerLoop:
		for {
			select {
			case <-signals:
				break ConsumerLoop
			case err := <-partitionConsumer.Errors():
				iLog.Error(fmt.Sprintf("Error consuming message: %v", err))

			case message := <-partitionConsumer.Messages():
				iLog.Info(fmt.Sprintf("Consumed message offset %d: %s", message.Offset, string(message.Value)))
				if data.Type == "local" {
					ID := uuid.New().String()
					msg := queue.Message{
						Id:        ID,
						UUID:      ID,
						Retry:     3,
						Execute:   0,
						Topic:     topic,
						PayLoad:   []byte(message.Value),
						Handler:   handler,
						CreatedOn: time.Now(),
					}
					iLog.Debug(fmt.Sprintf("Push message %s to queue: %s", msg, q.QueueID))
					q.Push(msg)
				} else {
					iLog.Debug(fmt.Sprintf("Call IAC Endpoint to handle the message %s with: %s", message.Value, handler))
					KafkaConsumer.CallWebService(message, topic, handler)
				}
			}
		}
	}()

	KafkaConsumer.waitForTerminationSignal()
}
func (KafkaConsumer *KafkaConsumer) CallWebService(msg *sarama.ConsumerMessage, topic string, handler string) {

	method := "POST"
	url := KafkaConsumer.AppServer + "/trancode/execute"

	client := &http.Client{}

	type MSGData struct {
		TranCode string                 `json:"code"`
		Inputs   map[string]interface{} `json:"inputs"`
	}

	var result map[string]interface{}
	err := json.Unmarshal(msg.Value, &result)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}
	var inputs map[string]interface{}

	inputs["Payload"] = result
	inputs["Topic"] = topic

	msgdata := &MSGData{
		TranCode: handler,
		Inputs:   inputs,
	}

	bytesdata, err := json.Marshal(msgdata)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bytesdata))

	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+KafkaConsumer.ApiKey)

	resp, err := client.Do(req)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}
	KafkaConsumer.iLog.Debug(fmt.Sprintf("Response data: %v", result))

}
func (KafkaConsumer *KafkaConsumer) waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	KafkaConsumer.Consumer.Close()

	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}
