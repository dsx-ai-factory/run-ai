package composer

import (
	"fmt"
	"net/http"

	"github.com/modern-go/reflect2"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

func BoolToBool(src *bool, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "StringToString: target cannot be nil")
	}

	if src != nil {
		switch t := target.(type) {
		case *bool:
			*t = *src
		case **bool:
			*t = src
		case *workloadsApi.NullableBool:
			t.Set(src)
		default:
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("BoolToBool: unsupported target type %T", target))
		}
	}
	return nil
}

func StringToString(src *string, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "StringToString: target cannot be nil")
	}

	if src != nil {
		switch t := target.(type) {
		case *string:
			*t = *src
		case **string:
			*t = src
		case *workloadsApi.NullableString:
			t.Set(src)
		case *corev1.PullPolicy:
			*t = corev1.PullPolicy(*src)
		case *corev1.RestartPolicy:
			*t = corev1.RestartPolicy(*src)
		case *corev1.DNSPolicy:
			*t = corev1.DNSPolicy(*src)
		case *corev1.SeccompProfileType:
			*t = corev1.SeccompProfileType(*src)
		default:
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToString: unsupported target type %T", target))
		}
	}
	return nil
}

func AddStringToArray(src *string, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "AddStringToArray: target cannot be nil")
	}

	if src != nil {
		switch t := target.(type) {
		case *[]string:
			*t = append(*t, *src)
		default:
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("AddStringToArray: unsupported target type %T", target))
		}
	}

	return nil
}

func AddStringsAoArray(src []string, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "AddStringsToArray: target cannot be nil")
	}

	if src != nil {
		switch t := target.(type) {
		case *[]string:
			*t = append(*t, src...)
		default:
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("AddStringsAoArray: unsupported target type %T", target))
		}
	}

	return nil
}

func UnixStrToStringArray(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	parsed, err := ParseCommandLine(*src)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("UnixStrToStringArray: failed to parse command line: %v", err))
	}

	targetPtr, ok := target.(*[]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("UnixStrToStringArray: target type %T is not *[]string", target))
	}

	*targetPtr = parsed
	return nil
}

func StringToMapString(src *string, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToMapString: target type %T is not *map[string]string", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	key, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToMapString: cookie type %T is not string", cookie))
	}

	(*targetPtr)[key] = *src

	return nil
}

// StringToMapStringLabelSafe is like StringToMapString but sanitizes the value
// to ensure it's valid for Kubernetes labels. Labels only allow alphanumeric
// characters, '-', '_', and '.', and must start/end with alphanumeric.
// Invalid characters like '@' are replaced with '_'.
func StringToMapStringLabelSafe(src *string, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToMapStringLabelSafe: target type %T is not *map[string]string", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	key, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToMapStringLabelSafe: cookie type %T is not string", cookie))
	}

	(*targetPtr)[key] = sanitizeLabelValue(*src)

	return nil
}

const maxLabelLength = 63

// sanitizeLabelValue replaces invalid label characters with underscores.
// Kubernetes labels allow: alphanumeric, '-', '_', '.'
// Must start and end with alphanumeric and not exceed 63 characters.
func sanitizeLabelValue(s string) string {
	if s == "" {
		return s
	}

	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isValidLabelChar(c) {
			result[i] = c
		} else {
			result[i] = '_'
		}
	}

	// Truncate to max length
	if len(result) > maxLabelLength {
		result = result[:maxLabelLength]
	}

	// Handle empty result after potential processing
	if len(result) == 0 {
		return ""
	}

	// Ensure starts with alphanumeric
	if !isAlphanumeric(result[0]) {
		result[0] = 'x'
	}

	// Ensure ends with alphanumeric
	if !isAlphanumeric(result[len(result)-1]) {
		result[len(result)-1] = 'x'
	}

	return string(result)
}

func isValidLabelChar(c byte) bool {
	return isAlphanumeric(c) || c == '-' || c == '_' || c == '.'
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func BoolToMapString(src *bool, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("BoolToMapString: target type %T is not *map[string]string", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	key, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToMapString: cookie type %T is not string", cookie))
	}

	(*targetPtr)[key] = "true"

	return nil
}

func MapToMapString(src map[string]*string, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapToMapString: unsupported target type %T", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	for k, v := range src {
		if v == nil {
			continue
		}
		(*targetPtr)[k] = *v
	}

	return nil
}

func NonZeroIntToInt[T IntType](src *T, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "NonZeroIntToInt: target cannot be nil")
	}

	val := convutils.Deref(src, 0)
	if val == 0 {
		return nil
	}

	switch t := target.(type) {
	case *int:
		*t = int(val)
	case *int32:
		*t = int32(val)
	case *int64:
		*t = int64(val)
	case **int:
		*t = convutils.Ref(int(val))
	case **int32:
		*t = convutils.Ref(int32(val))
	case **int64:
		*t = convutils.Ref(int64(val))
	default:
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NonZeroIntToInt: target type %T is not supported", target))
	}

	return nil
}

func IntToInt[T IntType](src *T, target any, _ any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "NonZeroIntToInt: target cannot be nil")
	}

	val := convutils.Deref(src, 0)

	switch t := target.(type) {
	case *int:
		*t = int(val)
	case *int32:
		*t = int32(val)
	case *int64:
		*t = int64(val)
	case **int:
		*t = convutils.Ref(int(val))
	case **int32:
		*t = convutils.Ref(int32(val))
	case **int64:
		*t = convutils.Ref(int64(val))
	default:
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NonZeroIntToInt: target type %T is not supported", target))
	}

	return nil
}

func NonZeroIntToMapString[T IntType](src *T, target any, cookie any) *workloadsApi.Error {
	if reflect2.IsNil(target) {
		return workloadsApi.NewError(http.StatusInternalServerError, "NonZeroIntToInt: target cannot be nil")
	}

	val := convutils.Deref(src, 0)
	if val == 0 {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NonZeroIntToMapString: unsupported target type %T", target))
	}

	if reflect2.IsNil(cookie) {
		return workloadsApi.NewError(http.StatusInternalServerError, "NonZeroIntToMapString: cookie cannot be nil")
	}

	key, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NonZeroIntToMapString: cookie type %T is not string", cookie))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	(*targetPtr)[key] = fmt.Sprintf("%d", val)

	return nil
}
