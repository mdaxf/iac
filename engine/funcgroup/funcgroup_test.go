package funcgroup

import (
	"testing"

	"github.com/mdaxf/iac/engine/types"
)

func TestFGroup_CheckRouter(t *testing.T) {
	type args struct {
		RouterDef types.RouterDef
	}
	tests := []struct {
		name string
		c    *FGroup
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.CheckRouter(tt.args.RouterDef); got != tt.want {
				t.Errorf("FGroup.CheckRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}
