package services

import (
	"fmt"
	"regexp"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

func normalizeListParams(offset int32, limit int32, defaultLimit int, maxLimit int) (int, int) {
	off := int(offset)
	if off < 0 {
		off = 0
	}
	lim := int(limit)
	if lim < 1 {
		lim = defaultLimit
	}
	if lim > maxLimit {
		lim = maxLimit
	}
	return off, lim
}

// Same rule as swagger/source/schemas/Mapping/MappingFieldName.yaml (OpenAPI cannot validate map keys under additionalProperties).
// and also Field groups, item and spec selectors have the same naming rules as field names
var mappingElementNameRegexp = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$`)

func validateMappingNameKey(elementName string, elementKind string) *api.Error {
	n := len(elementName)
	if n < 1 || n > 63 {
		return api.NewValidationError(fmt.Sprintf("%s name %q must be between 1 and 63 characters", elementKind, elementName))
	}
	if !mappingElementNameRegexp.MatchString(elementName) {
		return api.NewValidationError(fmt.Sprintf(
			"%s name %q is invalid: must be alphanumeric with optional internal hyphens, and must not start or end with a hyphen", elementKind, elementName))
	}
	return nil
}
