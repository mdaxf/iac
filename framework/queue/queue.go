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
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/com"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/logger"
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
	PayLoad     []byte
	Handler     string
	CreatedOn   time.Time
	ExecutedOn  time.Time
	CompletedOn time.Time
}

type PayLoad struct {
	Topic   string `json:"Topic"`
	Payload string `json:"Payload"`
}

// NewMessageQueue creates a new instance of MessageQueue with the specified Id and Name.
// It initializes the logger and sets the necessary connections.
// It also starts the execution of the MessageQueue in a separate goroutine.
// Returns a pointer to the created MessageQueue.

func NewMessageQueue(Id string, Name string) *MessageQueue {

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageQueue"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			iLog.PerformanceWithDuration("framework.queue.NewMessageQueue", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				iLog.Error(fmt.Sprintf("Error in framework.queue.NewMessageQueue: %s", r))
				return
			}
		}() */

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

// NewMessageQueuebyExternal creates a new instance of MessageQueue with the given parameters.
// It takes an Id string, a Name string, a DB *sql.DB, a DocDBconn *documents.DocDB, and a SignalRClient signalr.Client.
// It returns a pointer to the created MessageQueue.

func NewMessageQueuebyExternal(Id string, Name string, DB *sql.DB, DocDBconn *documents.DocDB, SignalRClient signalr.Client) *MessageQueue {

	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MessageQueue"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			iLog.PerformanceWithDuration("framework.queue.NewMessageQueuebyExternal", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				iLog.Error(fmt.Sprintf("Error in framework.queue.NewMessageQueuebyExternal: %s", r))
				return
			}
		}()  */
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

// Push adds a message to the message queue.
// It measures the performance duration of the operation and logs any errors that occur.
// The message is appended to the queue and can be accessed later.
// It takes a Message struct as a parameter.

func (mq *MessageQueue) Push(message Message) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.Push", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.Push: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Push message %v to queue: %s", message, mq.QueueID))
	mq.messages = append(mq.messages, message)
}

// Pop removes and returns the first message from the message queue.
// If the queue is empty, it returns an empty Message.
// It measures the performance duration of the operation and logs any errors that occur.
func (mq *MessageQueue) Pop() Message {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.Pop", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.Pop: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Pop message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("Pop message from queue: %v", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

// Length returns the number of messages in the message queue.
func (mq *MessageQueue) Length() int {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	return len(mq.messages)
}

// Clear removes all messages from the message queue.
// It acquires a lock to ensure thread safety and sets the messages slice to nil.
// Any error that occurs during the clearing process is recovered and logged.
// The performance duration of the Clear operation is also logged.

func (mq *MessageQueue) Clear() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.Clear", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.Clear: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.messages = nil
}

// WaitAndPop waits for a message to be available in the queue and then removes and returns it.
// If the queue is empty, it will return an empty Message struct.
// The timeout parameter specifies the maximum duration to wait for a message before returning.
// If a timeout of zero is provided, it will wait indefinitely until a message is available.
// It measures the performance duration of the operation and logs any errors that occur.
// It takes a timeout time.Duration as a parameter and returns a Message struct.
func (mq *MessageQueue) WaitAndPop(timeout time.Duration) Message {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.WaitAndPop", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.WaitAndPop: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue which queue length is %d", len(mq.messages)))
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("WaitAndPop message from queue: %v", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

// WaitAndPopWithTimeout waits for a specified duration and pops a message from the message queue.
// If the queue is empty, it returns an empty message.
// It also logs the performance duration of the function.
// If there is an error during the execution, it recovers and logs the error.
// This function is thread-safe.
//
// Parameters:
// - timeout: The duration to wait for a message before timing out.
//
// Returns:
// - Message: The popped message from the queue.

func (mq *MessageQueue) WaitAndPopWithTimeout(timeout time.Duration) Message {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.WaitAndPopWithTimeout", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.WaitAndPopWithTimeout: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue which queue length is %d", len(mq.messages)))
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("WaitAndPopWithTimeout message from queue: %v", mq.messages[0]))
	message := mq.messages[0]
	mq.messages = mq.messages[1:]
	return message
}

// Peek returns the first message in the message queue without removing it.
// If the queue is empty, it returns an empty Message.
// It measures the performance duration of the operation and logs any errors that occur.
func (mq *MessageQueue) Peek() Message {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.Peek", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.Peek: %s", r))
			return
		}
	}()
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.iLog.Debug(fmt.Sprintf("Peek message from queue which queue length is %d", len(mq.messages)))
	if len(mq.messages) == 0 {
		return Message{}
	}
	mq.iLog.Debug(fmt.Sprintf("Peek message from queue: %v", mq.messages[0]))
	message := mq.messages[0]
	return message
}

