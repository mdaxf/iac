package funcs

import "testing"

func TestTableUpdateFuncs_Execute(t *testing.T) {
	type args struct {
		f *Funcs
	}
	tests := []struct {
		name string
		cf   *TableUpdateFuncs
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
