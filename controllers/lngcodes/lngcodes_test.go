package lngcodes

import (
	"testing"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
)

func TestLCController_GetLngCodes(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		f    *LCController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.GetLngCodes(tt.args.c)
		})
	}
}

func TestLCController_UpdateLngCode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		f    *LCController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.UpdateLngCode(tt.args.c)
		})
	}
}

func TestLCController_insertlngcode(t *testing.T) {
	type args struct {
		db       *dbconn.DBOperation
		lngcode  string
		text     string
		language string
		User     string
	}
	tests := []struct {
		name    string
		f       *LCController
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.insertlngcode(tt.args.db, tt.args.lngcode, tt.args.text, tt.args.language, tt.args.User); (err != nil) != tt.wantErr {
				t.Errorf("LCController.insertlngcode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLCController_updatelngcodbyid(t *testing.T) {
	type args struct {
		db       *dbconn.DBOperation
		id       int
		lngcode  string
		text     string
		language string
		User     string
	}
	tests := []struct {
		name    string
		f       *LCController
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.updatelngcodbyid(tt.args.db, tt.args.id, tt.args.lngcode, tt.args.text, tt.args.language, tt.args.User); (err != nil) != tt.wantErr {
				t.Errorf("LCController.updatelngcodbyid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLCController_populatesinglelngcodes(t *testing.T) {
	type args struct {
		db       *dbconn.DBOperation
		lngcode  string
		text     string
		language string
		User     string
	}
	tests := []struct {
		name string
		f    *LCController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.populatesinglelngcodes(tt.args.db, tt.args.lngcode, tt.args.text, tt.args.language, tt.args.User)
		})
	}
}

func TestLCController_populatelngcodes(t *testing.T) {
	type args struct {
		lngcodes []string
		text     []string
		language string
		User     string
	}
	tests := []struct {
		name string
		f    *LCController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.populatelngcodes(tt.args.lngcodes, tt.args.text, tt.args.language, tt.args.User)
		})
	}
}
