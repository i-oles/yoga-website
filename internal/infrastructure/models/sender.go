package models

import models2 "main/internal/domain/models"

type PendingConfirmationTmplData struct {
	SenderName       string
	RecipientName    string
	ConfirmationLink string
}

type FinalConfirmationTmplData struct {
	SenderName    string
	RecipientName string
	ClassName     string
	ClassLevel    models2.ClassLevel
	DayOfWeek     string
	Hour          string
	Date          string
	Location      string
}
