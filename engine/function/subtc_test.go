package funcs

import (
	"reflect"
	"testing"
)

func TestSubTranCodeFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *SubTranCodeFuncs
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

func Test_convertSliceToMap(t *testing.T) {
	type args struct {
		slice []interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertSliceToMap(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertSliceToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
