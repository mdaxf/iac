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
	"strings"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

type DBOperation struct {
	DBTx       *sql.Tx
	ModuleName string
	iLog       logger.Log
	User       string
}

// NewDBOperation creates a new instance of DBOperation.
// It takes the following parameters:
// - User: the name of the user performing the database operation.
// - DBTx: the SQL transaction object.
// - moduleName: the name of the module associated with the database operation.
// If moduleName is empty, it defaults to logger.Database.
// It returns a pointer to the newly created DBOperation instance.
// The function also logs the performance duration of the operation.
// If there is an error during the operation, it is recovered and logged as an error.
// The function returns a pointer to the newly created DBOperation instance.

func NewDBOperation(User string, DBTx *sql.Tx, moduleName string) *DBOperation {
	startTime := time.Now()
	if moduleName == "" {
		moduleName = logger.Database
	}
	iLog := logger.Log{ModuleName: moduleName, User: User, ControllerName: "Database"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("dbconn.NewDBOperation", elapsed)
	}()

	return &DBOperation{
		DBTx:       DBTx,
		ModuleName: moduleName,
		iLog:       iLog,
		User:       User,
	}
}

// Query executes a SQL query with optional arguments and returns the resulting rows.
// It measures the performance of the query and logs any errors that occur.
// If a database transaction is not already in progress, it begins a new transaction.
// The transaction is committed if it was started locally.
// The returned rows must be closed after use to release associated resources.
// The function returns the following parameters:
// - rows: the resulting rows.
// - err: the error that occurred during the operation.
// If there is an error during the operation, it is recovered and logged as an error.
// The function returns the resulting rows and any error that occurred during the operation.
// The function also logs the performance duration of the operation.

func (db *DBOperation) Query(querystr string, args ...interface{}) (*sql.Rows, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.Query", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err))
			return
		}
	}()

	db.iLog.Debug(fmt.Sprintf("Query: %s %s...", querystr, args))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return nil, err
		}
		defer idbtx.Commit()
	}

	//fmt.Println(string(args))
	stmt, err := idbtx.Prepare(querystr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {

		idbtx.Rollback()

		db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err.Error()))
		return nil, err
	}
	defer rows.Close()

	if blocaltx {
		idbtx.Commit()
	}

	return rows, nil
}

// QuerybyList executes a database query with a list of parameters.
// It takes a query string, a list of parameter names, a map of parameter values, and a list of input types.
// The function returns a map of query results, the number of rows affected, the number of rows returned, and an error (if any).
// The function also logs the performance duration of the operation.
// If there is an error during the operation, it is recovered and logged as an error.

func (db *DBOperation) QuerybyList(querystr string, namelist []string, inputs map[string]interface{}, finputs []types.Input) (map[string][]interface{}, int, int, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.QuerybyList", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err))
			return
		}
	}()

	db.iLog.Debug(fmt.Sprintf("Query: %s {%s} {%s}", querystr, namelist, inputs))

	// create a slice to hold the parameter values in the same order as they appear in the SQL query
	var values []interface{}

	// Execute the SQL statement with the given inputs
	for i := range namelist {
		paramPlaceholder := "@" + namelist[i]
		paramValuePlaceholder := ""
		switch finputs[i].Datatype {
		case types.Integer:
			paramValuePlaceholder = fmt.Sprintf("%d", inputs[namelist[i]])
		case types.Float:
			paramValuePlaceholder = fmt.Sprintf("%f", inputs[namelist[i]])
		case types.Bool:
			paramValuePlaceholder = fmt.Sprintf("%t", inputs[namelist[i]])
		default:
			paramValuePlaceholder = fmt.Sprintf("'%v'", inputs[namelist[i]])
		}
		querystr = strings.Replace(querystr, paramPlaceholder, paramValuePlaceholder, -1)
		//	values = append(values, inputs[namelist[i]])
	}

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return nil, 0, 0, err
		}
		defer idbtx.Commit()
	}

	var stmt *sql.Stmt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()
	stmt, err = idbtx.PrepareContext(ctx, querystr)
	//stmt, err := idbtx.Prepare(querystr)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to prepare the query: %s with error: %s", querystr, err.Error()))
		return nil, 0, 0, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, values...)
	//	rows, err := stmt.Query(values...)
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to execute the query: %s with error: %s", querystr, err.Error()))
		return nil, 0, 0, err
	}
	defer rows.Close()

	if blocaltx {
		idbtx.Commit()
	}

	return db.Conto_JsonbyList(rows)
}

