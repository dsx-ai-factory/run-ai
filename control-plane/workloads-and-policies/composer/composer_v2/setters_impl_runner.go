/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/init_containers"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	pcreds "github.com/run-ai/runai/runai-cluster/cluster/common/personalcreds"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//////////////////////////////////////////////////////////////////////
// CONNECTIONS
//////////////////////////////////////////////////////////////////////

// serviceTypeOrder is the canonical order in which port Services are emitted.
// It mirrors the V1 cluster builder (GetPortsPerServiceType) so that the
// generated Service name suffixes ("0-clusterip", "1-loadbalancer", ...) stay
// aligned with how the cluster historically numbered them.
var serviceTypeOrder = []openapi.ServiceType{
	openapi.SERVICETYPE_CLUSTER_IP,
	openapi.SERVICETYPE_LOAD_BALANCER,
	openapi.SERVICETYPE_NODE_PORT,
}

// PortsServiceOptions is the cookie accepted by PortsToWorkloadRunner. It lets
// the (distributed) workload composer drive per-replica Service generation:
//   - Selector targets a specific replica's pods (Kubeflow labels). When nil,
//     the setter defaults to {"release": workloadName}, matching the V1 builder.
//   - NameQualifier (e.g. "worker", "master", "launcher") is inserted into
//     the Service name so each supported role's Services are unique and don't
//     collide. It is empty for non-distributed workloads.
type PortsServiceOptions struct {
	Selector      map[string]string
	NameQualifier string
}

// PortsToWorkloadRunner builds the networking-plugin Service configuration from
// the per-item Port spec. It mirrors the V1 cluster builder's behavior: ports
// are grouped by Service type and a single Service (one NetworkingEndpoint) is
// created per type, so that e.g. multiple LoadBalancer ports share a single
// external IP. No external access (Ingress/Route) is created for plain ports.
// The cookie is an optional *PortsServiceOptions; when nil the setter defaults
// to a {"release": workloadName} selector and no name qualifier.
func PortsToWorkloadRunner(info *composer.PortsInfo, target any, cookie any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	cfg, ok := target.(*openapi.WorkloadV2Configuration)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("PortsToWorkloadRunner: target type %T is not *openapi.WorkloadV2Configuration", target))
	}

	var selector map[string]string
	var nameQualifier string
	switch opts := cookie.(type) {
	case nil:
		selector = map[string]string{"release": info.WorkloadName}
	case *PortsServiceOptions:
		selector = opts.Selector
		if selector == nil {
			selector = map[string]string{"release": info.WorkloadName}
		}
		nameQualifier = opts.NameQualifier
	default:
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("PortsToWorkloadRunner: cookie type %T is not *PortsServiceOptions", cookie))
	}

	// Group the ports by their resolved Service type. Each port within a Service
	// is named after its container port (port-<containerPort>) - short, unique
	// within the Service, and within Kubernetes' 15-char IANA_SVC_NAME limit.
	portsByServiceType := make(map[openapi.ServiceType][]openapi.ServicePort)

	for _, nick := range composer.SortedMapKeys(info.Items) {
		p := info.Items[nick]
		if p == nil {
			continue
		}

		serviceType, typeErr := composer.PortServiceTypeToServiceType(p.ServiceType.Get())
		if typeErr != nil {
			return typeErr
		}

		containerPort := p.Container.Get()
		if containerPort == nil {
			return workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("PortsToWorkloadRunner: port %q is missing the mandatory container port", nick))
		}

		servicePort := buildServicePort(*containerPort, p.External.Get(), serviceType)
		portsByServiceType[serviceType] = append(portsByServiceType[serviceType], *servicePort)
	}

	// Emit one Service (one NetworkingEndpoint) per service type in canonical
	// order. The Service name is "<prefix>-<workload>-<typeshort>[-<qualifier>]"
	// where typeshort is cip/lb/np, mirroring the URL Service naming pattern
	// (kf-<wl>-url-<role>). Distributed callers pass the supported replica
	// role (for example "worker", "master", or "launcher") so per-role
	// Services never collide; non-distributed callers pass an empty qualifier.
	for _, serviceType := range serviceTypeOrder {
		ports := portsByServiceType[serviceType]
		if len(ports) == 0 {
			continue
		}

		suffix := composer.ServiceTypeShortName(serviceType)
		if nameQualifier != "" {
			suffix = fmt.Sprintf("%s-%s", suffix, nameQualifier)
		}
		serviceName := composer.ComposeNetworkingResourceName(info.ResourcePrefix, info.WorkloadName, suffix)

		serviceSpec := openapi.NewServiceSpec(serviceName, selector, ports)
		serviceSpec.SetType(serviceType)

		svc := openapi.NewNetworkingService()
		svc.SetSpec(*serviceSpec)

		cfg.Networking = append(cfg.Networking, openapi.NetworkingEndpoint{
			Services: []openapi.NetworkingService{*svc},
		})
	}

	return nil
}

