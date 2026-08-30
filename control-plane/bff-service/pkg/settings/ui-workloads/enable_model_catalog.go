package workloads

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type EnableModelCatalog struct {
	Enabled bool `json:"enabled"`
}

func NewEnableModelCatalog() setting_pkg.SettingerWithTransform {
	return &EnableModelCatalog{}
}

func (e *EnableModelCatalog) Category() string {
	return string(utils.UIWorkloadsCategory)
}

func (e *EnableModelCatalog) Description() string {
	return "Enable instant deployment of Hugging Face models via the Models view or the workload submission form. " +
		"For more information, see the <a href='https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai/using-inference/hugging-face-inference' target='_blank'>Deploy Inference Workloads from Hugging Face</a>"
}

func (e *EnableModelCatalog) DisplayName() string {
	return "Hugging face models"
}

func (e *EnableModelCatalog) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG)
}

func (e *EnableModelCatalog) SetDefault() {
	e.Enabled = true
}

func (e *EnableModelCatalog) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *EnableModelCatalog) Validate(value interface{}) bool {
	return true
}

func (e *EnableModelCatalog) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return EnableModelCatalog{
		Enabled: settingValue,
	}
}

func (e *EnableModelCatalog) TransformToResponse() interface{} {
	return api.GetEnableModelCatalog{
		Value: api.GetEnableModelCatalogValue{
			Enabled: &e.Enabled,
		},
		DisplayName: e.DisplayName(),
		Description: e.Description(),
		Stage:       string(e.Stage()),
	}
}

func GetEnableModelCatalog(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*EnableModelCatalog](tenantID)
}

func UpdateEnableModelCatalog(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateWorkloadsSettingByKeyRequest)
	if !ok {
		return utils.ErrInvalidInput
	}
	if req.Enabled == nil {
		return utils.ErrInvalidInput
	}
	settingValue := EnableModelCatalog{
		Enabled: *req.Enabled,
	}
	return setting.Update[*EnableModelCatalog](tenantID, settingValue).UpdatedByUser().Execute()
}
