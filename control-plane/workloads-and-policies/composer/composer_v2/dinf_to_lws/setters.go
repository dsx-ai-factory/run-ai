package dinf_to_lws

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/init_containers"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const servingPortServicePort = int32(80)

// lwsNetworkingCookie carries serving-port composition context. servingURL is
// resolved by the setter and consumed by the follow-up sidecar composition step.
type lwsNetworkingCookie struct {
	config       composer.ClusterStorageConfig
	projectName  string
	workloadName string
	lwsName      string
	serviceName  string
	servingURL   string
}

func servingPortToLWS(src *composer.ServingPortInfo, target any, rawCookie any) *workloadsApi.Error {
	lwsTarget, ok := target.(*LeaderWorkerSetTarget)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("servingPortToLWS: target type %T is not *LeaderWorkerSetTarget", target))
	}
	cookie, ok := rawCookie.(*lwsNetworkingCookie)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("servingPortToLWS: cookie type %T is not *lwsNetworkingCookie", rawCookie))
	}
	port := src.PortInfo.Container.Get()
	if port == nil {
		return workloadsApi.NewError(http.StatusInternalServerError, "servingPortToLWS: projected port is missing")
	}

	authEnabled := src.AuthInfo.AuthorizationType != nil &&
		*src.AuthInfo.AuthorizationType != composer.ConnectionAuthorizationType_PUBLIC
	targetPort := *port
	if authEnabled {
		targetPort = composer.Oauth2ContainerDefaultPort
	}
	servicePort := workloadsManagerApi.NewServicePort(servingPortServicePort)
	servicePort.SetName("serving-port")
	servicePort.SetTargetPort(targetPort)
	servicePort.SetProtocol(workloadsManagerApi.SERVICEPROTOCOL_TCP)
	serviceSpec := workloadsManagerApi.NewServiceSpec(
		cookie.serviceName,
		map[string]string{
			lwsv1.SetNameLabelKey:     cookie.lwsName,
			lwsv1.WorkerIndexLabelKey: "0",
		},
		[]workloadsManagerApi.ServicePort{*servicePort},
	)
	service := workloadsManagerApi.NewNetworkingService()
	service.SetSpec(*serviceSpec)

	endpoint := workloadsManagerApi.NetworkingEndpoint{
		Services: []workloadsManagerApi.NetworkingService{*service},
	}
	exposeExternally := convutils.Deref(src.ExposedExternally, true)
	urlString := convutils.Deref(src.ExposedUrl, "")
	if exposeExternally {
		if urlString == "" {
			autoURL, err := init_containers.ComposeAutoExposedURL(
				cookie.config.ClusterDomain(),
				cookie.projectName,
				cookie.workloadName,
				"",
				cookie.config.SubdomainSupportEnabled(),
			)
			if err != nil {
				return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose serving URL: %v", err))
			}
			urlString = autoURL.String()
		}
		if err := composer.ValidateUrlSyntax(&urlString, false); err != nil {
			return err
		}
		externalAccess, err := composer_v2.BuildExternalAccess(
			cookie.config,
			cookie.serviceName,
			urlString,
			servingPortServicePort,
		)
		if err != nil {
			return err
		}
		endpoint.SetExternalAccess(*externalAccess)
	} else {
		urlString = fmt.Sprintf("http://%s.%s.svc.cluster.local/", cookie.serviceName, lwsTarget.LeaderWorkerSet.Namespace)
	}
	cookie.servingURL = urlString
	lwsTarget.Config.Networking = append(lwsTarget.Config.Networking, endpoint)

	leaderContainer := &lwsTarget.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0]
	leaderContainer.Ports = append(leaderContainer.Ports, corev1.ContainerPort{
		Name:          fmt.Sprintf("port-%d", *port),
		ContainerPort: *port,
		Protocol:      corev1.ProtocolTCP,
	})
	if leaderContainer.ReadinessProbe == nil {
		leaderContainer.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(*port),
				},
			},
		}
	}
	return nil
}

func startupPolicyToLWS(src *string, target any, _ any) *workloadsApi.Error {
	policy, ok := target.(*lwsv1.StartupPolicyType)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("startupPolicyToLWS: target type %T is not *v1.StartupPolicyType", target))
	}
	if src != nil {
		*policy = lwsv1.StartupPolicyType(*src)
	}
	return nil
}

func restartPolicyToLWS(src *string, target any, _ any) *workloadsApi.Error {
	policy, ok := target.(*lwsv1.RestartPolicyType)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("restartPolicyToLWS: target type %T is not *v1.RestartPolicyType", target))
	}
	if src != nil {
		*policy = lwsv1.RestartPolicyType(*src)
	}
	return nil
}

func rolloutStrategyToLWS(src *string, target any, _ any) *workloadsApi.Error {
	strategy, ok := target.(*lwsv1.RolloutStrategyType)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("rolloutStrategyToLWS: target type %T is not *v1.RolloutStrategyType", target))
	}
	if src != nil {
		*strategy = lwsv1.RolloutStrategyType(*src)
	}
	return nil
}