// buildServicePort maps a single workload Port into a ServiceSpec port,
// mirroring the V1 cluster builder. For NodePort services the external port is
// the allocated nodePort and the service port equals the container port; for
// ClusterIP / LoadBalancer the external port (when supplied) becomes the
// service port. The targetPort always points at the container port. The port
// name is "port-<containerPort>" - short and unique within the Service.
func buildServicePort(containerPort int32, external *int32, serviceType openapi.ServiceType) *openapi.ServicePort {
	servicePort := openapi.NewServicePort(containerPort)
	servicePort.SetName(fmt.Sprintf("port-%d", containerPort))
	servicePort.SetTargetPort(containerPort)
	servicePort.SetProtocol(openapi.SERVICEPROTOCOL_TCP)

	if serviceType == openapi.SERVICETYPE_NODE_PORT {
		servicePort.SetPort(containerPort)
		if external != nil {
			servicePort.SetNodePort(*external)
		}
		return servicePort
	}

	if external != nil {
		servicePort.SetPort(*external)
	} else {
		servicePort.SetPort(containerPort)
	}
	return servicePort
}

// UrlsToWorkloadRunner builds the networking-plugin configuration from the
// per-item ExposedUrl spec. The body is filled by the networking-plugin work;
// the oauth2/username-loader init containers are handled by UrlsToK8sInitContainers.
// The cookie parameter can carry a custom label selector (map[string]string) for
// Services; if nil, defaults to {"release": workloadName}.
func UrlsToWorkloadRunner(info *composer.ExposedUrlInfo, target any, cookie any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	cfg, ok := target.(*openapi.WorkloadV2Configuration)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("UrlsToWorkloadRunner: target type %T is not *openapi.WorkloadV2Configuration", target))
	}

	if cfg.Networking == nil {
		cfg.Networking = make([]openapi.NetworkingEndpoint, 0, len(info.Items))
	}

	const defaultHttpPort = int32(80)

	// The cookie parameter can carry a custom label selector for Services.
	// For non-distributed workloads, cookie is typically nil and we default to "release".
	// For Kubeflow distributed workloads, the caller passes kubeflow-specific labels
	// (job-name, replica-type, replica-index) matching the target pod.
	var selector map[string]string
	if cookie == nil {
		selector = map[string]string{"release": info.WorkloadName}
	} else {
		var ok bool
		selector, ok = cookie.(map[string]string)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("UrlsToWorkloadRunner: cookie type %T is not map[string]string", cookie))
		}
	}

	// Track whether the first (no-suffix) URL has been generated.
	// We mirror the notification's connections-pre-processor logic, which
	// starts at false and lets the first auto URL be suffix-less, regardless of
	// whether the workload is jupyter/tensorboard. This keeps the Ingress URL
	// aligned with the notification's exposedUrlPath/exposedUrlSubDomain values.
	//
	// When ForceSuffixInUrls is set (distributed non-primary role), start at
	// true so every URL includes its nick suffix and avoids competing with the
	// primary role for the same Ingress host.
	noSuffixUrlGenerated := info.ForceSuffixInUrls

	for _, nick := range composer.SortedMapKeys(info.Items) {
		item := info.Items[nick]
		if item == nil {
			continue
		}

		// When this URL has no OAuth proxy, route the Service directly to the
		// workload application. Forced-proxy URLs continue to use the OAuth port
		// even when external URL auth is disabled cluster-wide.
		targetPort := item.ContainerPort
		if init_containers.ShouldBuildOAuth2Proxy(info.Config, item) {
			targetPort = item.OAuthPort
		}
		if targetPort == nil {
			continue
		}
		serviceTargetPort := *targetPort
		portName := fmt.Sprintf("port-%d", serviceTargetPort)

		// Determine the URL: either custom-provided or auto-generated
		urlStr := item.Url
		if urlStr == "" {
			// Auto-generate the URL
			urlSuffix := nick
			if !noSuffixUrlGenerated {
				urlSuffix = ""
				noSuffixUrlGenerated = true
			}
			autoUrl, autoErr := init_containers.ComposeAutoExposedURL(
				info.Config.ClusterDomain(),
				info.ProjectName,
				info.WorkloadName,
				urlSuffix,
				info.Config.SubdomainSupportEnabled(),
			)
			if autoErr != nil {
				return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose auto URL for %q: %v", nick, autoErr))
			}
			urlStr = autoUrl.String()
		}

		port := openapi.NewServicePort(defaultHttpPort)
		port.SetName(portName)
		port.SetTargetPort(serviceTargetPort)

		// K8s resource names follow the cluster builder convention
		// "<workloadInitials>-<workloadName>-<suffix>" so two workloads with the
		// same base name but different kinds (e.g. job vs distributed) cannot
		// collide on Service/Ingress names.
		serviceName := composer.ComposeNetworkingResourceName(info.ResourcePrefix, info.WorkloadName, item.Nick)
		serviceSpec := openapi.NewServiceSpec(
			serviceName,
			selector,
			[]openapi.ServicePort{*port},
		)

		svc := openapi.NewNetworkingService()
		svc.SetSpec(*serviceSpec)

		endpoint := openapi.NetworkingEndpoint{
			Services: []openapi.NetworkingService{*svc},
		}

		// Build external access (Ingress/Route/HTTPRoute)
		externalAccess, err := buildExternalAccess(info.Config, serviceName, urlStr, defaultHttpPort)
		if err != nil {
			return err
		}
		if externalAccess != nil {
			endpoint.ExternalAccess = *openapi.NewNullableExternalAccess(externalAccess)
		}

		cfg.Networking = append(cfg.Networking, endpoint)
	}

	return nil
}

