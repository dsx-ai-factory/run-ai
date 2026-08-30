package control_plane

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
	"k8s.io/utils/ptr"

	authzclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"

	authzapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
)

type ControlPlaneAuthorizationProvider interface {
	authorizers.Authorizer
}

type ControlPlaneAuthorizer struct {
	authzClient authzclient.IAuthorizationExternalClient
}

var _ ControlPlaneAuthorizationProvider = &ControlPlaneAuthorizer{}

func NewControlPlaneAuthorizerWithClient(authzClient authzclient.IAuthorizationExternalClient) *ControlPlaneAuthorizer {
	return &ControlPlaneAuthorizer{
		authzClient: authzClient,
	}
}

func NewControlPlaneAuthorizer() *ControlPlaneAuthorizer {
	return NewControlPlaneAuthorizerWithClient(authzclient.NewAuthorizationExternalClient())
}

func (cpa *ControlPlaneAuthorizer) Authorize(ctx context.Context, request *authorizers.Request) (allowed bool,
	reason authorizers.Reason, err error,
) {
	requestAction, resourceType := workloadActionAndTypeToAuthorizationActionAndResourceType(request.Action, request.Kind)

	req := cpa.authzClient.EvaluatePermission(ctx).BackendEvaluationRequest(authzapi.BackendEvaluationRequest{
		ResourceId:    *authzapi.NewNullableString(ptr.To(request.WorkloadUID)),
		RequestAction: *authzapi.NewNullableAction(&requestAction),
		ResourceType:  *authzapi.NewNullableResourceType(&resourceType),
	})

	_, res, err := req.Execute()
	if err != nil {
		return false, authorizers.ReasonError, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusNoContent:
		return true, authorizers.ReasonAuthorized, nil
	case http.StatusForbidden:
		return false, authorizers.ReasonUnauthorized,
			fmt.Errorf("authorization: user [%s] is forbidden from verb [%s] on workload uid [%s] with name [%s]",
				request.AuthInfo.Subject, request.Action, request.WorkloadUID, request.Workload)
	default:
		statusCode := res.StatusCode

		if res.Body == nil {
			return false, authorizers.ReasonError,
				fmt.Errorf("authorization: response with status code [%d] has no body", statusCode)
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return false, authorizers.ReasonError,
				fmt.Errorf("authorization: failed reading body for failure response with code [%d], error: %v",
					statusCode, err)
		}

		return false, authorizers.ReasonError,
			fmt.Errorf("authorization: unexpected error occurred: status code [%d], response body: %s",
				res.StatusCode, string(resBody))
	}
}

func workloadActionAndTypeToAuthorizationActionAndResourceType(
	wa api.WorkloadAction, wk api.WorkloadType,
) (authzapi.Action, authzapi.ResourceType) {
	authzAction := workloadActionToAuthorizationAction(wa)
	if authzAction == authzapi.ACTION_READ {
		return authzapi.ACTION_READ, authzapi.RESOURCETYPE_WORKLOADS
	} else {
		// TODO use this function once resolved:
		// workloadKindToResourceType(wk)
		return authzapi.ACTION_UPDATE, authzapi.RESOURCETYPE_WORKLOADS
	}
}

func workloadActionToAuthorizationAction(wa api.WorkloadAction) authzapi.Action {
	switch wa {
	case api.Status, api.PortForward, api.Logs, api.LogsDownload:
		return authzapi.ACTION_READ
	case api.Exec, api.Attach:
		return authzapi.ACTION_UPDATE
	default:
		panic(fmt.Sprintf("authorization: could not map workload action [%s] to any authorization action",
			wa.String()))
	}
}
