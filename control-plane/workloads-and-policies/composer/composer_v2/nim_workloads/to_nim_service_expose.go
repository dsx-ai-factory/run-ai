package nim_workloads

import (
	"fmt"
	"net/http"
	"strings"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	networkingv1 "k8s.io/api/networking/v1"
	"knative.dev/pkg/apis"
)

func (nsc *NimServiceComposer) composeService(src *workloadsManagerApi.NimServiceServingPort,
	target *nimv1alpha1.Service, setter FnNimServingPortSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !convutils.HasValue(src.Port) {
			return workloadsApi.NewError(http.StatusBadRequest, "servingPort: port must be provided")
		}

		return setter(src, target, cookie)
	}
}

type IngressInfo struct {
	ExposedExternally       bool
	WorkloadName            string
	ClusterDomain           string
	SubdomainSupportEnabled bool

	Url             *apis.URL
	PortName        string
	IngressPathType networkingv1.PathType
	IngressPath     string
	IngressClass    string
	IsOpenshift     bool
}

func resolveIngressController(ingressClass string, isOpenshift bool) string {
	if ingressClass == "" || isOpenshift {
		return defaultIngressClassName
	}

	ingressClass = strings.ToLower(ingressClass)
	switch {
	case strings.Contains(ingressClass, ingressControllerNginx):
		return ingressControllerNginx
	case strings.Contains(ingressClass, ingressControllerHAProxy):
		return ingressControllerHAProxy
	default:
		return defaultIngressClassName
	}
}

func (nsc *NimServiceComposer) composeIngress(src *workloadsManagerApi.NimServiceServingPort,
	target *nimv1alpha1.Ingress, setter FnNimIngressSetter, cookie any, ingressInfo *IngressInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return setter(nil, target, cookie)
		}

		if convutils.HasValue(src.ExposedUrl) {
			if validationErr := composer.ValidateUrlSyntax(src.ExposedUrl.Get(), false); validationErr != nil {
				return validationErr
			}
		}

		if nsc.useNewFlowIngress {
			*ingressInfo = IngressInfo{WorkloadName: nsc.fc.WorkloadName, ExposedExternally: false}
			return setter(ingressInfo, target, ingressInfo)
		}

		*ingressInfo = IngressInfo{WorkloadName: nsc.fc.WorkloadName}
		ingressInfo.ExposedExternally = convutils.Deref(src.ExposeExternally.Get(), true)
		nsc.getClusterNetworkingInfo(ingressInfo)

		var urlErr error
		if convutils.HasValue(src.ExposedUrl) {
			ingressInfo.Url, urlErr = apis.ParseURL(*src.ExposedUrl.Get())
		} else {
			ingressInfo.Url, urlErr = nsc.composeExposedUrl(src.ExposedUrl.Get(), nsc.fc.ProjectName, nsc.fc.WorkloadName,
				ingressInfo.ClusterDomain, ingressInfo.SubdomainSupportEnabled)
		}
		if urlErr != nil {
			return workloadsApi.NewError(http.StatusBadRequest, urlErr.Error())
		}

		if convutils.HasValue(src.GrpcPort) && convutils.Deref(src.ExposedProtocol.Get(), "") != workloadsManagerApi.SERVINGPORTEXPOSEDPROTOCOL_HTTP {
			ingressInfo.PortName = portNameGrpc
		} else {
			ingressInfo.PortName = portNameHttp
		}

		ingressInfo.IngressPath = ingressInfo.Url.Path
		ingressInfo.IngressPathType = networkingv1.PathTypePrefix
		if ingressInfo.PortName != portNameGrpc && ingressInfo.Url.Path != "" && ingressInfo.Url.Path != "/" {
			if resolveIngressController(ingressInfo.IngressClass, ingressInfo.IsOpenshift) == ingressControllerHAProxy {
				ingressInfo.IngressPath = strings.TrimSuffix(ingressInfo.Url.Path, "/")
			} else {
				ingressInfo.IngressPath = ingressInfo.Url.Path + "(/|$)(.*)"
				ingressInfo.IngressPathType = networkingv1.PathTypeImplementationSpecific
			}
		}

		return setter(ingressInfo, target, ingressInfo)
	}
}

