package cached

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
	control_plane "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers/control-plane"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type AuthorizationResponse struct {
	Allowed bool
	Reason  authorizers.Reason
	Error   error
}

type ControlPlaneAuthorizationProvider interface {
	authorizers.Authorizer
}

type ControlPlaneAuthorizer struct {
	cpa control_plane.ControlPlaneAuthorizationProvider
}

var _ ControlPlaneAuthorizationProvider = &ControlPlaneAuthorizer{}

func NewCachedControlPlaneAuthorizer(provider control_plane.ControlPlaneAuthorizationProvider) *ControlPlaneAuthorizer {
	return &ControlPlaneAuthorizer{
		cpa: provider,
	}
}

func (cpa *ControlPlaneAuthorizer) Authorize(
	ctx context.Context, request *authorizers.Request,
) (allowed bool, reason authorizers.Reason, err error) {
	key := cacheKeyForRequest(request)

	if res, err := cache.Get[AuthorizationResponse](ctx, key); err != nil {
		logging.Warnf("failed to get AuthorizationResponse from cache, key [%s]", key)
	} else if res != nil {
		return res.Allowed, res.Reason, res.Error
	}

	allowed, reason, err = cpa.cpa.Authorize(ctx, request)
	if err != nil {
		if reason == authorizers.ReasonError {
			// don't cache responses with unknown reasons
			return allowed, reason, err
		}
	}

	if err = cache.Set[AuthorizationResponse](ctx, key, AuthorizationResponse{
		Allowed: allowed,
		Reason:  reason,
		Error:   err,
	}); err != nil {
		logging.Warnf("failed to set AuthorizationResponse to cache, key [%s]", key)
	}

	return allowed, reason, err
}

func cacheKeyForRequest(request *authorizers.Request) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v",
		request.AuthInfo.Subject, request.Workload, request.Kind, request.Project, request.Action)
}
