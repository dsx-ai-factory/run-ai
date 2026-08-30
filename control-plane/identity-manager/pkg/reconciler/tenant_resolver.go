package reconciler

import (
	"context"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

//go:generate mockery --name TenantResolver
type TenantResolver interface {
	GetRealmTenantId(ctx context.Context, realm string) string
}

type RealmTenantResolver struct {
	realmDal business.RealmDal
}

func NewRealmTenantResolver(ctx context.Context, realmDal business.RealmDal) *RealmTenantResolver {
	return &RealmTenantResolver{
		realmDal: realmDal,
	}
}

func (r *RealmTenantResolver) GetRealmTenantId(ctx context.Context, realm string) string {
	tenantId, ok := ctx.Value(runaictx.CfTenantId).(string)
	if ok && tenantId != "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Tenant id %s was found in context", tenantId)
		return tenantId
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Tenant id was not found in context, searching in db %s", realm)
	tenantId, err := r.getTenantIdByRealm(ctx, realm)
	if err == nil {
		return tenantId
	}
	return ""
}

func (r *RealmTenantResolver) getTenantIdByRealm(ctx context.Context, realmName string) (string, error) {
	realm, err := r.realmDal.GetRealmByName(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get realm by name from db: %s. Error %s", realmName, err.Error())
		return "", err
	}
	return realm.TenantId, nil
}
