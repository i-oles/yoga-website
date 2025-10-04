package services

import (
	"context"
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type IClassesService interface {
	GetAllClasses(ctx context.Context) ([]models.Class, error)
	CreateClasses(ctx context.Context, class []models.Class) ([]models.Class, error)
}

type IBookingsService interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, id uuid.UUID, token string) error
	CancelBookingForm(ctx context.Context, id uuid.UUID, token string) (models.Booking, error)
}

type IPendingBookingsService interface {
	CreatePendingBooking(ctx context.Context, params models.PendingBookingParams) (uuid.UUID, error)
}

type ITokenGenerator interface {
	Generate(length int) (string, error)
}

type ISender interface {
	SendLinkToConfirmation(recipientEmail, recipientFirstName, linkToConfirmation string) error
	SendConfirmations(msg models.ConfirmationMsg) error
	SendInfoAboutCancellationToOwner(recipientFirstName, recipientLastName string, startTime time.Time) error
}
