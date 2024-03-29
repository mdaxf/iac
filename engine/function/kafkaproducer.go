package funcs

import (
	//	"context"
	"encoding/json"
	"fmt"
	"time"

	"os"
	"os/signal"

	"github.com/IBM/sarama"
)

type SendMessagebyKafka struct {
}

func (cf *SendMessagebyKafka) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendMessagebyKafka.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendMessagebyKafka.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.SendMessagebyKafka.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendMessagebyKafka.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("SendMessagebyKafka Execute: %v", f))

	namelist, valuelist, _ := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("SendMessagebyKafka Execute: %v, %v", namelist, valuelist))
	kafkaServer := ""
	Topic := ""
	data := make(map[string]interface{})
	for i, name := range namelist {
		if name == "Topic" {
			Topic = valuelist[i]

			continue
		} else if name == "Server" {
			kafkaServer = valuelist[i]
		}
		data[name] = valuelist[i]
	}

	if Topic == "" {
		f.iLog.Error(fmt.Sprintf("SendMessagebyKafka validate wrong: %v", "Topic is empty"))
		return
	}

	if kafkaServer == "" {
		f.iLog.Error(fmt.Sprintf("SendMessagebyKafka validate wrong: %v", "kafkaServer is empty"))
		return
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer([]string{kafkaServer}, config)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error creating producer: %v", err))
		f.ErrorMessage = fmt.Sprintf("Error creating producer: %v", err)
		return
	}
	defer producer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Produce messages to topic (asynchronously)

	jsonData, err := json.Marshal(data)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error:%v", err))
		return
	}
	// Convert JSON byte array to string
	jsonString := string(jsonData)

	f.iLog.Debug(fmt.Sprintf("SendMessagebyKafka Execute: topic, %s, message: %v", Topic, jsonString))

ProducerLoop:
	for {
		select {
		case <-signals:
			break ProducerLoop
		case err := <-producer.Errors():
			f.iLog.Error(fmt.Sprintf("Failed to produce message: %v", err))
		case success := <-producer.Successes():
			f.iLog.Info(fmt.Sprintf("Produced message to topic %s partition %d offset %d",
				success.Topic, success.Partition, success.Offset))
		}

		message := &sarama.ProducerMessage{
			Topic: Topic,
			Value: sarama.StringEncoder(jsonString),
		}

		producer.Input() <- message
	}

	outputs := make(map[string][]interface{})
	f.SetOutputs(f.convertMap(outputs))
}

// Validate validates the SendMessagebyKafka function.
// It checks if the namelist and valuelist are empty,
// and if the "Topic" name is present in the namelist.
// Returns true if the validation passes, otherwise returns false with an error.
// It also logs the performance of the function.
func (cf *SendMessagebyKafka) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendMessagebyKafka.Validate", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendMessagebyKafka.Validate with error: %s", err))
				f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendMessagebyKafka.Validate with error: %s", err)
				return
			}
		}()
	*/
	f.iLog.Debug(fmt.Sprintf("SendMessagebyKafka validate: %v", f))
	namelist, valuelist, _ := f.SetInputs()

	if len(namelist) == 0 {
		return false, fmt.Errorf("SendMessagebyKafka validate: %v", "namelist is empty")
	}

	if len(valuelist) == 0 {
		return false, fmt.Errorf("SendMessagebyKafka validate: %v", "valuelist is empty")
	}
	found := false
	for _, name := range namelist {
		if name == "" {
			return false, fmt.Errorf("SendMessagebyKafka validate: %v", "name is empty")
		}

		if name == "Topic" {
			found = true
		}
	}
	if !found {
		return false, fmt.Errorf("SendMessagebyKafka validate: %v", "Topic is not found")
	}

	return true, nil
}
