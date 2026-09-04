package auth

import (
	"net/http"

	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type MockService struct {
	*auth_mocks.IAuthorizationClient
}

func (m MockService) AuthorizeMiddleware(next http.Handler, _ string, _ string, _ bool, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader, err := runaictx.ParseSubjectId(r.Context())
		if err != nil || authHeader == "" {
			http.Error(w, `{ "message" : "forbidden"`, http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r.WithContext(r.Context()))
		}
	}
}
