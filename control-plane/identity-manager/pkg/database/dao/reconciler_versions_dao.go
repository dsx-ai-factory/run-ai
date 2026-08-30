package dao

import (
	"context"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	ReconcilerVersionsTable = "reconciler_versions"
)

func NewReconcilerVersionsDao() business.ReconcilerVersionsDal {
	return &ReconcilerVersions{}
}

type ReconcilerVersions struct{}

func (r ReconcilerVersions) TableName() string {
	return ReconcilerVersionsTable
}

func (r ReconcilerVersions) CreateReconcilerVersions(ctx context.Context, reconcilerVersions *models.ReconcilerVersions) (*models.ReconcilerVersions, error) {
	_, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Create(reconcilerVersions)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return reconcilerVersions, nil
}

func (r ReconcilerVersions) UpdateServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error {
	reconcilerVersionToUpdate := models.ReconcilerVersions{
		TenantId:        tenantId,
		ServiceAccounts: version,
	}
	return r.updateReconcilerVersion(ctx, &reconcilerVersionToUpdate)
}

func (r ReconcilerVersions) UpdateAccessKeysReconcilerVersions(ctx context.Context, tenantId string, version string) error {
	reconcilerVersionToUpdate := models.ReconcilerVersions{
		TenantId:   tenantId,
		AccessKeys: version,
	}
	return r.updateReconcilerVersion(ctx, &reconcilerVersionToUpdate)
}

func (r ReconcilerVersions) UpdateInternalServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error {
	reconcilerVersionToUpdate := models.ReconcilerVersions{
		TenantId:                tenantId,
		InternalServiceAccounts: version,
	}
	return r.updateReconcilerVersion(ctx, &reconcilerVersionToUpdate)
}

func (r ReconcilerVersions) updateReconcilerVersion(ctx context.Context, reconcilerVersions *models.ReconcilerVersions) error {
	rowsAffected, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Where(&models.ReconcilerVersions{
		TenantId: reconcilerVersions.TenantId,
	}).Updates(reconcilerVersions)
	if err != nil {
		return utils.ParseGormError(err)
	}
	if rowsAffected == 0 {
		_, err := r.CreateReconcilerVersions(ctx, reconcilerVersions)
		return err
	}
	return nil
}

func (r ReconcilerVersions) GetReconcilerVersionsByTenantId(ctx context.Context, tenantId string) (*models.ReconcilerVersions, error) {
	reconcilerVersion := &models.ReconcilerVersions{
		TenantId: tenantId,
	}
	return r.getReconcilerVersion(ctx, reconcilerVersion)
}

func (r ReconcilerVersions) GetReconcilerVersions(ctx context.Context) (*[]models.ReconcilerVersions, error) {
	reconcilerVersions := &[]models.ReconcilerVersions{}
	err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Find(reconcilerVersions)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return reconcilerVersions, nil
}

func (r ReconcilerVersions) getReconcilerVersion(ctx context.Context, ReconcilerVersion *models.ReconcilerVersions) (*models.ReconcilerVersions, error) {
	err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Where(ReconcilerVersion).First(ReconcilerVersion)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return ReconcilerVersion, nil
}
