package settings

import (
	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/settings/branding"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/settings/notifications"
	ui_appearance "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-appearance"
	ui_workloads "github.com/run-ai/runai/control-plane/bff-service/pkg/settings/ui-workloads"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type settingHandler struct {
	get    func(tenantID string) (manager.Settinger, error)
	update func(tenantID string, updateRequest any) error
}

var settingRegistry = map[string]*settingHandler{
	string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE): {
		get:    notifications.GetAdminMessage,
		update: notifications.UpdateAdminMessage,
	},
	string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM): {
		get:    ui_workloads.GetEnableNIM,
		update: ui_workloads.UpdateEnableNIM,
	},
	string(api.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW): {
		get:    ui_workloads.GetEnableDeletedWorkloadsView,
		update: ui_workloads.UpdateEnableDeletedWorkloadsView,
	},
	string(api.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL): {
		get:    ui_workloads.GetAllowGPT2Model,
		update: ui_workloads.UpdateAllowGPT2Model,
	},
	string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG): {
		get:    ui_workloads.GetEnableModelCatalog,
		update: ui_workloads.UpdateEnableModelCatalog,
	},
	string(api.APPEARANCESETTINGSKEYENUM_BANNER_CONFIG): {
		get:    ui_appearance.GetBannerConfig,
		update: ui_appearance.UpdateBannerConfig,
	},
	string(api.BRANDINGSETTINGSKEYENUM_LOGO): {
		get:    branding.GetLogo,
		update: branding.UpdateLogo,
	},
}

var transformerRegistry = map[utils.CategoryKey]func([]manager.Settinger) any{
	utils.NotificationsCategory: notifications.TransformToNotificationsSettingsResponse,
	utils.UIWorkloadsCategory:   ui_workloads.TransformToWorkloadsSettingsResponse,
	utils.UIAppearanceCategory:  ui_appearance.TransformToAppearanceSettingsResponse,
	utils.BrandingCategory:      branding.TransformToBrandingSettingsResponse,
}
