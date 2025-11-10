package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"main/internal/domain/models"
	infrastructureModels "main/internal/infrastructure/models"
	"main/pkg/converter"
	"main/pkg/translator"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	SenderName                     string
	SenderEmail                    string
	ConfirmationRequestTmplPath    string
	ConfirmationFinalEmailTmplPath string
	ClassCancellationTmplPath      string
	Dialer                         *gomail.Dialer
}

func NewSender(
	host string,
	port int,
	senderEmail string,
	password string,
	senderName string,
	baseSenderTmplPath string,
) *Sender {
	d := gomail.NewDialer(host, port, senderEmail, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // TODO: change to false in production

	return &Sender{
		SenderName:                     senderName,
		SenderEmail:                    senderEmail,
		ConfirmationRequestTmplPath:    baseSenderTmplPath + "confirmation_request_email.tmpl",
		ConfirmationFinalEmailTmplPath: baseSenderTmplPath + "confirmation_email.tmpl",
		ClassCancellationTmplPath:      baseSenderTmplPath + "class_cancellation.tmpl",
		Dialer:                         d,
	}
}

func (s Sender) SendLinkToConfirmation(
	recipientEmail string,
	recipientFirstName string,
	linkToConfirmation string,
) error {
	tmplData := infrastructureModels.ConfirmationRequestTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: recipientFirstName,
		LinkToConfirmation: linkToConfirmation,
	}

	tmpl, err := template.ParseFiles(s.ConfirmationRequestTmplPath)
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
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "Yoga - Prośba o potwierdzenie rezerwacji!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendConfirmations(msg models.ConfirmationMsg) error {
	startTimeDetails, err := getTimeDetails(msg.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := infrastructureModels.ConfirmationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: msg.RecipientFirstName,
		ClassName:          msg.ClassName,
		ClassLevel:         msg.ClassLevel,
		WeekDay:            startTimeDetails.weekDayInPolish,
		Hour:               startTimeDetails.startHour,
		Date:               startTimeDetails.startDate,
		Location:           msg.Location,
		CancellationLink:   msg.CancellationLink,
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
		startTimeDetails.weekDayInPolish,
		startTimeDetails.startDate,
		startTimeDetails.startHour,
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

func (s Sender) SendInfoAboutCancellationToOwner(
	recipientFirstName, recipientLastName string, startTime time.Time,
) error {
	startTimeDetails, err := getTimeDetails(startTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	subject := fmt.Sprintf("%s %s cancelled: %s (%s) at %s.",
		recipientFirstName,
		recipientLastName,
		startTimeDetails.weekDayInPolish,
		startTimeDetails.startDate,
		startTimeDetails.startHour,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", s.SenderEmail)
	m.SetHeader("Subject", subject)

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

type timeDetails struct {
	startHour       string
	startDate       string
	weekDayInPolish string
}

func getTimeDetails(t time.Time) (timeDetails, error) {
	timeWarsawUTC, err := converter.ConvertToWarsawTime(t)
	if err != nil {
		return timeDetails{}, fmt.Errorf("could not convert to warsaw time: %w", err)
	}

	weekDayInPolish, err := translator.TranslateToWeekDayToPolish(timeWarsawUTC.Weekday())
	if err != nil {
		return timeDetails{}, fmt.Errorf("could not translate: %s weekday: %w", weekDayInPolish, err)
	}

	startDate := timeWarsawUTC.Format(converter.DateLayout)
	startHour := timeWarsawUTC.Format(converter.HourLayout)

	return timeDetails{
		startHour:       startHour,
		startDate:       startDate,
		weekDayInPolish: weekDayInPolish,
	}, nil
}

func (s Sender) SendInfoAboutClassCancellation(
	recipientEmail, recipientFirstName string, class models.Class,
) error {
	classTimeDetails, err := getTimeDetails(class.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := infrastructureModels.ClassCancellationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: recipientFirstName,
		ClassName:          class.ClassName,
		Hour:               classTimeDetails.startHour,
		WeekDay:            classTimeDetails.weekDayInPolish,
		Date:               classTimeDetails.startDate,
		Location:           class.Location,
	}

	tmpl, err := template.ParseFiles(s.ClassCancellationTmplPath)
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
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "Yoga - Zajęcia Odwołane!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
