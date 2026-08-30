package settings

import (
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// EnableNgcIntegration controls whether users can browse/select NGC catalog container images.
type EnableNgcIntegration struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableNgcIntegration) Category() string {
	return CatalogCategory
}

func (e *EnableNgcIntegration) Description() string {
	return "Enable to browse and select container images from the NVIDIA NGC catalog for faster environment setup."
}

func (e *EnableNgcIntegration) DisplayName() string {
	return "NGC catalog"
}

func (e *EnableNgcIntegration) Key() string {
	return string(api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION)
}

func (e *EnableNgcIntegration) SetDefault() {
	e.Enabled = false
}

func (e *EnableNgcIntegration) Stage() setting_pkg.Stage {
	return setting_pkg.Beta
}

func (e *EnableNgcIntegration) Validate(value interface{}) bool {
	return true
}

func (e *EnableNgcIntegration) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return EnableNgcIntegration{
		Enabled: settingValue,
	}
}
