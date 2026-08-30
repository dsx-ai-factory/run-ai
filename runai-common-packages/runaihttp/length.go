package runaihttp

import (
	"fmt"
)

func AssertMinLength(optionalValue interface{}, fieldName string, restriction int) error {
	if restriction <= 0 {
		return nil
	}

	valuePtr, err := getValuePtr[string](optionalValue)
	if err != nil || valuePtr == nil {
		return nil
	}

	if len(*valuePtr) < restriction {
		return fmt.Errorf("value for %s must have at least %d characters", fieldName, restriction)
	}

	return nil
}

func AssertMaxLength(optionalValue interface{}, fieldName string, restriction int) error {
	if restriction <= 0 {
		return nil
	}

	valuePtr, err := getValuePtr[string](optionalValue)
	if err != nil || valuePtr == nil {
		return nil
	}

	if len(*valuePtr) > restriction {
		return fmt.Errorf("value for %s must have at most %d characters", fieldName, restriction)
	}

	return nil
}
