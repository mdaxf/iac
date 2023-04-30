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
	"log"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

func Query(querystr string, args ...interface{}) (*sql.Rows, error) {
	fmt.Println(string(querystr))
	//fmt.Println(string(args))
	stmt, err := DB.Prepare(querystr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows, nil
}

func QuerybyList(querystr string, namelist []string, valuelist []string, dbTx ...*sql.Tx) (map[string][]interface{}, error) {
	// create a slice to hold the parameter values in the same order as they appear in the SQL query
	var values []interface{}

	// Execute the SQL statement with the given inputs
	for i := range namelist {
		paramPlaceholder := "@" + namelist[i]
		paramValuePlaceholder := fmt.Sprintf("'%v'", valuelist[i])
		querystr = strings.Replace(querystr, paramPlaceholder, paramValuePlaceholder, -1)
		values = append(values, valuelist[i])
	}

	idbTx := append(dbTx, nil)[0]
	var stmt *sql.Stmt

	if idbTx != nil {
		stmt, err := idbTx.Prepare(querystr)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

	} else {
		// Prepare the SQL statement with placeholders
		stmt, err := DB.Prepare(querystr)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()
	}

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return Conto_JsonbyList(rows)
}

func Query_Json(querystr string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := Query(querystr, args...)
	if err != nil {
		return nil, err
	}
	return Conto_Json(rows)
}
func Conto_JsonbyList(rows *sql.Rows) (map[string][]interface{}, error) {
	cols, err := rows.ColumnTypes()
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
		for i, name := range colNames {
			data[name] = append(data[name], *(values[i].(*interface{})))
		}

	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return data, nil

}
func Conto_Json(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.ColumnTypes()
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
		for i, name := range colNames {
			row[name] = *(values[i].(*interface{}))
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return data, nil
}

func ExecSP(procedureName string, args ...interface{}) error {
	// Construct the stored procedure call with placeholders for each parameter
	placeholders := make([]string, len(args))
	for i := range args {
		placeholders[i] = "?"
	}
	call := fmt.Sprintf("CALL %s(%s)", procedureName, strings.Join(placeholders, ","))

	// Call the stored procedure with the dynamic parameters
	_, err := DB.Exec(call, args...)
	if err != nil {
		return err
	}
	return nil
}

func ExeSPwithRow(procedureName string, args ...interface{}) (*sql.Rows, error) {
	// Construct the stored procedure call with placeholders for each parameter and the output parameter
	var outputparameters []string
	placeholders := make([]string, len(args))
	for i := range args {
		output, parameter := chechoutputparameter(args[i].(string))
		if output {
			outputparameters = append(outputparameters, parameter)
		}
		placeholders[i] = "?"
	}
	//placeholders = append(placeholders, "@output_param")
	call := fmt.Sprintf("CALL %s(%s)", procedureName, strings.Join(placeholders, ","))

	// Call the stored procedure with the dynamic parameters and the output parameter

	rows, err := DB.Query(call, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func ExecSP_Json(procedureName string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ExeSPwithRow(procedureName, args...)
	if err != nil {
		return nil, err
	}
	return Conto_Json(rows)
}

func chechoutputparameter(str string) (bool, string) {
	output := false
	parameter := str
	if strings.Contains(str, " output") {
		parts := strings.Split(str, " ")
		output = true
		parameter = parts[0]

	}
	return output, parameter

}

func TableInsert(TableName string, Columns []string, Values []interface{}) (int64, error) {
	var querystr string
	var args []interface{}
	querystr = "INSERT INTO " + TableName + "(" + strings.Join(Columns, ",") + ") VALUES (" + strings.Repeat("?,", len(Columns)-1) + "?)"
	args = Values
	fmt.Println(querystr)
	fmt.Println(args)
	stmt, err := DB.Prepare(querystr)
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return lastId, err
}

func TableUpdate(TableName string, Columns []string, Values []interface{}, Where string, WhereArgs []interface{}) (int64, error) {

	fmt.Println(WhereArgs)
	fmt.Println(Values)
	var querystr string
	var args []interface{}

	switch DatabaseType {
	case "sqlserver":
		setPlaceholders := make([]string, len(Columns))
		for i, column := range Columns {
			setPlaceholders[i] = fmt.Sprintf("%s = '%s'", column, Values[i])
		}
		setClause := strings.Join(setPlaceholders, ", ")
		querystr := fmt.Sprintf("UPDATE %s SET %s WHERE %s", TableName, setClause, Where)
		args = []interface{}{}
		fmt.Println(querystr)
		fmt.Println(args)
		stmt, err := DB.Prepare(querystr)

		res, err := stmt.Exec(args...)

		rowcount, err := res.RowsAffected()
		return rowcount, err

	default:
	case "mysql":

		querystr = "UPDATE " + TableName + " SET " + strings.Join(Columns, "=?,") + "=? WHERE " + Where
		args = append(Values, WhereArgs...)
		fmt.Println(querystr)
		fmt.Println(args)

		stmt, err := DB.Prepare(querystr)
		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(args...)
		if err != nil {
			log.Fatal(err)
		}
		rowcount, err := res.RowsAffected()
		return rowcount, err
	}
	return 0, nil
}

func TableDelete(TableName string, Where string, WhereArgs []interface{}) (int64, error) {
	var querystr string
	var args []interface{}
	querystr = "DELETE FROM " + TableName + " WHERE " + Where
	args = WhereArgs
	fmt.Println(querystr)
	fmt.Println(args)
	stmt, err := DB.Prepare(querystr)
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return lastId, err
}
