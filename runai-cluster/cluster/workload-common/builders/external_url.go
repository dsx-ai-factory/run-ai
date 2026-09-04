/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	JupyterServiceSuffix     = "jupyter"
	TensorboardServiceSuffix = "tensorboard"
)

type ExternalUrlBuilder struct {
	k8sClient      client.Client
	isOpenshift    bool
	authPortOffset int
	services       []corev1.Service
	ingresses      []networkingv1.Ingress
	routes         []openshiftv1.Route
	httpRoutes     []gatewayv1.HTTPRoute
}

func (eb *ExternalUrlBuilder) WithAuthPortOffset(offset int) *ExternalUrlBuilder {
	eb.authPortOffset = offset
	return eb
}

func NewExternalUrlBuilder(k8sClient client.Client, isOpenshift bool) *ExternalUrlBuilder {
	return &ExternalUrlBuilder{
		k8sClient:   k8sClient,
		isOpenshift: isOpenshift,
	}
}

// GetPreview returns a list of resources (name/kind only) that will be created on build
func (eb *ExternalUrlBuilder) GetPreview(_ context.Context, workloadInitials string, objectMeta *metav1.ObjectMeta, exposedUrls *flds.ExposedUrlsField,
	isJupyter bool, isTensorboard bool) ([]Resource, error) {
	var resources []Resource

	resources = append(resources, eb.getPreviewServicesForExposedUrls(exposedUrls, workloadInitials, objectMeta.Name, objectMeta.Namespace)...)
	resources = append(resources, eb.getPreviewServicesForJupyterOrTensorflow(isJupyter, isTensorboard, workloadInitials, objectMeta.Name, objectMeta.Namespace)...)
	resources = append(resources, eb.getPreviewIngressesOrRoutesForServices(resources)...)

	return resources, nil
}

func (eb *ExternalUrlBuilder) getPreviewServicesForExposedUrls(exposedUrls *flds.ExposedUrlsField, workloadInitials string, workloadName string, namespace string) []Resource {
	var result []Resource

	for i, nick := range commonutils.SortedKeys(exposedUrls.Deref()) {
		exposedUrl := exposedUrls.Deref()[nick]
		if exposedUrl.Value == nil {
			continue
		}

		// validate consistency between the customUrl flag and the url value
		customUrlEnabled := ptr.Deref(exposedUrl.Value.CustomUrl, false)
		customUrlProvided := ptr.Deref(exposedUrl.Value.Url, "") != ""
		if (customUrlEnabled && !customUrlProvided) || (!customUrlEnabled && customUrlProvided) {
			continue
		}

		result = append(result, Resource{
			Kind:      serviceGroupVersion,
			Name:      getServiceName(workloadInitials, workloadName, fmt.Sprintf("%d", i)),
			Namespace: namespace,
		})

	}
	return result
}

func (eb *ExternalUrlBuilder) getPreviewServicesForJupyterOrTensorflow(isJupyter bool, isTensorboard bool, workloadInitials string, workloadName string, namespace string) []Resource {
	// Handle only if Jupyter or Tensorboard
	if !isJupyter && !isTensorboard {
		return nil
	}
	serviceNameSuffix := JupyterServiceSuffix
	if isTensorboard {
		serviceNameSuffix = TensorboardServiceSuffix
	}
	return []Resource{{
		Kind:      serviceGroupVersion,
		Name:      getServiceName(workloadInitials, workloadName, serviceNameSuffix),
		Namespace: namespace,
	}}
}

// getPreviewIngressesOrRoutesForServices returns routing resource previews for the given services
// based on the active routing mode. When Ingress is enabled, it returns Ingress (or Route on OpenShift)
// resources. When Gateway API is enabled, it returns HTTPRoute resources. In dual-mode (both enabled),
// it returns both, allowing zero-downtime migration between routing mechanisms.
func (eb *ExternalUrlBuilder) getPreviewIngressesOrRoutesForServices(services []Resource) []Resource {
	var kinds []metav1.GroupVersionKind
	if config.IngressEnabled() {
		if eb.isOpenshift {
			kinds = append(kinds, RouteGroupVersion)
		} else {
			kinds = append(kinds, ingressGroupVersion)
		}
	}
	if config.GatewayEnabled() {
		kinds = append(kinds, HTTPRouteGroupVersion)
	}

	var result []Resource
	for _, kind := range kinds {
		for _, service := range services {
			result = append(result, Resource{
				Kind:      kind,
				Name:      service.Name,
				Namespace: service.Namespace,
			})
		}
	}
	return result
}

