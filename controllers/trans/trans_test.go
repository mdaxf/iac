package trans

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/engine/types"
)

func TestTranCodeController_ExecuteTranCode(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		e    *TranCodeController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e.ExecuteTranCode(tt.args.ctx)
		})
	}
}

func TestTranCodeController_Execute(t *testing.T) {
	type args struct {
		Code           string
		externalinputs map[string]interface{}
		user           string
		clientid       string
	}
	tests := []struct {
		name    string
		e       *TranCodeController
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Execute(tt.args.Code, tt.args.externalinputs, tt.args.user, tt.args.clientid)
			if (err != nil) != tt.wantErr {
				t.Errorf("TranCodeController.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranCodeController.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranCodeController_getTransCode(t *testing.T) {
	type args struct {
		name     string
		user     string
		clientid string
	}
	tests := []struct {
		name    string
		e       *TranCodeController
		args    args
		want    types.TranCode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.getTransCode(tt.args.name, tt.args.user, tt.args.clientid)
			if (err != nil) != tt.wantErr {
				t.Errorf("TranCodeController.getTransCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranCodeController.getTransCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranCodeController_GetTranCodeListFromRespository(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		e    *TranCodeController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e.GetTranCodeListFromRespository(tt.args.ctx)
		})
	}
}

func TestTranCodeController_GetTranCodeDetailFromRespository(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		e    *TranCodeController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e.GetTranCodeDetailFromRespository(tt.args.ctx)
		})
	}
}

func Test_getDataFromRequest(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		args    args
		want    TranCodeData
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDataFromRequest(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDataFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