// runaiHostTrimPrefix is stripped from the URL host on OpenShift+subdomain
// clusters so the hostname matches the *.<domain> wildcard certificate. The
// V1 cluster builder applies the same trim (RunaiTrimStr) to Routes/HTTPRoutes
// and the oauth2-proxy redirect/whitelist domains; we keep the value identical
// so the generated resources line up.
const runaiHostTrimPrefix = "runai."

// buildExternalAccess builds the per-endpoint ExternalAccess for the cluster's
// enabled external-access mechanisms, mirroring the V1 cluster builder's
// dispatch in workload-common/builders/external_url.go::buildIngressOrRoute.
func buildExternalAccess(config composer.ClusterStorageConfig, serviceName, urlStr string, servicePort int32) (*openapi.ExternalAccess, *workloadsApi.Error) {
	externalAccess := openapi.NewExternalAccess()
	ingressEnabled := true
	gatewayAPIEnabled := false
	if config != nil {
		ingressEnabled = config.IngressEnabled()
		gatewayAPIEnabled = config.GatewayAPIEnabled()
	}

	if ingressEnabled {
		if config != nil && config.IsOpenshift() {
			route, err := buildRouteForExposedUrl(config, serviceName, urlStr)
			if err != nil {
				return nil, err
			}
			externalAccess.SetRoute(route)
		} else {
			ingress, err := buildIngressForExposedUrl(serviceName, urlStr, servicePort)
			if err != nil {
				return nil, err
			}
			externalAccess.SetIngress(ingress)
		}
	}

	if gatewayAPIEnabled {
		if config == nil || config.GatewayName() == "" {
			return nil, workloadsApi.NewError(http.StatusServiceUnavailable, "cluster Gateway API is enabled but gateway name is not reported")
		}
		httpRoute, err := buildHTTPRouteForExposedUrl(config, serviceName, urlStr, servicePort)
		if err != nil {
			return nil, err
		}
		externalAccess.SetHttpRoute(httpRoute)
	}

	if !externalAccess.HasIngress() && !externalAccess.HasRoute() && !externalAccess.HasHttpRoute() {
		return nil, workloadsApi.NewError(http.StatusServiceUnavailable, "cluster has no enabled external URL routing mechanism")
	}

	return externalAccess, nil
}

