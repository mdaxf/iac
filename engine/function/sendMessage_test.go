package funcs

import (
	"testing"
)

func TestSendMessageFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *SendMessageFuncs
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cf.Execute(tt.args.f)
		})
	}
}

func TestSendMessageFuncs_Validate(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name    string
		cf      *SendMessageFuncs
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cf.Validate(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessageFuncs.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SendMessageFuncs.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
