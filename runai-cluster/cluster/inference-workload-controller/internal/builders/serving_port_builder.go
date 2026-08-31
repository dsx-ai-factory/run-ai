/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"strings"
	"time"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	HttpPort        = 80
	servingPortName = "serving-port" // must match the NIMServiceMonitor that expects this name for metrics scraping
)

type ServingPortBuilder struct {
	k8sClient   client.Client
	isOpenshift bool
}

func NewServingPortBuilder(k8sClient client.Client, isOpenshift bool) *ServingPortBuilder {
	return &ServingPortBuilder{
		k8sClient:   k8sClient,
		isOpenshift: isOpenshift,
	}
}

func (b *ServingPortBuilder) GetPreview(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	servingPort, err := b.GetServingPort(workload.Spec.ServingPorts)
	if err != nil {
		return nil, err
	}

	if servingPort == nil {
		return []client.Object{}, nil
	}

	objects := []client.Object{}

	service, err := b.previewService(workload)
	if err != nil {
		return nil, err
	}
	if service != nil {
		objects = append(objects, service)
	}

	routingObjects, err := b.previewIngressOrRoute(ctx, workload, servingPort)
	if err != nil {
		return nil, err
	}
	objects = append(objects, routingObjects...)

	return objects, nil
}

func (b *ServingPortBuilder) Build(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	servingPort, err := b.GetServingPort(workload.Spec.ServingPorts)
	if err != nil {
		return nil, err
	}

	if servingPort == nil {
		return []client.Object{}, nil
	}

	objects := []client.Object{}

	service, err := b.buildService(workload, servingPort)
	if err != nil {
		return nil, err
	}
	if service != nil {
		objects = append(objects, service)
	}

	routingObjects, err := b.buildIngressOrRoute(ctx, workload, servingPort)
	if err != nil {
		return nil, err
	}
	objects = append(objects, routingObjects...)

	return objects, nil
}

func (b *ServingPortBuilder) GetServingPort(servingPorts *flds.ServingPortsField) (*flds.ServingPortItemValue, error) {
	if servingPorts == nil {
		return nil, nil
	}

	// for now, only a single serving port is supported
	if len(servingPorts.Deref()) != 1 {
		return nil, fmt.Errorf("only a single serving port is supported")
	}

	// get the serving port
	var servingPort flds.ServingPortItemValue
	for _, port := range servingPorts.Deref() {
		if port.Value == nil {
			return nil, fmt.Errorf("serving port value is not set")
		}

		servingPort = *port.Value
		break
	}

	if servingPort.Port == nil {
		return nil, fmt.Errorf("serving port: mandatory field 'port' is not set")
	}

	return &servingPort, nil
}

func (b *ServingPortBuilder) GetPorts(servingPorts *flds.ServingPortsField) ([]corev1.ContainerPort, error) {
	servingPort, err := b.GetServingPort(servingPorts)
	if err != nil {
		return nil, err
	}

	if servingPort == nil {
		return []corev1.ContainerPort{}, nil
	}

	containerPort := corev1.ContainerPort{
		Name:          getPortName(*servingPort.Port),
		ContainerPort: int32(*servingPort.Port),
		Protocol:      corev1.ProtocolTCP,
	}

	return []corev1.ContainerPort{containerPort}, nil
}

func (b *ServingPortBuilder) previewService(workload *dinfw.DistributedInferenceWorkload) (*corev1.Service, error) {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CoreV1,
			Kind:       constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServingPortNetworkResourceName(workload.Name),
			Namespace: workload.Namespace,
		},
	}
	return service, nil
}

