package function

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExecFunction(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		f    *FunctionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.TestExecFunction(tt.args.c)
		})
	}
}
