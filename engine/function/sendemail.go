package funcs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sendemail "github.com/mdaxf/iac/framework/notification"
)

type EmailFuncs struct{}

type EmailStru struct {
	ToEmailList  []string
	FromEmail    string
	Subject      string
	Body         string
	Attachment   string
	SmtpServer   string
	SmtpPort     int
	SmtpUser     string
	SmtpPassword string
}

func (cf *EmailFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmail.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendEmail.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendEmail.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("EmailFuncs Execute"))

	_, email := cf.Validate(f)

	ecfg := sendemail.EmailConfiguration{
		Host:     email.SmtpServer,
		Port:     email.SmtpPort,
		Username: email.SmtpUser,
		Password: email.SmtpPassword,
		From:     email.FromEmail,
	}

	if email.Attachment != "" {
		sendemail.SendEmailWithAttachment(ecfg, email.ToEmailList, email.Subject, email.Body, email.Attachment)
		return
	} else {
		sendemail.SendEmail(ecfg, email.ToEmailList, email.Subject, email.Body)
	}
}

func (cf *EmailFuncs) Validate(f *Funcs) (bool, EmailStru) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmail.Validate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendEmail.Validate with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendEmail.Validate with error: %s", err)
			return
		}
	}()

	namelist, valuelist, _ := f.SetInputs()

	email := EmailStru{}

	for i, name := range namelist {
		value := valuelist[i]

		if name == "ToEmails" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "ToEmail is empty"))

			} else {
				email.ToEmailList = strings.Split(value, ";")
				if len(email.ToEmailList) == 0 {
					f.iLog.Debug(fmt.Sprintf("Error: %v", "ToEmail is empty"))

				}
			}
		}
		if name == "FromEmail" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "FromEmail is empty"))

			} else {
				email.FromEmail = value
			}
		}
		if name == "Subject" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Subject is empty"))

			} else {
				email.Subject = value
			}
		}

		if name == "Body" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Body is empty"))

			} else {
				email.Body = value
			}
		}

		if name == "SmtpServer" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpServer is empty"))

			} else {
				email.SmtpServer = value
			}
		}

		if name == "SmtpPort" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpPort is empty"))

			} else {
				var err error
				email.SmtpPort, err = strconv.Atoi(value)
				if err != nil {
					f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpPort is not a number"))
					email.SmtpPort = 0
				}
			}
		}

		if name == "SmtpUser" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpUser is empty"))

			} else {
				email.SmtpUser = value
			}
		}

		if name == "SmtpPassword" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpPassword is empty"))

			} else {
				email.SmtpPassword = value
			}
		}

		if name == "Attachment" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Attachment is empty"))

			} else {
				email.Attachment = value
			}
		}

	}

	if len(email.ToEmailList) == 0 || email.Subject == "" || email.Body == "" {
		f.iLog.Debug(fmt.Sprintf("Error: %v", "ToEmails or Subject or Body is empty"))
		return false, email
	} else {
		return true, email
	}
}
