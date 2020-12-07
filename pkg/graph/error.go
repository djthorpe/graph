package graph

/////////////////////////////////////////////////////////////////////
// TYPES

type Error struct {
	err error
	obj bool
}

/////////////////////////////////////////////////////////////////////
// NEW

func NewError(err error, obj bool) *Error {
	return &Error{err, obj}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e *Error) IsErr() bool {
	return e.err != nil
}

func (e *Error) Error() string {
	if e.err == nil {
		return ""
	} else {
		return e.err.Error()
	}
}

func (e *Error) Obj() bool {
	return e.obj
}

func (e *Error) Unwrap() error {
	return e.err
}
