package kafka

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/queue"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type KafkasConfig struct {
	Kafkas []KafkaConfig `json:"kafkas"`
}

type KafkaConfig struct {
	Server string       `json:"server"`
	Topics []KafkaTopic `json:"topics"`
}

type KafkaTopic struct {
	Topic   string `json:"topic"`
	Handler string `json:"handler"`
}

type KafkaConsumer struct {
	Config   KafkaConfig
	Queue    *queue.MessageQueue
	iLog     logger.Log
	Consumer sarama.Consumer
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

func NewKafkaConsumerExternal(config KafkaConfig, q *queue.MessageQueue, docDBconn *documents.DocDB, db *sql.DB, signalRClient signalr.Client) *KafkaConsumer {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "KafkaConsumer"}

	iLog.Debug(fmt.Sprintf(("Create Kafkaconsumer with configuration : %s"), logger.ConvertJson(config)))

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

	consumer, err := sarama.NewConsumer([]string{KafkaConsumer.Config.Server}, config)
	if err != nil {
		KafkaConsumer.iLog.Error(fmt.Sprintf("Error creating consumer: %v", err))
		return
	}

	KafkaConsumer.Consumer = consumer

	for _, data := range KafkaConsumer.Config.Topics {
		topic := data.Topic
		handler := data.Handler
		KafkaConsumer.initKafkaConsumerbyTopic(consumer, topic, handler, KafkaConsumer.iLog, KafkaConsumer.Queue)
	}
}

func (KafkaConsumer *KafkaConsumer) initKafkaConsumerbyTopic(consumer sarama.Consumer, topic string, handler string, iLog logger.Log, q *queue.MessageQueue) {

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
			}
		}
	}()

	KafkaConsumer.waitForTerminationSignal()
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
