package health

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCheckSystemHealth(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	r := gin.Default()
	r.Run(":8080")

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TestCheckSystemHealth",
			args: args{
				c: &gin.Context{},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckSystemHealth(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSystemHealth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckSystemHealth() = %v, want %v", got, tt.want)
			}
		})
	}
}
