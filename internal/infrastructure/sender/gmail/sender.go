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

func (s Sender) SendConfirmationCreateLink(msg models.ConfirmationCreateMsg) error {
	tmplData := infrastructureModels.PendingConfirmationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: msg.RecipientFirstName,
		ConfirmationLink:   msg.ConfirmationCreateLink,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationCreateEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msgContent strings.Builder
	err = tmpl.Execute(&msgContent, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", msg.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Prośba o potwierdzenie rezerwacji!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendConfirmationCancelLink(msg models.ConfirmationCancelMsg) error {
	tmplData := infrastructureModels.PendingConfirmationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: msg.RecipientFirstName,
		ConfirmationLink:   msg.ConfirmationCancelLink,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationCancelEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msgContent strings.Builder
	err = tmpl.Execute(&msgContent, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", msg.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Prośba o potwierdzenie odwołania rezerwacji")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendFinalConfirmations(msg models.ConfirmationMessage) error {
	tmplData := infrastructureModels.FinalConfirmationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: msg.RecipientFirstName,
		ClassName:          msg.ClassName,
		ClassLevel:         msg.ClassLevel,
		WeekDay:            msg.WeekDay,
		Hour:               msg.Hour,
		Date:               msg.Date,
		Location:           msg.Location,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationFinalEmailTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var msgContent strings.Builder
	err = tmpl.Execute(&msgContent, tmplData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	msgToRecipient := gomail.NewMessage()
	msgToRecipient.SetHeader("From", s.SenderEmail)
	msgToRecipient.SetHeader("To", msg.RecipientEmail)
	msgToRecipient.SetHeader("Subject", "Yoga - Rezerwacja potwierdzona!")
	msgToRecipient.SetBody("text/html", msgContent.String())

	subject := fmt.Sprintf("%s %s booked: %s (%s) at %s.",
		msg.RecipientFirstName,
		msg.RecipientLastName,
		msg.WeekDay,
		msg.Date,
		msg.Hour,
	)

	msgToOwner := gomail.NewMessage()
	msgToOwner.SetHeader("From", s.SenderEmail)
	msgToOwner.SetHeader("To", s.SenderEmail)
	msgToOwner.SetHeader("Subject", subject)

	if err = s.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send emails: %w", err)
	}

	return nil
}

func (s Sender) SendInfoAboutCancellationToOwner(msg models.ConfirmationToOwnerMsg) error {
	subject := fmt.Sprintf("%s %s cancelled: %s (%s) at %s.",
		msg.RecipientFirstName,
		msg.RecipientLastName,
		msg.WeekDay,
		msg.Date,
		msg.Hour,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", s.SenderEmail)
	m.SetHeader("Subject", subject)

	if err := s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
