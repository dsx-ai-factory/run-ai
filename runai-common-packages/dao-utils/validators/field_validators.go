package validators

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Validator interface {
	Validate(value string) error
}

type FieldValidator interface {
	GetColumnsName() string
	GetValidator() Validator
}

type FieldValidation struct {
	columnsName string
	validator   Validator
}

func (fc *FieldValidation) GetColumnsName() string {
	return fc.columnsName
}

func (fc *FieldValidation) GetValidator() Validator {
	return fc.validator
}

type UUIDValidator struct {
}

func (*UUIDValidator) Validate(value string) error {
	_, err := uuid.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %s", value)
	}
	return nil
}

type DateValidator struct {
}

func (*DateValidator) Validate(value string) error {
	// Try common date formats
	dateFormats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, value); err == nil {
			return nil
		}
	}
	return fmt.Errorf("invalid date format: %s", value)
}

type UserInputValidator struct {
}

func (*UserInputValidator) Validate(value string) error {
	if len(value) > 1000 {
		return fmt.Errorf("input too long")
	}

	return nil
}

func NewAlwaysValidFieldValidation(columnName string) *FieldValidation {
	return NewFieldValidation(columnName, &UserInputValidator{})
}

func NewUUIDFieldValidation(columnName string) *FieldValidation {
	return NewFieldValidation(columnName, &UUIDValidator{})
}

func NewDateFieldValidation(columnName string) *FieldValidation {
	return NewFieldValidation(columnName, &DateValidator{})
}
