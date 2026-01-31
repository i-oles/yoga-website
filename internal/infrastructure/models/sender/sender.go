package sender

type ConfirmationRequestTmplData struct {
	SenderName         string
	RecipientFirstName string
	LinkToConfirmation string
}

type ConfirmationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	CancellationLink   string
	PassState          []bool
}

type BookingCancellationTmplData struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	PassState          []bool
}

type TmplWithMsg struct {
	SenderName         string
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	Message            string
	PassState          []bool
}

type PassActivationTmplData struct {
	SenderName string
	PassState  []bool
}
