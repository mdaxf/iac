package testing

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

// TestContext provides common testing utilities
type TestContext struct {
	T             *testing.T
	DB            *sql.DB
	Ctx           context.Context
	CancelFunc    context.CancelFunc
	Logger        logger.Log
	TestData      map[string]interface{}
	Cleanup       []func()
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	return &TestContext{
		T:          t,
		Ctx:        ctx,
		CancelFunc: cancel,
		Logger: logger.Log{
			ModuleName:     "TestFramework",
			User:           "TestUser",
			ControllerName: "Test",
		},
		TestData: make(map[string]interface{}),
		Cleanup:  make([]func(), 0),
	}
}

// AddCleanup adds a cleanup function to be called after the test
func (tc *TestContext) AddCleanup(fn func()) {
	tc.Cleanup = append(tc.Cleanup, fn)
}

// RunCleanup runs all cleanup functions
func (tc *TestContext) RunCleanup() {
	for i := len(tc.Cleanup) - 1; i >= 0; i-- {
		tc.Cleanup[i]()
	}
	tc.CancelFunc()
}

// AssertEqual asserts that two values are equal
func (tc *TestContext) AssertEqual(expected, actual interface{}, message string) {
	tc.T.Helper()
	if !reflect.DeepEqual(expected, actual) {
		tc.T.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func (tc *TestContext) AssertNotEqual(expected, actual interface{}, message string) {
	tc.T.Helper()
	if reflect.DeepEqual(expected, actual) {
		tc.T.Errorf("%s: expected values to be different, but both are %v", message, expected)
	}
}

// AssertNil asserts that a value is nil
func (tc *TestContext) AssertNil(value interface{}, message string) {
	tc.T.Helper()
	if value != nil && !reflect.ValueOf(value).IsNil() {
		tc.T.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil asserts that a value is not nil
func (tc *TestContext) AssertNotNil(value interface{}, message string) {
	tc.T.Helper()
	if value == nil || reflect.ValueOf(value).IsNil() {
		tc.T.Errorf("%s: expected non-nil value", message)
	}
}

// AssertTrue asserts that a condition is true
func (tc *TestContext) AssertTrue(condition bool, message string) {
	tc.T.Helper()
	if !condition {
		tc.T.Errorf("%s: expected true, got false", message)
	}
}

// AssertFalse asserts that a condition is false
func (tc *TestContext) AssertFalse(condition bool, message string) {
	tc.T.Helper()
	if condition {
		tc.T.Errorf("%s: expected false, got true", message)
	}
}

// AssertError asserts that an error occurred
func (tc *TestContext) AssertError(err error, message string) {
	tc.T.Helper()
	if err == nil {
		tc.T.Errorf("%s: expected error, got nil", message)
	}
}

// AssertNoError asserts that no error occurred
func (tc *TestContext) AssertNoError(err error, message string) {
	tc.T.Helper()
	if err != nil {
		tc.T.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertContains asserts that a string contains a substring
func (tc *TestContext) AssertContains(haystack, needle string, message string) {
	tc.T.Helper()
	if !contains(haystack, needle) {
		tc.T.Errorf("%s: expected '%s' to contain '%s'", message, haystack, needle)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

// indexString finds the index of a substring
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestDataGenerator generates test data
type TestDataGenerator struct{}

// GenerateTestFunction creates a test function definition
func (tdg *TestDataGenerator) GenerateTestFunction(name string, funcType types.FunctionType) *types.Function {
	return &types.Function{
		Name:     name,
		Functype: funcType,
		Content:  "test content",
		Inputs:   []types.Input{},
		Outputs:  []types.Output{},
	}
}

// GenerateTestInput creates a test input definition
func (tdg *TestDataGenerator) GenerateTestInput(name string, source types.InputSource, datatype types.DataType) types.Input {
	return types.Input{
		Name:         name,
		Aliasname:    name,
		Source:       source,
		Datatype:     datatype,
		Defaultvalue: "",
		Inivalue:     "",
		List:         false,
	}
}

// GenerateTestOutput creates a test output definition
func (tdg *TestDataGenerator) GenerateTestOutput(name string, datatype types.DataType) types.Output {
	return types.Output{
		Name:     name,
		Datatype: datatype,
	}
}

// GenerateSession creates test session data
func (tdg *TestDataGenerator) GenerateSession() map[string]interface{} {
	return map[string]interface{}{
		"UserNo":   "testuser",
		"ClientID": "testclient",
		"TenantID": "testtenant",
		"SessionID": "test-session-123",
	}
}

// MockDatabase provides a mock database for testing
type MockDatabase struct {
	Queries     []string
	Results     []map[string]interface{}
	Errors      []error
	QueryCount  int
	CommitCount int
	RollbackCount int
	InTransaction bool
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Queries:     make([]string, 0),
		Results:     make([]map[string]interface{}, 0),
		Errors:      make([]error, 0),
		QueryCount:  0,
		CommitCount: 0,
		RollbackCount: 0,
		InTransaction: false,
	}
}

// AddQueryResult adds a query result to return
func (md *MockDatabase) AddQueryResult(result map[string]interface{}) {
	md.Results = append(md.Results, result)
}

// AddError adds an error to return on next query
func (md *MockDatabase) AddError(err error) {
	md.Errors = append(md.Errors, err)
}

// Query simulates a database query
func (md *MockDatabase) Query(query string, args ...interface{}) (map[string]interface{}, error) {
	md.Queries = append(md.Queries, query)
	md.QueryCount++

	if len(md.Errors) > 0 {
		err := md.Errors[0]
		md.Errors = md.Errors[1:]
		return nil, err
	}

	if len(md.Results) > 0 {
		result := md.Results[0]
		md.Results = md.Results[1:]
		return result, nil
	}

	return make(map[string]interface{}), nil
}

// BeginTransaction simulates starting a transaction
func (md *MockDatabase) BeginTransaction() error {
	if md.InTransaction {
		return fmt.Errorf("already in transaction")
	}
	md.InTransaction = true
	return nil
}

// Commit simulates committing a transaction
func (md *MockDatabase) Commit() error {
	if !md.InTransaction {
		return fmt.Errorf("not in transaction")
	}
	md.CommitCount++
	md.InTransaction = false
	return nil
}

// Rollback simulates rolling back a transaction
func (md *MockDatabase) Rollback() error {
	if !md.InTransaction {
		return fmt.Errorf("not in transaction")
	}
	md.RollbackCount++
	md.InTransaction = false
	return nil
}

// Reset resets the mock database state
func (md *MockDatabase) Reset() {
	md.Queries = make([]string, 0)
	md.Results = make([]map[string]interface{}, 0)
	md.Errors = make([]error, 0)
	md.QueryCount = 0
	md.CommitCount = 0
	md.RollbackCount = 0
	md.InTransaction = false
}

// BenchmarkHelper provides utilities for benchmarking
type BenchmarkHelper struct {
	B *testing.B
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{B: b}
}

// Run runs a benchmark with setup and teardown
func (bh *BenchmarkHelper) Run(name string, setup func(), teardown func(), fn func()) {
	bh.B.Run(name, func(b *testing.B) {
		if setup != nil {
			setup()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fn()
		}
		b.StopTimer()

		if teardown != nil {
			teardown()
		}
	})
}

// MeasureMemory measures memory allocation
func (bh *BenchmarkHelper) MeasureMemory(fn func()) {
	bh.B.ReportAllocs()
	for i := 0; i < bh.B.N; i++ {
		fn()
	}
}

// IntegrationTestSuite provides a suite for integration tests
type IntegrationTestSuite struct {
	T           *testing.T
	DBConnected bool
	DBConn      *sql.DB
	TestCtx     *TestContext
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		T:           t,
		DBConnected: false,
		TestCtx:     NewTestContext(t),
	}
}

// SetupDatabase sets up a test database
func (its *IntegrationTestSuite) SetupDatabase(connectionString string) error {
	// This would connect to an actual test database
	// For now, we'll skip actual DB connection in tests
	its.DBConnected = true
	return nil
}

// TeardownDatabase cleans up the test database
func (its *IntegrationTestSuite) TeardownDatabase() {
	if its.DBConn != nil {
		its.DBConn.Close()
	}
	its.DBConnected = false
}

// RunTest runs a test with setup and teardown
func (its *IntegrationTestSuite) RunTest(name string, test func(*TestContext)) {
	its.T.Run(name, func(t *testing.T) {
		ctx := NewTestContext(t)
		defer ctx.RunCleanup()

		test(ctx)
	})
}

// TableDrivenTest provides utilities for table-driven tests
type TableDrivenTest struct {
	T *testing.T
}

// NewTableDrivenTest creates a new table-driven test helper
func NewTableDrivenTest(t *testing.T) *TableDrivenTest {
	return &TableDrivenTest{T: t}
}

// TestCase represents a single test case
type TestCase struct {
	Name     string
	Input    interface{}
	Expected interface{}
	Error    bool
	Skip     bool
}

// Run runs table-driven tests
func (tdt *TableDrivenTest) Run(testCases []TestCase, testFunc func(interface{}) (interface{}, error)) {
	for _, tc := range testCases {
		tdt.T.Run(tc.Name, func(t *testing.T) {
			if tc.Skip {
				t.Skip("Test case marked as skip")
				return
			}

			result, err := testFunc(tc.Input)

			if tc.Error {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if !reflect.DeepEqual(result, tc.Expected) {
					t.Errorf("Expected %v, got %v", tc.Expected, result)
				}
			}
		})
	}
}

// PerformanceTest provides utilities for performance testing
type PerformanceTest struct {
	T              *testing.T
	MaxDuration    time.Duration
	MaxMemoryMB    int64
	EnableProfiling bool
}

// NewPerformanceTest creates a new performance test
func NewPerformanceTest(t *testing.T, maxDuration time.Duration, maxMemoryMB int64) *PerformanceTest {
	return &PerformanceTest{
		T:           t,
		MaxDuration: maxDuration,
		MaxMemoryMB: maxMemoryMB,
	}
}

// Run runs a performance test
func (pt *PerformanceTest) Run(name string, fn func()) {
	pt.T.Run(name, func(t *testing.T) {
		start := time.Now()

		fn()

		duration := time.Since(start)

		if duration > pt.MaxDuration {
			t.Errorf("Performance test '%s' took %v, exceeding maximum of %v",
				name, duration, pt.MaxDuration)
		} else {
			t.Logf("Performance test '%s' completed in %v", name, duration)
		}
	})
}

// Global test data generator
var TestGen = &TestDataGenerator{}
