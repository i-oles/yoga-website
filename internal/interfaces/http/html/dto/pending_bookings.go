package dto

type PendingBookingForm struct {
	Email     string `binding:"required,email" form:"email"`
	ClassID   string `binding:"required,uuid" form:"class_id"`
	LastName  string `binding:"required,max=30" form:"last_name"`
	FirstName string `binding:"required,min=3,max=30" form:"first_name"`
}
