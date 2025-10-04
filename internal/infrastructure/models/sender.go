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
}

type ConfirmationToOwnerTmplData struct {
	SenderName         string
	RecipientFirstName string
	RecipientLastName  string
	WeekDay            string
	Hour               string
	Date               string
}
