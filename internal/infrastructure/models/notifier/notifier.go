package notifier

type BaseTmpl struct {
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

type BaseTmplWithMsg struct {
	BaseTmplData BaseTmpl
	Message      string
}

type BookingConfirmationTmpl struct {
	BaseTmplData     BaseTmpl
	CancellationLink string
}

type BookingConfirmationRequestTmpl struct {
	RecipientFirstName string
	ConfirmationLink   string
	Signature          string
}

type PassActivationTmpl struct {
	PassState []bool
	Signature string
}
