package notifier

import "main/internal/domain/models"

type PassItemView struct {
	Status         models.PassStatus
	ClassStartDate string
}

type BaseTmpl struct {
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	PassItemsView      []PassItemView
	Signature          string
}

type BaseTmplWithMsg struct {
	BaseTmplData BaseTmpl
	Message      string
}

type BaseTmplWithCancellationLink struct {
	BaseTmplData     BaseTmpl
	CancellationLink string
}

type BookingConfirmationRequestTmpl struct {
	RecipientFirstName string
	ConfirmationLink   string
	Signature          string
}

type PassActivationTmpl struct {
	PassItemsView []PassItemView
	Signature     string
}
