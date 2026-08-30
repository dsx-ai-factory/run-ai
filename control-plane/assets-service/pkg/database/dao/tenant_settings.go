package dao

import (
	"context"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func tenantSettingsSet(tx orm.DB, rec dbmodel.TenantSettingsRec) error {
	_, err := tx.Where(dbmodel.TenantSettingsRec{
		TenantId: rec.TenantId,
		Key:      rec.Key,
	}).
		Assign(rec).
		FirstOrCreate(&rec)
	return err
}

func tenantSettingsGet(tx orm.DB, rec *dbmodel.TenantSettingsRec) (int64, error) {
	return tx.Model(rec).Where(dbmodel.TenantSettingsRec{
		TenantId: rec.TenantId,
		Key:      rec.Key,
	}).Scan(rec)
}

func setTenantPresetsRevision(tx orm.DB, tenantId int32, revision int32) error {
	return tenantSettingsSet(tx, dbmodel.TenantSettingsRec{
		TenantId: tenantId,
		Key:      dbmodel.KeyPresetsInstalled,
		Value:    fmt.Sprintf("%v", revision),
		Type:     dbmodel.TenantSettingsTypeNumber,
	})
}

func getTenantPresetsRevision(ctx context.Context, tx orm.DB, tenantId int32) (int32, error) {
	rec := dbmodel.TenantSettingsRec{
		TenantId: tenantId,
		Key:      dbmodel.KeyPresetsInstalled,
	}
	rows, err := tenantSettingsGet(tx, &rec)
	if err != nil {
		return 0, err
	}
	if rows == 0 {
		return 0, nil
	}
	revision, err := strconv.Atoi(rec.Value)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid presets revision '%s' for tenant %v: %v", rec.Value, tenantId, err)
		return 0, err
	}
	return int32(revision), nil
}
