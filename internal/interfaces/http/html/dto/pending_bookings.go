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

type PendingOperationCancelRequest struct {
	ClassID string `form:"class_id" binding:"required,uuid"`
	Email   string `form:"email" binding:"required,email"`
}

type PendingOperationCancelResponse struct {
	ClassID uuid.UUID `json:"class_id"`
}
