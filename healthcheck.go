package main

import (
	"fmt"
	"time"
)

func checkconnection() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.CheckConnection", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to main.CheckConnection with error: %s", err))

			return
		}
	}()

	// check db connection

	// check document db connection
}

func checkservices() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.CheckServices", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to main.CheckServices with error: %s", err))

			return
		}
	}()

	ilog.Debug(fmt.Sprintf("Check services"))
	// check message bus connection

	// check singalR connection

	// check kafka connection

	//check activeMQ connection

	// check mqtt connection

	// check redis connection

	// check rabbitMQ connection
}
