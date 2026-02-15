package sender

import (
	"main/internal/domain/models"
)

type INotifier interface {
	NotifyPassActivation(pass models.Pass) error
	NotifyConfirmationLink(email, firstName, confirmationLink string) error
	NotifyBookingConfirmation(params models.NotifierParams, cancellationLink string) error
	NotifyBookingCancellation(params models.NotifierParams) error
	NotifyClassUpdate(recipientEmail, recipientFirstName, msg string, class models.Class) error
	NotifyClassCancellation(params models.NotifierParams, msg string) error
}