func (b *ServingPortBuilder) buildService(workload *dinfw.DistributedInferenceWorkload,
	servingPort *flds.ServingPortItemValue) (*corev1.Service, error) {
	// When we build the service port, we must support NIM scraping.
	// Therefore, we create the service with:
	// 1) the serving port name ("serving-port")
	// 2) the NIM labels (constants.LabelNIMWorkload), if they exist in the workload labels (see CopyNimLabel).

	// determine the target port - use auth port if authorization is required
	authType := getAuthorizationType(servingPort)
	var targetPort int32
	if authType != constants.ServingPortAuthorizationTypePublic {
		targetPort = constants.ExternalUrlAuthContainerDefaultPort
	} else {
		targetPort = int32(*servingPort.Port)
	}

	// create service port
	servicePort := corev1.ServicePort{
		Name:       servingPortName,
		Port:       HttpPort,
		TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: targetPort},
		Protocol:   corev1.ProtocolTCP,
	}

	serviceLabels := workload.Labels
	CopyNimLabel(workload.LeaderSpec.Labels, serviceLabels)

	// create the service
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CoreV1,
			Kind:       constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              GetServingPortNetworkResourceName(workload.Name),
			Namespace:         workload.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            serviceLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{servicePort},
			// the selector should select only the leader pods, but from all replicas
			Selector: map[string]string{
				lwsv1.SetNameLabelKey:     GetUniqueLwsName(workload),
				lwsv1.WorkerIndexLabelKey: "0", // the leader
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// set controller reference
	err := controllerutils.SetControllerReference(workload, service, b.k8sClient.Scheme())
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (b *ServingPortBuilder) previewIngressOrRoute(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	servingPort *flds.ServingPortItemValue) ([]client.Object, error) {
	if !exposeExternally(servingPort) {
		return nil, nil
	}

	serviceName := GetServingPortNetworkResourceName(workload.Name)
	var objects []client.Object

	if workloadconfig.IngressEnabled() {
		if b.isOpenshift {
			objects = append(objects, b.previewRoute(ctx, workload, serviceName))
		} else {
			objects = append(objects, b.previewIngress(ctx, workload, serviceName))
		}
	}

	if workloadconfig.GatewayEnabled() {
		objects = append(objects, b.previewHTTPRoute(workload, serviceName))
	}

	return objects, nil
}

func (b *ServingPortBuilder) buildIngressOrRoute(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	servingPort *flds.ServingPortItemValue) ([]client.Object, error) {
	if !exposeExternally(servingPort) {
		return nil, nil
	}

	serviceName := GetServingPortNetworkResourceName(workload.Name)

	url, err := getExposedUrl(ctx, workload.Name, servingPort)
	if err != nil {
		return nil, err
	}

	var objects []client.Object

	if workloadconfig.IngressEnabled() {
		var resource client.Object
		if b.isOpenshift {
			resource, err = b.buildRoute(ctx, workload, serviceName, url)
		} else {
			resource, err = b.buildIngress(ctx, workload, serviceName, url, servingPort)
		}
		if err != nil {
			return nil, err
		}
		if err = controllerutils.SetControllerReference(workload, resource, b.k8sClient.Scheme()); err != nil {
			return nil, err
		}
		objects = append(objects, resource)
	}

	if workloadconfig.GatewayEnabled() {
		httpRoute, err := b.buildHTTPRoute(workload, serviceName, url)
		if err != nil {
			return nil, err
		}
		if err = controllerutils.SetControllerReference(workload, httpRoute, b.k8sClient.Scheme()); err != nil {
			return nil, err
		}
		objects = append(objects, httpRoute)
	}

	return objects, nil
}

func (b *ServingPortBuilder) previewIngress(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	serviceName string) client.Object {
	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.NetworkingV1,
			Kind:       constants.KindIngress,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: workload.Namespace,
		},
	}
}

func (b *ServingPortBuilder) previewRoute(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	serviceName string) client.Object {
	return &openshiftv1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.RouteV1,
			Kind:       constants.KindRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: workload.Namespace,
		},
	}
}

func (b *ServingPortBuilder) buildIngress(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	serviceName string, url *apis.URL, servingPort *flds.ServingPortItemValue) (*networkingv1.Ingress, error) {
	// Prepare the path with capture group for path rewriting (inference-specific)
	ingressPath, pathType := AddPathRewriteIngressRegex(url.Path, true, workloadconfig.IngressClass())

	// Add path rewrite annotations (inference-specific)
	pathRewriteAnnotations := AddPathRewriteIngressAnnotations(nil, PathRewriteConfig{
		OriginalPath: url.Path,
		IngressClass: workloadconfig.IngressClass(),
	})

	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.NetworkingV1,
			Kind:       constants.KindIngress,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              serviceName,
			Namespace:         workload.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            workload.Labels,
			Annotations:       pathRewriteAnnotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: url.URL().Hostname(),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     ingressPath,
									PathType: ptr.To(pathType),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: HttpPort,
											},
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

	// Apply cluster-wide defaults (ingressClassName, body size, default annotations)
	ingress = workloadbuilders.EnrichIngressWithDefaults(ingress)

	return ingress, nil
}

