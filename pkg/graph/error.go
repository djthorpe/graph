package graph

/////////////////////////////////////////////////////////////////////
// TYPES

type result struct {
	err error
	obj bool
}

/////////////////////////////////////////////////////////////////////
// NEW

func newResult(err error, obj bool) *result {
	return &result{err, obj}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e *result) IsErr() bool {
	return e.err != nil
}

func (e *result) Error() string {
	if e.err == nil {
		return ""
	} else {
		return e.err.Error()
	}
}

func (e *result) Obj() bool {
	return e.obj
}

func (e *result) Unwrap() error {
	return e.err
}
