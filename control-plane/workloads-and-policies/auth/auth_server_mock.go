package auth

import (
	"context"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
)

// SimpleMockAuthService mocks the authorization service middleware for testing purposes.
// the mock is very trivial: if subject-id header is given - approve the request, otherwise - reject it.
// it works well with the functions:
// headerForAdmin, headerForNonAdmin (both approved) and headerWithMissingInfo (rejected)
// an enhanced option is to set a permitted context that is loaded to the context

type SimpleMockAuthService struct {
	*auth_mocks.IAuthorizationClient

	permittedScopes *authClient.PermittedScopes
}

func NewSimpleMockAuthService() *SimpleMockAuthService {
	return &SimpleMockAuthService{
		permittedScopes: &authClient.PermittedScopes{},
	}
}

func (m *SimpleMockAuthService) SetNoPermissions() *SimpleMockAuthService {
	*m.permittedScopes = authClient.PermittedScopes{}
	return m
}

func (m *SimpleMockAuthService) AddSystemScope() *SimpleMockAuthService {
	m.permittedScopes.System = convutils.Ref(true)
	return m
}

func (m *SimpleMockAuthService) AddPermittedTenants(randoms ...*testutils.RandomValues) *SimpleMockAuthService {
	for _, random := range randoms {
		if !random.IsForbidden {
			m.permittedScopes.Tenants = append(m.permittedScopes.Tenants, random.TenantIdStr)
		}
	}

	return m
}

func (m *SimpleMockAuthService) AddPermittedClusters(randoms ...*testutils.RandomValues) *SimpleMockAuthService {
	for _, random := range randoms {
		if !random.IsForbidden {
			m.permittedScopes.Clusters = append(m.permittedScopes.Clusters, random.ClusterId)
		}
	}

	return m
}

func (m *SimpleMockAuthService) AddPermittedDepartments(randoms ...*testutils.RandomValues) *SimpleMockAuthService {
	for _, random := range randoms {
		if !random.IsForbidden {
			m.permittedScopes.Departments = append(m.permittedScopes.Departments, random.DepartmentId)
		}
	}

	return m
}

func (m *SimpleMockAuthService) AddPermittedProjects(randoms ...*testutils.RandomValues) *SimpleMockAuthService {
	for _, random := range randoms {
		if !random.IsForbidden {
			m.permittedScopes.Projects = append(m.permittedScopes.Projects, random.ProjectIdStr)
		}
	}

	return m
}

func (m SimpleMockAuthService) AuthorizeMiddleware(next http.Handler, _ /*resourceType*/ string, _ /*requestAction*/ string,
	_ /*hasResourceId*/ bool, _ /*resourceIdParam*/ string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader, err := runaictx.ParseSubjectId(r.Context())
		if err != nil || authHeader == "" {
			http.Error(w, "{ \"message\" : \"forbidden\"", http.StatusForbidden)
		} else {
			if m.permittedScopes == nil {
				next.ServeHTTP(w, r.WithContext(r.Context()))
			} else {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), structs.PermittedScopesKey, m.permittedScopes)))
			}
		}
	}
}