// execute is a method of the MessageQueue struct that starts the execution of message processing.
// It creates worker goroutines to process messages from the queue until a termination signal is received.
// The method also measures the performance of the execution and handles any panics that occur during processing.
// It is called by the NewMessageQueue function.

func (mq *MessageQueue) execute() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.execute", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("Error in framework.queue.execute: %s", r))
			return
		}
	}()
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
			mq.iLog.Debug(fmt.Sprintf("creating worker %d has %v jobs, %v", n, len(workermessageQueue), workermessageQueue))

			close(workermessageQueue)
			mq.iLog.Debug(fmt.Sprintf("complete create worker %d has %d jobs, %v", n, len(workermessageQueue), workermessageQueue))
			go mq.worker(n, workermessageQueue, &wg)

			time.Sleep(time.Millisecond * 500)
		}

	}()

	wg.Wait()
	mq.waitForTerminationSignal()
}

// waitForTerminationSignal waits for a termination signal and performs cleanup or graceful shutdown logic.
// It listens for an interrupt signal or SIGTERM signal, and upon receiving the signal, it prints a shutdown message,
// sleeps for 2 seconds, and exits the program with a status code of 0.
// It measures the performance duration of the operation and logs any errors that occur.
// It is called by the execute method.
// It is also called by the NewMessageQueue function.

func (mq *MessageQueue) waitForTerminationSignal() {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			mq.iLog.PerformanceWithDuration("framework.queue.waitForTerminationSignal", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				mq.iLog.Error(fmt.Sprintf("Error in framework.queue.waitForTerminationSignal: %s", r))
				return
			}
		}() */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}

func (mq *MessageQueue) processMessage(message Message) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		mq.iLog.PerformanceWithDuration("framework.queue.processMessage", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			mq.iLog.Error(fmt.Sprintf("framework.queue.processMessage failed to process message: %v", r))
			return
		}
	}()

	mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %v", message))

	if message.Handler != "" && message.Topic != "" {
		// Unmarshal the JSON data into a Message object
		message.ExecutedOn = time.Now().UTC()
		//	mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %v", message))

		//handler := M.handler
		//data := M.data
		//handler(data)
		var err error
		message.Execute = message.Execute + 1

		mq.iLog.Debug(fmt.Sprintf("execute the message %s with data %v with handler %v", message.Id, message.PayLoad, message.Handler))

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
		data["Payload"] = string(message.PayLoad)

		data["ID"] = message.Id
		data["UUID"] = message.UUID
		data["CreatedOn"] = message.CreatedOn

		//		data = com.ConvertstructToMap(message)
		mq.iLog.Debug(fmt.Sprintf("Message data %v", data))

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
		mq.iLog.Debug(fmt.Sprintf("execute the message %s with data %s with handler %s with output %s", message.Id, message.PayLoad, message.Handler, outputs))

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
			mq.iLog.Debug(fmt.Sprintf("execute the message %s failed, and retry time: %d  retry set value %d with data %s with handler %s",
				message.Id, message.Execute, message.Retry, message.PayLoad, message.Handler))
			mq.Push(message)
		}

		return err
	}
	return nil
}

// bytesToMap converts a Message object to a map.
// It is currently unimplemented and will cause a panic.
func bytesToMap(message Message) {
	panic("unimplemented")
}

// worker is a function that represents a worker in the MessageQueue.
// It processes messages from the jobs channel and calls the processMessage function for each message.
// The worker logs debug messages for its start, progress, and finish.
// It also logs performance metrics for the time taken to process the messages.
// If any panic occurs during message processing, it logs an error message.
// The worker waits for all jobs to be processed before returning.
// It is called by the execute method.
func (mq *MessageQueue) worker(id int, jobs <-chan Message, wg *sync.WaitGroup) {
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			mq.iLog.PerformanceWithDuration("framework.queue.worker", elapsed)
		}()

		defer func() {
			if r := recover(); r != nil {
				mq.iLog.Error(fmt.Sprintf("framework.queue.processMessage failed to process message: %v", r))
				return
			}
		}()
	*/
	mq.iLog.Debug(fmt.Sprintf("worker %d started", id))
	mq.iLog.Debug(fmt.Sprintf("worker %d has %d jobs ", id, len(jobs)))
	defer wg.Done()
	mq.iLog.Debug(fmt.Sprintf("start process worker %d ", id))
	for msg := range jobs {
		mq.iLog.Debug(fmt.Sprintf("worker %d started to process message %v", id, msg))
		mq.processMessage(msg)
	}
	mq.iLog.Debug(fmt.Sprintf("worker %d finished", id))
	wg.Wait()
	wg.Done()
}
