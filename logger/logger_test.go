package logger

import (
	"testing"

	"github.com/mdaxf/iac/framework/logs"
)

func TestInit(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.args.config)
		})
	}
}

func Test_setLogger(t *testing.T) {
	type args struct {
		loger   *logs.IACLogger
		config  map[string]interface{}
		logtype string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLogger(tt.args.loger, tt.args.config, tt.args.logtype)
		})
	}
}

func TestLog_Debug(t *testing.T) {
	type args struct {
		logmsg string
	}
	tests := []struct {
		name string
		l    *Log
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Debug(tt.args.logmsg)
		})
	}
}

func TestLog_Info(t *testing.T) {
	type args struct {
		logmsg string
	}
	tests := []struct {
		name string
		l    *Log
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Info(tt.args.logmsg)
		})
	}
}

func TestLog_Error(t *testing.T) {
	type args struct {
		logmsg string
	}
	tests := []struct {
		name string
		l    *Log
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Error(tt.args.logmsg)
		})
	}
}
