package heartbeat

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"
)

var (
	ErrServiceUnavailable = errors.New("service is unavailable")
)

// HeartbeatManager manages heartbeat checks for various services.
type HeartbeatManager struct {
	checkers map[string]com.HeartbeatChecker
	mu       sync.RWMutex
}

// NewHeartbeatManager creates a new HeartbeatManager instance.
func NewHeartbeatManager() *HeartbeatManager {
	return &HeartbeatManager{
		checkers: make(map[string]com.HeartbeatChecker),
	}
}

// RegisterChecker registers a new HeartbeatChecker with a unique name.
func (m *HeartbeatManager) RegisterChecker(name string, checker com.HeartbeatChecker) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HeartbeatManager"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("RegisterChecker", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to RegisterChecker %s with error: %s", name, err))

			return
		}
	}()

	ilog.Debug(fmt.Sprintf("RegisterChecker %s", name))

	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[name] = checker
}

// StartHeartbeatChecks starts the heartbeat checks for all registered services.
func (m *HeartbeatManager) StartHeartbeatChecks(interval time.Duration) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "StartHeartbeatChecks"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("StartHeartbeatChecks", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to StartHeartbeatChecks with error: %s", err))

			return
		}
	}()

	for name, checker := range m.checkers {
		go m.startHeartbeatCheck(name, checker, interval)
	}
}

func (m *HeartbeatManager) startHeartbeatCheck(name string, checker com.HeartbeatChecker, interval time.Duration) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "startHeartbeatCheck"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("startHeartbeatCheck", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to startHeartbeatCheck %s with error: %s", name, err))

			return
		}
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := checker.Ping()
			if err != nil {
				m.handleServiceUnavailable(name, checker)
			}
		}
	}
}

func (m *HeartbeatManager) handleServiceUnavailable(name string, checker com.HeartbeatChecker) {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "handleServiceUnavailable"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("handleServiceUnavailable", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			ilog.Error(fmt.Sprintf("There is error to handleServiceUnavailable %s with error: %s", name, err))

			return
		}
	}()

	m.mu.RLock()
	defer m.mu.RUnlock()

	ilog.Error(fmt.Sprintf("Service %s is unavailable", name))

	err := checker.Disconnect()
	if err != nil {
		ilog.Error(fmt.Sprintf("There is error to handleServiceUnavailable %s with error: %s", name, err))
	}

	tries := 0
	maxTries := 3

	for {
		err := checker.ReConnect()
		if err == nil {
			// Service reconnected
			break
		}
		// Handle connection error and retry
		ilog.Error(fmt.Sprintf("There is error to handleServiceUnavailable %s with error: %s", name, err))

		tries++
		if tries >= maxTries {
			ilog.Error(fmt.Sprintf("Service %s is still unavailable after %d tries", name, tries))
			return
		}
		time.Sleep(5 * time.Second)
	}
}
