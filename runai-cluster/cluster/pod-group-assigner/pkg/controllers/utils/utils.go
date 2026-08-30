package utils

import (
	"reflect"
)

func IndexOfItemInList[T any](list []T, item T) int {
	for i, element := range list {
		if reflect.DeepEqual(element, item) {
			return i
		}
	}

	return -1
}
