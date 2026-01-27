package models

import "time"

type SenderParams struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  *string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	UsedPassCredits    *int
	TotalPassCredits   *int
}
