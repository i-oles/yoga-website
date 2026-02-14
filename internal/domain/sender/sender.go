package sender

import (
	"main/internal/domain/models"
)

type ISender interface {
	SendLinkToConfirmation(recipientEmail, recipientFirstName, linkToConfirmation string) error
	SendConfirmations(params models.SenderParams, cancellationLink string) error
	SendInfoAboutClassCancellation(params models.SenderParams, msg string) error
	SendInfoAboutUpdate(recipientEmail, recipientFirstName, message string, class models.Class) error
	SendInfoAboutBookingCancellation(params models.SenderParams) error
	SendPass(pass models.Pass) error
}