func (b *ServingPortBuilder) buildRoute(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	serviceName string, url *apis.URL) (*openshiftv1.Route, error) {
	// OpenShift rejects passthrough Routes with a non-empty spec.path, and rewrite
	// annotations don't apply to encrypted traffic — drop both in passthrough mode.
	termination := workloadbuilders.DefaultRouteTermination()
	pathForRoute := url.Path
	pathForAnnotations := url.Path
	if termination == openshiftv1.TLSTerminationPassthrough {
		pathForRoute = ""
		pathForAnnotations = ""
	}

	routeAnnotations := AddPathRewriteRouteAnnotations(nil, pathForAnnotations, true)

	route := &openshiftv1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.RouteV1,
			Kind:       constants.KindRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              serviceName,
			Namespace:         workload.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            workload.Labels,
			Annotations:       routeAnnotations,
		},
		Spec: openshiftv1.RouteSpec{
			Host: url.URL().Hostname(),
			Path: pathForRoute,
			To: openshiftv1.RouteTargetReference{
				Kind: constants.KindService,
				Name: serviceName,
			},
			TLS: &openshiftv1.TLSConfig{
				InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   termination,
			},
		},
	}

	route = workloadbuilders.EnrichRouteWithDefaults(route)

	return route, nil
}

func (b *ServingPortBuilder) previewHTTPRoute(workload *dinfw.DistributedInferenceWorkload,
	serviceName string) client.Object {
	return &gatewayv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.GatewayV1,
			Kind:       constants.KindHTTPRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: workload.Namespace,
		},
	}
}

func (b *ServingPortBuilder) buildHTTPRoute(workload *dinfw.DistributedInferenceWorkload,
	serviceName string, url *apis.URL) (*gatewayv1.HTTPRoute, error) {
	hostname := url.URL().Hostname()
	if workloadconfig.SubdomainSupportEnabled() && hostname != "" {
		hostname = strings.Replace(hostname, workloadconfig.RunaiTrimStr, "", 1)
	}

	// Gateway name and namespace are guaranteed non-empty by the go-operator's
	// deployment builder and CRD validation when GATEWAY_API_ENABLED=true.
	gwName := gatewayv1.ObjectName(workloadconfig.GatewayName())
	gwNamespace := gatewayv1.Namespace(workloadconfig.GatewayNamespace())
	gatewayHostname := gatewayv1.Hostname(hostname)
	pathPrefix := url.Path
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	port := gatewayv1.PortNumber(HttpPort)

	var filters []gatewayv1.HTTPRouteFilter
	if pathPrefix != "/" {
		filters = append(filters, gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterURLRewrite,
			URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
				Path: &gatewayv1.HTTPPathModifier{
					Type:               gatewayv1.PrefixMatchHTTPPathModifier,
					ReplacePrefixMatch: ptr.To("/"),
				},
			},
		})
	}

	httpRoute := &gatewayv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.GatewayV1,
			Kind:       constants.KindHTTPRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              serviceName,
			Namespace:         workload.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            workload.Labels,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{
					{
						Name:      gwName,
						Namespace: &gwNamespace,
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{gatewayHostname},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
								Value: &pathPrefix,
							},
						},
					},
					Filters: filters,
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(serviceName),
									Port: &port,
								},
							},
						},
					},
				},
			},
		},
	}

	return httpRoute, nil
}

func getPortName(port int) string {
	return fmt.Sprintf("port-%d", port)
}

func GetServingPortNetworkResourceName(workloadName string) string {
	// using "sp" to have a unique name for the serving port network resource
	// using "0" as we currently only support a single serving port per workload
	return fmt.Sprintf("dinfw-%s-sp-0", workloadName)
}

func getExposedUrl(ctx context.Context, workloadName string, servingPort *flds.ServingPortItemValue) (*apis.URL, error) {

	var url *apis.URL
	var err error

	if servingPort.ExposedUrl != nil {
		url, err = apis.ParseURL(*servingPort.ExposedUrl)
	} else {
		url, err = workloadbuilders.ComposeExposedUrlInt(ctx, "", workloadName)
	}
	if err != nil {
		return nil, err
	}

	if url.Path == "" {
		url.Path = "/"
	}

	return url, nil
}

func getAuthorizationType(servingPort *flds.ServingPortItemValue) string {
	if servingPort != nil && servingPort.AuthorizationType != nil {
		return *servingPort.AuthorizationType
	}
	return constants.ServingPortAuthorizationTypeDefault
}

func exposeExternally(servingPort *flds.ServingPortItemValue) bool {
	if servingPort != nil && servingPort.ExposeExternally != nil {
		return *servingPort.ExposeExternally
	}
	return true // default value
}

func (b *ServingPortBuilder) IsAuthRequired(servingPorts *flds.ServingPortsField) bool {
	servingPort, err := b.GetServingPort(servingPorts)
	if err != nil {
		return false
	}
	return getAuthorizationType(servingPort) != constants.ServingPortAuthorizationTypePublic
}

