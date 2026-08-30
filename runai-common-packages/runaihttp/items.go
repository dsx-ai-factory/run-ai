package runaihttp

import (
	"fmt"
	"reflect"
)

func AssertMinItems(optionalValue interface{}, fieldName string, restriction int) error {
	if restriction <= 0 {
		return nil
	}

	length, err := getNumberOfItems(optionalValue)
	if err != nil {
		return nil
	}

	if length < restriction {
		return fmt.Errorf("value for %s must have at least %d items", fieldName, restriction)
	}

	return nil
}

func AssertMaxItems(optionalValue interface{}, fieldName string, restriction int) error {
	if restriction < 0 {
		return nil
	}

	length, err := getNumberOfItems(optionalValue)
	if err != nil {
		return nil
	}

	if length > restriction {
		return fmt.Errorf("value for %s must have at most %d items", fieldName, restriction)
	}

	return nil
}

func getNumberOfItems(optionalValue interface{}) (int, error) {
	rt := reflect.ValueOf(optionalValue)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rt.Len(), nil
	default:
		return -1, fmt.Errorf("invalid type %T", optionalValue)
	}
}
