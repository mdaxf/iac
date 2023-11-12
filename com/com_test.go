package com

import (
	"reflect"
	"testing"
)

type Person struct {
	Name    string
	Age     int
	Address string
}

type Message struct {
	Id      string
	UUID    string
	Retry   int
	Execute int
	Topic   string
	PayLoad interface{}
	Handler string
}

func Test_ConvertstructToMap(t *testing.T) {
	person := Person{
		Name:    "John Doe",
		Age:     30,
		Address: "123 Main St",
	}
	msg := Message{
		Id:      "1",
		UUID:    "2",
		Retry:   3,
		Execute: 4,
		Topic:   "iac/gauge/values",
		PayLoad: `[{
            "gauge":"Gauge001",
            "parameter":"CircuitBreaker",
			"value":"1"
            },
            {
            "gauge":"Gauge001",
            "parameter":"ElectricMeter",
			"value":"110"
            }]`,
		Handler: "test",
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{name: "test1",
			args:    args{input: person},
			want:    map[string]interface{}{"Name": "John Doe", "Age": 30, "Address": "123 Main St"},
			wantErr: false,
		},
		{name: "test2",
			args: args{input: msg},
			want: map[string]interface{}{"Id": "1", "UUID": "2", "Retry": 3, "Execute": 4, "Topic": "iac/gauge/values", "PayLoad": `[{
			}]`, "Handler": "test"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertstructToMap(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertstructToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Convertbase64ToMap(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				input: "W3sKICAgICAgICAgICAgImdhdWdlIjoiR2F1Z2UwMDEiLAogICAgICAgICAgICAicGFyYW1ldGVyIjoiQ2lyY3VpdEJyZWFrZXIiLAoJCQkidmFsdWUiOiIxIgogICAgICAgICAgICB9LAogICAgICAgICAgICB7CiAgICAgICAgICAgICJnYXVnZSI6IkdhdWdlMDAxIiwKICAgICAgICAgICAgInBhcmFtZXRlciI6IkVsZWN0cmljTWV0ZXIiLAoJCQkidmFsdWUiOiIxMTAiCiAgICAgICAgICAgIH1dCg==",
			},
			want: map[string]interface{}{
				"gauge":     "Gauge001",
				"parameter": "CircuitBreaker",
				"value":     "1",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				input: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convertbase64ToMap(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convertbase64ToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Convertbase64ToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConvertbytesToMap(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				data: []byte(`{"gauge":"Gauge001","parameter":"CircuitBreaker","value":"1"}`),
			},
			want: map[string]interface{}{
				"gauge":     "Gauge001",
				"parameter": "CircuitBreaker",
				"value":     "1",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				data: []byte{91, 123, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 34, 103, 97, 117, 103, 101, 34, 58, 34, 71, 97, 117, 103, 101, 48, 48, 49, 34, 44, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 34, 112, 97, 114, 97, 109, 101, 116, 101, 114, 34, 58, 34, 67, 105, 114, 99, 117, 105, 116, 66, 114, 101, 97, 107, 101, 114, 34, 44, 10, 9, 9, 9, 34, 118, 97, 108, 117, 101, 34, 58, 34, 49, 34, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 125, 44, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 123, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 34, 103, 97, 117, 103, 101, 34, 58, 34, 71, 97, 117, 103, 101, 48, 48, 49, 34, 44, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 34, 112, 97, 114, 97, 109, 101, 116, 101, 114, 34, 58, 34, 69, 108, 101, 99, 116, 114, 105, 99, 77, 101, 116, 101, 114, 34, 44, 10, 9, 9, 9, 34, 118, 97, 108, 117, 101, 34, 58, 34, 49, 49, 48, 34, 10, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 125, 93, 10},
			},
			want: map[string]interface{}{
				"gauge":     "Gauge001",
				"parameter": "CircuitBreaker",
				"value":     "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertbytesToMap(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertbytesToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertbytesToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