// Query_Json executes a database query with the provided query string and arguments,
// and returns the result as a slice of maps, where each map represents a row of the result set.
// The query is executed within a transaction, and the transaction is automatically committed
// if it was initiated locally. If an error occurs during the query or conversion to JSON,
// the transaction is rolled back and the error is returned.
//
// Parameters:
// - querystr: The query string to execute.
// - args: Optional arguments to be passed to the query.
//
// Returns:
// - []map[string]interface{}: The result set as a slice of maps.
// - error: Any error that occurred during the query or conversion to JSON.

func (db *DBOperation) Query_Json(querystr string, args ...interface{}) ([]map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.Query_Json", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err))
			return
		}
	}()

	db.iLog.Debug(fmt.Sprintf("Query with json object result: %s %s...", querystr, args))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return nil, err
		}
		defer idbtx.Commit()
	}

	//fmt.Println(string(args))
	//stmt, err := idbtx.Prepare(querystr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()
	stmt, err := idbtx.PrepareContext(ctx, querystr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	//rows, err := stmt.Query(args...)

	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err.Error()))
		return nil, err
	}
	defer rows.Close()

	db.iLog.Debug(fmt.Sprintf("Query with json object result:%v...", rows))
	jsondata, err := db.Conto_Json(rows)
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to convert the rows to json with error: %s", err.Error()))
		return nil, err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return jsondata, nil
}

// ExecSP executes a stored procedure with the given procedureName and arguments.
// It measures the execution time and logs the performance.
// If an error occurs during execution, it logs the error and rolls back the transaction.
// If a local transaction is used, it commits the transaction at the end.
// Parameters:
//   - procedureName: the name of the stored procedure to execute
//   - args: the arguments to pass to the stored procedure
// Returns:
//   - error: an error if there was a problem executing the stored procedure

func (db *DBOperation) ExecSP(procedureName string, args ...interface{}) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.ExecSP", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute store procedure %s in database with error: %s", procedureName, err))
			return
		}
	}()
	db.iLog.Debug(fmt.Sprintf("start execute the Store procedure: %s with parameters %s...", procedureName, args))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return err
		}
		defer idbtx.Commit()
	}

	// Construct the stored procedure call with placeholders for each parameter
	placeholders := make([]string, len(args))
	for i := range args {
		placeholders[i] = "?"
	}
	call := fmt.Sprintf("CALL %s(%s)", procedureName, strings.Join(placeholders, ","))

	db.iLog.Debug(fmt.Sprintf("Call the stored procedure %s with the dynamic parameters %s...", call, args))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()

	// Call the stored procedure with the dynamic parameters
	_, err := idbtx.ExecContext(ctx, call, args...)
	//_, err := idbtx.Exec(call, args...)
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return nil
}

// ExeSPwithRow executes a stored procedure and returns the result set as a *sql.Rows object.
// It takes the procedureName as a string and the args as variadic parameters.
// The function measures the execution time and logs it using the PerformanceWithDuration method of the db.iLog object.
// If there is an error during execution, it logs the error using the Error method of the db.iLog object.
// The function handles panics and recovers from them, logging the error if any.
// If a database transaction is not already in progress, it begins a new transaction and commits it at the end.
// The function constructs the stored procedure call with placeholders for each parameter and the output parameter.
// It checks if any of the parameters are output parameters and stores their names in the outputparameters slice.
// The function uses the call string to call the stored procedure with the dynamic parameters and the output parameter.
// It returns the result set as a *sql.Rows object and nil error if successful, otherwise it returns nil and the error.

func (db *DBOperation) ExeSPwithRow(procedureName string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.ExeSPwithRow", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute store procedure %s in database with error: %s", procedureName, err))
			return
		}
	}()

	// Construct the stored procedure call with placeholders for each parameter and the output parameter
	db.iLog.Debug(fmt.Sprintf("start execute the Store procedure to return rows: %s with parameters %s...", procedureName, args))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return nil, err
		}
		defer idbtx.Commit()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()

	var outputparameters []string
	placeholders := make([]string, len(args))
	for i := range args {
		output, parameter := db.chechoutputparameter(args[i].(string))
		if output {
			outputparameters = append(outputparameters, parameter)
		}
		placeholders[i] = "?"
	}
	//placeholders = append(placeholders, "@output_param")
	call := fmt.Sprintf("CALL %s(%s)", procedureName, strings.Join(placeholders, ","))

	// Call the stored procedure with the dynamic parameters and the output parameter

	rows, err := idbtx.QueryContext(ctx, call, args...)
	//rows, err := idbtx.Query(call, args...)
	defer rows.Close()

	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return nil, err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return rows, nil
}

