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
	"database/sql"
	"fmt"
	"strings"

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

func NewDBOperation(User string, DBTx *sql.Tx, moduleName string) *DBOperation {
	if moduleName == "" {
		moduleName = logger.Database
	}
	return &DBOperation{
		DBTx:       DBTx,
		ModuleName: moduleName,
		iLog:       logger.Log{ModuleName: moduleName, User: User, ControllerName: "Database"},
		User:       User,
	}
}

func (db *DBOperation) Query(querystr string, args ...interface{}) (*sql.Rows, error) {

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
		db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err.Error()))
		return nil, err
	}
	defer rows.Close()

	if blocaltx {
		idbtx.Commit()
	}

	return rows, nil
}

func (db *DBOperation) QuerybyList(querystr string, namelist []string, inputs map[string]interface{}, finputs []types.Input) (map[string][]interface{}, error) {

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
		values = append(values, inputs[namelist[i]])
	}

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

	var stmt *sql.Stmt

	stmt, err := idbtx.Prepare(querystr)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to prepare the query: %s with error: %s", querystr, err.Error()))
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(values...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the query: %s with error: %s", querystr, err.Error()))
		return nil, err
	}
	defer rows.Close()

	if blocaltx {
		idbtx.Commit()
	}

	return db.Conto_JsonbyList(rows)
}

func (db *DBOperation) Query_Json(querystr string, args ...interface{}) ([]map[string]interface{}, error) {
	db.iLog.Debug(fmt.Sprintf("Query with json object result: %s %s...", querystr, args))
	rows, err := db.Query(querystr, args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to query database with error: %s", err.Error()))
		return nil, err
	}
	return db.Conto_Json(rows)
}

func (db *DBOperation) ExecSP(procedureName string, args ...interface{}) error {

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

	// Call the stored procedure with the dynamic parameters
	_, err := idbtx.Exec(call, args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return nil
}

func (db *DBOperation) ExeSPwithRow(procedureName string, args ...interface{}) (*sql.Rows, error) {
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

	rows, err := idbtx.Query(call, args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return nil, err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return rows, nil
}

func (db *DBOperation) ExecSP_Json(procedureName string, args ...interface{}) ([]map[string]interface{}, error) {
	db.iLog.Debug(fmt.Sprintf("start execute the Store procedure: %s with parameters %s...", procedureName, args))
	rows, err := db.ExeSPwithRow(procedureName, args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the Store procedure: %s with parameters %s with error: %s", procedureName, args, err.Error()))
		return nil, err
	}
	return db.Conto_Json(rows)
}

func (db *DBOperation) chechoutputparameter(str string) (bool, string) {
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

func (db *DBOperation) TableInsert(TableName string, Columns []string, Values []string) (int64, error) {
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

	var querystr string

	args := make([]interface{}, len(Values))
	querystr = "INSERT INTO " + TableName + "(" + strings.Join(Columns, ",") + ") VALUES (" + strings.Repeat("?,", len(Columns)-1) + "?)"

	for i, s := range Values {
		args[i] = s
	}

	fmt.Println(querystr)
	fmt.Println(args)
	stmt, err := idbtx.Prepare(querystr)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to prepare the insert statement with error: %s", err.Error()))
		return 0, err
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the insert statement with error: %s", err.Error()))
		return 0, err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the last insert id with error: %s", err.Error()))
	}

	if blocaltx {
		idbtx.Commit()
	}

	return lastId, err
}

func (db *DBOperation) TableUpdate(TableName string, Columns []string, Values []string, datatypes []int, Where string) (int64, error) {

	db.iLog.Debug(fmt.Sprintf("start to update the table: %s with columns: %s and values: %s data type: %s", TableName, Columns, Values, datatypes))

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

		stmt, err := DB.Prepare(querystr)
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to prepare the update statement with error: %s", err.Error()))
			idbtx.Rollback()
			return 0, err
		}

		res, err := stmt.Exec(args...)
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

		stmt, err := idbtx.Prepare(querystr)
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to prepare the update statement with error: %s", err.Error()))
			return 0, err
		}
		res, err := stmt.Exec(args...)
		if err != nil {
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

func (db *DBOperation) TableDelete(TableName string, Where string) (int64, error) {

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

	var querystr string
	var args []interface{}
	querystr = "DELETE FROM " + TableName + " WHERE " + Where

	db.iLog.Debug(fmt.Sprintf("The delete query string is: %s  parametrs: %s...", querystr, args))

	//fmt.Println(querystr)
	//fmt.Println(args)
	stmt, err := idbtx.Prepare(querystr)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to prepare the delete statement with error: %s", err.Error()))
		return 0, err
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to execute the delete statement with error: %s", err.Error()))
		return 0, err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the last insert id with error: %s", err.Error()))
		return 0, err
	}

	if blocaltx {
		idbtx.Commit()
	}

	return lastId, err
}

func (db *DBOperation) Conto_JsonbyList(rows *sql.Rows) (map[string][]interface{}, error) {

	db.iLog.Debug(fmt.Sprintf("Start to convert the rows to json...%s", rows))
	cols, err := rows.ColumnTypes()
	if err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the column types with error: %s", err.Error()))
		return nil, err
	}
	data := make(map[string][]interface{})
	colNames := make([]string, len(cols))

	for i, col := range cols {
		colNames[i] = col.Name()
		data[col.Name()] = []interface{}{}
	}

	for rows.Next() {
		values := make([]interface{}, len(colNames))
		for i := range values {
			values[i] = new(interface{})
		}
		err := rows.Scan(values...)
		if err != nil {
			db.iLog.Debug(fmt.Sprintf("There is error to scan the row with error: %s", err.Error()))
			return nil, err

		}
		for i, name := range colNames {
			data[name] = append(data[name], *(values[i].(*interface{})))
		}

	}

	if err := rows.Err(); err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the rows with error: %s", err.Error()))
	}

	db.iLog.Debug(fmt.Sprintf("The result of the conversion is: %s", data))

	return data, nil

}
func (db *DBOperation) Conto_Json(rows *sql.Rows) ([]map[string]interface{}, error) {

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
	data := []map[string]interface{}{}

	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(colNames))
		for i := range values {
			values[i] = new(interface{})
		}
		err := rows.Scan(values...)
		if err != nil {
			db.iLog.Error(fmt.Sprintf("There is error to scan the row with error: %s", err.Error()))
			return nil, err

		}
		for i, name := range colNames {
			row[name] = *(values[i].(*interface{}))
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		db.iLog.Error(fmt.Sprintf("There is error to get the rows with error: %s", err.Error()))
	}
	db.iLog.Debug(fmt.Sprintf("The result of the conversion is: %s", data))

	return data, nil
}
