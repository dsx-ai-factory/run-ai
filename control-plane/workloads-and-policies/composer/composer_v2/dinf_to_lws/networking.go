package dinf_to_lws

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/init_containers"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func (dinfComposer *DistributedInferenceToLWSComposer) processNetworkingFields(
	src *workloadsApi.DistributedInferenceServingPort,
	target *LeaderWorkerSetTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return workloadsApi.NewError(http.StatusBadRequest, "servingPort is required for distributed inference workloads")
		}

		projected := workloadsApi.SupersetServingPort{}
		if err := convutils.CopySpecs(&projected, src); err != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to project servingPort: %v", err))
		}

		config := dinfComposer.fc.ClusterStorageConfig
		if config == nil {
			config = composer.DefaultClusterStorageConfig()
		}

		cookie := lwsNetworkingCookie{
			config:       config,
			projectName:  dinfComposer.fc.ProjectName,
			workloadName: dinfComposer.fc.WorkloadName,
			lwsName:      dinfComposer.leaderWorkerSetName(),
			serviceName:  fmt.Sprintf("dinfw-%s-sp-0", dinfComposer.fc.WorkloadName),
		}

		runner := &composer.StepRunner{}

		// Shared out-parameter: composeServingPort fills authInfo (and
		// cookie.servingURL) during the first step; buildAuthProxySidecar
		// reads both to build the oauth2-proxy sidecar. Closures capture the
		// pointers so the batch sees consistent values at execution time.
		authInfo := composer.PortAuthInfo{}

		runner.Batch(
			dinfComposer.composeServingPort(&projected, target, &cookie, &authInfo),
			dinfComposer.buildAuthProxySidecar(&projected, target, &cookie, &authInfo),
		)

		return runner.Error()
	}
}

func (dinfComposer *DistributedInferenceToLWSComposer) composeServingPort(
	src *workloadsApi.SupersetServingPort,
	target *LeaderWorkerSetTarget,
	cookie *lwsNetworkingCookie,
	authInfo *composer.PortAuthInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.Port.Get() == nil || *src.Port.Get() <= 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "servingPort: mandatory port is not provided")
		}

		protocol := src.Protocol.Get()
		if protocol != nil && *protocol != string(workloadsApi.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP) {
			return workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("servingPort: unsupported protocol %q, only 'http' is supported", *protocol))
		}

		return dinfComposer.fc.ComposeServingPort(
			convutils.Ref(workloadsApi.REPLICATYPE_LEADER),
			src,
			target,
			servingPortToLWS,
			cookie,
			authInfo,
		)()
	}
}

// buildAuthProxySidecar returns a batch step that installs an oauth2-proxy
// sidecar next to the leader container when the serving port is not public.
// No-op for public serving ports. Must run after composeServingPort so that
// cookie.servingURL and authInfo have been populated.
func (dinfComposer *DistributedInferenceToLWSComposer) buildAuthProxySidecar(
	src *workloadsApi.SupersetServingPort,
	target *LeaderWorkerSetTarget,
	cookie *lwsNetworkingCookie,
	authInfo *composer.PortAuthInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if authInfo.AuthorizationType == nil ||
			*authInfo.AuthorizationType == composer.ConnectionAuthorizationType_PUBLIC {
			return nil
		}
		port := src.Port.Get()
		if port == nil {
			return workloadsApi.NewError(http.StatusInternalServerError,
				"servingPort: projected port is missing")
		}
		return init_containers.BuildOauth2ProxySidecar(
			&target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec,
			init_containers.ServingProxySpec{
				ContainerPort:    *port,
				ServingURL:       cookie.servingURL,
				CookiePrefix:     cookie.serviceName,
				AuthorizedUsers:  strings.Join(authInfo.AuthorizedUsers, ","),
				AuthorizedGroups: strings.Join(authInfo.AuthorizedGroups, ","),
				WorkloadName:     cookie.workloadName,
				ProjectName:      cookie.projectName,
				Config:           cookie.config,
			},
		)
	}
}
