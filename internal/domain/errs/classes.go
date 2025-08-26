package errs

const (
	BadRequestCode = iota
)

type ClassError struct {
	Code int
	Err  error
}

func ErrClassBadRequest(err error) *ClassError {
	return &ClassError{
		Code: BadRequestCode,
		Err:  err,
	}
}

func (e ClassError) Error() string {
	return e.Err.Error()
}
