package models

type PendingConfirmationTmplData struct {
	SenderName       string
	RecipientName    string
	ConfirmationLink string
}

type FinalConfirmationTmplData struct {
	SenderName    string
	RecipientName string
	ClassName     string
	ClassLevel    string
	WeekDay       string
	Hour          string
	Date          string
	Location      string
}
