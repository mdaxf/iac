package checks

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultTimeoutConnect    = 5 * time.Second
	defaultTimeoutDisconnect = 5 * time.Second
	defaultTimeoutPing       = 5 * time.Second
)

// Config is the MongoDB checker configuration settings container.
type MongoDbCheck struct {
	Ctx context.Context
	// DSN is the MongoDB instance connection DSN. Required.
	ConnectionString string

	// TimeoutConnect defines timeout for establishing mongo connection, if not set - default value is used
	TimeoutConnect time.Duration
	// TimeoutDisconnect defines timeout for closing connection, if not set - default value is used
	TimeoutDisconnect time.Duration
	// TimeoutDisconnect defines timeout for making ping request, if not set - default value is used
	TimeoutPing time.Duration
	// return the check error
	Error error
}

func CheckMongoDBStatus(ctx context.Context, connectionString string, timeoutConnect time.Duration, timeoutDisconnect time.Duration, timeoutPing time.Duration) error {
	check := NewMongoDbCheck(ctx, connectionString, timeoutConnect, timeoutDisconnect, timeoutPing)
	return check.CheckStatus()
}

func NewMongoDbCheck(ctx context.Context, connectionString string, timeoutConnect time.Duration, timeoutDisconnect time.Duration, timeoutPing time.Duration) MongoDbCheck {
	if timeoutConnect == 0 {
		timeoutConnect = defaultTimeoutConnect
	}

	if timeoutDisconnect == 0 {
		timeoutDisconnect = defaultTimeoutDisconnect
	}

	if timeoutPing == 0 {
		timeoutPing = defaultTimeoutPing
	}

	return MongoDbCheck{
		Ctx:               ctx,
		ConnectionString:  connectionString,
		Error:             nil,
		TimeoutConnect:    timeoutConnect,
		TimeoutDisconnect: timeoutDisconnect,
		TimeoutPing:       timeoutPing,
	}
}

func (check MongoDbCheck) CheckStatus() error {

	var checkErr error
	checkErr = nil
	ctx := check.Ctx
	client, err := mongo.NewClient(options.Client().ApplyURI(check.ConnectionString))
	if err != nil {
		checkErr = fmt.Errorf("mongoDB health check failed on client creation: %w", err)
		check.Error = checkErr
		return checkErr
	}

	ctxConn, cancelConn := context.WithTimeout(ctx, check.TimeoutConnect)
	defer cancelConn()

	err = client.Connect(ctxConn)
	if err != nil {
		checkErr = fmt.Errorf("mongoDB health check failed on connect: %w", err)
		check.Error = checkErr
		return checkErr
	}

	defer func() {
		ctxDisc, cancelDisc := context.WithTimeout(ctx, check.TimeoutDisconnect)
		defer cancelDisc()

		// override checkErr only if there were no other errors
		if err := client.Disconnect(ctxDisc); err != nil && checkErr == nil {
			checkErr = fmt.Errorf("mongoDB health check failed on closing connection: %w", err)
			check.Error = checkErr
			return
		}
	}()

	ctxPing, cancelPing := context.WithTimeout(ctx, check.TimeoutPing)
	defer cancelPing()

	err = client.Ping(ctxPing, readpref.Primary())
	if err != nil {
		checkErr = fmt.Errorf("mongoDB health check failed on ping: %w", err)
		check.Error = checkErr
		return checkErr
	}

	return nil
}
