package queryapi

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var dateFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC3339Nano,
}

type Validator interface {
	Validate(value string) error
}

type UUIDValidator struct{}

func (UUIDValidator) Validate(value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("invalid UUID format: %s", value)
	}
	return nil
}

type DateValidator struct{}

type UserInputValidator struct{}

func (UserInputValidator) Validate(value string) error {
	if len(value) > 1000 {
		return fmt.Errorf("input too long")
	}

	return nil
}

func (DateValidator) Validate(value string) error {
	for _, format := range dateFormats {
		if _, err := time.Parse(format, value); err == nil {
			return nil
		}
	}

	return fmt.Errorf("invalid date format: %s", value)
}
