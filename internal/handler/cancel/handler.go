package cancel

import (
	"context"
	"fmt"
	"main/internal/errs"
	"main/internal/generator"
	"main/internal/repository"
	"main/internal/sender"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	ClassesRepo          repository.Classes
	ConfirmedBookingRepo repository.ConfirmedBookings
	PendingBookingRepo   repository.PendingBookings
	TokenGenerator       generator.Token
	MessageSender        sender.Message
	ErrorHandler         errs.ErrorHandler
	DomainAddr           string
	Operation            repository.Operation
}

func NewHandler(
	classesRepo repository.Classes,
	confirmedBookingsRepo repository.ConfirmedBookings,
	pendingBookingRepo repository.PendingBookings,
	tokenGenerator generator.Token,
	messageSender sender.Message,
	ErrorHandler errs.ErrorHandler,
	domainAddr string,
) *Handler {
	return &Handler{
		ClassesRepo:          classesRepo,
		ConfirmedBookingRepo: confirmedBookingsRepo,
		PendingBookingRepo:   pendingBookingRepo,
		TokenGenerator:       tokenGenerator,
		MessageSender:        messageSender,
		ErrorHandler:         ErrorHandler,
		DomainAddr:           domainAddr,
		Operation:            repository.CancelBooking,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	classIDStr := c.PostForm("classID")

	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	// TODO: this should take ctx
	class, err := h.ClassesRepo.Get(classID)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	err = h.submitCancelBookingOperation(ctx, c, class)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}

func (h *Handler) submitCancelBookingOperation(ctx context.Context, c *gin.Context, class repository.Class) error {
	firstName := c.PostForm("first_name")
	email := c.PostForm("email")

	token, err := h.TokenGenerator.Generate(32)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	////TODO: remove - only for tests
	//token = fmt.Sprintf("test%d", rand.Int())
	//fmt.Printf("token: %s\n", token)

	expiry := time.Now().Add(24 * time.Hour)

	cancelPendingOperation := repository.PendingOperation{
		ID:             uuid.New(),
		ClassID:        class.ID,
		Operation:      repository.CancelBooking,
		Email:          email,
		FirstName:      firstName,
		AuthToken:      token,
		TokenExpiresAt: expiry,
	}

	err = h.PendingBookingRepo.Insert(ctx, cancelPendingOperation)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	cancelBookingConfirmationData := sender.ConfirmationData{
		RecipientEmail:   email,
		RecipientName:    firstName,
		ConfirmationLink: fmt.Sprintf("%s/confirmation?token=%s", h.DomainAddr, token),
	}

	err = h.MessageSender.SendConfirmationLink(cancelBookingConfirmationData)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return nil
}
