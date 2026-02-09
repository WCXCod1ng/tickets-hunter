package xerr

type ValidationError struct {
	code uint32
	msg  string
}

func (e *ValidationError) Code() uint32 {
	return e.code
}

func (e *ValidationError) Error() string {
	return e.msg
}

func NewValidationError(code uint32, msg string) *ValidationError {
	return &ValidationError{code: code, msg: msg}
}
