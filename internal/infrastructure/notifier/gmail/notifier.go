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

	"gopkg.in/gomail.v2"
)

const PassLabel = "KARNET"

type notifier struct {
	dialer                             *gomail.Dialer
	login                              string
	bookingConfirmationRequestTmplPath string
	bookingConfirmationTmplPath        string
	classCancellationTmplPath          string
	classUpdateTmplPath                string
	bookingCancellationTmplPath        string
	passActivationTmplPath             string
	classReminderTmplPath              string
	passTmplPath                       string
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
		classReminderTmplPath:              baseTmplPath + "class_reminder.tmpl",
		passTmplPath:                       baseTmplPath + "pass.tmpl",
	}
}

func (n *notifier) NotifyPassActivation(email string, passItems []models.PassItem) error {
	tmplData := notifierModels.PassActivationTmplData{
		Signature:     n.signature,
		PassItemsView: n.getPassItemsView(passItems),
	}

	tmpl, err := template.ParseFiles(n.passActivationTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := "Yoga - Twój karnet jest aktywny!"

	msgToRecipient, err := n.buildMsgToRecipient(email, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", email, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyConfirmationLink(
	email, firstName, confirmationLink string, classStartTime time.Time,
) error {
	tmplData := notifierModels.BookingConfirmationRequestTmplData{
		RecipientFirstName: firstName,
		ConfirmationLink:   confirmationLink,
		Signature:          n.signature,
	}

	tmpl, err := template.ParseFiles(n.bookingConfirmationRequestTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	classStartTimeDetails, err := getClassStartTimeDetails(classStartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	subject := fmt.Sprintf("Yoga (%s) - Potwierdź swoją rezerwację!", classStartTimeDetails.startDate)

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

	tmplData := notifierModels.BookingConfirmationTmplData{
		BaseTmplData:     baseTmplData,
		CancellationLink: cancellationLink,
		PassItemsView:    n.getPassItemsView(params.PassItems),
	}

	tmpl, err := template.ParseFiles(n.bookingConfirmationTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := fmt.Sprintf("Yoga (%s) - rezerwacja potwierdzona!", classStartTimeDetails.startDate)

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(
		models.StatusBooked,
		params.RecipientFirstName,
		params.RecipientLastName,
		n.getPassItemsView(params.PassItems),
		classStartTimeDetails,
	)

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

	tmplData := notifierModels.BookingCancellationTmplData{
		BaseTmplData:  n.getBaseTmplData(params, classStartTimeDetails),
		PassItemsView: n.getPassItemsView(params.PassItems),
	}

	tmpl, err := template.ParseFiles(n.bookingCancellationTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := fmt.Sprintf("Yoga (%s) - rezerwacja odwołana!", classStartTimeDetails.startDate)

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	msgToOwner := n.buildMsgToOwner(
		models.StatusCancelled,
		params.RecipientFirstName,
		params.RecipientLastName,
		n.getPassItemsView(params.PassItems),
		classStartTimeDetails,
	)

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

	tmplData := notifierModels.ClassUpdateTmplData{
		BaseTmplData: n.getBaseTmplData(params, classStartTimeDetails),
		Message:      msg,
	}

	tmpl, err := template.ParseFiles(n.classUpdateTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := fmt.Sprintf(
		"Yoga (%s) Musiałem wprowadzić zmiany w zajęciach!",
		classStartTimeDetails.startDate,
	)

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

	tmplData := notifierModels.ClassCancellationTmplData{
		BaseTmplData:  n.getBaseTmplData(params, classStartTimeDetails),
		Message:       msg,
		PassItemsView: n.getPassItemsView(params.PassItems),
	}

	tmpl, err := template.ParseFiles(n.classCancellationTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := fmt.Sprintf("Yoga (%s) - zajęcia odwołane!", classStartTimeDetails.startDate)

	msgToRecipient, err := n.buildMsgToRecipient(params.RecipientEmail, subject, tmpl, tmplData)
	if err != nil {
		return fmt.Errorf("could not build msg to recipient %s: %w", params.RecipientEmail, err)
	}

	if err = n.dialer.DialAndSend(msgToRecipient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *notifier) NotifyBookingReminder(
	params models.NotifierParams, cancellationLink string,
) error {
	classStartTimeDetails, err := getClassStartTimeDetails(params.StartTime)
	if err != nil {
		return fmt.Errorf("could not get class start time details: %w", err)
	}

	tmplData := notifierModels.BookingReminderTmplData{
		BaseTmplData:     n.getBaseTmplData(params, classStartTimeDetails),
		CancellationLink: cancellationLink,
		PassItemsView:    n.getPassItemsView(params.PassItems),
	}

	tmpl, err := template.ParseFiles(n.classReminderTmplPath, n.passTmplPath)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	subject := fmt.Sprintf("Yoga (%s) - przypomnienie o zajęciach!", classStartTimeDetails.startDate)

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
	recipientFirstName, recipientLastName string,
	passItemsView []notifierModels.PassItemView,
	classTimeDetails timeDetails,
) *gomail.Message {
	subject := fmt.Sprintf("%s %s %s",
		recipientFirstName,
		recipientLastName,
		status,
	)

	if passItemsView != nil {
		usedPassItemsCount := 0

		for _, item := range passItemsView {
			if item.Status == models.FuturePassStatus || item.Status == models.PastPassStatus {
				usedPassItemsCount++
			}
		}

		subject += fmt.Sprintf(
			"%s: %d/%d", PassLabel, usedPassItemsCount, len(passItemsView),
		)
	}

	msg := fmt.Sprintf("%s (%s) - %s",
		classTimeDetails.weekDayInPolish,
		classTimeDetails.startDate,
		classTimeDetails.startHour,
	)

	msgToOwner := gomail.NewMessage()
	msgToOwner.SetHeader("From", n.login)
	msgToOwner.SetHeader("To", n.login)
	msgToOwner.SetHeader("Subject", subject)
	msgToOwner.SetBody("text/html", msg)

	return msgToOwner
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
	return notifierModels.BaseTmplData{
		RecipientFirstName: params.RecipientFirstName,
		ClassName:          params.ClassName,
		ClassLevel:         params.ClassLevel,
		Hour:               classStartTimeDetails.startHour,
		WeekDay:            classStartTimeDetails.weekDayInPolish,
		Date:               classStartTimeDetails.startDate,
		Location:           params.Location,
		Signature:          n.signature,
	}
}

func (n *notifier) getPassItemsView(passItems []models.PassItem) []notifierModels.PassItemView {
	passItemsView := make([]notifierModels.PassItemView, 0, len(passItems))

	for _, item := range passItems {
		viewItem := notifierModels.PassItemView{
			Status: item.Status,
		}

		if item.ClassStartTime != nil {
			viewItem.ClassStartDate = item.ClassStartTime.Format("02.01")
		}

		passItemsView = append(passItemsView, viewItem)
	}

	return passItemsView
}
