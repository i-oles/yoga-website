package gmail

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"strings"
	"time"

	"main/internal/domain/models"
	notifierModels "main/internal/infrastructure/models/notifier"
	"main/pkg/converter"
	"main/pkg/translator"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

const PassValue = "KARNET"

type notifier struct {
	dialer                             *gomail.Dialer
	login                              string
	bookingConfirmationRequestTmplPath string
	bookingConfirmationTmplPath        string
	classCancellationTmplPath          string
	classUpdateTmplPath                string
	bookingCancellationTmplPath        string
	passActivationTmplPath             string
	signature                          string
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
		dialer:                             dialer,
		login:                              login,
		signature:                          signature,
		bookingConfirmationRequestTmplPath: baseTmplPath + "booking_confirmation_request.tmpl",
		bookingConfirmationTmplPath:        baseTmplPath + "booking_confirmation.tmpl",
		classCancellationTmplPath:          baseTmplPath + "class_cancellation.tmpl",
		classUpdateTmplPath:                baseTmplPath + "class_update.tmpl",
		bookingCancellationTmplPath:        baseTmplPath + "booking_cancellation.tmpl",
		passActivationTmplPath:             baseTmplPath + "pass_activation.tmpl",
	}
}

func (n *notifier) NotifyPassActivation(pass models.Pass) error {
	passState := getPassState(pass.UsedBookingIDs, pass.TotalBookings)

	tmplData := notifierModels.PassActivationTmplData{
		Signature: n.signature,
		PassState: passState,
	}

	tmpl, err := template.ParseFiles(n.passActivationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Twój karnet jest aktywny!"

	msgToRecipient, err := n.buildMsgToRecipient(pass.Email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", pass.Email, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyConfirmationLink(email, firstName, confirmationLink string) error {
	tmplData := notifierModels.BookingConfirmationRequestTmpl{
		RecipientFirstName: firstName,
		ConfirmationLink:   confirmationLink,
		Signature:          n.signature,
	}

	tmpl, err := template.ParseFiles(n.bookingConfirmationRequestTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Potwierdź swoją rezerwację!"

	msgToRecipient, err := n.buildMsgToRecipient(email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", email, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
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

	tmpl, err := template.ParseFiles(n.bookingConfirmationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Rezerwacja potwierdzona!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(models.StatusBooked, params, baseTmplData)

	if err = n.dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
		return fmt.Errorf("failed to send emails: %w", err)
	}

	return nil
}

func (n *notifier) NotifyBookingCancellation(params models.NotifierParams) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	tmplData := n.getBaseTmplData(params, classStartTimeDetails)

	tmpl, err := template.ParseFiles(n.bookingCancellationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Rezerwacja odwołana!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(models.StatusCancelled, params, tmplData)

	if err = n.dialer.DialAndSend(msgToRecipient, msgToOwner); err != nil {
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

	tmpl, err := template.ParseFiles(n.classUpdateTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Musiałem wprowadzić zmiany w zajęciach, na które się wybierasz!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
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

	tmpl, err := template.ParseFiles(n.classCancellationTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Zajęcia Odwołane!"

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
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
	msg.SetHeader("From", n.login)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	return msg, nil
}

func (n *notifier) buildMsgToOwner(
	status models.OperationStatus,
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
	msgToOwner.SetHeader("From", n.login)
	msgToOwner.SetHeader("To", n.login)
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
		Signature:          n.signature,
	}

	if params.PassUsedBookingIDs != nil && params.PassTotalBookings != nil {
		tmplData.PassState = getPassState(params.PassUsedBookingIDs, *params.PassTotalBookings)
	}

	return tmplData
}
