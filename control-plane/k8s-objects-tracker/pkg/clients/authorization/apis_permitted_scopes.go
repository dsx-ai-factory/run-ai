package authorization

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/utils/ptr"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func (c Client) CheckPermittedScopes(ctx context.Context, resourceType authorizationClient.ResourceType, action *authorizationClient.Action) (*authorizationClient.PermittedScopesActions, *api.Error) {
	requestInfo := fmt.Sprintf("permitted scopes of %s", resourceType)

	if action != nil {
		requestInfo += fmt.Sprintf(" [%s]", *action)
	}

	requestPayload := authorizationClient.GetPermittedScopesRequest{
		ResourceType: resourceType,
	}

	if action != nil {
		requestPayload.Action = *authorizationClient.NewNullableString(ptr.To(string(*action)))
	}

	request := InitClient().PermissionsAPI.GetPermittedScopes(ctx).GetPermittedScopesRequest(requestPayload)

	resp, r, err := InitClient().PermissionsAPI.GetPermittedScopesExecute(request)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		if r.StatusCode == http.StatusNotFound {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", requestInfo), nil)
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get permitted scopes of %v on %s: %v", action, resourceType, r.Body)
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to receive %s", requestInfo), nil)
	}

	return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request %s", requestInfo), err)
}
