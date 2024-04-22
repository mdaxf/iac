package callback

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac-signalr/signalr"
)

func TestRegisterCallBack(t *testing.T) {
	type args struct {
		key      string
		callBack interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterCallBack(tt.args.key, tt.args.callBack)
		})
	}
}

func TestExecuteTranCode(t *testing.T) {
	type args struct {
		key       string
		tcode     string
		inputs    map[string]interface{}
		ctx       context.Context
		ctxcancel context.CancelFunc
		dbTx      *sql.Tx
		DBCon     *documents.DocDB
		sc        signalr.Client
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExecuteTranCode(tt.args.key, tt.args.tcode, tt.args.inputs, tt.args.ctx, tt.args.ctxcancel, tt.args.dbTx, tt.args.DBCon, tt.args.sc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecuteTranCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