// Build creates the resources
func (eb *ExternalUrlBuilder) Build(ctx context.Context, workloadKind string, jobType string, workloadInitials string,
	objectMeta *metav1.ObjectMeta, workloadBaseName string, owner client.Object, exposedUrls *flds.ExposedUrlsField,
	ingressUrl *flds.StringField, calculatedPriorityClassNameLabel string, isJupyter bool,
	isTensorboard bool, isAuthDisabled bool) ([]corev1.Service, []networkingv1.Ingress, []openshiftv1.Route, []gatewayv1.HTTPRoute, error) {
	var objects []client.Object

	eb.services = make([]corev1.Service, 0)
	eb.ingresses = make([]networkingv1.Ingress, 0)
	eb.routes = make([]openshiftv1.Route, 0)
	eb.httpRoutes = make([]gatewayv1.HTTPRoute, 0)

	err := SetObjectMetaLabels(ctx, workloadKind, jobType, objectMeta.Name, objectMeta)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = SetPriorityClassLabel(ctx, eb.k8sClient, calculatedPriorityClassNameLabel, objectMeta)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = eb.buildExposedUrls(ctx, workloadInitials, objectMeta, workloadBaseName, exposedUrls, isJupyter, isTensorboard, isAuthDisabled)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// backward compatibility creation of jupyter and tensorboard
	err = eb.buildJupyterOrTensorboard(ctx, workloadInitials, workloadBaseName, objectMeta, ingressUrl, isJupyter, isTensorboard, isAuthDisabled)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// append services
	for i := range eb.services {
		if owner != nil {
			err = controllerutils.SetControllerReference(owner, &eb.services[i], eb.k8sClient.Scheme())
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}
		objects = append(objects, &eb.services[i])
	}

	// append ingresses or routes (legacy routing)
	if config.IngressEnabled() {
		if !eb.isOpenshift {
			for i := range eb.ingresses {
				if owner != nil {
					err = controllerutils.SetControllerReference(owner, &eb.ingresses[i], eb.k8sClient.Scheme())
					if err != nil {
						return nil, nil, nil, nil, err
					}
				}
				objects = append(objects, &eb.ingresses[i])
			}
		} else {
			for i := range eb.routes {
				if owner != nil {
					err = controllerutils.SetControllerReference(owner, &eb.routes[i], eb.k8sClient.Scheme())
					if err != nil {
						return nil, nil, nil, nil, err
					}
				}
				objects = append(objects, &eb.routes[i])
			}
		}
	}

	// append HTTPRoutes (Gateway API routing)
	if config.GatewayEnabled() {
		for i := range eb.httpRoutes {
			if owner != nil {
				err = controllerutils.SetControllerReference(owner, &eb.httpRoutes[i], eb.k8sClient.Scheme())
				if err != nil {
					return nil, nil, nil, nil, err
				}
			}
			objects = append(objects, &eb.httpRoutes[i])
		}
	}

	return eb.services, eb.ingresses, eb.routes, eb.httpRoutes, nil
}

