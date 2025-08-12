package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"main/internal/domain/models"
	"strings"

	"gopkg.in/gomail.v2"
)

type templateData struct {
	SenderName       string
	RecipientName    string
	ConfirmationLink string
}

type Sender struct {
	SenderName         string
	SenderEmail        string
	TemplateCreatePath string
	TemplateCancelPath string
	Dialer             *gomail.Dialer
}

func NewSender(
	host string,
	port int,
	senderEmail string,
	password string,
	senderName string,
	templateCreatePath string,
	templateCancelPath string,
) *Sender {
	d := gomail.NewDialer(host, port, senderEmail, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // TODO: change to false in production

	return &Sender{
		SenderName:         senderName,
		SenderEmail:        senderEmail,
		TemplateCreatePath: templateCreatePath,
		TemplateCancelPath: templateCancelPath,
		Dialer:             d,
	}
}

//TODO: refactor - wydziel jedna wspolna funkcje.

func (s Sender) SendConfirmationCreateLink(data models.ConfirmationCreateParams) error {
	tmplData := templateData{
		SenderName:       s.SenderName,
		RecipientName:    data.RecipientName,
		ConfirmationLink: data.ConfirmationCreateLink,
	}

	tmpl, err := template.ParseFiles(s.TemplateCreatePath)
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
	m.SetHeader("To", data.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Potwierdzenie rezerwacji")
	m.SetBody("text/html", msg.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendConfirmationCancelLink(data models.ConfirmationCancelParams) error {
	tmplData := templateData{
		SenderName:       s.SenderName,
		RecipientName:    data.RecipientName,
		ConfirmationLink: data.ConfirmationCancelLink,
	}

	tmpl, err := template.ParseFiles(s.TemplateCancelPath)
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
	m.SetHeader("To", data.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Odwołanie rezerwacji")
	m.SetBody("text/html", msg.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
