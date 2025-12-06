package dto

import "github.com/google/uuid"

type PendingBookingForm struct {
	ClassID   string `binding:"required,uuid" form:"class_id"`
	FirstName string `form:"first_name" binding:"required,min=3,max=30"`
	LastName  string `form:"last_name" binding:"required,max=30"`
	Email     string `form:"email" binding:"required,email"`
}

type PendingBookingView struct {
	ClassID uuid.UUID
}