func (eb *ExternalUrlBuilder) buildExposedUrls(ctx context.Context, workloadInitials string, objectMeta *metav1.ObjectMeta, workloadBaseName string,
	exposedUrls *flds.ExposedUrlsField, isJupyter bool, isTensorboard bool, isAuthDisabled bool) error {
	// typically we add a suffix to the url to make it unique. we want at least one url which
	// will not have this suffix, as it is more comfortable to use such url. note that jupyter/tensorboard
	// already generate such url so when one of them is enabled all exposed URLs will have suffix
	noSuffixUrlGenerated := false
	if isJupyter || isTensorboard {
		noSuffixUrlGenerated = true
	}

	clusterDomain := config.GetClusterDomain()

	projectName, err := project.GetProjectFromContext(ctx)
	if err != nil {
		return fmt.Errorf("project name not found in context")
	}

	options := ComposeUrlOptions{
		SubdomainSupported: config.SubdomainSupportEnabled(),
	}

	urlsPerNick, err := GetUrlPerNick(exposedUrls, projectName, workloadBaseName, clusterDomain, &noSuffixUrlGenerated, options)
	if err != nil {
		return err
	}

	exposedUrlsArr := exposedUrls.Deref()

	for idx, nick := range commonutils.SortedKeys(exposedUrls.Deref()) {
		exposedUrl := exposedUrlsArr[nick]
		if exposedUrl.Value == nil {
			continue
		}

		url := urlsPerNick[nick]
		if url == nil {
			return fmt.Errorf("failed to determine external url for %s", nick)
		}

		serviceName := getServiceName(workloadInitials, workloadBaseName, fmt.Sprintf("%d", idx))

		if exposedUrl.Value.ContainerPort == nil {
			return fmt.Errorf("container port is mandatory for exposed url %v", url.String())
		}
		service, err := createServiceExposedUrl(url, workloadBaseName,
			objectMeta.Namespace, *exposedUrl.Value.ContainerPort,
			objectMeta, serviceName, idx+eb.authPortOffset, true, isAuthDisabled)
		if err != nil {
			return err
		}
		eb.services = append(eb.services, service)

		err = eb.buildIngressOrRoute(objectMeta, *url, &service)
		if err != nil {
			return err
		}
	}

	return nil
}

func (eb *ExternalUrlBuilder) buildJupyterOrTensorboard(ctx context.Context, workloadInitials string, workloadBaseName string, objectMeta *metav1.ObjectMeta,
	ingressUrl *flds.StringField, isJupyter bool, isTensorboard bool, isAuthDisabled bool) error {
	if !isJupyter && !isTensorboard {
		return nil
	}

	var serviceNameSuffix string
	if isJupyter {
		objectMeta.Labels[constants.LabelJupyter] = "true"
		serviceNameSuffix = JupyterServiceSuffix
	} else {
		objectMeta.Labels[constants.LabelTensorboard] = "true"
		serviceNameSuffix = TensorboardServiceSuffix
	}

	if ingressUrl != nil {
		jobUrl, err := ComposeJupyterOrTensorboardJobUrl(ctx, workloadBaseName)
		if err != nil {
			return fmt.Errorf("failed to compose job extneral url: %w", err)
		}

		serviceName := getServiceName(workloadInitials, workloadBaseName, serviceNameSuffix)
		jupyterService, err := createServiceExposedUrl(jobUrl, workloadBaseName, objectMeta.Namespace, DefaultJupyterAndTensorContainerPort, objectMeta, serviceName, 0, false, isAuthDisabled)
		if err != nil {
			return err
		}

		if eb.services == nil {
			eb.services = make([]corev1.Service, 0, 1)
		}
		eb.services = append(eb.services, jupyterService)

		err = eb.buildIngressOrRoute(objectMeta, *jobUrl, &jupyterService)
		if err != nil {
			return err
		}
	}

	return nil
}

func (eb *ExternalUrlBuilder) buildIngressOrRoute(objectMeta *metav1.ObjectMeta, jobUrl apis.URL, service *corev1.Service) error {
	if config.IngressEnabled() {
		if eb.isOpenshift {
			eb.buildRouteForExposedUrl(objectMeta, jobUrl, service)
		} else {
			if err := eb.buildIngressForExposedUrl(objectMeta, jobUrl, service); err != nil {
				return err
			}
		}
	}

	if config.GatewayEnabled() {
		if err := eb.buildHTTPRouteForExposedUrl(objectMeta, jobUrl, service); err != nil {
			return err
		}
	}

	AddJobUrlToAnnotations(objectMeta, jobUrl.String())

	return nil
}

