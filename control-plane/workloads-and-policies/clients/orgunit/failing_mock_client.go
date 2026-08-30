package orgunit

import (
	"context"
	"net/http"

	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

type FailingMockClient struct {
	FailGetOrgTree bool
}

func (fmc *FailingMockClient) Client() ClientInterface {
	return fmc
}

func (fmc *FailingMockClient) GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error) {
	if fmc.FailGetOrgTree {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "mock: failed to get org tree")
	}

	return []openapi.OrgUnit{}, nil
}

func (fmc *FailingMockClient) ListProjects(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error) {
	return []openapi.ProjectMetadata{}, nil
}

func (fmc *FailingMockClient) GetAncestors(ctx context.Context, scopeType, scopeId string) (*openapi.Ancestors, *openapi.Error) {
	return &openapi.Ancestors{}, nil
}