func nimIngressSetter(src *IngressInfo, target any, _ any) *workloadsApi.Error {
	targetPtr, ok := target.(*nimv1alpha1.Ingress)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPortToIngressSetter: target type %T is not a *nim1alpha1.Ingress", target))
	}

	if src == nil {
		return nil
	}

	(*targetPtr).Enabled = convutils.Ref(src.ExposedExternally)

	if !src.ExposedExternally || src.Url == nil {
		return nil
	}

	(*targetPtr).Spec = networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				Host: src.Url.URL().Hostname(),
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     src.IngressPath,
								PathType: convutils.Ref(src.IngressPathType),
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: src.WorkloadName,
										Port: networkingv1.ServiceBackendPort{
											Name: src.PortName,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if src.IngressClass != "" && !src.IsOpenshift {
		(*targetPtr).Spec.IngressClassName = &src.IngressClass
	}
	if src.IsOpenshift {
		(*targetPtr).Spec.TLS = []networkingv1.IngressTLS{{}}
	}

	ingressAnnotations := make(map[string]string)
	ingressController := resolveIngressController(src.IngressClass, src.IsOpenshift)

	if src.PortName != portNameGrpc && src.Url.Path != "" && src.Url.Path != "/" {
		switch ingressController {
		case ingressControllerNginx:
			ingressAnnotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
		case ingressControllerHAProxy:
			pathPrefix := strings.TrimSuffix(src.Url.Path, "/")
			ingressAnnotations["haproxy.org/path-rewrite"] = fmt.Sprintf("%s(/|$)(.*) /\\2", pathPrefix)
		}
	}
	if src.PortName == portNameGrpc {
		switch ingressController {
		case ingressControllerNginx:
			ingressAnnotations["nginx.ingress.kubernetes.io/backend-protocol"] = "GRPC"
		case ingressControllerHAProxy:
			ingressAnnotations["haproxy.ingress.kubernetes.io/backend-protocol"] = "grpc"
		}
	}
	if len(ingressAnnotations) != 0 {
		(*targetPtr).Annotations = ingressAnnotations
	}

	return nil
}

func (nsc *NimServiceComposer) composeExposedUrl(exposedUrl *string, projectName string, workloadName string, clusterDomain string, subdomainSupportEnabled bool) (*apis.URL, error) {
	if exposedUrl != nil {
		return apis.ParseURL(*exposedUrl)
	}
	url, err := apis.ParseURL(clusterDomain)
	if err != nil || url == nil {
		return nil, fmt.Errorf("invalid cluster domain '%s': %v", clusterDomain, err)
	}

	if subdomainSupportEnabled {
		subdomain := fmt.Sprintf("%s-%s", projectName, workloadName)
		url.Host = fmt.Sprintf("%s.%s", subdomain, url.Host)
	} else {
		path := fmt.Sprintf("/%s/%s", projectName, workloadName)
		if url.Path != "" {
			path = fmt.Sprintf("%s/%s", path, url.Path)
		}
		url.Path = path
	}

	if url.Path == "" {
		url.Path = "/"
	}

	return url, nil
}

func (nsc *NimServiceComposer) getClusterNetworkingInfo(ingressInfo *IngressInfo) {
	cluster := nsc.clusterInfo
	ingressInfo.ClusterDomain = convutils.Deref(cluster.Domain.Get(), "")
	if cluster.Status.IsSet() && cluster.Status.Get().Config.IsSet() {
		ingressInfo.SubdomainSupportEnabled = convutils.Deref(cluster.Status.Get().Config.Get().SubdomainEnabled.Get(), false)
		ingressInfo.IngressClass = convutils.Deref(cluster.Status.Get().Config.Get().IngressClass.Get(), "")
	}
	if ingressInfo.IngressClass == "" {
		ingressInfo.IngressClass = defaultIngressClassName
	}
	if cluster.Status.IsSet() && cluster.Status.Get().Platform.IsSet() {
		platformType := convutils.Deref(cluster.Status.Get().Platform.Get().Type.Get(), "")
		ingressInfo.IsOpenshift = platformType == "openshift"
	}
}

// GetAutoExposureUrl returns the auto-generated URL when exposure is wanted and no exposedUrl; nil otherwise.
func (nsc *NimServiceComposer) GetAutoExposureUrl(src *workloadsManagerApi.NimServiceServingPort) (*string, error) {
	if src == nil || convutils.HasValue(src.ExposedUrl) {
		return nil, nil
	}
	var ingressInfo IngressInfo
	nsc.getClusterNetworkingInfo(&ingressInfo)
	if ingressInfo.ClusterDomain == "" {
		return nil, nil
	}
	u, err := nsc.composeExposedUrl(src.ExposedUrl.Get(), nsc.fc.ProjectName, nsc.fc.WorkloadName,
		ingressInfo.ClusterDomain, ingressInfo.SubdomainSupportEnabled)
	if err != nil || u == nil {
		return nil, err
	}
	s := u.URL().String()
	return &s, nil
}
