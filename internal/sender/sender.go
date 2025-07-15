package sender

type Message interface {
	SendConfirmationLink(data BookingConfirmationData) error
}

type BookingConfirmationData struct {
	RecipientEmail   string
	RecipientName    string
	ConfirmationLink string
}
