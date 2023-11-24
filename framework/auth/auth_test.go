package auth

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenerate_authentication_token(t *testing.T) {
	type args struct {
		userID    string
		loginName string
		ClientID  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := Generate_authentication_token(tt.args.userID, tt.args.loginName, tt.args.ClientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate_authentication_token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Generate_authentication_token() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Generate_authentication_token() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Generate_authentication_token() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetUserInformation(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := GetUserInformation(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserInformation() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetUserInformation() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetUserInformation() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateToken(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtendexptime(t *testing.T) {
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := Extendexptime(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extendexptime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Extendexptime() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Extendexptime() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Extendexptime() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name string
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AuthMiddleware(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_protectedHandler(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protectedHandler(tt.args.c)
		})
	}
}
