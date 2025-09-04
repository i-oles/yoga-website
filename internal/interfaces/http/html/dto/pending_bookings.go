package dto

import "github.com/google/uuid"

// TODO: w routach pending jest classID - ale nie jest parsowane z URL - zobacz co lepsze
type PendingBookingForm struct {
	ClassID   string `form:"class_id" binding:"required,uuid"`
	FirstName string `form:"first_name" binding:"required,min=3,max=30"`
	LastName  string `form:"last_name" binding:"required,max=30"`
	Email     string `form:"email" binding:"required,email"`
}

type PendingBookingView struct {
	ClassID uuid.UUID
}

// TODO: chce tu zwracac nazwisko i imie?
type CancelBookingView struct {
	WeekDay         string
	StartDate       string
	StartHour       string
	ClassLevel      string
	ClassName       string
	CurrentCapacity int
	MaxCapacity     int
	Location        string
}
