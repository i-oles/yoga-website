package models

type PendingConfirmationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ConfirmationLink   string
}

type FinalConfirmationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
}

type ConfirmationToOwnerData struct {
	SenderName         string
	RecipientFirstName string
	RecipientLastName  string
	WeekDay            string
	Hour               string
	Date               string
}
