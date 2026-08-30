package networking

import (
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	OAuth2ProxyContainerName = "oauth2-proxy"
	AuthProxyServicePort     = 80
)

func (p *Plugin) buildAuthProxyDeployment(
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
	upstreamURL string,
	externalHost string,
	externalPath string,
) (*appsv1.Deployment, error) {
	name := getAuthProxyName(runner)
	container, err := p.buildOAuth2ProxyContainer(config, upstreamURL, name, externalHost, externalPath)
	if err != nil {
		return nil, err
	}
	initContainers, err := p.buildAuthProxyInitContainers(config, name)
	if err != nil {
		return nil, err
	}
	return composeAuthProxyDeployment(
		runner, name, container,
		initContainers,
		p.buildAuthProxyVolumes(),
	), nil
}

func composeAuthProxyDeployment(
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	container *corev1.Container,
	initContainers []corev1.Container,
	volumes []corev1.Volume,
) *appsv1.Deployment {
	labels := map[string]string{
		LabelApp:            name,
		LabelWorkloadRunner: runner.Name,
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.AppsV1,
			Kind:       constants.KindDeployment,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: runner.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers:     []corev1.Container{*container},
					Volumes:        volumes,
				},
			},
		},
	}
}

func (p *Plugin) buildAuthProxyVolumes() []corev1.Volume {
	return workloadbuilders.ComposeAuthProxyVolumes()
}

func (p *Plugin) buildAuthProxyInitContainers(config *dromosv1alpha1.NetworkingConfig, nick string) ([]corev1.Container, error) {
	if !hasService(config) {
		return nil, nil
	}

	auth := config.Services[0].Auth
	if auth == nil || auth.AuthorizedUsers == "" {
		return nil, nil
	}

	initContainer, err := workloadbuilders.BuildUsernameLoaderInitContainerWithError(auth.AuthorizedUsers, nick)
	if err != nil {
		return nil, err
	}
	return []corev1.Container{*initContainer}, nil
}

func (p *Plugin) buildOAuth2ProxyContainer(
	config *dromosv1alpha1.NetworkingConfig,
	upstreamURL string,
	cookiePrefix string,
	externalHost string,
	externalPath string,
) (*corev1.Container, error) {
	var authorizedUsers, authorizedGroups *string
	forceJsonResponse := false
	if hasService(config) && config.Services[0].Auth != nil {
		auth := config.Services[0].Auth
		if auth.AuthorizedUsers != "" {
			authorizedUsers = &auth.AuthorizedUsers
		}
		if auth.AuthorizedGroups != "" {
			authorizedGroups = &auth.AuthorizedGroups
		}
		if auth.ForceJsonResponse != nil {
			forceJsonResponse = *auth.ForceJsonResponse
		}
	}
	return buildOAuth2ProxyContainerFromAuth(
		upstreamURL, cookiePrefix, externalHost, externalPath,
		authorizedUsers, authorizedGroups, forceJsonResponse,
	)
}

func buildOAuth2ProxyContainerFromAuth(
	upstreamURL, cookiePrefix, externalHost, externalPath string,
	authorizedUsers, authorizedGroups *string,
	forceJsonResponse bool,
) (*corev1.Container, error) {
	cfg := workloadbuilders.OAuth2ProxyConfig{
		ContainerName:              OAuth2ProxyContainerName,
		UpstreamURL:                upstreamURL,
		ExternalHost:               externalHost,
		ExternalPath:               externalPath,
		ProxyPrefix:                "/oauth2",
		Port:                       constants.ExternalUrlAuthContainerDefaultPort,
		CookiePrefix:               cookiePrefix,
		AuthorizedUsers:            authorizedUsers,
		AuthorizedGroups:           authorizedGroups,
		Nick:                       cookiePrefix,
		OmitPortDeclaration:        false, // Always declare ports for standalone deployment
		ForceJsonResponse:          forceJsonResponse,
		EnableSidecarRestartPolicy: false,
		ReverseProxy:               true,
	}
	return workloadbuilders.BuildOAuth2ProxyContainer(cfg)
}

