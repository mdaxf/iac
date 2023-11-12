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
	}{}
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

func TestJSFuncs_Testfunction(t *testing.T) {
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
		{
			name: "test",
			cf:   &JSFuncs{},
			args: args{
				content: `var result;
				var x = 3;
		
				if (x > 0) {
					result = "Positive";
				} else if (x < 0) {
					result = "Negative";
				} else {
					result = "Zero";
				}
		
				result;`,
				inputs: map[string]interface{}{
					"test": "test",
				},
				outputs: []string{"result"},
			},
			want: map[string]interface{}{
				"result": "Positive",
			},
			wantErr: false,
		},
		{
			name: "test1",
			cf:   &JSFuncs{},
			args: args{
				content: `DataDecimal = 0.0; DataInt = 0; DataBoolean = false; DataString = ""; 							 					
				if(DataType == 1 || DataType == "1")
				  DataInt = parseInt(Value)
				else if(DataType == 2)
						DataDecimal = parseFloat(Value);
				else if(DataType == 3)
						DataBoolean = (Value == "true" || Value== "1")? true: false;
				else
				  DataString = Value;`,
				inputs: map[string]interface{}{
					"DataType": 1,
					"Value":    "456",
				},
				outputs: []string{"DataInt", "DataDecimal", "DataBoolean"},
			},
			want: map[string]interface{}{
				"DataInt":     456,
				"DataDecimal": 0.0,
				"DataBoolean": false,
			},
			wantErr: false,
		},
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
