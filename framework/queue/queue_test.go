package queue

import (
	"database/sql"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac-signalr/signalr"
)

func TestNewMessageQueuebyExternal(t *testing.T) {
	type args struct {
		Id            string
		Name          string
		DB            *sql.DB
		DocDBconn     *documents.DocDB
		SignalRClient signalr.Client
	}
	tests := []struct {
		name string
		args args
		want *MessageQueue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMessageQueuebyExternal(tt.args.Id, tt.args.Name, tt.args.DB, tt.args.DocDBconn, tt.args.SignalRClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMessageQueuebyExternal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_Push(t *testing.T) {
	type args struct {
		message Message
	}
	tests := []struct {
		name string
		mq   *MessageQueue
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mq.Push(tt.args.message)
		})
	}
}

func TestMessageQueue_Pop(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
		want Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mq.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageQueue.Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_Length(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mq.Length(); got != tt.want {
				t.Errorf("MessageQueue.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_Clear(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mq.Clear()
		})
	}
}

func TestMessageQueue_WaitAndPop(t *testing.T) {
	type args struct {
		timeout time.Duration
	}
	tests := []struct {
		name string
		mq   *MessageQueue
		args args
		want Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mq.WaitAndPop(tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageQueue.WaitAndPop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_WaitAndPopWithTimeout(t *testing.T) {
	type args struct {
		timeout time.Duration
	}
	tests := []struct {
		name string
		mq   *MessageQueue
		args args
		want Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mq.WaitAndPopWithTimeout(tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageQueue.WaitAndPopWithTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_Peek(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
		want Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mq.Peek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageQueue.Peek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageQueue_execute(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mq.execute()
		})
	}
}

func TestMessageQueue_waitForTerminationSignal(t *testing.T) {
	tests := []struct {
		name string
		mq   *MessageQueue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mq.waitForTerminationSignal()
		})
	}
}

func TestMessageQueue_processMessage(t *testing.T) {
	type args struct {
		message Message
	}
	tests := []struct {
		name    string
		mq      *MessageQueue
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mq.processMessage(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("MessageQueue.processMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageQueue_worker(t *testing.T) {
	type args struct {
		id   int
		jobs <-chan Message
		wg   *sync.WaitGroup
	}
	tests := []struct {
		name string
		mq   *MessageQueue
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mq.worker(tt.args.id, tt.args.jobs, tt.args.wg)
		})
	}
}
