package submit

import (
	"context"
	"fmt"
	"main/internal/errs"
	"main/internal/generator"
	"main/internal/repository"
	"main/internal/sender"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ClassesRepo          repository.Classes
	ConfirmedBookingRepo repository.ConfirmedBookings
	PendingBookingRepo   repository.PendingBookings
	TokenGenerator       generator.Token
	MessageSender        sender.Message
	ErrorHandler         errs.ErrorHandler
	DomainAddr           string
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
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	classIDStr := c.PostForm("classID")

	classID, err := strconv.Atoi(classIDStr)
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

	err = h.submitPendingBooking(ctx, c, class)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}

func (h *Handler) submitPendingBooking(ctx context.Context, c *gin.Context, class repository.Class) error {
	if class.SpotsLeft == 0 {
		return errs.ErrClassFullyBooked(fmt.Errorf("no spots left in class with id: %d", class.ID))
	}

	name := c.PostForm("name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	token, err := h.TokenGenerator.Generate(32)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	token = fmt.Sprintf("olaf%d", rand.Int())
	fmt.Printf("token: %s\n", token)

	expiry := time.Now().Add(24 * time.Hour)

	pendingBooking := repository.PendingBooking{
		ClassID:   class.ID,
		ClassType: class.Type,
		Date:      class.Datetime,
		Place:     class.Place,
		Name:      name,
		LastName:  lastName,
		Email:     email,
		Token:     token,
		ExpiresAt: expiry,
	}

	err = h.PendingBookingRepo.Insert(ctx, pendingBooking)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	bookingConfirmationData := sender.BookingConfirmationData{
		RecipientEmail:   email,
		RecipientName:    name,
		ConfirmationLink: fmt.Sprintf("%s/confirmation?token=%s", h.DomainAddr, token),
	}

	err = h.MessageSender.SendConfirmationLink(bookingConfirmationData)
	if err != nil {
		return &errs.BookingError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return nil
}
