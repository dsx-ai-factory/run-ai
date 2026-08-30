package authorization

import (
	"context"

	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
)

type ClientInterface interface {
	// CheckPermittedScopes returns the scopes in which given action is permitted on a given resource type
	CheckPermittedScopes(ctx context.Context, resourceType authApi.ResourceType, action authApi.Action) (*authApi.PermittedScopes, *authApi.Error)
}
