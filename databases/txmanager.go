// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
)

// TransactionManager manages database transactions
type TransactionManager struct {
	db           RelationalDB
	iLog         logger.Log
	maxRetries   int
	retryDelay   time.Duration
	isolation    sql.IsolationLevel
	mu           sync.RWMutex
	activeTxns   map[string]*ManagedTransaction
}

// ManagedTransaction represents a managed database transaction
type ManagedTransaction struct {
	ID           string
	Tx           *sql.Tx
	StartTime    time.Time
	Isolation    sql.IsolationLevel
	Savepoints   []string
	IsCommitted  bool
	IsRolledBack bool
	mu           sync.RWMutex
}

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	Isolation   sql.IsolationLevel
	ReadOnly    bool
	MaxRetries  int
	RetryDelay  time.Duration
	Timeout     time.Duration
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db RelationalDB) *TransactionManager {
	return &TransactionManager{
		db:         db,
		iLog:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "TransactionManager"},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
		isolation:  sql.LevelReadCommitted,
		activeTxns: make(map[string]*ManagedTransaction),
	}
}

// SetDefaultIsolation sets the default isolation level
func (tm *TransactionManager) SetDefaultIsolation(isolation sql.IsolationLevel) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.isolation = isolation
}

// SetMaxRetries sets the maximum number of retries for retryable errors
func (tm *TransactionManager) SetMaxRetries(maxRetries int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.maxRetries = maxRetries
}

// SetRetryDelay sets the delay between retries
func (tm *TransactionManager) SetRetryDelay(delay time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.retryDelay = delay
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin(ctx context.Context, opts *TransactionOptions) (*ManagedTransaction, error) {
	if opts == nil {
		opts = &TransactionOptions{
			Isolation: tm.isolation,
		}
	}

	txOpts := &sql.TxOptions{
		Isolation: opts.Isolation,
		ReadOnly:  opts.ReadOnly,
	}

	tx, err := tm.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, NewDatabaseError("begin_transaction", err, tm.db.GetType(), "")
	}

	mtx := &ManagedTransaction{
		ID:        generateTxID(),
		Tx:        tx,
		StartTime: time.Now(),
		Isolation: opts.Isolation,
	}

	tm.mu.Lock()
	tm.activeTxns[mtx.ID] = mtx
	tm.mu.Unlock()

	tm.iLog.Debug(fmt.Sprintf("Started transaction %s with isolation %v", mtx.ID, opts.Isolation))

	return mtx, nil
}

// ExecuteInTransaction executes a function within a transaction
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, opts *TransactionOptions, fn func(*sql.Tx) error) error {
	if opts == nil {
		opts = &TransactionOptions{}
	}

	maxRetries := tm.maxRetries
	if opts.MaxRetries > 0 {
		maxRetries = opts.MaxRetries
	}

	retryDelay := tm.retryDelay
	if opts.RetryDelay > 0 {
		retryDelay = opts.RetryDelay
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			tm.iLog.Warn(fmt.Sprintf("Retrying transaction (attempt %d/%d) after error: %v", attempt, maxRetries, lastErr))
			time.Sleep(retryDelay)
		}

		// Apply timeout if specified
		txCtx := ctx
		var cancel context.CancelFunc
		if opts.Timeout > 0 {
			txCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}

		err := tm.executeTransaction(txCtx, opts, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
}

// executeTransaction executes a single transaction attempt
func (tm *TransactionManager) executeTransaction(ctx context.Context, opts *TransactionOptions, fn func(*sql.Tx) error) (err error) {
	mtx, err := tm.Begin(ctx, opts)
	if err != nil {
		return err
	}

	// Ensure transaction is cleaned up
	defer func() {
		tm.cleanupTransaction(mtx, err)
	}()

	// Handle panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in transaction: %v", r)
			tm.iLog.Error(fmt.Sprintf("Panic in transaction %s: %v", mtx.ID, r))
			mtx.Rollback()
		}
	}()

	// Execute the transaction function
	err = fn(mtx.Tx)
	if err != nil {
		tm.iLog.Error(fmt.Sprintf("Transaction %s failed: %v", mtx.ID, err))
		if rbErr := mtx.Rollback(); rbErr != nil {
			tm.iLog.Error(fmt.Sprintf("Failed to rollback transaction %s: %v", mtx.ID, rbErr))
		}
		return err
	}

	// Commit the transaction
	err = mtx.Commit()
	if err != nil {
		tm.iLog.Error(fmt.Sprintf("Failed to commit transaction %s: %v", mtx.ID, err))
		return err
	}

	tm.iLog.Debug(fmt.Sprintf("Transaction %s committed successfully", mtx.ID))
	return nil
}

