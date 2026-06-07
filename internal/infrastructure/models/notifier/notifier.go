package notifier

import "main/internal/domain/models"

type BaseTmplData struct {
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	Signature          string
}

type ClassUpdateTmplData struct {
	BaseTmplData BaseTmplData
	Message      string
}

type ClassCancellationTmplData struct {
	BaseTmplData  BaseTmplData
	Message       string
	PassItemsView []PassItemView
}

type BookingConfirmationTmplData struct {
	BaseTmplData     BaseTmplData
	CancellationLink string
	PassItemsView    []PassItemView
}

type BookingCancellationTmplData struct {
	BaseTmplData  BaseTmplData
	PassItemsView []PassItemView
}

type BookingReminderTmplData struct {
	BaseTmplData     BaseTmplData
	CancellationLink string
	PassItemsView    []PassItemView
}

type BookingConfirmationRequestTmplData struct {
	RecipientFirstName string
	ConfirmationLink   string
	Signature          string
}

type PassActivationTmplData struct {
	PassItemsView []PassItemView
	Signature     string
}

type PassItemView struct {
	Status         models.PassStatus
	ClassStartDate string
}
