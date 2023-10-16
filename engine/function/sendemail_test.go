package funcs

import (
	"reflect"
	"testing"
)

func TestEmailFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *EmailFuncs
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

func TestEmailFuncs_Validate(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name  string
		cf    *EmailFuncs
		args  args
		want  bool
		want1 EmailStru
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.cf.Validate(tt.args.f)
			if got != tt.want {
				t.Errorf("EmailFuncs.Validate() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("EmailFuncs.Validate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
