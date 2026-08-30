package page

import (
	"errors"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
)

// static errors
var (
	ErrNilFilter           = errors.New("filter is nil")
	ErrNilPage             = errors.New("page is nil")
	ErrNilTxn              = errors.New("txn is nil")
	ErrPageSizeNotPositive = errors.New("page size must be a positive number")
	ErrorNextTokenIsZero   = errors.New("next token cannot be with 0 timestamp")
)

const (
	errNextTokenInvalid = "next token timestamp (%s) must be after current timestamp (%s)"
)

func CheckNextTimeStampAfterCurrent(current, next time.Time) error {
	if current.Equal(next) || current.After(next) {
		return fmt.Errorf(errNextTokenInvalid, next.Format(structs.TimeType), current.Format(structs.TimeType))
	}
	return nil
}

func GetPageSize(maxItems, pageSize int32) int32 {
	if maxItems <= 0 {
		return pageSize
	}
	if maxItems < pageSize {
		return maxItems
	}
	return pageSize
}
