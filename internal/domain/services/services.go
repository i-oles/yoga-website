package services

import (
	"context"
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type IClassesService interface {
	GetClasses(
		ctx context.Context,
		onlyUpcomingClasses bool,
		classesLimit *int, 
	) ([]models.Class, error)
	CreateClasses(ctx context.Context, class []models.Class) ([]models.Class, error)
	DeleteClass(ctx context.Context, classID uuid.UUID) error
}

type IBookingsService interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, id uuid.UUID, token string) error
	CancelBookingForm(ctx context.Context, id uuid.UUID, token string) (models.Booking, error)
	DeleteBooking(ctx context.Context, id uuid.UUID) error
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
	SendInfoAboutClassCancellation(recipientEmail, recipientFirstName string, class models.Class) error
}
