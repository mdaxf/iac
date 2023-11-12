package queue

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	//	"github.com/mdaxf/iac/controllers/trans"
	"github.com/mdaxf/iac/com"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
)

type MessageQueue struct {
	QueueID       string
	QueuName      string
	messages      []Message
	lock          sync.Mutex
	iLog          logger.Log
	DocDBconn     *documents.DocDB
	DB            *sql.DB
	SignalRClient signalr.Client
}

type Message struct {
	Id          string
	UUID        string
	Retry       int
	Execute     int
	Topic       string
	PayLoad     interface{}
	Handler     string
	CreatedOn   time.Time
	ExecutedOn  time.Time
	CompletedOn time.Time
}

type PayLoad struct {
	Topic   string `json:"Topic"`
	Payload string `json:"Payload"`
}

func NewMessageQueue(Id string, Name string) *MessageQueue {

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageQueue"}

	iLog.Debug(fmt.Sprintf(("Create MessageQueue %s %s"), Id, Name))

	mq := &MessageQueue{
		QueueID:  Id,
		QueuName: Name,
		iLog:     iLog,
	}

	mq.DocDBconn = documents.DocDBCon
	mq.SignalRClient = com.IACMessageBusClient
	mq.DB = dbconn.DB
	go mq.execute()

	return mq
}
func NewMessageQueuebyExternal(Id string, Name string, DB *sql.DB, DocDBconn *documents.DocDB, SignalRClient signalr.Client) *MessageQueue {

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageQueue"}

	iLog.Debug(fmt.Sprintf(("Create MessageQueue %s %s"), Id, Name))

	mq := &MessageQueue{
		QueueID:       Id,
		QueuName:      Name,
		iLog:          iLog,
		DocDBconn:     DocDBconn,
		SignalRClient: SignalRClient,
		DB:            DB,
	}

	go mq.execute()

	return mq
}
func (mq *MessageQueue) Push(message Message) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Push message %s to queue: %s", message, mq.QueueID))
	mq.messages = append(mq.messages, message)
}

