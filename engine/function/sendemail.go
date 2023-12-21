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

// Execute executes the SendEmail function.
// It validates the email configuration, sets up the email parameters,
// and sends the email using the provided SMTP server.
// If an attachment is specified, it sends the email with the attachment,
// otherwise it sends the email without any attachment.
func (cf *EmailFuncs) Execute(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance duration
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmail.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics and handle the error
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendEmail.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.SendEmail.Execute with error: %s", err))
			// set the error message
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendEmail.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("EmailFuncs Execute"))

	// validate the email configuration
	_, email := cf.Validate(f)

	// set up the email configuration
	ecfg := sendemail.EmailConfiguration{
		Host:     email.SmtpServer,
		Port:     email.SmtpPort,
		Username: email.SmtpUser,
		Password: email.SmtpPassword,
		From:     email.FromEmail,
	}

	// send the email with or without attachment
	if email.Attachment != "" {
		sendemail.SendEmailWithAttachment(ecfg, email.ToEmailList, email.Subject, email.Body, email.Attachment)
		return
	} else {
		sendemail.SendEmail(ecfg, email.ToEmailList, email.Subject, email.Body)
	}
}

// Validate validates the inputs for sending an email.
// It checks if the required fields (ToEmails, Subject, Body) are not empty.
// If any of the required fields are empty, it returns false along with the empty email structure.
// Otherwise, it returns true along with the populated email structure.
func (cf *EmailFuncs) Validate(f *Funcs) (bool, EmailStru) {
	// function execution time measurement
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmail.Validate", elapsed)
	}()

	/*	// recover from any panics and log the error
		defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is an error in engine.funcs.SendEmail.Validate: %s", err))
				f.ErrorMessage = fmt.Sprintf("There is an error in engine.funcs.SendEmail.Validate: %s", err)
				return
			}
		}()
	*/
	// get the input values
	namelist, valuelist, _ := f.SetInputs()

	// create an empty email structure
	email := EmailStru{}

	// iterate over the input names and values
	for i, name := range namelist {
		value := valuelist[i]

		// check if the name is "ToEmails"
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

		// check if the name is "FromEmail"
		if name == "FromEmail" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "FromEmail is empty"))
			} else {
				email.FromEmail = value
			}
		}

		// check if the name is "Subject"
		if name == "Subject" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Subject is empty"))
			} else {
				email.Subject = value
			}
		}

		// check if the name is "Body"
		if name == "Body" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Body is empty"))
			} else {
				email.Body = value
			}
		}

		// check if the name is "SmtpServer"
		if name == "SmtpServer" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpServer is empty"))
			} else {
				email.SmtpServer = value
			}
		}

		// check if the name is "SmtpPort"
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

		// check if the name is "SmtpUser"
		if name == "SmtpUser" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpUser is empty"))
			} else {
				email.SmtpUser = value
			}
		}

		// check if the name is "SmtpPassword"
		if name == "SmtpPassword" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "SmtpPassword is empty"))
			} else {
				email.SmtpPassword = value
			}
		}

		// check if the name is "Attachment"
		if name == "Attachment" {
			if value == "" {
				f.iLog.Debug(fmt.Sprintf("Error: %v", "Attachment is empty"))
			} else {
				email.Attachment = value
			}
		}
	}

	// check if any of the required fields are empty
	if len(email.ToEmailList) == 0 || email.Subject == "" || email.Body == "" {
		f.iLog.Debug(fmt.Sprintf("Error: %v", "ToEmails or Subject or Body is empty"))
		return false, email
	} else {
		return true, email
	}
}
