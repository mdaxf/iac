package funcs

import (
	"reflect"
	"testing"
	"time"
)

func TestFuncs_HandleInputs(t *testing.T) {
	tests := []struct {
		name    string
		f       *Funcs
		want    []string
		want1   []string
		want2   map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := tt.f.HandleInputs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Funcs.HandleInputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Funcs.HandleInputs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Funcs.HandleInputs() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Funcs.HandleInputs() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestFuncs_SetInputs(t *testing.T) {
	tests := []struct {
		name  string
		f     *Funcs
		want  []string
		want1 []string
		want2 map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := tt.f.SetInputs()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Funcs.SetInputs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Funcs.SetInputs() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Funcs.SetInputs() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_isArray(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isArray(tt.args.value); got != tt.want {
				t.Errorf("isArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_checkifRepeatExecution(t *testing.T) {
	tests := []struct {
		name    string
		f       *Funcs
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.checkifRepeatExecution()
			if (err != nil) != tt.wantErr {
				t.Errorf("Funcs.checkifRepeatExecution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Funcs.checkifRepeatExecution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_checkinputvalue(t *testing.T) {
	type args struct {
		Aliasname string
		variables map[string]interface{}
	}
	tests := []struct {
		name    string
		f       *Funcs
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.checkinputvalue(tt.args.Aliasname, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("Funcs.checkinputvalue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Funcs.checkinputvalue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_customMarshal(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := customMarshal(tt.args.v); got != tt.want {
				t.Errorf("customMarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_ConverttoInt(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ConverttoInt(tt.args.str); got != tt.want {
				t.Errorf("Funcs.ConverttoInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_ConverttoFloat(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ConverttoFloat(tt.args.str); got != tt.want {
				t.Errorf("Funcs.ConverttoFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_ConverttoBool(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ConverttoBool(tt.args.str); got != tt.want {
				t.Errorf("Funcs.ConverttoBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_ConverttoDateTime(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ConverttoDateTime(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Funcs.ConverttoDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_SetOutputs(t *testing.T) {
	type args struct {
		outputs map[string]interface{}
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.SetOutputs(tt.args.outputs)
		})
	}
}

func TestFuncs_SetfuncOutputs(t *testing.T) {
	tests := []struct {
		name string
		f    *Funcs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.SetfuncOutputs()
		})
	}
}

func TestFuncs_SetfuncSingleOutputs(t *testing.T) {
	type args struct {
		outputs map[string]interface{}
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.SetfuncSingleOutputs(tt.args.outputs)
		})
	}
}

func TestFuncs_ConvertfromBytes(t *testing.T) {
	type args struct {
		bytesbuffer []byte
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ConvertfromBytes(tt.args.bytesbuffer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Funcs.ConvertfromBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFuncs_Execute(t *testing.T) {
	tests := []struct {
		name string
		f    *Funcs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Execute()
		})
	}
}

func TestFuncs_CancelExecution(t *testing.T) {
	type args struct {
		errormessage string
	}
	tests := []struct {
		name string
		f    *Funcs
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.CancelExecution(tt.args.errormessage)
		})
	}
}
