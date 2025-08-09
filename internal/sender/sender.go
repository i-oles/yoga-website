package sender

type Message interface {
	SendConfirmationLink(data ConfirmationData) error
}

type ConfirmationData struct {
	RecipientEmail   string
	RecipientName    string
	ConfirmationLink string
}
