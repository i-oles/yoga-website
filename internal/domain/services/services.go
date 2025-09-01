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

type IPendingOperationsService interface {
	CreateBooking(ctx context.Context, createParams models.CreateParams) (uuid.UUID, error)
	CancelBooking(ctx context.Context, cancelParams models.CancelParams) (uuid.UUID, error)
}

type IConfirmationService interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, token string) (models.Class, error)
}

type ITokenGenerator interface {
	Generate(length int) (string, error)
}

type ISender interface {
	SendConfirmationCreateLink(params models.ConfirmationCreateParams) error
	SendConfirmationCancelLink(params models.ConfirmationCancelParams) error
	SendFinalConfirmations(params models.ConfirmationFinalParams) error
	SendInfoAboutCancellationToOwner(msgParams models.ConfirmationCancelToOwnerParams) error
}
