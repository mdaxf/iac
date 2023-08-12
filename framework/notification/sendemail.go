package email

import (
	"fmt"
	"strings"

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

func SendEmail(emailConfig EmailConfiguration, to []string, subject string, body string) error {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Notification"}
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

func SendEmailWithAttachment(emailConfig EmailConfiguration, to []string, subject string, body string, attachment string) error {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Notification"}
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
