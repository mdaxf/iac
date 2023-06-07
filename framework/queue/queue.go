package queue

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

type MessageQueue struct {
	QueueID  string
	QueuName string
	messages []Message
	lock     sync.Mutex
	iLog     logger.Log
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

	mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %s", message))

	if message.Handler != "" && message.Topic != "" {
		// Unmarshal the JSON data into a Message object
		message.ExecutedOn = time.Now()
		mq.iLog.Debug(fmt.Sprintf("handlemessagefromqueue message from queue: %s", message))

		//handler := M.handler
		//data := M.data
		//handler(data)
		var err error
		message.Execute = message.Execute + 1

		status := "Success"
		errormessage := ""
		message.CompletedOn = time.Now()

		if err != nil {
			status = "Failed"
			errormessage = err.Error()
		}

		msghis := map[string]interface{}{
			"message":      message,
			"executedon":   time.Now(),
			"executedby":   "System",
			"status":       status,
			"errormessage": errormessage,
			"messagequeue": mq.QueuName,
		}

		_, err = documents.DocDBCon.InsertCollection("Job_History", msghis)

		if err != nil && message.Execute < message.Retry {
			message.Retry++
			mq.iLog.Debug(fmt.Sprintf("execute the message %d failed, and retry time: %d  retry set value %d with data %s with handler %s",
				message.Id, message.Execute, message.Retry, message.PayLoad, message.Handler))
			mq.Push(message)
		}

		return err
	}
	return nil
}

func (mq *MessageQueue) worker(id int, jobs <-chan Message, wg *sync.WaitGroup) {
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
