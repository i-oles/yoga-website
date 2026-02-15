package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"strings"
	"time"

	"main/internal/domain/models"
	notifierModels "main/internal/infrastructure/models/sender"
	"main/pkg/converter"
	"main/pkg/translator"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

const PassValue = "KARNET"

type notifier struct {
	Dialer                             *gomail.Dialer
	Login                              string
	Signature                          string
	BookingConfirmationRequestTmplPath string
	BookingConfirmationTmplPath        string
	ClassCancellationTmplPath          string
	ClassUpdateTmplPath                string
	BookingCancellationTmplPath        string
	PassActivationTmplPath             string
}

func NewNotifier(
	host string,
	port int,
	login string,
	password string,
	signature string,
	baseTmplPath string,
) *notifier {
	dialer := gomail.NewDialer(host, port, login, password)
	dialer.TLSConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	return &notifier{
		Dialer:                             dialer,
		Login:                              login,
		Signature:                          signature,
		BookingConfirmationRequestTmplPath: baseTmplPath + "booking_confirmation_request.tmpl",
		BookingConfirmationTmplPath:        baseTmplPath + "booking_confirmation.tmpl",
		ClassCancellationTmplPath:          baseTmplPath + "class_cancellation.tmpl",
		ClassUpdateTmplPath:                baseTmplPath + "class_update.tmpl",
		BookingCancellationTmplPath:        baseTmplPath + "booking_cancellation.tmpl",
		PassActivationTmplPath:             baseTmplPath + "pass_activation.tmpl",
	}
}

func (s *notifier) NotifyConfirmationLink(email, firstName, confirmationLink string) error {
	tmplData := notifierModels.BookingConfirmationRequestTmpl{
		RecipientFirstName: firstName,
		ConfirmationLink:   confirmationLink,
		Signture:           s.Signature,
	}

	tmpl, err := template.ParseFiles(s.BookingConfirmationRequestTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Potwierdź swoją rezerwację!"

	msgToRecipient, err := s.buildMsgToRecipient(email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", email, err)
	}

	if err = s.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *notifier) buildMsgToRecipient(
	email, subject string,
	tmpl *template.Template,
	tmplData any,
) (*gomail.Message, error) {
	var body strings.Builder

	err := tmpl.Execute(&body, tmplData)
	if err != nil {
		return nil, fmt.Errorf("could not execute template: %w", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.Login)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	return msg, nil
}

func (s *notifier) NotifyBookingConfirmation(params models.NotifierParams, cancellationLink string) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	baseTmplData := notifierModels.BaseTmplData{
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		WeekDay:            classStartTimeDetails.weekDayInPolish,
		Hour:               classStartTimeDetails.startHour,
		Date:               classStartTimeDetails.startDate,
		Location:           params.Location,
		Signature:          s.Signature,
	}

	var isPass bool

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		baseTmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
		isPass = true
	}

	tmplData := notifierModels.BookingConfirmationTmpl{
		BaseTmplData:     baseTmplData,
		CancellationLink: cancellationLink,
	}

	tmpl, err := template.ParseFiles(s.BookingConfirmationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Rezerwacja potwierdzona!"

	msgToRecipient, err := s.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := s.buildMsgToOwner(isPass, models.StatusBooked, params, classStartTimeDetails)

	if err = s.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send emails: %w", err)
	}

	return nil
}

func (s *notifier) buildMsgToOwner(
	isPass bool,
	status models.BookingStatus,
	params models.NotifierParams,
	startTimeDetails timeDetails,
) *gomail.Message {
	passDetails := ""
	if isPass {
		passDetails = fmt.Sprintf("%s: %d/%d", PassValue, len(params.PassUsedBookingIDs), *params.PassTotalBookings)
	}

	subject := fmt.Sprintf("%s %s %s %s",
		params.RecipientFirstName,
		params.RecipientLastName,
		status,
		passDetails,
	)

	msg := fmt.Sprintf("%s (%s) - %s",
		startTimeDetails.weekDayInPolish,
		startTimeDetails.startDate,
		startTimeDetails.startHour,
	)

	msgToOwner := gomail.NewMessage()
	msgToOwner.SetHeader("From", s.Login)
	msgToOwner.SetHeader("To", s.Login)
	msgToOwner.SetHeader("Subject", subject)
	msgToOwner.SetBody("text/html", msg)

	return msgToOwner
}

