package services

import (
	"context"
	"main/internal/domain/models"

	"github.com/google/uuid"
)

type IServiceClasses interface {
	GetAllClasses(ctx context.Context) ([]models.Class, error)
}

type IServicePendingOperations interface {
	CreateBooking(ctx context.Context, createParams models.CreateParams) (uuid.UUID, error)
	CancelBooking(ctx context.Context, cancelParams models.CancelParams) (uuid.UUID, error)
}

type IServiceConfirmation interface {
	CreateBooking(ctx context.Context, token string) (models.Class, error)
	CancelBooking(ctx context.Context, token string) (models.Class, error)
}

type Token interface {
	Generate(length int) (string, error)
}

type Message interface {
	SendConfirmationLink(params models.ConfirmationMsgParams) error
}
