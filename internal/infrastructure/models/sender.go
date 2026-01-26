package models

type ConfirmationRequestTmplData struct {
	SenderName         string
	RecipientFirstName string
	LinkToConfirmation string
}

type ConfirmationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	CancellationLink   string
	PassState          []bool
}

type ConfirmationToOwnerTmplData struct {
	SenderName         string
	RecipientFirstName string
	RecipientLastName  string
	WeekDay            string
	Hour               string
	Date               string
}

type ClassCancellationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	ReasonMsg          string
}

type BookingCancellationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	PassState          []bool
}

type ClassUpdateTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	Message            string
}

type PassActivationTmplData struct {
	SenderName string
	PassState  []bool
}
