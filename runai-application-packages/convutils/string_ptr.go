package convutils

func IsEmptyOrNull(stringPtr *string) bool {
	return Deref(stringPtr, "") == ""
}

func IsStringHasValue(stringPtr *string) bool {
	return Deref(stringPtr, "") != ""
}

// BoolIsTrue returns true if the bool pointer is true, false otherwise
func BoolIsTrue(boolPtr *bool) bool {
	return Deref(boolPtr, false)
}

func BoolIsFalse(boolPtr *bool) bool {
	return !Deref(boolPtr, false)
}

func EmptyIfNull(stringPtr *string) *string {
	if stringPtr == nil {
		return Ref("")
	}
	return stringPtr
}
