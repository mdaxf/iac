package workflow

import (
	"testing"
	// "github.com/mdaxf/iac/controllers/databaseop"
)

func TestExplodionEngine_Explode(t *testing.T) {
	explosion := NewExplosion("NCR Flow", "NCR", "NCR", "Sys", "abcd")
	data := make(map[string]interface{})
	tests := []struct {
		name        string
		e           *ExplodionEngine
		description string
		data        map[string]interface{}
		wantErr     bool
	}{
		{
			name:        "Test Case 1",
			e:           explosion,
			description: "Test",
			data:        data,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Explode(tt.description, tt.data); (err != nil) != tt.wantErr {
				t.Errorf("ExplodionEngine.Explode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
