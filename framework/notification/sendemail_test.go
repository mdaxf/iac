package email

import (
	"testing"
)

func TestSendEmail(t *testing.T) {
	type args struct {
		emailConfig EmailConfiguration
		to          []string
		subject     string
		body        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendEmail(tt.args.emailConfig, tt.args.to, tt.args.subject, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendEmailWithAttachment(t *testing.T) {
	type args struct {
		emailConfig EmailConfiguration
		to          []string
		subject     string
		body        string
		attachment  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendEmailWithAttachment(tt.args.emailConfig, tt.args.to, tt.args.subject, tt.args.body, tt.args.attachment); (err != nil) != tt.wantErr {
				t.Errorf("SendEmailWithAttachment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
