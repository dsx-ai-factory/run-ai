package business

import (
	"context"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
)

//go:generate mockery --name RealmSettingsDal
type RealmSettingsDal interface {
	SaveRealmSetting(ctx context.Context, realmSetting *models.RealmSettings) error
	GetRealmSettingByKey(ctx context.Context, tenantId string, settingKey string) (*models.RealmSettings, error)
	GetRealmSettingsByTenantId(ctx context.Context, tenantId string) (*[]models.RealmSettings, error)
}