func (eb *ExternalUrlBuilder) buildIngressForExposedUrl(objectMeta *metav1.ObjectMeta, jobUrl apis.URL, service *corev1.Service) error {
	if eb.ingresses == nil {
		eb.ingresses = make([]networkingv1.Ingress, 0)
	}
	hostPort, err := UrlToPort(&jobUrl)
	if err != nil {
		return err
	}

	ingressAnnotations := config.IngressAnnotations()
	if ingressAnnotations == nil {
		ingressAnnotations = make(map[string]string)
	}
	ingressAnnotations[constants.IngressBodySizeAnnotationKey] = constants.IngressBodySizeAnnotationDefaultValue

	ingressClass := config.IngressClass()
	var ingressClassName *string
	if ingressClass != "" {
		ingressClassName = &ingressClass
	}

	ingress := networkingv1.Ingress{
		TypeMeta: networkingV1TypeMeta(&networkingv1.Ingress{}),
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   objectMeta.Namespace,
			Annotations: ingressAnnotations,
			Labels:      objectMeta.Labels,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: jobUrl.URL().Hostname(),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{},
					},
				},
			},
		},
	}

	ingressPort := networkingv1.HTTPIngressPath{
		Path:     jobUrl.Path,
		PathType: (*networkingv1.PathType)(ptr.To(string(networkingv1.PathTypePrefix))),
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: service.ObjectMeta.Name,
				Port: networkingv1.ServiceBackendPort{
					Number: hostPort,
				},
			},
		},
	}

	ingress.Spec.Rules[0].HTTP.Paths = append(ingress.Spec.Rules[0].HTTP.Paths, ingressPort)

	eb.ingresses = append(eb.ingresses, ingress)

	return nil
}

func (eb *ExternalUrlBuilder) buildRouteForExposedUrl(objectMeta *metav1.ObjectMeta, jobUrl apis.URL, service *corev1.Service) error {
	if eb.routes == nil {
		eb.routes = make([]openshiftv1.Route, 0)
	}

	hostUrl := jobUrl.URL().Hostname()
	// OpenShift host-based routing mechanism already supports wildcard certificate authorization.
	// We just need to trim the 'runai' str from our subdomain routes.
	if config.SubdomainSupportEnabled() {
		hostUrl = strings.Replace(hostUrl, config.RunaiTrimStr, "", 1)
	}

	termination := DefaultRouteTermination()
	route := openshiftv1.Route{
		TypeMeta: routeTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            objectMeta.Labels,
			Name:              service.Name,
			Namespace:         objectMeta.Namespace,
		},
		Spec: openshiftv1.RouteSpec{
			Host: hostUrl,
			To: openshiftv1.RouteTargetReference{
				Kind: constants.KindService,
				Name: service.Name,
			},
			TLS: &openshiftv1.TLSConfig{
				InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   termination,
			},
		},
	}
	// OpenShift rejects passthrough Routes with a non-empty spec.path because the
	// Router cannot do path-based routing on encrypted traffic.
	if termination != openshiftv1.TLSTerminationPassthrough {
		route.Spec.Path = jobUrl.Path
	}

	eb.routes = append(eb.routes, route)

	return nil
}

func (eb *ExternalUrlBuilder) buildHTTPRouteForExposedUrl(objectMeta *metav1.ObjectMeta, jobUrl apis.URL, service *corev1.Service) error {
	// Gateway name and namespace are guaranteed non-empty by the go-operator's
	// deployment builder and CRD validation when GATEWAY_API_ENABLED=true.
	if eb.httpRoutes == nil {
		eb.httpRoutes = make([]gatewayv1.HTTPRoute, 0)
	}

	hostUrl := jobUrl.URL().Hostname()
	if eb.isOpenshift && config.SubdomainSupportEnabled() {
		hostUrl = strings.Replace(hostUrl, config.RunaiTrimStr, "", 1)
	}

	hostPort, err := UrlToPort(&jobUrl)
	if err != nil {
		return err
	}

	path := jobUrl.Path
	if path == "" {
		path = "/"
	}

	pathPrefix := gatewayv1.PathMatchPathPrefix
	port := gatewayv1.PortNumber(hostPort)
	gwNamespace := gatewayv1.Namespace(config.GatewayNamespace())

	httpRoute := gatewayv1.HTTPRoute{
		TypeMeta: httpRouteTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: objectMeta.Namespace,
			Labels:    objectMeta.Labels,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{
					{
						Name:      gatewayv1.ObjectName(config.GatewayName()),
						Namespace: &gwNamespace,
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(hostUrl)},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  &pathPrefix,
								Value: &path,
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(service.ObjectMeta.Name),
									Port: &port,
								},
							},
						},
					},
				},
			},
		},
	}

	eb.httpRoutes = append(eb.httpRoutes, httpRoute)
	return nil
}

