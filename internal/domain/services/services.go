package services

import (
	"context"

	"main/internal/domain/models"

	"github.com/google/uuid"
)

type IClassesService interface {
	ListClasses(
		ctx context.Context,
		onlyUpcomingClasses bool,
		classesLimit *int,
	) ([]models.ClassWithCurrentCapacity, error)
	CreateClasses(ctx context.Context, classes []models.Class) ([]models.Class, error)
	UpdateClass(ctx context.Context, id uuid.UUID, update models.UpdateClass) (models.Class, error)
	DeleteClass(ctx context.Context, classID uuid.UUID, msg *string) error
}

type IBookingsService interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, id uuid.UUID, token string) error
	GetBookingForCancellation(ctx context.Context, id uuid.UUID, token string) (models.Booking, error)
	DeleteBooking(ctx context.Context, id uuid.UUID) error
}

type IPendingBookingsService interface {
	CreatePendingBooking(ctx context.Context, params models.PendingBookingParams) (uuid.UUID, error)
}

type IPassesService interface {
	ActivatePass(ctx context.Context, params models.PassActivationParams) (models.Pass, error)
}

type ITokenGenerator interface {
	Generate(length int) (string, error)
}