// cleanupTransaction removes the transaction from active transactions
func (tm *TransactionManager) cleanupTransaction(mtx *ManagedTransaction, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.activeTxns, mtx.ID)
}

// GetActiveTxnCount returns the number of active transactions
func (tm *TransactionManager) GetActiveTxnCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.activeTxns)
}

// GetActiveTxnIDs returns IDs of all active transactions
func (tm *TransactionManager) GetActiveTxnIDs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0, len(tm.activeTxns))
	for id := range tm.activeTxns {
		ids = append(ids, id)
	}
	return ids
}

// ManagedTransaction methods

// Commit commits the transaction
func (mtx *ManagedTransaction) Commit() error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()

	if mtx.IsCommitted {
		return ErrTransactionAlreadyCommitted
	}
	if mtx.IsRolledBack {
		return ErrTransactionAlreadyRolledBack
	}

	err := mtx.Tx.Commit()
	if err != nil {
		return err
	}

	mtx.IsCommitted = true
	return nil
}

// Rollback rolls back the transaction
func (mtx *ManagedTransaction) Rollback() error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()

	if mtx.IsCommitted {
		return ErrTransactionAlreadyCommitted
	}
	if mtx.IsRolledBack {
		return nil // Already rolled back, not an error
	}

	err := mtx.Tx.Rollback()
	if err != nil {
		return err
	}

	mtx.IsRolledBack = true
	return nil
}

// Savepoint creates a savepoint within the transaction
func (mtx *ManagedTransaction) Savepoint(name string) error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()

	if mtx.IsCommitted || mtx.IsRolledBack {
		return ErrTransactionFinished
	}

	// Execute savepoint command (database-specific)
	query := fmt.Sprintf("SAVEPOINT %s", name)
	_, err := mtx.Tx.Exec(query)
	if err != nil {
		return err
	}

	mtx.Savepoints = append(mtx.Savepoints, name)
	return nil
}

// RollbackToSavepoint rolls back to a specific savepoint
func (mtx *ManagedTransaction) RollbackToSavepoint(name string) error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()

	if mtx.IsCommitted || mtx.IsRolledBack {
		return ErrTransactionFinished
	}

	// Check if savepoint exists
	found := false
	for _, sp := range mtx.Savepoints {
		if sp == name {
			found = true
			break
		}
	}

	if !found {
		return ErrSavepointNotFound
	}

	// Execute rollback to savepoint command
	query := fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)
	_, err := mtx.Tx.Exec(query)
	return err
}

// ReleaseSavepoint releases a savepoint
func (mtx *ManagedTransaction) ReleaseSavepoint(name string) error {
	mtx.mu.Lock()
	defer mtx.mu.Unlock()

	if mtx.IsCommitted || mtx.IsRolledBack {
		return ErrTransactionFinished
	}

	// Execute release savepoint command
	query := fmt.Sprintf("RELEASE SAVEPOINT %s", name)
	_, err := mtx.Tx.Exec(query)
	if err != nil {
		return err
	}

	// Remove from savepoints list
	for i, sp := range mtx.Savepoints {
		if sp == name {
			mtx.Savepoints = append(mtx.Savepoints[:i], mtx.Savepoints[i+1:]...)
			break
		}
	}

	return nil
}

// GetDuration returns the duration of the transaction
func (mtx *ManagedTransaction) GetDuration() time.Duration {
	return time.Since(mtx.StartTime)
}

// IsActive returns whether the transaction is still active
func (mtx *ManagedTransaction) IsActive() bool {
	mtx.mu.RLock()
	defer mtx.mu.RUnlock()
	return !mtx.IsCommitted && !mtx.IsRolledBack
}

// Helper functions

var txIDCounter uint64
var txIDMutex sync.Mutex

func generateTxID() string {
	txIDMutex.Lock()
	defer txIDMutex.Unlock()
	txIDCounter++
	return fmt.Sprintf("txn_%d_%d", time.Now().Unix(), txIDCounter)
}

// Additional error types
var (
	ErrTransactionAlreadyCommitted  = fmt.Errorf("transaction already committed")
	ErrTransactionAlreadyRolledBack = fmt.Errorf("transaction already rolled back")
	ErrTransactionFinished          = fmt.Errorf("transaction already finished")
	ErrSavepointNotFound            = fmt.Errorf("savepoint not found")
)
