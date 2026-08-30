package hard_delete

import (
	"context"

	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
)

type HardDeletionService struct{}

// NewHardDeletionService creates a default api service
func NewHardDeletionService() *HardDeletionService {
	return &HardDeletionService{}
}

// DeleteTenant - Hard Deletes Tenant Access Rules
func (h *HardDeletionService) DeleteTenant(ctx context.Context, tenantId string) error {
	return dao.HardDeleteTenantAccessRules(ctx, tenantId)
}
