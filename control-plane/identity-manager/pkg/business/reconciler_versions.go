package business

import (
	"context"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
)

//go:generate mockery --name ReconcilerVersionsDal
type ReconcilerVersionsDal interface {
	CreateReconcilerVersions(ctx context.Context, realm *models.ReconcilerVersions) (*models.ReconcilerVersions, error)
	UpdateServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error
	UpdateAccessKeysReconcilerVersions(ctx context.Context, tenantId string, version string) error
	UpdateInternalServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error
	GetReconcilerVersionsByTenantId(ctx context.Context, tenantId string) (*models.ReconcilerVersions, error)
	GetReconcilerVersions(ctx context.Context) (*[]models.ReconcilerVersions, error)
}
