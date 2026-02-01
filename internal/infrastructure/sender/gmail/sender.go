package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"strings"
	"time"

	"main/internal/domain/models"
	senderModels "main/internal/infrastructure/models/sender"
	"main/pkg/converter"
	"main/pkg/translator"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

const Pass = "KARNET"

type Sender struct {
	SenderName                         string
	SenderEmail                        string
	BookingConfirmationRequestTmplPath string
	BookingConfirmationTmplPath        string
	ClassCancellationTmplPath          string
	ClassUpdateTmplPath                string
	BookingCancellationTmplPath        string
	PassActivationTmplPath             string
	Dialer                             *gomail.Dialer
}

func NewSender(
	host string,
	port int,
	senderEmail string,
	password string,
	senderName string,
	baseSenderTmplPath string,
) *Sender {
	dialer := gomail.NewDialer(host, port, senderEmail, password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true} // TODO: change to false in production

	return &Sender{
		SenderName:                         senderName,
		SenderEmail:                        senderEmail,
		BookingConfirmationRequestTmplPath: baseSenderTmplPath + "booking_confirmation_request.tmpl",
		BookingConfirmationTmplPath:        baseSenderTmplPath + "booking_confirmation.tmpl",
		ClassCancellationTmplPath:          baseSenderTmplPath + "class_cancellation.tmpl",
		ClassUpdateTmplPath:                baseSenderTmplPath + "class_update.tmpl",
		BookingCancellationTmplPath:        baseSenderTmplPath + "booking_cancellation.tmpl",
		PassActivationTmplPath:             baseSenderTmplPath + "pass_activation.tmpl",
		Dialer:                             dialer,
	}
}

func (s Sender) SendLinkToConfirmation(
	recipientEmail string,
	recipientFirstName string,
	linkToConfirmation string,
) error {
	tmplData := senderModels.ConfirmationRequestTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: recipientFirstName,
		LinkToConfirmation: linkToConfirmation,
	}

	tmpl, err := template.ParseFiles(s.BookingConfirmationRequestTmplPath)
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

func (s Sender) SendConfirmations(params models.SenderParams, cancellationLink string) error {
	startTimeDetails, err := getTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := senderModels.ConfirmationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		WeekDay:            startTimeDetails.weekDayInPolish,
		Hour:               startTimeDetails.startHour,
		Date:               startTimeDetails.startDate,
		Location:           params.Location,
		CancellationLink:   cancellationLink,
	}

	var isPass bool
	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
		isPass = true
	}

	tmpl, err := template.ParseFiles(s.BookingConfirmationTmplPath)
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
	msgToRecipient.SetHeader("To", params.RecipientEmail)
	msgToRecipient.SetHeader("Subject", "Yoga - Rezerwacja potwierdzona!")
	msgToRecipient.SetBody("text/html", msgContent.String())

	paymentType := ""
	if isPass {
		paymentType = Pass
	}

	subject := fmt.Sprintf("%s %s booked %s",
		params.RecipientFirstName,
		*params.RecipientLastName,
		paymentType,
	)

	msg := fmt.Sprintf("%s (%s) - %s",
		startTimeDetails.weekDayInPolish,
		startTimeDetails.startDate,
		startTimeDetails.startHour,
	)

	msgToOwner := gomail.NewMessage()
	msgToOwner.SetHeader("From", s.SenderEmail)
	msgToOwner.SetHeader("To", s.SenderEmail)
	msgToOwner.SetHeader("Subject", subject)
	msgToOwner.SetBody("text/html", msg)

	if err = s.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send emails: %w", err)
	}

	return nil
}

func getPassState(usedBookingIDs []uuid.UUID, totalBookings int) []bool {
	result := make([]bool, totalBookings)

	for i := range usedBookingIDs {
		result[i] = true
	}

	return result
}

func (s Sender) SendInfoAboutCancellationToOwner(
	recipientFirstName, recipientLastName string, startTime time.Time,
) error {
	startTimeDetails, err := getTimeDetails(startTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	subject := fmt.Sprintf("%s %s cancelled", recipientFirstName, recipientLastName)

	msg := fmt.Sprintf("%s (%s) - %s",
		startTimeDetails.weekDayInPolish,
		startTimeDetails.startDate,
		startTimeDetails.startHour,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", s.SenderEmail)
	m.SetHeader("To", s.SenderEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", msg)

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
	params models.SenderParams, msg string,
) error {
	classTimeDetails, err := getTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := senderModels.TmplWithMsg{
		SenderName:         s.SenderName,
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classTimeDetails.startHour,
		WeekDay:            classTimeDetails.weekDayInPolish,
		Date:               classTimeDetails.startDate,
		Location:           params.Location,
		Message:            msg,
	}

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
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
	m.SetHeader("To", params.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Zajęcia Odwołane!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendInfoAboutUpdate(
	recipientEmail, recipientFirstName, message string, class models.Class,
) error {
	classTimeDetails, err := getTimeDetails(class.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := senderModels.TmplWithMsg{
		SenderName:         s.SenderName,
		RecipientFirstName: recipientFirstName,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		Hour:               classTimeDetails.startHour,
		WeekDay:            classTimeDetails.weekDayInPolish,
		Date:               classTimeDetails.startDate,
		Location:           class.Location,
		Message:            message,
	}

	tmpl, err := template.ParseFiles(s.ClassUpdateTmplPath)
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
	m.SetHeader("Subject", "Yoga - Musiałem wprowadzić zmiany w zajęciach, na które się wybierasz!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendInfoAboutBookingCancellation(params models.SenderParams) error {
	classTimeDetails, err := getTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := senderModels.BookingCancellationTmplData{
		SenderName:         s.SenderName,
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classTimeDetails.startHour,
		WeekDay:            classTimeDetails.weekDayInPolish,
		Date:               classTimeDetails.startDate,
		Location:           params.Location,
	}

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
	}

	tmpl, err := template.ParseFiles(s.BookingCancellationTmplPath)
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
	m.SetHeader("To", params.RecipientEmail)
	m.SetHeader("Subject", "Yoga - Rezerwacja odwołana!")
	m.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s Sender) SendPass(pass models.Pass) error {
	passState := getPassState(pass.UsedBookingIDs, pass.TotalBookings)

	tmplData := senderModels.PassActivationTmplData{
		SenderName: s.SenderName,
		PassState:  passState,
	}

	tmpl, err := template.ParseFiles(s.PassActivationTmplPath)
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
	msgToRecipient.SetHeader("To", pass.Email)
	msgToRecipient.SetHeader("Subject", "Yoga - twój karnet!")
	msgToRecipient.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
