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
	"reflect"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mdaxf/iac/engine/types"
)

func TestDBOperation_Query(t *testing.T) {
	type args struct {
		querystr string
		args     []interface{}
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    *sql.Rows
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Query(tt.args.querystr, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_QuerybyList(t *testing.T) {
	type args struct {
		querystr string
		namelist []string
		inputs   map[string]interface{}
		finputs  []types.Input
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    map[string][]interface{}
		want1   int
		want2   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := tt.db.QuerybyList(tt.args.querystr, tt.args.namelist, tt.args.inputs, tt.args.finputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.QuerybyList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.QuerybyList() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DBOperation.QuerybyList() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("DBOperation.QuerybyList() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestDBOperation_Query_Json(t *testing.T) {
	type args struct {
		querystr string
		args     []interface{}
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Query_Json(tt.args.querystr, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.Query_Json() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.Query_Json() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_ExecSP(t *testing.T) {
	type args struct {
		procedureName string
		args          []interface{}
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.ExecSP(tt.args.procedureName, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.ExecSP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBOperation_ExeSPwithRow(t *testing.T) {
	type args struct {
		procedureName string
		args          []interface{}
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    *sql.Rows
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.ExeSPwithRow(tt.args.procedureName, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.ExeSPwithRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.ExeSPwithRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_ExecSP_Json(t *testing.T) {
	type args struct {
		procedureName string
		args          []interface{}
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.ExecSP_Json(tt.args.procedureName, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.ExecSP_Json() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.ExecSP_Json() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_chechoutputparameter(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		db    *DBOperation
		args  args
		want  bool
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.db.chechoutputparameter(tt.args.str)
			if got != tt.want {
				t.Errorf("DBOperation.chechoutputparameter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DBOperation.chechoutputparameter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestDBOperation_TableInsert(t *testing.T) {
	type args struct {
		TableName string
		Columns   []string
		Values    []string
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.TableInsert(tt.args.TableName, tt.args.Columns, tt.args.Values)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.TableInsert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBOperation.TableInsert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_TableUpdate(t *testing.T) {
	type args struct {
		TableName string
		Columns   []string
		Values    []string
		datatypes []int
		Where     string
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.TableUpdate(tt.args.TableName, tt.args.Columns, tt.args.Values, tt.args.datatypes, tt.args.Where)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.TableUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBOperation.TableUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_TableDelete(t *testing.T) {
	type args struct {
		TableName string
		Where     string
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.TableDelete(tt.args.TableName, tt.args.Where)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.TableDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBOperation.TableDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBOperation_Conto_JsonbyList(t *testing.T) {
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    map[string][]interface{}
		want1   int
		want2   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := tt.db.Conto_JsonbyList(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.Conto_JsonbyList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.Conto_JsonbyList() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DBOperation.Conto_JsonbyList() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("DBOperation.Conto_JsonbyList() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestDBOperation_Conto_Json(t *testing.T) {
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		db      *DBOperation
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.Conto_Json(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBOperation.Conto_Json() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBOperation.Conto_Json() = %v, want %v", got, tt.want)
			}
		})
	}
}
