package auth

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
)

// SimpleMockAuthService mocks the authorization service middleware for testing purposes.
// the mock is very trivial: if subject-id header is given - approve the request, otherwise - reject it.
// it works well with the functions:
// headerForAdmin, headerForNonAdmin (both approved) and headerWithMissingInfo (rejected)

type SimpleMockAuthService struct {
	*auth_mocks.IAuthorizationClient
}

func (m SimpleMockAuthService) AuthorizeMiddleware(next http.Handler, _ /*resourceType*/ string, _ /*requestAction*/ string,
	_ /*hasResourceId*/ bool, _ /*resourceIdParam*/ string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader, err := runaictx.ParseSubjectId(r.Context())
		if err != nil || authHeader == "" {
			http.Error(w, "{ \"message\" : \"forbidden\"", http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r.WithContext(r.Context()))
		}
	}
}
