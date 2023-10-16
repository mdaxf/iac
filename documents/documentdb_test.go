package documents

import "testing"

func TestConnectDB(t *testing.T) {
	type args struct {
		DatabaseType       string
		DatabaseConnection string
		DatabaseName       string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConnectDB(tt.args.DatabaseType, tt.args.DatabaseConnection, tt.args.DatabaseName)
		})
	}
}
