package sender

type BaseTmplData struct {
	RecipientFirstName string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	PassState          []bool
	Signature          string
}

type BookingConfirmationTmpl struct {
	BaseTmplData     BaseTmplData
	CancellationLink string
}

type TmplWithMsg struct {
	BaseTmplData BaseTmplData
	Message      string
}

type BookingConfirmationRequestTmpl struct {
	RecipientFirstName string
	ConfirmationLink   string
	Signture           string
}

type PassActivationTmplData struct {
	SenderName string
	PassState  []bool
}
