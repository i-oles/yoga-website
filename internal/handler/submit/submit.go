package submit

import (
	"errors"
	"fmt"
	"main/internal/generator"
	"main/internal/repository"
	"main/internal/sender"
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
}

func NewHandler(
	classesRepo repository.Classes,
	confirmedBookingsRepo repository.ConfirmedBookings,
	pendingBookingRepo repository.PendingBookings,
	tokenGenerator generator.Token,
	messageSender sender.Message,
) *Handler {
	return &Handler{
		ClassesRepo:          classesRepo,
		ConfirmedBookingRepo: confirmedBookingsRepo,
		PendingBookingRepo:   pendingBookingRepo,
		TokenGenerator:       tokenGenerator,
		MessageSender:        messageSender,
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

	class, err := h.ClassesRepo.Get(classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

		return
	}

	if class.SpotsLeft == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("this class is fully booked")})
	}

	name := c.PostForm("name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	token, err := h.TokenGenerator.Generate(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	expiry := time.Now().Add(24 * time.Hour)

	pendingBooking := repository.PendingBooking{
		ClassID:   classID,
		Name:      name,
		LastName:  lastName,
		Email:     email,
		Token:     token,
		ExpiresAt: expiry,
	}

	err = h.PendingBookingRepo.Insert(ctx, pendingBooking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	bookingConfirmationData := sender.BookingConfirmationData{
		RecipientEmail: email,
		RecipientName:  name,
		//TODO: move address to config
		ConfirmationLink: fmt.Sprintf("localhost:8080/confirmation/%d?token=%s", classID, token),
	}

	err = h.MessageSender.SendConfirmationLink(bookingConfirmationData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}
