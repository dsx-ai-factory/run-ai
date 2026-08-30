package services

import (
	"context"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
)

type HardTenantDeletionService struct{}

func NewHardTenantDeletionService() *HardTenantDeletionService {
	return &HardTenantDeletionService{}
}

func (h *HardTenantDeletionService) HardDelete(ctx context.Context, tenantId string) error {
	projects, err := dao.GetAllTenantProjects(ctx, tenantId)
	if err != nil {
		return err
	}
	departments, err := dao.GetAllTenantDepartments(ctx, tenantId)
	if err != nil {
		return err
	}

	err = dao.HardDeleteTenantNodeTypes(ctx, tenantId)
	if err != nil {
		return err
	}
	err = dao.HardDeleteOvertimeData(ctx, tenantId, projects, departments)
	if err != nil {
		return err
	}
	err = dao.HardDeleteTenantProjects(ctx, tenantId, projects)
	if err != nil {
		return err
	}
	err = dao.HardDeleteTenantDepartments(ctx, tenantId, departments)
	if err != nil {
		return err
	}
	return nil
}
