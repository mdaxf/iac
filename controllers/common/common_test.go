package common

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRequestBody(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequestBody(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequestBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRequestBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapToStruct(t *testing.T) {
	type args struct {
		data      map[string]interface{}
		outStruct interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mapToStruct(tt.args.data, tt.args.outStruct); (err != nil) != tt.wantErr {
				t.Errorf("mapToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
