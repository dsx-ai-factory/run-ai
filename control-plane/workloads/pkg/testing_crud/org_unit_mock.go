package testing_crud

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

// OrgUnitMock returns an orgunit.FactoryInterface mock backed by this client's data
func (mc *RandomsBasedAuthMockClient) OrgUnitMock() orgunit.FactoryInterface {
	return &orgUnitMockAdapter{backendMock: mc}
}

// orgUnitMockAdapter adapts RandomsBasedAuthMockClient to orgunit.FactoryInterface and orgunit.ClientInterface
type orgUnitMockAdapter struct {
	backendMock *RandomsBasedAuthMockClient
}

func (a *orgUnitMockAdapter) Client() orgunit.ClientInterface {
	return a
}

func (a *orgUnitMockAdapter) GetOrgTree(ctx context.Context) ([]orgUnitApi.OrgUnit, *orgUnitApi.Error) {
	result := make([]orgUnitApi.OrgUnit, 0)

	for key, entry := range a.backendMock.RandomsByKey {
		if entry.IsForbidden {
			continue
		}

		parent := a.backendMock.ParentByKey[key]

		orgUnit := orgUnitApi.OrgUnit{
			Id:         key.ScopeId,
			Type:       orgUnitApi.ScopeType(key.ScopeType),
			ParentId:   convutils.Ref(parent.ScopeId),
			ParentType: convutils.Ref(orgUnitApi.ScopeType(parent.ScopeType)),
		}

		result = append(result, orgUnit)
	}

	return result, nil
}

func (a *orgUnitMockAdapter) ListProjects(ctx context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	return []orgUnitApi.ProjectMetadata{}, nil
}

func (a *orgUnitMockAdapter) GetAncestors(ctx context.Context, scopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	return a.backendMock.GetAncestors(ctx, scopeinfo.Scope(scopeType), scopeId)
}
