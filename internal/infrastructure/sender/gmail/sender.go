package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"main/internal/domain/models"
	infrastructureModels "main/internal/infrastructure/models"
	"strings"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	SenderName                      string
	SenderEmail                     string
	ConfirmationCreateEmailTmplPath string
	ConfirmationCancelEmailTmplPath string
	ConfirmationFinalEmailTmplPath  string
	Dialer                          *gomail.Dialer
}

func NewSender(
	host string,
	port int,
	senderEmail string,
	password string,
	senderName string,
	confirmationCreateEmailTmplPath string,
	confirmationCancelEmailTmplPath string,
	confirmationFinalEmailTmplPath string,
) *Sender {
	d := gomail.NewDialer(host, port, senderEmail, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // TODO: change to false in production

	return &Sender{
		SenderName:                      senderName,
		SenderEmail:                     senderEmail,
		ConfirmationCreateEmailTmplPath: confirmationCreateEmailTmplPath,
		ConfirmationCancelEmailTmplPath: confirmationCancelEmailTmplPath,
		ConfirmationFinalEmailTmplPath:  confirmationFinalEmailTmplPath,
		Dialer:                          d,
	}
}

//TODO: refactor - one common function

func (s Sender) SendConfirmationCreateLink(msgParams models.ConfirmationCreateParams) error {
	tmplData := infrastructureModels.PendingConfirmationTmplData{
		SenderName:       s.SenderName,
		RecipientName:    msgParams.RecipientName,
		ConfirmationLink: msgParams.ConfirmationCreateLink,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationCreateEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msg strings.Builder
	err = tmpl.Execute(&msg, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", msgParams.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Potwierdź rezerwację!")
	m.SetBody("text/html", msg.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendConfirmationCancelLink(msgParams models.ConfirmationCancelParams) error {
	tmplData := infrastructureModels.PendingConfirmationTmplData{
		SenderName:       s.SenderName,
		RecipientName:    msgParams.RecipientName,
		ConfirmationLink: msgParams.ConfirmationCancelLink,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationCancelEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msg strings.Builder
	err = tmpl.Execute(&msg, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", msgParams.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Odwołanie rezerwacji")
	m.SetBody("text/html", msg.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendFinalConfirmation(msgParams models.ConfirmationFinalParams) error {
	tmplData := infrastructureModels.FinalConfirmationTmplData{
		SenderName:    s.SenderName,
		RecipientName: msgParams.RecipientName,
		ClassName:     msgParams.ClassName,
		ClassLevel:    msgParams.ClassLevel,
		WeekDay:       msgParams.WeekDay,
		Hour:          msgParams.Hour,
		Date:          msgParams.Date,
		Location:      msgParams.Location,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationFinalEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msg strings.Builder
	err = tmpl.Execute(&msg, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", msgParams.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Potwierdzenie rezerwacji!")
	m.SetBody("text/html", msg.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
