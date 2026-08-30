package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department/model"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

type DepartmentClient interface {
	GetDepartmentList(ctx context.Context, filters []string, limit, offset int32) (*openapi.GetDepartments200Response, error)
}

type DepartmentService struct {
	departmentClient DepartmentClient
}

func NewDepartmentService(departmentClient DepartmentClient) *DepartmentService {
	return &DepartmentService{departmentClient: departmentClient}
}

func (bns *DepartmentService) GetDepartmentList(ctx context.Context, filters []model.Filter, limit, offset int32) (*BriefDepartmentHandler, error) {
	var filterList []string
	for _, wlFilter := range filters {
		filterList = append(filterList, wlFilter.Get())
	}

	list, err := bns.departmentClient.GetDepartmentList(ctx, filterList, limit, offset)
	if err != nil {
		return nil, err
	}
	return NewBriefDepartmentHandler(list), nil
}