func (b *ServingPortBuilder) ValidateAndGetServingPortAuthFields(servingPort *flds.ServingPortItemValue) (bool, *string, *string, error) {
	if servingPort == nil {
		return false, nil, nil, nil
	}

	authType := getAuthorizationType(servingPort)

	// if it's public or authenticatedUsers, no additional fields should be provided
	if authType == constants.ServingPortAuthorizationTypePublic || authType == constants.ServingPortAuthorizationTypeAuthenticatedUsers {
		// make sure no auth fields are set when public
		if servingPort.AuthorizedUsers != nil && *servingPort.AuthorizedUsers != "" {
			return false, nil, nil, fmt.Errorf("authorizedUsers is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedUsers)
		}
		if servingPort.AuthorizedGroups != nil && *servingPort.AuthorizedGroups != "" {
			return false, nil, nil, fmt.Errorf("authorizedGroups is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedGroups)
		}

		authRequired := authType != constants.ServingPortAuthorizationTypePublic
		return authRequired, nil, nil, nil
	}

	// for authorizedUsers, only users should be provided
	if authType == constants.ServingPortAuthorizationTypeAuthorizedUsers {
		if servingPort.AuthorizedUsers == nil || *servingPort.AuthorizedUsers == "" {
			return false, nil, nil, fmt.Errorf("authorizedUsers is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedUsers)
		}
		if servingPort.AuthorizedGroups != nil && *servingPort.AuthorizedGroups != "" {
			return false, nil, nil, fmt.Errorf("authorizedGroups is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedGroups)
		}
		return true, servingPort.AuthorizedUsers, nil, nil
	}

	// for authorizedGroups, only groups should be provided
	if authType == constants.ServingPortAuthorizationTypeAuthorizedGroups {
		if servingPort.AuthorizedGroups == nil || *servingPort.AuthorizedGroups == "" {
			return false, nil, nil, fmt.Errorf("authorizedGroups is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedGroups)
		}
		if servingPort.AuthorizedUsers != nil && *servingPort.AuthorizedUsers != "" {
			return false, nil, nil, fmt.Errorf("authorizedUsers is only allowed when authorizationType is set to %s",
				constants.ServingPortAuthorizationTypeAuthorizedUsers)
		}
		return true, nil, servingPort.AuthorizedGroups, nil
	}

	return false, nil, nil, fmt.Errorf("invalid authorizationType: %s. Valid values are: %s, %s, %s, %s",
		authType,
		constants.ServingPortAuthorizationTypePublic,
		constants.ServingPortAuthorizationTypeAuthenticatedUsers,
		constants.ServingPortAuthorizationTypeAuthorizedUsers,
		constants.ServingPortAuthorizationTypeAuthorizedGroups,
	)
}

func (b *ServingPortBuilder) GetAuthContainers(ctx context.Context, workloadName string,
	servingPorts *flds.ServingPortsField) ([]corev1.Container, []corev1.Container, error) {
	servingPort, err := b.GetServingPort(servingPorts)
	if err != nil {
		return nil, nil, err
	}

	// validate and get auth fields
	authRequired, authorizedUsers, authorizedGroups, err := b.ValidateAndGetServingPortAuthFields(servingPort)
	if err != nil {
		return nil, nil, err
	}

	// if no auth required, return empty list
	if !authRequired {
		return []corev1.Container{}, []corev1.Container{}, nil
	}

	containers := []corev1.Container{}
	initContainers := []corev1.Container{}

	// if authorizedUsers type is used, add username loader init container
	if authorizedUsers != nil {
		usernameLoaderContainer, err := workloadbuilders.GetUsernameLoaderInitContainerWithError(nil, authorizedUsers)
		if err != nil {
			return nil, nil, err
		}
		if usernameLoaderContainer != nil {
			initContainers = append(initContainers, *usernameLoaderContainer)
		}
	}

	serviceName := GetServingPortNetworkResourceName(workloadName)

	// prepare URL for OAuth2 proxy container
	var jobUrl *apis.URL
	if !exposeExternally(servingPort) {
		// internal URL - use cluster URL
		clusterUrlStr, err := getServingPortClusterUrl(ctx, serviceName)
		if err != nil {
			return nil, nil, err
		}
		jobUrl, err = apis.ParseURL(clusterUrlStr)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// external URL - use exposed URL
		var err error
		jobUrl, err = getExposedUrl(ctx, workloadName, servingPort)
		if err != nil {
			return nil, nil, err
		}
	}

	// create OAuth2 proxy container
	cookiePrefix := serviceName
	oauth2Container, err := workloadbuilders.GetOauth2ProxyContainer(
		jobUrl,
		int32(*servingPort.Port),
		authorizedUsers,
		authorizedGroups,
		0,              // idx
		"serving-port", // nick
		true,           // isInference
		true,           // isInferenceServingPort
		cookiePrefix,
	)
	if err != nil {
		return nil, nil, err
	}

	containers = append(containers, *oauth2Container)

	return containers, initContainers, nil
}
