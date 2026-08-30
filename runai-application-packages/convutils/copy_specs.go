package convutils

import (
	"encoding/json"
	"reflect"
)

// CopySpecs should be used to copy specs of workload from one type to another
// copier.copy should NOT be used for this purpose, as it does not handle the
// specs variations that we have, like in distributed inference, etc.
func CopySpecs(dest interface{}, src interface{}) error {
	asBytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(asBytes, &dest)
}

// IsEmpty checks if a struct has all zero values
func IsEmpty(v any) bool {
	return reflect.ValueOf(v).IsZero()
}
