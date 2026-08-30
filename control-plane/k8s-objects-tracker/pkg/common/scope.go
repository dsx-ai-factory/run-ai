package common

import (
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
)

// ValidateScopeType rejects scopeType values outside the closed set used by
// BuildScopeQuery. Empty scopeType is treated as valid so callers can skip
// validation when the parameter is optional and omitted.
func ValidateScopeType(scopeType string) *api.Error {
	if scopeType == "" {
		return nil
	}

	switch authorizationClient.ScopeType(scopeType) {
	case authorizationClient.SCOPETYPE_TENANT,
		authorizationClient.SCOPETYPE_CLUSTER,
		authorizationClient.SCOPETYPE_DEPARTMENT,
		authorizationClient.SCOPETYPE_PROJECT:
		return nil
	default:
		return api.NewValidationError("invalid scopeType")
	}
}