// BuildExternalAccess exposes the shared Ingress, Route, and HTTPRoute construction to typed V2 composers.
func BuildExternalAccess(config composer.ClusterStorageConfig, serviceName, urlStr string, servicePort int32) (*openapi.ExternalAccess, *workloadsApi.Error) {
	return buildExternalAccess(config, serviceName, urlStr, servicePort)
}

func buildIngressForExposedUrl(serviceName, urlStr string, servicePort int32) (map[string]interface{}, *workloadsApi.Error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid URL %q: %v", urlStr, err))
	}

	host := parsed.Hostname()
	path := parsed.Path
	if path == "" {
		path = "/"
	}

	ingress := map[string]interface{}{
		"apiVersion": "networking.k8s.io/v1",
		"kind":       "Ingress",
		"metadata": map[string]interface{}{
			"name": serviceName,
		},
		"spec": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"host": host,
					"http": map[string]interface{}{
						"paths": []map[string]interface{}{
							{
								"path":     path,
								"pathType": "Prefix",
								"backend": map[string]interface{}{
									"service": map[string]interface{}{
										"name": serviceName,
										"port": map[string]interface{}{
											"number": servicePort,
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

	return ingress, nil
}

// buildRouteForExposedUrl builds the OpenShift Route that exposes the
// per-endpoint Service externally. Routes by Host with HTTP-to-HTTPS redirect
// and uses the cluster-configured TLS termination policy (default edge;
// passthrough enables customer-terminated end-to-end TLS).
func buildRouteForExposedUrl(config composer.ClusterStorageConfig, serviceName, urlStr string) (map[string]interface{}, *workloadsApi.Error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid URL %q: %v", urlStr, err))
	}

	host := parsed.Hostname()
	if config != nil && config.SubdomainSupportEnabled() {
		host = strings.Replace(host, runaiHostTrimPrefix, "", 1)
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	termination := string(clustersApi.OPENSHIFTROUTETERMINATIONPOLICY_EDGE)
	if config != nil {
		termination = config.OpenshiftRouteTerminationPolicy()
	}

	route := map[string]interface{}{
		"apiVersion": "route.openshift.io/v1",
		"kind":       "Route",
		"metadata": map[string]interface{}{
			"name": serviceName,
		},
		"spec": map[string]interface{}{
			"host": host,
			"to": map[string]interface{}{
				"kind": "Service",
				"name": serviceName,
			},
			"tls": map[string]interface{}{
				"insecureEdgeTerminationPolicy": "Redirect",
				"termination":                   termination,
			},
		},
	}
	// OpenShift rejects passthrough Routes with a non-empty spec.path because the
	// Router cannot do path-based routing on encrypted traffic.
	if termination != string(clustersApi.OPENSHIFTROUTETERMINATIONPOLICY_PASSTHROUGH) {
		route["spec"].(map[string]interface{})["path"] = path
	}

	return route, nil
}

// buildHTTPRouteForExposedUrl builds the Gateway API HTTPRoute that exposes
// the per-endpoint Service through the cluster Gateway. Mirrors the V1 cluster
// builder (buildHTTPRouteForExposedUrl): the HTTPRoute attaches to the
// cluster's configured Gateway via parentRefs, matches the URL path with
// PathPrefix, and forwards to the Service. On OpenShift+subdomain we strip
// the "runai." host prefix so traffic lands on the platform's wildcard.
func buildHTTPRouteForExposedUrl(config composer.ClusterStorageConfig, serviceName, urlStr string, servicePort int32) (map[string]interface{}, *workloadsApi.Error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid URL %q: %v", urlStr, err))
	}

	host := parsed.Hostname()
	if config != nil && config.IsOpenshift() && config.SubdomainSupportEnabled() {
		host = strings.Replace(host, runaiHostTrimPrefix, "", 1)
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	parentRef := map[string]interface{}{
		"name": config.GatewayName(),
	}
	if ns := config.GatewayNamespace(); ns != "" {
		parentRef["namespace"] = ns
	}

	httpRoute := map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]interface{}{
			"name": serviceName,
		},
		"spec": map[string]interface{}{
			"parentRefs": []map[string]interface{}{parentRef},
			"hostnames":  []string{host},
			"rules": []map[string]interface{}{
				{
					"matches": []map[string]interface{}{
						{
							"path": map[string]interface{}{
								"type":  "PathPrefix",
								"value": path,
							},
						},
					},
					"backendRefs": []map[string]interface{}{
						{
							"name": serviceName,
							"port": servicePort,
						},
					},
				},
			},
		},
	}

	return httpRoute, nil
}

//////////////////////////////////////////////////////////////////////
// STORAGE
//////////////////////////////////////////////////////////////////////

func MapPvcToWorkloadRunner(src map[string]*workloadsApi.PvcInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*composer.VolumesAndMounts)
	if !ok || targetPtr == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("MapPvcToWorkloadRunner: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		volume, err := pvcVolumeFromInstance(key, instance)
		if err != nil {
			return err
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		readOnly := convutils.BoolIsTrue(instance.ReadOnly.Get())
		targetPtr.VolumeMount = append(targetPtr.VolumeMount, corev1.VolumeMount{
			Name:      key,
			MountPath: convutils.Deref(instance.Path.Get(), ""),
			ReadOnly:  readOnly,
		})

		if !convutils.BoolIsTrue(instance.ExistingPvc.Get()) && !convutils.BoolIsTrue(instance.Ephemeral.Get()) {
			if targetPtr.Config == nil {
				return workloadsApi.NewError(http.StatusInternalServerError,
					"MapPvcToWorkloadRunner: target Config must not be nil for new persistent PVCs")
			}

			pvcConfig, err := pvcConfigFromInstance(instance)
			if err != nil {
				return err
			}
			appendPvcConfig(targetPtr.Config, pvcConfig)
		}
	}

	return nil
}

func pvcConfigFromInstance(instance *workloadsApi.PvcInstance) (openapi.WorkloadV2UpdateConfigurationPvc, *workloadsApi.Error) {
	claimInfo := instance.ClaimInfo.Get()
	if claimInfo == nil {
		return openapi.WorkloadV2UpdateConfigurationPvc{},
			workloadsApi.NewError(http.StatusBadRequest, "pvc: claimInfo is mandatory for non-existing PVC")
	}

	accessModes := pvcAccessModesToWorkloadV2(claimInfo.AccessModes.Get())
	claimName := convutils.Deref(instance.ClaimName.Get(), "")
	if claimName == "" {
		return openapi.WorkloadV2UpdateConfigurationPvc{},
			workloadsApi.NewError(http.StatusBadRequest, "pvc: claimName is mandatory for non-existing PVC")
	}

	size := convutils.Deref(claimInfo.Size.Get(), "")
	if size == "" {
		return openapi.WorkloadV2UpdateConfigurationPvc{},
			workloadsApi.NewError(http.StatusBadRequest, "pvcSize is mandatory for non-existing PVC")
	}

	annotations := pvcAnnotationsFromAddedAttrs(claimInfo.AddedAttrValues)
	result := openapi.WorkloadV2UpdateConfigurationPvc{
		Metadata: openapi.WorkloadV2UpdateConfigurationPvcMetadata{
			Name: claimName,
		},
		Spec: openapi.WorkloadV2UpdateConfigurationPvcSpec{
			AccessModes: accessModes,
			Resources: openapi.WorkloadV2UpdateConfigurationPvcResources{
				Requests: openapi.WorkloadV2UpdateConfigurationPvcResourcesRequests{
					Storage: size,
				},
			},
		},
	}
	if len(annotations) != 0 {
		result.Metadata.Annotations = &annotations
	}

	if convutils.HasValue(claimInfo.StorageClass) {
		result.Spec.SetStorageClassName(*claimInfo.StorageClass.Get())
	}

	if convutils.HasValue(claimInfo.VolumeMode) {
		result.Spec.SetVolumeMode(pvcVolumeModeToWorkloadV2(*claimInfo.VolumeMode.Get()))
	}

	return result, nil
}

func pvcVolumeFromInstance(volumeName string, instance *workloadsApi.PvcInstance) (corev1.Volume, *workloadsApi.Error) {
	if convutils.BoolIsTrue(instance.ExistingPvc.Get()) || !convutils.BoolIsTrue(instance.Ephemeral.Get()) {
		claimName := convutils.Deref(instance.ClaimName.Get(), "")
		if claimName == "" {
			return corev1.Volume{},
				workloadsApi.NewError(http.StatusBadRequest, "pvc: claimName is mandatory")
		}

		return corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
					ReadOnly:  convutils.BoolIsTrue(instance.ReadOnly.Get()),
				},
			},
		}, nil
	}

	claimTemplate, err := pvcClaimTemplateFromInstance(instance)
	if err != nil {
		return corev1.Volume{}, err
	}

	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Ephemeral: &corev1.EphemeralVolumeSource{
				VolumeClaimTemplate: claimTemplate,
			},
		},
	}, nil
}

