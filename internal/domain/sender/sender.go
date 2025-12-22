package sender

import (
	"time"

	"main/internal/domain/models"
)

type ISender interface {
	SendLinkToConfirmation(recipientEmail, recipientFirstName, linkToConfirmation string) error
	SendConfirmations(msg models.ConfirmationMsg) error
	SendInfoAboutCancellationToOwner(
		recipientFirstName, recipientLastName string, startTime time.Time,
	) error
	SendInfoAboutClassCancellation(
		recipientEmail, recipientFirstName, reasonMsg string, class models.Class,
	) error
	SendInfoAboutUpdate(
		recipientEmail, recipientFirstName, message string, class models.Class,
	) error
	SendInfoAboutBookingCancellation(
		recipientEmail, recipientFirstName string, class models.Class,
	) error
}
