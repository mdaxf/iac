package funcs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// ExecutionNode represents a function in the execution graph
type ExecutionNode struct {
	Function     *types.Function
	Dependencies []string // Names of functions this depends on
	Outputs      map[string]interface{}
	Error        error
	Executed     bool
	mu           sync.RWMutex
}

// ExecutionGraph manages parallel function execution with dependency resolution
type ExecutionGraph struct {
	Nodes           map[string]*ExecutionNode
	ExecutionOrder  [][]string // Groups of functions that can execute in parallel
	MaxConcurrency  int
	Timeout         time.Duration
	Log             logger.Log
	funcExecutor    FunctionExecutor
}

// FunctionExecutor defines how to execute a single function
type FunctionExecutor interface {
	Execute(ctx context.Context, fn *types.Function, inputs map[string]interface{}) (map[string]interface{}, error)
}

// NewExecutionGraph creates a new execution graph
func NewExecutionGraph(maxConcurrency int, timeout time.Duration, log logger.Log) *ExecutionGraph {
	return &ExecutionGraph{
		Nodes:          make(map[string]*ExecutionNode),
		MaxConcurrency: maxConcurrency,
		Timeout:        timeout,
		Log:            log,
	}
}

// AddFunction adds a function to the execution graph
func (eg *ExecutionGraph) AddFunction(fn *types.Function, dependencies []string) {
	eg.Nodes[fn.Name] = &ExecutionNode{
		Function:     fn,
		Dependencies: dependencies,
		Outputs:      make(map[string]interface{}),
		Executed:     false,
	}
}

// BuildExecutionOrder performs topological sort to determine parallel execution groups
func (eg *ExecutionGraph) BuildExecutionOrder() error {
	// Track in-degree (number of dependencies) for each node
	inDegree := make(map[string]int)
	for name, node := range eg.Nodes {
		inDegree[name] = len(node.Dependencies)
	}

	// Find all nodes with no dependencies (can execute immediately)
	var executionOrder [][]string
	remaining := make(map[string]bool)
	for name := range eg.Nodes {
		remaining[name] = true
	}

	for len(remaining) > 0 {
		// Find all nodes with no remaining dependencies
		currentLevel := []string{}
		for name := range remaining {
			if inDegree[name] == 0 {
				currentLevel = append(currentLevel, name)
			}
		}

		if len(currentLevel) == 0 {
			// Circular dependency detected
			return types.NewValidationError("Circular dependency detected in function execution graph", nil).
				WithDetail("remaining_functions", fmt.Sprintf("%v", remaining))
		}

		executionOrder = append(executionOrder, currentLevel)

		// Remove executed nodes and update in-degrees
		for _, name := range currentLevel {
			delete(remaining, name)

			// Reduce in-degree for nodes that depend on this one
			for depName, depNode := range eg.Nodes {
				for _, dep := range depNode.Dependencies {
					if dep == name {
						inDegree[depName]--
					}
				}
			}
		}
	}

	eg.ExecutionOrder = executionOrder
	return nil
}

// ExecuteParallel executes functions in parallel according to dependency order
func (eg *ExecutionGraph) ExecuteParallel(ctx context.Context) error {
	if len(eg.ExecutionOrder) == 0 {
		if err := eg.BuildExecutionOrder(); err != nil {
			return err
		}
	}

	eg.Log.Info(fmt.Sprintf("Starting parallel execution of %d function groups", len(eg.ExecutionOrder)))

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, eg.Timeout)
	defer cancel()

	for levelIdx, level := range eg.ExecutionOrder {
		eg.Log.Info(fmt.Sprintf("Executing level %d with %d functions in parallel", levelIdx+1, len(level)))

		if err := eg.executeLevel(execCtx, level); err != nil {
			return err
		}
	}

	eg.Log.Info("Parallel execution completed successfully")
	return nil
}

// executeLevel executes a group of functions in parallel
func (eg *ExecutionGraph) executeLevel(ctx context.Context, functionNames []string) error {
	// Create worker pool
	semaphore := make(chan struct{}, eg.MaxConcurrency)
	var wg sync.WaitGroup
	errChan := make(chan error, len(functionNames))

	for _, name := range functionNames {
		wg.Add(1)

		go func(funcName string) {
			defer wg.Done()

			// Acquire semaphore (limit concurrency)
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute function
			if err := eg.executeFunction(ctx, funcName); err != nil {
				errChan <- err
			}
		}(name)
	}

	// Wait for all functions in this level to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return err // Return first error encountered
	}

	return nil
}

