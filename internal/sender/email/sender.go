package email

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"main/internal/sender"
	"strings"

	"gopkg.in/gomail.v2"
)

type templateData struct {
	FromName         string
	FromEmail        string
	RecipientName    string
	RecipientEmail   string
	ConfirmationLink string
}

type Sender struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
}

func NewSender(Host string, Port int, Username string, Password string, FromName string) *Sender {
	return &Sender{
		Host:     Host,
		Port:     Port,
		Username: Username,
		Password: Password,
		FromName: FromName,
	}
}

func (s Sender) SendConfirmationLink(data sender.BookingConfirmationData) error {
	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // TODO: change to false in production

	tmplData := templateData{
		FromName:         s.FromName,
		FromEmail:        s.Username,
		RecipientName:    data.RecipientName,
		RecipientEmail:   data.RecipientEmail,
		ConfirmationLink: data.ConfirmationLink,
	}

	//TODO: move path to config
	tmpl, err := template.ParseFiles("internal/sender/templates/confirmation_email.tmpl")
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msg strings.Builder
	err = tmpl.Execute(&msg, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.Username)
	m.SetHeader("To", data.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Potwierdzenie rezerwacji")
	m.SetBody("text/html", msg.String())

	if err = d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
