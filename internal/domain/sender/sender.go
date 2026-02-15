package sender

import (
	"main/internal/domain/models"
)

type INotifier interface {
	NotifyPassState(pass models.Pass) error
	NotifyBookingCancellation(params models.NotifierParams) error
	NotifyClassCancellation(params models.NotifierParams, msg string) error
	NotifyClassUpdate(recipientEmail, recipientFirstName, msg string, class models.Class) error
	NotifyBookingConfirmation(params models.NotifierParams, cancellationLink string) error
	NotifyConfirmationLink(email, firstName, confirmationLink string) error
}