func (p *Plugin) buildAuthProxyService(
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
) *corev1.Service {
	labels := map[string]string{
		LabelApp:            name,
		LabelWorkloadRunner: runner.Name,
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CoreV1,
			Kind:       constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: runner.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       AuthProxyServicePort,
					TargetPort: intstr.FromInt32(constants.ExternalUrlAuthContainerDefaultPort),
				},
			},
		},
	}
}

func (p *Plugin) buildWorkloadService(
	runner *dromosv1alpha1.WorkloadRunner,
	serviceCreate *corev1.Service,
) *corev1.Service {
	svc := serviceCreate.DeepCopy()
	svc.TypeMeta = metav1.TypeMeta{
		APIVersion: constants.CoreV1,
		Kind:       constants.KindService,
	}
	svc.Namespace = runner.Namespace
	svc.Name = getResourceName(svc.Name, runner)
	return svc
}

func (p *Plugin) buildIngress(
	runner *dromosv1alpha1.WorkloadRunner,
	ingressSpec *networkingv1.Ingress,
	backendServiceName string,
	backendServicePort int32,
) *networkingv1.Ingress {
	ingress := ingressSpec.DeepCopy()
	ingress.TypeMeta = metav1.TypeMeta{
		APIVersion: constants.NetworkingV1,
		Kind:       constants.KindIngress,
	}
	ingress.Namespace = runner.Namespace
	ingress.Name = getResourceName(ingress.Name, runner)

	if backendServiceName != "" {
		for i := range ingress.Spec.Rules {
			if ingress.Spec.Rules[i].HTTP != nil {
				for j := range ingress.Spec.Rules[i].HTTP.Paths {
					ingress.Spec.Rules[i].HTTP.Paths[j].Backend = networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: backendServiceName,
							Port: networkingv1.ServiceBackendPort{
								Number: backendServicePort,
							},
						},
					}
				}
			}
		}
	}

	return workloadbuilders.EnrichIngressWithDefaults(ingress)
}

func (p *Plugin) buildRoute(
	runner *dromosv1alpha1.WorkloadRunner,
	routeSpec *openshiftv1.Route,
	backendServiceName string,
) *openshiftv1.Route {
	route := routeSpec.DeepCopy()
	route.TypeMeta = metav1.TypeMeta{
		APIVersion: constants.RouteV1,
		Kind:       constants.KindRoute,
	}
	route.Namespace = runner.Namespace
	route.Name = getResourceName(route.Name, runner)

	if backendServiceName != "" {
		route.Spec.To = openshiftv1.RouteTargetReference{
			Kind: "Service",
			Name: backendServiceName,
		}
	}

	route = workloadbuilders.EnrichRouteWithDefaults(route)

	return route
}

func (p *Plugin) buildHTTPRoute(
	runner *dromosv1alpha1.WorkloadRunner,
	httpRouteSpec *gatewayv1.HTTPRoute,
	backendServiceName string,
	backendServicePort int32,
) *gatewayv1.HTTPRoute {
	httpRoute := httpRouteSpec.DeepCopy()
	httpRoute.TypeMeta = metav1.TypeMeta{
		APIVersion: constants.GatewayV1,
		Kind:       constants.KindHTTPRoute,
	}
	httpRoute.Namespace = runner.Namespace
	httpRoute.Name = getResourceName(httpRoute.Name, runner)

	if backendServiceName != "" {
		port := gatewayv1.PortNumber(backendServicePort)
		for i := range httpRoute.Spec.Rules {
			for j := range httpRoute.Spec.Rules[i].BackendRefs {
				httpRoute.Spec.Rules[i].BackendRefs[j].BackendRef = gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: gatewayv1.ObjectName(backendServiceName),
						Port: &port,
					},
				}
			}
		}
	}

	return httpRoute
}
