package services

import (
	"context"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

type IClassesService interface {
	GetAllClasses(ctx context.Context) ([]models.Class, error)
	CreateClasses(ctx context.Context, class []models.Class) ([]models.Class, error)
}

type IBookingsService interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, token string) (models.Class, error)
}

type IPendingBookingsService interface {
	CreatePendingBooking(ctx context.Context, params models.PendingBookingParams) (uuid.UUID, error)
	CancelPendingBooking(ctx context.Context, params models.CancelBookingParams) (uuid.UUID, error)
}

type ITokenGenerator interface {
	Generate(length int) (string, error)
}

type ISender interface {
	SendConfirmationCreateLink(msg models.ConfirmationCreateMsg) error
	SendConfirmationCancelLink(msg models.ConfirmationCancelMsg) error
	SendFinalConfirmations(msg models.ConfirmationMsg) error
	SendInfoAboutCancellationToOwner(msg models.ConfirmationToOwnerMsg) error
}
