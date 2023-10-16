package funcs

import (
	"reflect"
	"testing"
)

func TestJSFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *JSFuncs
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

func TestJSFuncs_Validate(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name    string
		cf      *JSFuncs
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
				t.Errorf("JSFuncs.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSFuncs.Validate() = %v, want %v", got, tt.want)
			}
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
		cf      *JSFuncs
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
				t.Errorf("JSFuncs.Testfunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSFuncs.Testfunction() = %v, want %v", got, tt.want)
			}
		})
	}
}
