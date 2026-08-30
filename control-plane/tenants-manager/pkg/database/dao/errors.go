package dao

import "errors"

type ClientErrorReason int32

const (
	General ClientErrorReason = iota
	NotFound
	Conflict
	NoContent
)

var ErrNotFound = errors.New("not found")

type DaoError struct {
	Reason ClientErrorReason
	Err    error
}

func (e *DaoError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}