// ExecSP_Json executes a stored procedure with the given name and arguments,
// and returns the result as a slice of maps, where each map represents a row
// of the result set. If an error occurs during execution, it returns nil and
// the error.
//
// The execution time of the stored procedure is logged using the PerformanceWithDuration
// method of the associated logger.
//
// If a panic occurs during execution, it is recovered and logged as an error.
//
// The execution of the stored procedure is logged using the Debug method of the
// associated logger.
//
// The result set is obtained by calling the ExeSPwithRow method of the DBOperation
// instance, and the rows are closed before returning.
//
// If an error occurs during execution, it is logged as an error and returned.
//
// The result set is converted to a JSON representation using the Conto_Json method
// of the DBOperation instance, and returned.

func (db *DBOperation) ExecSP_Json(procedureName string, args ...interface{}) ([]map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.ExecSP_Json", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				db.iLog.Error(fmt.Sprintf("There is error to execute store procedure %s in database with error: %s", procedureName, err))
				return
			}
		}()
	*/
	db.iLog.Debug(fmt.Sprintf("start execute the Store procedure: %s with parameters %s...", procedureName, args))
	rows, err := db.ExeSPwithRow(procedureName, args...)
	defer rows.Close()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return nil, err
	}
	return db.Conto_Json(rows)
}

// chechoutputparameter checks the output parameter of a given string.
// It splits the string by space and determines if it contains the word "output".
// If it does, it sets the output flag to true and returns the parameter without the word "output".
// The function also logs debug messages for the start and result of the check.
// It measures the performance duration of the function using the iLog.PerformanceWithDuration method.
// If there is a panic during the execution, it logs an error message with the error details.
// The function returns a boolean value indicating if the string contains an output parameter,
// and the parameter itself without the word "output".

func (db *DBOperation) chechoutputparameter(str string) (bool, string) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.chechoutputparameter", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				db.iLog.Error(fmt.Sprintf("There is error to chechoutputparameter with error: %s", err))
				return
			}
		}()
	*/
	db.iLog.Debug(fmt.Sprintf("start to check the output parameter: %s...", str))
	output := false
	parameter := str
	if strings.Contains(str, " output") {
		parts := strings.Split(str, " ")
		output = true
		parameter = parts[0]

	}
	db.iLog.Debug(fmt.Sprintf("the output parameter: %s is %s...", parameter, output))
	return output, parameter

}

// TableInsert inserts data into a specified table in the database.
// It takes the table name, column names, and corresponding values as input.
// It returns the last insert ID and any error encountered during the operation.
// The function measures the performance duration of the operation using the PerformanceWithDuration method of the db.iLog object.
func (db *DBOperation) TableInsert(TableName string, Columns []string, Values []string) (int64, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.TableInsert", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute table %s insert data with error: %s", TableName, err))
			return
		}
	}()

	db.iLog.Debug(fmt.Sprintf("start to insert the table: %s with columns: %s and values: %s...", TableName, Columns, Values))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return 0, err
		}
		defer idbtx.Commit()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()

	var querystr string

	args := make([]interface{}, len(Values))
	querystr = "INSERT INTO " + TableName + "(" + strings.Join(Columns, ",") + ") VALUES (" + strings.Repeat("?,", len(Columns)-1) + "?)"

	for i, s := range Values {
		args[i] = s
	}

	fmt.Println(querystr)
	fmt.Println(args)
	stmt, err := idbtx.PrepareContext(ctx, querystr)
	//	stmt, err := idbtx.Prepare(querystr)
	defer stmt.Close()
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to prepare the insert statement with error: %s", err.Error()))
		return 0, err
	}
	res, err := stmt.ExecContext(ctx, args...)
	//res, err := stmt.Exec(args...)

	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to execute the insert statement with error: %s", err.Error()))
		return 0, err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to get the last insert id with error: %s", err.Error()))
	}

	if blocaltx {
		idbtx.Commit()
	}

	return lastId, err
}

