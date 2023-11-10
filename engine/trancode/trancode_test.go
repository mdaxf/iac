package trancode

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/signalrsrv/signalr"
)

func TestExecutebyExternal(t *testing.T) {
	type args struct {
		trancode string
		data     map[string]interface{}
		DBTx     *sql.Tx
		DBCon    *documents.DocDB
		sc       signalr.Client
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExecutebyExternal(tt.args.trancode, tt.args.data, tt.args.DBTx, tt.args.DBCon, tt.args.sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecutebyExternal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecutebyExternal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranFlow_Execute(t *testing.T) {
	tests := []struct {
		name    string
		tr      *TranFlow
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("TranFlow.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranFlow.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranFlow_getFGbyName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		tr    *TranFlow
		args  args
		want  types.FuncGroup
		want1 int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tr.getFGbyName(tt.args.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranFlow.getFGbyName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TranFlow.getFGbyName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetTransCode(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    types.TranCode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTransCode(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytetoobj(t *testing.T) {
	type args struct {
		config []byte
	}
	tests := []struct {
		name    string
		args    args
		want    types.TranCode
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				config: []byte(`{
					"code": "test",
					"functiongroups": [
						{
							"name": "test",
							"functions": [
								{
									"name": "test",
									"function": "test",
									"inputs": [
										{
											"name": "test",
											"type": "test",
											"required": true,
											"validation": "test",
											"validationmessage": "test",
											"validationtype": "test",
											"validationvalue": "test",
											"validationvalue2": "test",
											"validationvalue3": "test",
										}]
								}]
						}]
				}`),
			},
			want:    types.TranCode{},
			wantErr: true, // TODO: Add test cases
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Bytetoobj(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytetoobj() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bytetoobj() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigtoobj(t *testing.T) {
	type args struct {
		config string
	}
	tests := []struct {
		name    string
		args    args
		want    types.TranCode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Configtoobj(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Configtoobj() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Configtoobj() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranFlowstr_Execute(t *testing.T) {
	type args struct {
		tcode     string
		inputs    map[string]interface{}
		ctx       context.Context
		ctxcancel context.CancelFunc
		dbTx      []*sql.Tx
	}
	tests := []struct {
		name    string
		tr      *TranFlowstr
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.Execute(tt.args.tcode, tt.args.inputs, tt.args.ctx, tt.args.ctxcancel, tt.args.dbTx...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TranFlowstr.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranFlowstr.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTranCodeData(t *testing.T) {
	type args struct {
		Code string
	}
	tests := []struct {
		name    string
		args    args
		want    types.TranCode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTranCodeData(tt.args.Code)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTranCodeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTranCodeData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTranCodeData(t *testing.T) {
	type args struct {
		Code   string
		DBConn *documents.DocDB
	}
	tests := []struct {
		name    string
		args    args
		want    types.TranCode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTranCodeData(tt.args.Code, tt.args.DBConn)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTranCodeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTranCodeData() = %v, want %v", got, tt.want)
			}
		})
	}
}
