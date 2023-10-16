package funcs

import (
	"testing"
)

func TestInputMapFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *InputMapFuncs
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

func TestInputMapFuncs_Validate(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *InputMapFuncs
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cf.Validate(tt.args.f); got != tt.want {
				t.Errorf("InputMapFuncs.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
