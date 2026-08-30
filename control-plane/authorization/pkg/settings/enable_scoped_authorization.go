package settings

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// EnableScopedAuthorization controls whether scoped authorization is enabled for the tenant.
type EnableScopedAuthorization struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableScopedAuthorization) Category() string {
	return AuthorizationCategory
}

func (e *EnableScopedAuthorization) Description() string {
	return "Enable scoped authorization for fine-grained access control across organizational scopes."
}

func (e *EnableScopedAuthorization) DisplayName() string {
	return "Scoped authorization"
}

func (e *EnableScopedAuthorization) Key() string {
	return string(api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION)
}

func (e *EnableScopedAuthorization) SetDefault() {
	e.Enabled = false
}

func (e *EnableScopedAuthorization) Stage() setting_pkg.Stage {
	return setting_pkg.Alpha
}

func (e *EnableScopedAuthorization) Validate(value interface{}) bool {
	return true
}

func (e *EnableScopedAuthorization) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return EnableScopedAuthorization{
		Enabled: settingValue,
	}
}
