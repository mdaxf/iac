package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	//	"go.mongodb.org/mongo-driver/bson"
	//	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"
)

// DocDB is the interface for document database
type MyDocDB struct {
	com.DocDB
}

var once sync.Once

// Connect establishes a connection to the document database.
// It returns an error if the connection fails.
func (db *MyDocDB) Connect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDatabase.Connect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documentdb.Connect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Connect Document Database defer error: %s", err))
		}
	}()

	// Log the database connection details
	iLog.Info(fmt.Sprintf("Connect Document Database: %s %s", db.DatabaseType, db.DatabaseConnection))

	// Establish the database connection if it hasn't been done before
	once.Do(func() {
		// Set client options
		clientOptions := options.Client().ApplyURI(db.DatabaseConnection)

		// Connect to MongoDB
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Document Database Error: %s", err.Error()))
			return
		}

		// Check the connection
		err = client.Ping(context.Background(), nil)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Document Database Error: %s", err.Error()))
			return
		}

		db.MongoDBClient = client
	})

	return nil
}

// Disconnect closes the connection to the document database.
// It returns an error if the disconnection fails.
func (db *MyDocDB) Disconnect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDatabase.Disconnect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documentdb.Disconnect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Disconnect Document Database defer error: %s", err))
		}
	}()

	// Log the database disconnection details
	iLog.Info(fmt.Sprintf("Disconnect Document Database: %s %s", db.DatabaseType, db.DatabaseConnection))

	// Disconnect from MongoDB
	err := db.MongoDBClient.Disconnect(context.Background())
	if err != nil {
		iLog.Error(fmt.Sprintf("Disconnect Document Database Error: %s", err.Error()))
	}

	return err
}

// ReConnect reconnects to the document database.
// It returns an error if the reconnection fails.
func (db *MyDocDB) ReConnect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDatabase.ReConnect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documentdb.ReConnect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("ReConnect Document Database defer error: %s", err))
		}
	}()

	// Log the database reconnection details
	iLog.Info(fmt.Sprintf("ReConnect Document Database: %s %s", db.DatabaseType, db.DatabaseConnection))

	// Reconnect to MongoDB
	err := db.Connect()
	if err != nil {
		iLog.Error(fmt.Sprintf("ReConnect Document Database Error: %s", err.Error()))
	}

	return err
}

func (db *MyDocDB) Ping() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DocumentDatabase.Ping"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documentdb.Ping", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Ping Document Database defer error: %s", err))
		}
	}()

	// Ping the database
	err := db.MongoDBClient.Ping(context.Background(), nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("Ping Document Database Error: %s", err.Error()))
		return err
	}

	return nil
}