func pvcClaimTemplateFromInstance(instance *workloadsApi.PvcInstance) (*corev1.PersistentVolumeClaimTemplate, *workloadsApi.Error) {
	claimInfo := instance.ClaimInfo.Get()
	if claimInfo == nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "pvc: claimInfo is mandatory for non-existing PVC")
	}

	accessModes := pvcAccessModesToK8s(claimInfo.AccessModes.Get())
	size := convutils.Deref(claimInfo.Size.Get(), "")
	if size == "" {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "pvcSize is mandatory for non-existing PVC")
	}

	storageQuantity, parseErr := resource.ParseQuantity(size)
	if parseErr != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("pvc: invalid storage quantity %q: %v", size, parseErr))
	}

	template := &corev1.PersistentVolumeClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: pvcAnnotationsFromAddedAttrs(claimInfo.AddedAttrValues),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageQuantity,
				},
			},
		},
	}

	if convutils.HasValue(claimInfo.StorageClass) {
		template.Spec.StorageClassName = claimInfo.StorageClass.Get()
	}

	if convutils.HasValue(claimInfo.VolumeMode) {
		volumeMode := corev1.PersistentVolumeMode(*claimInfo.VolumeMode.Get())
		template.Spec.VolumeMode = &volumeMode
	}

	return template, nil
}

