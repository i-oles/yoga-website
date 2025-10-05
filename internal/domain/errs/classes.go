package errs

const (
	BadRequestCode = iota
	ConflictCode
)

// TODO: could I name this better?
type ClassError struct {
	Code int
	Err  error
}

func (e ClassError) Error() string {
	return e.Err.Error()
}

func ErrClassValidation(err error) *ClassError {
	return &ClassError{
		Code: BadRequestCode,
		Err:  err,
	}
}

func ErrClassNotEmpty(err error) *ClassError {
	return &ClassError{
		Code: ConflictCode,
		Err: err, 
	}
}