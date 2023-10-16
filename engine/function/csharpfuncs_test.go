package funcs

import (
	"reflect"
	"testing"
)

func TestCSharpFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *CSharpFuncs
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

func Testfunction(t *testing.T) {
	type args struct {
		content string
		inputs  interface{}
		outputs []string
	}
	tests := []struct {
		name    string
		cf      *CSharpFuncs
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cf.Testfunction(tt.args.content, tt.args.inputs, tt.args.outputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CSharpFuncs.Testfunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CSharpFuncs.Testfunction() = %v, want %v", got, tt.want)
			}
		})
	}
}
