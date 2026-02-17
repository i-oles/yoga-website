package notifier

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
	Signature          string
}

type PassActivationTmplData struct {
	PassState []bool
	Signature string
}
