package engine

import "testing"

func TestEngine_Execute(t *testing.T) {
	tests := []struct {
		name string
		e    *Engine
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e.Execute()
		})
	}
}