// executeFunction executes a single function
func (eg *ExecutionGraph) executeFunction(ctx context.Context, name string) error {
	node := eg.Nodes[name]
	if node == nil {
		return fmt.Errorf("function %s not found in execution graph", name)
	}

	startTime := time.Now()
	eg.Log.Info(fmt.Sprintf("Executing function: %s", name))

	// Gather inputs from dependencies
	inputs := make(map[string]interface{})
	for _, depName := range node.Dependencies {
		depNode := eg.Nodes[depName]
		if depNode == nil {
			return fmt.Errorf("dependency %s not found for function %s", depName, name)
		}

		depNode.mu.RLock()
		if !depNode.Executed {
			depNode.mu.RUnlock()
			return fmt.Errorf("dependency %s not executed for function %s", depName, name)
		}

		// Copy outputs from dependency
		for k, v := range depNode.Outputs {
			inputs[k] = v
		}
		depNode.mu.RUnlock()
	}

	// Execute the function
	var outputs map[string]interface{}
	var err error

	if eg.funcExecutor != nil {
		outputs, err = eg.funcExecutor.Execute(ctx, node.Function, inputs)
	} else {
		// Default execution logic would go here
		outputs = make(map[string]interface{})
		err = nil
	}

	// Update node with results
	node.mu.Lock()
	node.Outputs = outputs
	node.Error = err
	node.Executed = true
	node.mu.Unlock()

	duration := time.Since(startTime)
	if err != nil {
		eg.Log.Error(fmt.Sprintf("Function %s failed after %v: %s", name, duration, err.Error()))
		return err
	}

	eg.Log.Info(fmt.Sprintf("Function %s completed successfully in %v", name, duration))
	return nil
}

// GetResults retrieves the outputs from all executed functions
func (eg *ExecutionGraph) GetResults() map[string]map[string]interface{} {
	results := make(map[string]map[string]interface{})

	for name, node := range eg.Nodes {
		node.mu.RLock()
		if node.Executed && node.Error == nil {
			results[name] = node.Outputs
		}
		node.mu.RUnlock()
	}

	return results
}

// WorkerPool manages a pool of workers for parallel execution
type WorkerPool struct {
	workers    int
	tasks      chan func()
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	log        logger.Log
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, log logger.Log) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers: workers,
		tasks:   make(chan func(), workers*2), // Buffered channel
		ctx:     ctx,
		cancel:  cancel,
		log:     log,
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

// worker processes tasks from the task channel
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.log.Debug(fmt.Sprintf("Worker %d started", id))

	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				wp.log.Debug(fmt.Sprintf("Worker %d stopped (channel closed)", id))
				return
			}

			// Execute task
			task()

		case <-wp.ctx.Done():
			wp.log.Debug(fmt.Sprintf("Worker %d stopped (context cancelled)", id))
			return
		}
	}
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task func()) error {
	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.log.Info("Shutting down worker pool")
	close(wp.tasks)
	wp.wg.Wait()
	wp.log.Info("Worker pool shutdown complete")
}

// ShutdownNow forcefully shuts down the worker pool
func (wp *WorkerPool) ShutdownNow() {
	wp.log.Info("Force shutting down worker pool")
	wp.cancel()
	wp.wg.Wait()
	wp.log.Info("Worker pool force shutdown complete")
}

// ParallelExecutor executes a batch of functions in parallel
type ParallelExecutor struct {
	maxConcurrency int
	timeout        time.Duration
	log            logger.Log
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxConcurrency int, timeout time.Duration, log logger.Log) *ParallelExecutor {
	return &ParallelExecutor{
		maxConcurrency: maxConcurrency,
		timeout:        timeout,
		log:            log,
	}
}

// ExecuteBatch executes a batch of independent functions in parallel
func (pe *ParallelExecutor) ExecuteBatch(ctx context.Context, functions []*types.Function, executor FunctionExecutor) ([]map[string]interface{}, []error) {
	results := make([]map[string]interface{}, len(functions))
	errors := make([]error, len(functions))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, pe.maxConcurrency)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, pe.timeout)
	defer cancel()

	for i, fn := range functions {
		wg.Add(1)

		go func(idx int, function *types.Function) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute function
			output, err := executor.Execute(execCtx, function, make(map[string]interface{}))
			results[idx] = output
			errors[idx] = err
		}(i, fn)
	}

	wg.Wait()
	return results, errors
}

// AsyncExecutor executes functions asynchronously and returns results via channels
type AsyncExecutor struct {
	workerPool *WorkerPool
	log        logger.Log
}

// NewAsyncExecutor creates a new async executor
func NewAsyncExecutor(workers int, log logger.Log) *AsyncExecutor {
	return &AsyncExecutor{
		workerPool: NewWorkerPool(workers, log),
		log:        log,
	}
}

// ExecuteAsync executes a function asynchronously
func (ae *AsyncExecutor) ExecuteAsync(ctx context.Context, fn *types.Function, executor FunctionExecutor) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	ae.workerPool.Submit(func() {
		output, err := executor.Execute(ctx, fn, make(map[string]interface{}))
		resultChan <- AsyncResult{
			Output: output,
			Error:  err,
		}
		close(resultChan)
	})

	return resultChan
}

// AsyncResult holds the result of an async execution
type AsyncResult struct {
	Output map[string]interface{}
	Error  error
}

// Shutdown gracefully shuts down the async executor
func (ae *AsyncExecutor) Shutdown() {
	ae.workerPool.Shutdown()
}

// ConcurrencyLimiter limits concurrent operations
type ConcurrencyLimiter struct {
	semaphore chan struct{}
}

// NewConcurrencyLimiter creates a new concurrency limiter
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Acquire acquires a slot for concurrent execution
func (cl *ConcurrencyLimiter) Acquire() {
	cl.semaphore <- struct{}{}
}

// Release releases a slot
func (cl *ConcurrencyLimiter) Release() {
	<-cl.semaphore
}

// Execute executes a function with concurrency limiting
func (cl *ConcurrencyLimiter) Execute(fn func()) {
	cl.Acquire()
	defer cl.Release()
	fn()
}
