package workflow

import "testing"

func TestExplodionEngine_Explode(t *testing.T) {
	explosion := NewExplosion("NCR Flow", "NCR", "NCR", "Sys", "abcd")

	tests := []struct {
		name    string
		e       *ExplodionEngine
		wantErr bool
	}{
		{
			name:    "Test Case 1",
			e:       explosion,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Explode(); (err != nil) != tt.wantErr {
				t.Errorf("ExplodionEngine.Explode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
