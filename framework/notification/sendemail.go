package email

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/mdaxf/iac/logger"
)

type EmailConfiguration struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SendEmail sends an email using the provided email configuration, recipient(s), subject, and body.
// It returns an error if the email sending fails.
// The email body is expected to be in HTML format.
func SendEmail(emailConfig EmailConfiguration, to []string, subject string, body string) error {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Notification"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("framework.email.SendEmail", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in framework.email.SendEmail: %s", r))
			return
		}
	}()

	log.Debug(fmt.Sprintf("Send the notification by email to: %s  ", strings.Join(to, ",")))

	m := gomail.NewMessage()
	m.SetHeader("From", emailConfig.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(emailConfig.Host, emailConfig.Port, emailConfig.Username, emailConfig.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Error(fmt.Sprintf("Failed to send email:%s", err.Error()))
		return err
	}
	return nil
}

// SendEmailWithAttachment sends an email with an attachment.
// It takes the following parameters:
// - emailConfig: EmailConfiguration struct containing email server configuration details.
// - to: a slice of strings representing the email recipients.
// - subject: a string representing the email subject.
// - body: a string representing the email body in HTML format.
// - attachment: a string representing the file path of the attachment.
// It returns an error if the email fails to send.

func SendEmailWithAttachment(emailConfig EmailConfiguration, to []string, subject string, body string, attachment string) error {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Notification"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("framework.email.SendEmailWithAttachment", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in framework.email.SendEmailWithAttachment: %s", r))
			return
		}
	}()
	log.Debug(fmt.Sprintf("Send the notification by email to: %s  ", strings.Join(to, ",")))

	m := gomail.NewMessage()
	m.SetHeader("From", emailConfig.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	m.Attach(attachment)

	d := gomail.NewDialer(emailConfig.Host, emailConfig.Port, emailConfig.Username, emailConfig.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Error(fmt.Sprintf("Failed to send email:%s", err.Error()))
		return err
	}
	return nil
}