func createServiceExposedUrl(url *apis.URL, workloadName string, namespace string, containerPort int32,
	topmostResource *metav1.ObjectMeta,
	serviceName string, idx int, isExposedUrlPrivate, isAuthDisabled bool) (corev1.Service, error) {
	hostPort, err := UrlToPort(url)
	if err != nil {
		return corev1.Service{}, err
	}

	if !isAuthDisabled {
		if isExposedUrlPrivate && config.ExternalAuthUrlEnabled() {
			containerPort = constants.ExternalUrlAuthContainerDefaultPort + int32(idx)
		}
	}

	sps := []corev1.ServicePort{
		{
			Name:       fmt.Sprintf("port-%s-%d-0", workloadName, containerPort),
			Port:       hostPort,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: containerPort},
			Protocol:   corev1.ProtocolTCP,
		},
	}

	return corev1.Service{
		TypeMeta: serviceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            topmostResource.Labels,
			Name:              serviceName,
			Namespace:         namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    sps,
			Selector: map[string]string{constants.LabelRelease: topmostResource.Name},
			Type:     constants.ServiceTypeClusterIP,
		},
	}, nil
}

// ////////////////////////////////////////////////////////////////////////////////////
// GetUrlPerNick receives the exposedUrls defined in the workload, and returns
// a map nick name to the url that was created.
//
//	**************
//	**** NOTE ****
//	**************
//
// This function is used by both workload controller (cluster) AND workloads service (control plane)
// The CP uses it for mapping between the nickname to the url and it stores it in the connections
// of the workload spec. Therefore, any change in the url structure must take into account existing
// workloads that may depend on this mapping.
// ////////////////////////////////////////////////////////////////////////////////////
func GetUrlPerNick(exposedUrls *flds.ExposedUrlsField, projectName string, workloadBaseName string,
	clusterDomain string, noSuffixUrlGenerated *bool, options ComposeUrlOptions,
) (map[string]*apis.URL, error) {
	result := make(map[string]*apis.URL)

	for _, nick := range commonutils.SortedKeys(exposedUrls.Deref()) {
		exposedUrl := exposedUrls.Deref()[nick]
		if exposedUrl.Value == nil {
			continue
		}

		// validate consistency b/w the customUrl flag and the url value
		customUrlEnabled := ptr.Deref(exposedUrl.Value.CustomUrl, false)
		customUrlProvided := ptr.Deref(exposedUrl.Value.Url, "") != ""
		if customUrlEnabled && !customUrlProvided {
			return nil, errors.New("url is mandatory when customUrl is enabled")
		}
		if !customUrlEnabled && customUrlProvided {
			return nil, errors.New("url should only be provided for customUrl")
		}

		var url *apis.URL
		var urlErr error
		if customUrlProvided {
			url, urlErr = apis.ParseURL(*exposedUrl.Value.Url)
		} else {
			urlSuffix := nick
			if !*noSuffixUrlGenerated {
				urlSuffix = ""
				*noSuffixUrlGenerated = true
			}
			url, urlErr = ComposeExposedUrl(urlSuffix, workloadBaseName, projectName, clusterDomain, options)
		}
		if urlErr != nil {
			return nil, urlErr
		}

		result[nick] = url
	}

	return result, nil
}