// pvcAccessModesToWorkloadV2 mirrors the V1 PvcAccessModes schema, which
// declares readWriteOnce with a default of true. When a V1 request omits
// access modes we preserve that default so the translated WorkloadV2 request
// still satisfies workloads-manager's non-empty AccessModes requirement.
func pvcAccessModesToWorkloadV2(accessModes *workloadsApi.PvcAccessModes) []openapi.WorkloadV2UpdateConfigurationPvcAccessMode {
	result := make([]openapi.WorkloadV2UpdateConfigurationPvcAccessMode, 0, 3)
	if accessModes == nil {
		return append(result, openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE)
	}

	if convutils.BoolIsTrue(accessModes.ReadWriteOnce.Get()) {
		result = append(result, openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE)
	}
	if convutils.BoolIsTrue(accessModes.ReadOnlyMany.Get()) {
		result = append(result, openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_ONLY_MANY)
	}
	if convutils.BoolIsTrue(accessModes.ReadWriteMany.Get()) {
		result = append(result, openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_MANY)
	}
	return result
}

// pvcAccessModesToK8s mirrors the V1 PvcAccessModes schema default of
// readWriteOnce=true so ephemeral PVC templates keep passing k8s validation
// when the V1 request omits access modes.
func pvcAccessModesToK8s(accessModes *workloadsApi.PvcAccessModes) []corev1.PersistentVolumeAccessMode {
	result := make([]corev1.PersistentVolumeAccessMode, 0, 3)
	if accessModes == nil {
		return append(result, corev1.ReadWriteOnce)
	}

	if convutils.BoolIsTrue(accessModes.ReadWriteOnce.Get()) {
		result = append(result, corev1.ReadWriteOnce)
	}
	if convutils.BoolIsTrue(accessModes.ReadOnlyMany.Get()) {
		result = append(result, corev1.ReadOnlyMany)
	}
	if convutils.BoolIsTrue(accessModes.ReadWriteMany.Get()) {
		result = append(result, corev1.ReadWriteMany)
	}
	return result
}

func pvcVolumeModeToWorkloadV2(volumeMode workloadsApi.PvcVolumeMode) openapi.WorkloadV2UpdateConfigurationPvcVolumeMode {
	return openapi.WorkloadV2UpdateConfigurationPvcVolumeMode(volumeMode)
}