func getPassState(usedBookingIDs []uuid.UUID, totalBookings int) []bool {
	result := make([]bool, totalBookings)

	for i := range usedBookingIDs {
		result[i] = true
	}

	return result
}

type timeDetails struct {
	startHour       string
	startDate       string
	weekDayInPolish string
}

func getClassStartTimeDetails(t time.Time) (timeDetails, error) {
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

func (s *notifier) NotifyClassCancellation(params models.NotifierParams, msg string) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	baseTmplData := notifierModels.BaseTmplData{
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classStartTimeDetails.startHour,
		WeekDay:            classStartTimeDetails.weekDayInPolish,
		Date:               classStartTimeDetails.startDate,
		Location:           params.Location,
		Signature:          s.Signature,
	}

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		baseTmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
	}

	tmplData := notifierModels.TmplWithMsg{
		BaseTmplData: baseTmplData,
		Message:      msg,
	}

	tmpl, err := template.ParseFiles(s.ClassCancellationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Zajęcia Odwołane!"

	msgToRecipient, err := s.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = s.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *notifier) NotifyClassUpdate(
	email, firstName, msg string, class models.Class,
) error {
	classStartTimeDetails, err := getClassStartTimeDetails(class.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	baseTmplData := notifierModels.BaseTmplData{
		Signature:          s.Signature,
		RecipientFirstName: firstName,
		ClassName:          class.ClassName,
		ClassLevel:         class.ClassLevel,
		Hour:               classStartTimeDetails.startHour,
		WeekDay:            classStartTimeDetails.weekDayInPolish,
		Date:               classStartTimeDetails.startDate,
		Location:           class.Location,
	}

	tmplData := notifierModels.TmplWithMsg{
		BaseTmplData: baseTmplData,
		Message:      msg,
	}

	tmpl, err := template.ParseFiles(s.ClassUpdateTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Musiałem wprowadzić zmiany w zajęciach, na które się wybierasz!"

	msgToRecipient, err := s.buildMsgToRecipient(email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", email, err)
	}

	if err = s.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *notifier) SendInfoAboutBookingCancellation(params models.NotifierParams) error {
	classTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	tmplData := notifierModels.BookingCancellationTmplData{
		SenderName:         s.Signature,
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classTimeDetails.startHour,
		WeekDay:            classTimeDetails.weekDayInPolish,
		Date:               classTimeDetails.startDate,
		Location:           params.Location,
	}

	var isPass bool

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
		isPass = true
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

	msgToRecipient := gomail.NewMessage()
	msgToRecipient.SetHeader("From", s.Login)
	msgToRecipient.SetHeader("To", params.RecipientEmail)
	msgToRecipient.SetHeader("Subject", "Yoga - Rezerwacja odwołana!")
	msgToRecipient.SetBody("text/html", msgContent.String())

	msgToOwner := s.buildMsgToOwner(isPass, models.StatusCancelled, params, classTimeDetails)

	if err = s.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *notifier) SendPass(pass models.Pass) error {
	passState := getPassState(pass.UsedBookingIDs, pass.TotalBookings)

	tmplData := notifierModels.PassActivationTmplData{
		SenderName: s.Signature,
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
	msgToRecipient.SetHeader("From", s.Login)
	msgToRecipient.SetHeader("To", pass.Email)
	msgToRecipient.SetHeader("Subject", "Yoga - twój karnet!")
	msgToRecipient.SetBody("text/html", msgContent.String())

	if err = s.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