// TableUpdate updates the specified table with the given columns, values, data types, and WHERE clause.
// It returns the number of rows affected and any error encountered during the update operation.
// The function measures the performance duration of the operation using the PerformanceWithDuration method of the db.iLog object.
// If there is a panic during the execution, it logs an error message with the error details.
func (db *DBOperation) TableUpdate(TableName string, Columns []string, Values []string, datatypes []int, Where string) (int64, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.TableUpdate", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute table %s update data with error: %s", TableName, err))
			return
		}
	}()
	db.iLog.Debug(fmt.Sprintf("start to update the table: %s with columns: %s and values: %s data type: %v", TableName, Columns, Values, datatypes))

	//fmt.Println(WhereArgs)
	//fmt.Println(Values)
	var querystr string
	var args []interface{}

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return 0, err
		}
		defer idbtx.Commit()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()

	switch DatabaseType {
	case "sqlserver":
		setPlaceholders := make([]string, len(Columns))
		for i, column := range Columns {

			switch datatypes[i] {
			case int(types.Integer):
				setPlaceholders[i] = fmt.Sprintf("%s = %d", column, Values[i])
			case int(types.Float):
				setPlaceholders[i] = fmt.Sprintf("%s = %f", column, Values[i])

			case int(types.Bool):
				setPlaceholders[i] = fmt.Sprintf("%s = %t", column, Values[i])

			default:
				setPlaceholders[i] = fmt.Sprintf("%s = '%v'", column, Values[i])

			}

			//	setPlaceholders[i] = fmt.Sprintf("%s = '%s'", column, Values[i])
		}
		setClause := strings.Join(setPlaceholders, ", ")
		querystr := fmt.Sprintf("UPDATE %s SET %s WHERE %s", TableName, setClause, Where)
		args = []interface{}{}

		db.iLog.Debug(fmt.Sprintf("The update query string is: %s  parametrs: %s...", querystr, args))

		stmt, err := DB.PrepareContext(ctx, querystr)
		//		stmt, err := DB.Prepare(querystr)
		defer stmt.Close()
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to prepare the update statement with error: %s", err.Error()))
			idbtx.Rollback()
			return 0, err
		}

		res, err := stmt.ExecContext(ctx, args...)
		//res, err := stmt.Exec(args...)
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute the update statement with error: %s", err.Error()))
			idbtx.Rollback()
			return 0, err
		}

		rowcount, err := res.RowsAffected()
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to get the affected rows with error: %s", err.Error()))
			idbtx.Rollback()
			return 0, err
		}

		if blocaltx {
			idbtx.Commit()
		}

		return rowcount, err

	default:
		//	case "mysql":

		querystr = "UPDATE " + TableName + " SET " + strings.Join(Columns, "=?,") + "=? WHERE " + Where

		args := make([]interface{}, len(Values))

		for i, s := range Values {
			args[i] = s
		}

		//fmt.Println(querystr)
		//fmt.Println(args)
		db.iLog.Debug(fmt.Sprintf("The update query string is: %s  parametrs: %s...", querystr, args))

		stmt, err := idbtx.PrepareContext(ctx, querystr)
		//stmt, err := idbtx.Prepare(querystr)
		defer stmt.Close()
		if err != nil {
			idbtx.Rollback()
			db.iLog.Error(fmt.Sprintf("There is error to prepare the update statement with error: %s", err.Error()))
			return 0, err
		}
		res, err := stmt.ExecContext(ctx, args...)
		//res, err := stmt.Exec(args...)
		if err != nil {
			idbtx.Rollback()
			db.iLog.Error(fmt.Sprintf("There is error to execute the update statement with error: %s", err.Error()))
			return 0, err
		}
		rowcount, err := res.RowsAffected()

		if blocaltx {
			idbtx.Commit()
		}

		return rowcount, err
	}

}

// TableDelete deletes records from a table based on the provided WHERE clause.
// It returns the number of affected rows and an error, if any.
// The function measures the performance duration of the operation using the PerformanceWithDuration method of the db.iLog object.
func (db *DBOperation) TableDelete(TableName string, Where string) (int64, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.TableDelete", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to execute table %s delete with error: %s", TableName, err))
			return
		}
	}()

	db.iLog.Debug(fmt.Sprintf("Start to delete the table: %s with where: %s and whereargs: ", TableName, Where))

	idbtx := db.DBTx
	blocaltx := false

	if idbtx == nil {
		idbtx, err = DB.Begin()
		blocaltx = true
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to begin database transaction with error: %s", err.Error()))
			return 0, err
		}
		defer idbtx.Commit()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(com.DBTransactionTimeout))
	defer cancel()

	var querystr string
	var args []interface{}
	querystr = "DELETE FROM " + TableName + " WHERE " + Where

	db.iLog.Debug(fmt.Sprintf("The delete query string is: %s  parametrs: %s...", querystr, args))

	//fmt.Println(querystr)
	//fmt.Println(args)
	stmt, err := idbtx.PrepareContext(ctx, querystr)
	//	stmt, err := idbtx.Prepare(querystr)
	defer stmt.Close()
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to prepare the delete statement with error: %s", err.Error()))
		return 0, err
	}
	res, err := stmt.ExecContext(ctx, args...)
	//	res, err := stmt.Exec(args...)
	if err != nil {
		idbtx.Rollback()
		db.iLog.Error(fmt.Sprintf("There is error to execute the delete statement with error: %s", err.Error()))
		return 0, err
	}
	lastId, err := res.RowsAffected()
	if err != nil {
		//idbtx.Commit()
		db.iLog.Error(fmt.Sprintf("There is error to get the last insert id with error: %s", err.Error()))
		//	return 0, err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return lastId, err
}

