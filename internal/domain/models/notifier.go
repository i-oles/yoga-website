package models

import (
	"time"
)

type NotifierParams struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	PassItems          []PassItem
}

type OperationStatus string

const (
	StatusBooked    OperationStatus = "booked"
	StatusCancelled OperationStatus = "cancelled"
)
