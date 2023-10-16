package databaseop

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDBController_GetDatabyQuery(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		db   *DBController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.db.GetDatabyQuery(tt.args.ctx)
		})
	}
}

func TestDBController_GetDataFromTables(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		db   *DBController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.db.GetDataFromTables(tt.args.ctx)
		})
	}
}

func TestDBController_getDataStructForQuery(t *testing.T) {
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.db.getDataStructForQuery(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBController.getDataStructForQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBController.getDataStructForQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DBController.getDataStructForQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestDBController_getmysqlsubtabls(t *testing.T) {
	type args struct {
		tablename  string
		data       map[string]interface{}
		markasJson bool
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.db.getmysqlsubtabls(tt.args.tablename, tt.args.data, tt.args.markasJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBController.getmysqlsubtabls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBController.getmysqlsubtabls() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DBController.getmysqlsubtabls() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestDBController_getsubtabls(t *testing.T) {
	type args struct {
		tablename  string
		data       map[string]interface{}
		markasJson bool
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.getsubtabls(tt.args.tablename, tt.args.data, tt.args.markasJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBController.getsubtabls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBController.getsubtabls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBController_InsertDataToTable(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.InsertDataToTable(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DBController.InsertDataToTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBController_UpdateDataToTable(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.UpdateDataToTable(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DBController.UpdateDataToTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBController_DeleteDataFromTable(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.db.DeleteDataFromTable(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DBController.DeleteDataFromTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBController_GetDataFromRequest(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		db      *DBController
		args    args
		want    DBData
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.GetDataFromRequest(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBController.GetDataFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBController.GetDataFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
