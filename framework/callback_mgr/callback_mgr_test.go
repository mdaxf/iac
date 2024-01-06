package callback_mgr

import (
	"reflect"
	"testing"
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

func TestCallBackFunc(t *testing.T) {
	type args struct {
		key  string
		args []interface{}
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
			if got, _ := CallBackFunc(tt.args.key, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CallBackFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
