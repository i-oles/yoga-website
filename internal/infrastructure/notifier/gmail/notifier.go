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

func (n *notifier) NotifyPassActivation(pass models.Pass) error {
	passState := getPassState(pass.UsedBookingIDs, pass.TotalBookings)

	tmplData := notifierModels.PassActivationTmplData{
		SenderName: n.Signature,
		PassState:  passState,
	}

	tmpl, err := template.ParseFiles(n.PassActivationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Twój karnet jest aktywny!"

	msgToRecipient, err := n.buildMsgToRecipient(pass.Email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", pass.Email, err)
	}

	if err = n.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyConfirmationLink(email, firstName, confirmationLink string) error {
	tmplData := notifierModels.BookingConfirmationRequestTmpl{
		RecipientFirstName: firstName,
		ConfirmationLink:   confirmationLink,
		Signture:           n.Signature,
	}

	tmpl, err := template.ParseFiles(n.BookingConfirmationRequestTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Potwierdź swoją rezerwację!"

	msgToRecipient, err := n.buildMsgToRecipient(email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", email, err)
	}

	if err = n.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyBookingConfirmation(
	params models.NotifierParams, cancellationLink string,
) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	baseTmplData := n.getBaseTmplData(params, classStartTimeDetails)

	tmplData := notifierModels.BookingConfirmationTmpl{
		BaseTmplData:     baseTmplData,
		CancellationLink: cancellationLink,
	}

	tmpl, err := template.ParseFiles(n.BookingConfirmationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Rezerwacja potwierdzona!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(models.StatusBooked, params, baseTmplData)

	if err = n.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send emails: %w", err)
	}

	return nil
}

func (n *notifier) NotifyBookingCancellation(params models.NotifierParams) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	baseTmplData := n.getBaseTmplData(params, classStartTimeDetails)

	tmpl, err := template.ParseFiles(n.BookingCancellationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Rezerwacja odwołana!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, baseTmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(models.StatusCancelled, params, baseTmplData)

	if err = n.Dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyClassUpdate(
	params models.NotifierParams, msg string,
) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	baseTmplData := n.getBaseTmplData(params, classStartTimeDetails)

	tmplData := notifierModels.TmplWithMsg{
		BaseTmplData: baseTmplData,
		Message:      msg,
	}

	tmpl, err := template.ParseFiles(n.ClassUpdateTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Musiałem wprowadzić zmiany w zajęciach, na które się wybierasz!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = n.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyClassCancellation(params models.NotifierParams, msg string) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get date details: %w", err)
	}

	basetmpldata := n.getBaseTmplData(params, classStartTimeDetails)

	tmplData := notifierModels.TmplWithMsg{
		BaseTmplData: basetmpldata,
		Message:      msg,
	}

	tmpl, err := template.ParseFiles(n.ClassCancellationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Zajęcia Odwołane!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = n.Dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) buildMsgToRecipient(
	email,
	subject string,
	tmpl *template.Template,
	tmplData any,
) (*gomail.Message, error) {
	var body strings.Builder

	err := tmpl.Execute(&body, tmplData)
	if err != nil {
		return nil, fmt.Errorf("could not execute template: %w", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", n.Login)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	return msg, nil
}

func (n *notifier) buildMsgToOwner(
	status models.BookingStatus,
	params models.NotifierParams,
	baseTmplData notifierModels.BaseTmplData,
) *gomail.Message {
	subject := fmt.Sprintf("%s %s %s",
		params.RecipientFirstName,
		params.RecipientLastName,
		status,
	)

	if baseTmplData.PassState != nil {
		subject += fmt.Sprintf(
			" %s: %d/%d", PassValue, len(params.PassUsedBookingIDs), *params.PassTotalBookings,
		)
	}

	msg := fmt.Sprintf("%s (%s) - %s",
		baseTmplData.WeekDay,
		baseTmplData.Date,
		baseTmplData.Hour,
	)

	msgToOwner := gomail.NewMessage()
	msgToOwner.SetHeader("From", n.Login)
	msgToOwner.SetHeader("To", n.Login)
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

// TODO: is this code duplicated?
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

func (n *notifier) getBaseTmplData(
	params models.NotifierParams, classStartTimeDetails timeDetails,
) notifierModels.BaseTmplData {
	tmplData := notifierModels.BaseTmplData{
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classStartTimeDetails.startHour,
		WeekDay:            classStartTimeDetails.weekDayInPolish,
		Date:               classStartTimeDetails.startDate,
		Location:           params.Location,
		Signature:          n.Signature,
	}

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
	}

	return tmplData
}
