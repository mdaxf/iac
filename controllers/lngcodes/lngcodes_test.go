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
		db         *dbconn.DBOperation
		lngcodeid  int64
		text       string
		languageid int64
		User       string
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
			if err := tt.f.insertlngcontent(tt.args.db, tt.args.lngcodeid, tt.args.text, tt.args.languageid, tt.args.User); (err != nil) != tt.wantErr {
				t.Errorf("LCController.insertlngcode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLCController_updatelngcodbyid(t *testing.T) {
	type args struct {
		db         *dbconn.DBOperation
		id         int
		text       string
		languageid int
		User       string
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
			if err := tt.f.updatelngcontent(tt.args.db, tt.args.id, tt.args.text, int64(tt.args.languageid), tt.args.User); (err != nil) != tt.wantErr {
				t.Errorf("LCController.updatelngcodbyid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLCController_populatesinglelngcodes(t *testing.T) {
	type args struct {
		db         *dbconn.DBOperation
		lngcode    string
		text       string
		languageid int64
		User       string
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
			tt.f.populatesinglelngcodes(tt.args.db, tt.args.lngcode, tt.args.text, tt.args.languageid, tt.args.User)
		})
	}
}

func TestLCController_populatelngcodes(t *testing.T) {
	type args struct {
		lngcodes   []string
		text       []string
		languageid int64
		User       string
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
			tt.f.populatelngcodes(tt.args.lngcodes, tt.args.text, tt.args.languageid, tt.args.User)
		})
	}
}