func (mq *MessageQueue) Pop() Message {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Pop message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("Pop message from queue: %s", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

func (mq *MessageQueue) Length() int {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	return len(mq.messages)
}

func (mq *MessageQueue) Clear() {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.messages = nil
}

func (mq *MessageQueue) WaitAndPop(timeout time.Duration) Message {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue which queue length is %d", len(mq.messages)))
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue: %s", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

func (mq *MessageQueue) WaitAndPopWithTimeout(timeout time.Duration) Message {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue which queue length is %d", len(mq.messages)))
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue: %s", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

func (mq *MessageQueue) Peek() Message {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Peek message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("Peek message from queue: %s", mq.messages[0]))
	message := mq.messages[0]
	return message
}

func (mq *MessageQueue) execute() {
	numMessages := 10
	//maxWorkers := 10

	// Create a wait group to synchronize the workers
	var wg sync.WaitGroup
	n := 0
	go func() {
		// Start the workers
		for {

			defer wg.Done()

			if mq.Length() == 0 {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			n += 1
			wg.Add(1)
			workermessageQueue := make(chan Message, numMessages)
			for i := 0; i < numMessages; i++ {
				workermessageQueue <- mq.Pop()
				if mq.Length() == 0 {
					break
				}
			}
			mq.iLog.Debug(fmt.Sprintf("creating worker %d has %s jobs, %s", n, len(workermessageQueue), workermessageQueue))

			close(workermessageQueue)
			mq.iLog.Debug(fmt.Sprintf("complete create worker %d has %s jobs, %s", n, len(workermessageQueue), workermessageQueue))
			go mq.worker(n, workermessageQueue, &wg)

			time.Sleep(time.Millisecond * 500)
		}
		wg.Wait()
	}()
	mq.waitForTerminationSignal()
}

func (mq *MessageQueue) waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}

func (mq *MessageQueue) processMessage(message Message) error {

	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to process message: %v", r))
			return
		}
	}()

	mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %v", message))

	if message.Handler != "" && message.Topic != "" {
		// Unmarshal the JSON data into a Message object
		message.ExecutedOn = time.Now()
		//	mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %v", message))

		//handler := M.handler
		//data := M.data
		//handler(data)
		var err error
		message.Execute = message.Execute + 1

		mq.iLog.Debug(fmt.Sprintf("execute the message %d with data %v with handler %s", message.Id, message.PayLoad, message.Handler))

		data := make(map[string]interface{})
		/*
			jsondata, err := json.Marshal(message.PayLoad)
			if err != nil {
				mq.iLog.Error(fmt.Sprintf("Failed to convert json to map: %v", err))
				return err
			}
			mq.iLog.Debug(fmt.Sprintf("Message payload data %s, %v", jsondata, message.PayLoad))
		*/
		data["Topic"] = message.Topic
		data["Payload"], err = com.ConvertInterfaceToString(message.PayLoad)
		if err != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to convert json to map: %v", err))
			return err
		}

		data["ID"] = message.Id
		data["UUID"] = message.UUID
		data["CreatedOn"] = message.CreatedOn

		//		data = com.ConvertstructToMap(message)
		mq.iLog.Debug(fmt.Sprintf("Message data %v", data))

		return nil
		if mq.DB == nil {
			mq.DB = dbconn.DB
		}

		if mq.DB == nil {
			mq.iLog.Error(fmt.Sprintf("Failed to get database connection"))
			return err
		}

		tx, err := mq.DB.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
		if err != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to begin transaction: %v", err))
			return err
		}
		defer tx.Rollback()
		mq.iLog.Debug(fmt.Sprintf("execute the transaction %s with data %v ", message.Handler, data))
		outputs, err := trancode.ExecutebyExternal(message.Handler, data, tx, mq.DocDBconn, mq.SignalRClient)
		if err != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to execute transaction: %v", err))
			return err
		}

		err = tx.Commit()
		if err != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to commit transaction: %v", err))
			return err
		}

		status := "Success"
		errormessage := ""
		if err != nil {
			status = "Failed"
			errormessage = err.Error()
		}
		mq.iLog.Debug(fmt.Sprintf("execute the message %d with data %s with handler %s with output %s", message.Id, message.PayLoad, message.Handler, outputs))

		message.CompletedOn = time.Now().UTC()

		msghis := map[string]interface{}{
			"message":      message,
			"executedon":   time.Now().UTC(),
			"executedby":   "System",
			"status":       status,
			"errormessage": errormessage,
			"messagequeue": mq.QueuName,
			"outputs":      outputs,
		}

		if mq.DocDBconn != nil {

			_, err = mq.DocDBconn.InsertCollection("Job_History", msghis)
		}

		if status != "Success" && message.Execute < message.Retry {
			message.Retry++
			mq.iLog.Debug(fmt.Sprintf("execute the message %d failed, and retry time: %d  retry set value %d with data %s with handler %s",
				message.Id, message.Execute, message.Retry, message.PayLoad, message.Handler))
			mq.Push(message)
		}

		return err
	}
	return nil
}

func bytesToMap(message Message) {
	panic("unimplemented")
}

func (mq *MessageQueue) worker(id int, jobs <-chan Message, wg *sync.WaitGroup) {

	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Failed to process message: %v", r))
			return
		}
	}()

	mq.iLog.Debug(fmt.Sprintf("worker %d started", id))
	mq.iLog.Debug(fmt.Sprintf("worker %d has %d jobs ", id, len(jobs)))
	defer wg.Done()
	mq.iLog.Debug(fmt.Sprintf("start process worker %d ", id))
	for msg := range jobs {
		mq.iLog.Debug(fmt.Sprintf("worker %d started to process message %s", id, msg))
		mq.processMessage(msg)
	}
	mq.iLog.Debug(fmt.Sprintf("worker %d finished", id))
	wg.Wait()
	wg.Done()
}