func (db *DBOperation) Conto_JsonbyList(rows *sql.Rows) (map[string][]interface{}, int, int, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.Conto_JsonbyList", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				db.iLog.Error(fmt.Sprintf("There is error to Conto_JsonbyList with error: %s", err))
				return
			}
		}()
	*/
	db.iLog.Debug(fmt.Sprintf("Start to convert the rows to json...%s", rows))
	cols, err := rows.ColumnTypes()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the column types with error: %s", err.Error()))
		return nil, 0, 0, err
	}
	data := make(map[string][]interface{})
	colNames := make([]string, len(cols))
	valuetmps := make([]interface{}, len(colNames))

	ColumnNumbers := 0
	for i, col := range cols {
		colNames[i] = col.Name()
		data[col.Name()] = []interface{}{}
		ColumnNumbers = ColumnNumbers + 1
	}

	RowNumbers := 0
	for rows.Next() {
		values := make([]interface{}, len(colNames))
		for i := range values {
			//values[i] = new(interface{})
			values[i] = &valuetmps[i]
		}
		err := rows.Scan(values...)
		if err != nil {
			db.iLog.Debug(fmt.Sprintf("There is error to scan the row with error: %s", err.Error()))
			return nil, 0, 0, err

		}
		for i, name := range colNames {

			var v interface{}

			val := valuetmps[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			//data[name] = append(data[name], *(values[i].(*interface{})))
			data[name] = append(data[name], v)
		}
		RowNumbers = RowNumbers + 1
	}

	if err := rows.Err(); err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the rows with error: %s", err.Error()))
	}

	db.iLog.Debug(fmt.Sprintf("The result of the conversion is: %s", data))

	return data, ColumnNumbers, RowNumbers, nil

}
func (db *DBOperation) Conto_Json(rows *sql.Rows) ([]map[string]interface{}, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		db.iLog.PerformanceWithDuration("dbconn.Conto_Json", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				db.iLog.Error(fmt.Sprintf("There is error to Conto_Json with error: %s", err))
				return
			}
		}()
	*/
	db.iLog.Debug(fmt.Sprintf("Start to convert the rows to json...%s", rows))

	cols, err := rows.ColumnTypes()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the column types with error: %s", err.Error()))
		return nil, err

	}

	colNames := make([]string, len(cols))
	for i, col := range cols {
		colNames[i] = col.Name()
	}
	data := make([]map[string]interface{}, 0)
	db.iLog.Debug(fmt.Sprintf("The column names are: %s", colNames))
	//	db.iLog.Debug(fmt.Sprintf("rows : %s", rows))
	valuetmps := make([]interface{}, len(colNames))

	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(colNames))
		for i := range values {

			values[i] = &valuetmps[i]
		}

		err := rows.Scan(values...)
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to scan the row with error: %s", err.Error()))
			return nil, err

		}
		//	db.iLog.Debug(fmt.Sprintf("The values of the row is: %s", values))
		for i, name := range colNames {
			var v interface{}
			val := valuetmps[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			//	db.iLog.Debug(fmt.Sprintf("The row field %s is: %s", name, v))
			row[name] = v
			//row[name] = *(values[i].(*interface{}))
			//	db.iLog.Debug(fmt.Sprintf("The row field %s is: %s", name, row[name]))
		}
		db.iLog.Debug(fmt.Sprintf("The row is: %s", row))
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the rows with error: %s", err.Error()))
	}
	db.iLog.Debug(fmt.Sprintf("The result of the conversion is: %s", data))
	//jsondata, err := json.Marshal(data)
	return data, nil
}