func pvcAnnotationsFromAddedAttrs(addedAttrValues []workloadsApi.PvcAddedAttrValue) map[string]string {
	if len(addedAttrValues) == 0 {
		return nil
	}

	result := make(map[string]string, len(addedAttrValues))
	for _, attr := range addedAttrValues {
		result[attr.Key] = convutils.Deref(attr.Value, "")
	}
	return result
}

func appendPvcConfig(config *openapi.WorkloadV2Configuration, pvc openapi.WorkloadV2UpdateConfigurationPvc) {
	for idx, existing := range config.Pvcs {
		if existing.Metadata.Name == pvc.Metadata.Name {
			config.Pvcs[idx] = pvc
			return
		}
	}

	config.Pvcs = append(config.Pvcs, pvc)
}

// BoolToPvcsCleanupConfig stores the cleanup-on-completion flag in the runner when the workload creates PVCs
func BoolToPvcsCleanupConfig(src *bool, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}
	config, ok := target.(*openapi.WorkloadV2Configuration)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("BoolToPvcsCleanupConfig: target type %T is not *WorkloadV2Configuration", target))
	}
	if config == nil || len(config.Pvcs) == 0 {
		return nil
	}

	pvcsConfig := config.PvcsConfig.Get()
	if pvcsConfig == nil {
		pvcsConfig = &openapi.WorkloadV2UpdateConfigurationPvcsConfig{}
	}
	if pvcsConfig.CleanupOnCompletion.IsSet() {
		return nil
	}
	pvcsConfig.CleanupOnCompletion = *openapi.NewNullableBool(src)
	config.PvcsConfig.Set(pvcsConfig)
	return nil
}

//////////////////////////////////////////////////////////////////////
// Image Pull Secrets
//////////////////////////////////////////////////////////////////////

type PersonalCredsCookie struct {
	WorkloadName string
}

type PersonalCredsTarget struct {
	Config  *openapi.WorkloadV2Configuration
	PodSpec *corev1.PodSpec
}

func ArrayToImagePullSecretsSetter(src []string, target any, cookie any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*PersonalCredsTarget)
	if !ok || targetPtr == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("arrayToImagePullSecrets: target type %T is not a non-nil *PersonalCredsTarget", target))
	}

	if targetPtr.PodSpec == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			"arrayToImagePullSecrets: target PodSpec must not be nil")
	}

	credCtx, ok := cookie.(*PersonalCredsCookie)
	if !ok || credCtx == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			"arrayToImagePullSecrets: cookie must be a non-nil *PersonalCredsCookie")
	}

	for _, secretName := range src {
		name := secretName
		if strings.HasPrefix(secretName, composer.UserCredentialsPrefix+":") {
			sourceSecret, _, parseErr := composer.ParseUserCredentialRef(secretName)
			if parseErr != nil {
				return parseErr
			}
			composer.RecordUserCredSource(targetPtr.Config, sourceSecret)
			name = pcreds.DestName(credCtx.WorkloadName, sourceSecret)
		}
		targetPtr.PodSpec.ImagePullSecrets = append(targetPtr.PodSpec.ImagePullSecrets, corev1.LocalObjectReference{Name: name})
	}

	return nil
}

type EnvVarUsingCredTarget struct {
	EnvVars *[]corev1.EnvVar
	Config  *openapi.WorkloadV2Configuration
}

func EnvVarsUsingCredSetter(src []composer.EnvVarInfo, target any, cookie any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*EnvVarUsingCredTarget)
	if !ok || targetPtr == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("EnvVarsUsingCredSetter: target type %T is not *EnvVarUsingCredCookie", target))
	}

	credCtx, ok := cookie.(*PersonalCredsCookie)
	if !ok || credCtx == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("envVarsUsingCredSetter: cookie type %T is not *PersonalCredsCookie", cookie))
	}

	if targetPtr.EnvVars == nil {
		targetPtr.EnvVars = &[]corev1.EnvVar{}
	}

	for _, info := range src {
		if info.Deleted {
			continue
		}
		source, _, parseErr := composer.ParseUserCredentialRef(info.ValueFrom)
		if parseErr != nil {
			return parseErr
		}
		composer.RecordUserCredSource(targetPtr.Config, source)

		*targetPtr.EnvVars = append(*targetPtr.EnvVars, corev1.EnvVar{
			Name: info.Name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pcreds.DestName(credCtx.WorkloadName, source),
					},
					Key: info.Key,
				},
			},
		})
	}

	return nil
}
