/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	ssv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	commonutils "github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

var MetricNames = map[string]string{
	"throughput":  "rps",
	"concurrency": "concurrency",
	"latency":     "knative_revision_request_latency",
}

const knativeProtocolGRPC = "h2c" // Knative supports grpc protocol by adding a port with name "h2c"

func createRevisionTemplateAnnotations(ctx context.Context, is ssv1.InferenceSpec) (map[string]string, error) {
	rawAnnotations, err := createRevisionTemplateRawAnnotations(is)
	if err != nil {
		return nil, err
	}

	specialAnnotations, err := createRevisionTemplateSpecialAnnotations(ctx, is)
	if err != nil {
		return nil, err
	}

	ret := utils.ConcatStringMaps(specialAnnotations, rawAnnotations)

	// Autoscaling class should be determined only after both Metric and MetricName fields were processed
	if isHPAAutoScalerMetric(ret[annotationMetric]) {
		ret[annotationAutoscalingClass] = annotationValueAutoscalingClassHPA
	}

	// Add all the annotations from the user
	for key, value := range is.Annotations.Deref() {
		ret[key] = ptr.Deref(value.Value, "")
	}

	return ret, nil
}

func createRevisionTemplateRawAnnotations(is ssv1.InferenceSpec) (map[string]string, error) {
	ret := map[string]string{}

	rawAnnotations := map[string]field_types.Field{
		annotationMinScale:                    is.MinScale,
		annotationMaxScale:                    is.MaxScale,
		annotationInitialScale:                is.InitialScale,
		annotationActivationScale:             is.ActivationScale,
		annotationMetric:                      is.MetricName,
		annotationTarget:                      is.Target,
		annotationTargetUtilizationPercentage: is.TargetUtilizationPercentage,
		annotationAutoscalingClass:            is.Class,
		constants.AnnotationUserName:          is.UserName,
	}

	for k, v := range rawAnnotations {
		if v.IsValueInitialized() {
			ret[k] = v.String()
		}
	}

	if is.ScaleToZeroRetentionSeconds != nil {
		ret[annotationScaleToZeroRetentionPeriod] = fmt.Sprintf("%ds", ptr.Deref(is.ScaleToZeroRetentionSeconds.Value, 0))
	}

	if is.ScaleDownDelaySeconds != nil {
		ret[annotationScaleDownDelay] = fmt.Sprintf("%ds", ptr.Deref(is.ScaleDownDelaySeconds.Value, 0))
	}

	if is.InitializationTimeoutSeconds != nil {
		ret[annotationProgressDeadline] = fmt.Sprintf("%ds", ptr.Deref(is.InitializationTimeoutSeconds.Value, 0))
	}

	return ret, nil
}

func createRevisionTemplateSpecialAnnotations(ctx context.Context, is ssv1.InferenceSpec) (map[string]string, error) {
	ret := map[string]string{}

	if is.Metric.IsValueInitialized() {
		if metricName, ok := MetricNames[is.Metric.Deref()]; ok {
			ret[annotationMetric] = metricName
		} else {
			return nil, fmt.Errorf("metric '%s' is not supported", is.Metric.Deref())
		}
	}

	resourceAnnotations, err := createResourceAllocationAnnotations(ctx, &is.ResourceAllocationFields, &is.RunaiSchedulerFields, is.Mps.Deref())
	if err != nil {
		return nil, err
	}

	return utils.ConcatStringMaps(ret, resourceAnnotations), nil
}

func createGpuResourceAllocationAnnotations(ctx context.Context, ra *field_groups.ResourceAllocationFields) (map[string]string, error) {
	var (
		memoryQuantity resource.Quantity
		ret            = map[string]string{}
		err            error
	)

	if ra == nil {
		return ret, nil
	}

	if ra.GpuMemory.IsValueInitialized() {
		memoryQuantity, err = ra.GpuMemory.DerefQuantity(0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gpu memory '%s'", ra.GpuMemory.Deref())
		}
	}

	if workloadbuilders.IsFractionalGpuOrMemroy(ctx, ra.Gpu.Deref(), memoryQuantity) {
		workloadbuilders.AddPartialGpuAnnotations(ret, ra.Gpu, memoryQuantity, ra.GpuDevices.Deref())
	}

	return ret, nil
}

func createResourceAllocationAnnotations(ctx context.Context, ra *field_groups.ResourceAllocationFields, rs *field_groups.RunaiSchedulerFields, mps bool) (map[string]string, error) {
	if ra == nil || rs == nil {
		// not supposed to happen, but just in case
		return nil, nil
	}

	ret, err := createGpuResourceAllocationAnnotations(ctx, ra)
	if err != nil {
		return nil, err
	}

	if mps {
		ret[annotationMps] = "true"
	}

	// set the used node pools in an annotation
	nodepoolsValue := field_groups.GetNodepoolsValue(*rs)
	if nodepoolsValue != "" {
		ret[commonconstants.AnnotationNodepools] = nodepoolsValue
	}

	return ret, nil
}

func createResourceAllocationAnnotationsForDistributedInference(ctx context.Context, ra *field_groups.ResourceAllocationFields, rs *field_groups.DistributedInferenceSchedulerFields) (map[string]string, error) {
	if ra == nil || rs == nil {
		// not supposed to happen, but just in case
		return nil, nil
	}

	ret, err := createGpuResourceAllocationAnnotations(ctx, ra)
	if err != nil {
		return nil, err
	}

	// set the used node pools in an annotation
	if rs.NodePools.IsValueInitialized() {
		nodepoolsValue := rs.NodePools.Deref()
		if nodepoolsValue == "" {
			return nil, fmt.Errorf("nodepools value is empty")
		}
		ret[commonconstants.AnnotationNodepools] = nodepoolsValue
	}

	return ret, nil
}

func isHPAAutoScalerMetric(metricAnnotation string) bool {
	return metricAnnotation != "rps" && metricAnnotation != "concurrency" && metricAnnotation != ""
}

func createRevisionTemplateLabels(ctx context.Context, k8sClient client.Client, workload *infv2.InferenceWorkload) (map[string]string, error) {
	ret := map[string]string{}

	// Not using workload.GetChildName() (for workloadLabelKey and LabelWorkloadName) as they should be workload's metadata name
	// even if there is a spec.name
	workloadName := workload.Name

	workloadLabelKey := fmt.Sprintf("%s%s", constants.LabelWorkloadPrefix, strings.ToLower(workload.Kind))

	ret[workloadLabelKey] = workloadName

	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err == nil {
		ret[constants.RunaiProjectLabel] = projectName
	}
	ret[constants.LabelRelease] = workloadName

	// we add WorkloadName and WorkloadKind labels to the pods
	ret[constants.LabelWorkloadName] = workloadName
	ret[constants.LabelWorkloadKind] = workload.Kind

	for key, value := range workload.Spec.Labels.Deref() {
		ret[key] = ptr.Deref(value.Value, "")
	}

	err = workloadbuilders.SetPriorityClassLabelOnLabels(ctx, k8sClient, workload.Spec.PriorityClassName.String(), ret)
	if err != nil {
		return ret, err
	}

	if value := workload.Spec.Preemptibility.Deref(); value != "" {
		ret[constants.LabelPreemptibilityName] = value
	}

	return ret, nil
}

func createContainerSecurityContext(containerFields *field_groups.ContainerFields, privileged *field_types.BoolField) *corev1.SecurityContext {
	sc := &corev1.SecurityContext{}

	if containerFields.RunAsUser.Deref() {
		if containerFields.RunAsUid != nil && containerFields.RunAsUid.Value != nil {
			sc.RunAsUser = ptr.To(int64(containerFields.RunAsUid.Deref()))
		}
		if containerFields.RunAsGid != nil && containerFields.RunAsGid.Value != nil {
			sc.RunAsGroup = ptr.To(int64(containerFields.RunAsGid.Deref()))
		}
	}

	if runAsNonRoot := containerFields.RunAsNonRoot.Deref(); runAsNonRoot {
		sc.RunAsNonRoot = ptr.To(true)
	}

	if readOnlyRootFilesystem := containerFields.ReadOnlyRootFilesystem.Deref(); readOnlyRootFilesystem {
		sc.ReadOnlyRootFilesystem = ptr.To(true)
	}

	if seccompProfileType := containerFields.SeccompProfileType.Deref(); seccompProfileType != "" {
		sc.SeccompProfile = &corev1.SeccompProfile{
			Type: corev1.SeccompProfileType(seccompProfileType),
		}
	}

	if len(containerFields.Capabilities.Deref()) != 0 || containerFields.DropAllCapabilities.Deref() {
		sc.Capabilities = &corev1.Capabilities{}

		if dropAllCapabilities := containerFields.DropAllCapabilities.Deref(); dropAllCapabilities {
			sc.Capabilities.Drop = []corev1.Capability{"all"}
		}

		if len(containerFields.Capabilities.Deref()) != 0 {
			sc.Capabilities.Add = []corev1.Capability{}
			capabilitiesMap := containerFields.Capabilities.Deref()

			for _, nick := range field_types.SortedCapabilitiesMapKeys(capabilitiesMap) {
				capability := capabilitiesMap[nick].Value
				if capability != nil {
					sc.Capabilities.Add = append(sc.Capabilities.Add, corev1.Capability(ptr.Deref(capability.Capability, "")))
				}
			}
		}
	}

	if privileged != nil {
		sc.Privileged = privileged.Value
	}

	if *sc == (corev1.SecurityContext{}) {
		sc = nil
	}

	return sc
}

func createPodSecurityContext(containerFields *field_groups.ContainerFields) *corev1.PodSecurityContext {
	psc := &corev1.PodSecurityContext{}

	if containerFields.RunAsGid != nil && containerFields.RunAsGid.Value != nil {
		psc.FSGroup = ptr.To(int64(containerFields.RunAsGid.Deref()))
	}

	if containerFields.SupplementalGroups.Deref() != "" {
		if groups, err := utils.DecodeInt64List(containerFields.SupplementalGroups.Deref(), ";"); err == nil {
			psc.SupplementalGroups = groups
		}
	}

	if reflect.DeepEqual(*psc, corev1.PodSecurityContext{}) {
		// if no fields were set, return nil
		psc = nil
	}
	return psc
}
func createGitSyncInitContainer(ctx context.Context, containerFields *field_groups.ContainerFields, storageFields *field_groups.StorageFields, templateSpec *servingv1.RevisionSpec) error {
	gitSyncInitContainer, err := workloadbuilders.GetGitSyncInitContainersWithError(ctx, containerFields, storageFields, false)
	if err != nil {
		return err
	}
	if len(gitSyncInitContainer) > 0 {
		if templateSpec.InitContainers == nil {
			templateSpec.InitContainers = make([]corev1.Container, 0, 1)
		}

		templateSpec.InitContainers = append(templateSpec.InitContainers, gitSyncInitContainer...)
	}
	return nil
}

func isServingPortAuthRequired(routingAndAuthorizationFields *field_groups.RoutingAndAuthorizationFields) bool {
	return routingAndAuthorizationFields.ServingPortAuthorizationType != nil &&
		routingAndAuthorizationFields.ServingPortAuthorizationType.Deref() != constants.ServingPortAuthorizationTypePublic
}

func validateAndGetServingPortAuthFields(routingAndAuthorizationFields *field_groups.RoutingAndAuthorizationFields) (bool, *string, *string, error) {
	if routingAndAuthorizationFields == nil {
		return false, nil, nil, nil
	}

	servingPortAuth := routingAndAuthorizationFields.ServingPortAuthorizationType.DerefOrDefault(constants.ServingPortAuthorizationTypeDefault)

	if servingPortAuth != constants.ServingPortAuthorizationTypeAuthorizedUsersOrGroups {
		if routingAndAuthorizationFields.ServingPortAuthorizedUsers.Deref() != "" {
			return false, nil, nil, fmt.Errorf("servingPortAuthorizedUsers is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups")
		}

		if routingAndAuthorizationFields.ServingPortAuthorizedGroups.Deref() != "" {
			return false, nil, nil, fmt.Errorf("servingPortAuthorizedGroups is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups")
		}
	} else {
		if routingAndAuthorizationFields.ServingPortAuthorizedUsers.Deref() == "" && routingAndAuthorizationFields.ServingPortAuthorizedGroups.Deref() == "" {
			return false, nil, nil, fmt.Errorf("servingPortAuthorizedUsers or servingPortAuthorizedGroups must be set when servingPortAuthorizationType is set to authorizedUsersOrGroups")
		}

		if routingAndAuthorizationFields.ServingPortAuthorizedUsers.Deref() != "" && routingAndAuthorizationFields.ServingPortAuthorizedGroups.Deref() != "" {
			return false, nil, nil, fmt.Errorf("servingPortAuthorizedUsers and servingPortAuthorizedGroups are mutually exclusive")
		}
	}

	// for authorized users and groups, it's important to return nil here and not a pointer to an empty string

	var authorizedUsers *string
	if routingAndAuthorizationFields.ServingPortAuthorizedUsers.Deref() != "" {
		authorizedUsers = commonutils.Ref(routingAndAuthorizationFields.ServingPortAuthorizedUsers.Deref())
	}

	var authorizedGroups *string
	if routingAndAuthorizationFields.ServingPortAuthorizedGroups.Deref() != "" {
		authorizedGroups = commonutils.Ref(routingAndAuthorizationFields.ServingPortAuthorizedGroups.Deref())
	}

	return servingPortAuth != constants.ServingPortAuthorizationTypePublic, authorizedUsers, authorizedGroups, nil
}

func getServingPortClusterUrl(ctx context.Context, workloadName string) (string, error) {
	// example: http://inference-0.runai-team-a.svc.cluster.local

	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err != nil {
		return "", err
	}

	servingPortUrl := fmt.Sprintf("http://%s.runai-%s.svc.cluster.local/", workloadName, projectName)

	return servingPortUrl, nil
}

func addContainerToPodSpec(podSpec *corev1.PodSpec, container corev1.Container, featureGateUtil *featuregate.FeatureGateUtil) {
	if featureGateUtil.FeatureGateSidecarContainersEnable {
		podSpec.InitContainers = append(podSpec.InitContainers, container)
	} else {
		podSpec.Containers = append(podSpec.Containers, container)
	}
}

func createAuthUsernameLoaderInitContainer(networkingFields *field_groups.NetworkingFields, servingPortAuthorizedUsers *string, podSpec *corev1.PodSpec) error {
	authUsernameLoaderInitContainer, err := workloadbuilders.GetUsernameLoaderInitContainerWithError(networkingFields.ExposedUrls.Deref(), servingPortAuthorizedUsers)
	if err != nil {
		return err
	}
	if authUsernameLoaderInitContainer != nil {
		if podSpec.InitContainers == nil {
			podSpec.InitContainers = make([]corev1.Container, 0, 1)
		}

		podSpec.InitContainers = append(podSpec.InitContainers, *authUsernameLoaderInitContainer)
	}
	return nil
}

func handleNetworkingFields(
	ctx context.Context, workloadName string, networkingFields *field_groups.NetworkingFields,
	routingAndAuthorizationFields *field_groups.RoutingAndAuthorizationFields,
	podSpec *corev1.PodSpec, workloadKind string, objectNaming *metav1.ObjectMeta,
	requestTimeoutSeconds ...int,
) (servingPort *corev1.ContainerPort, hasServingPortAuth bool, containerWithPortIndex int, err error) {
	if networkingFields == nil {
		return nil, false, 0, nil
	}

	// map hostIPC and hostNetwork to the Knative pod spec
	// requires Knative config-features: kubernetes.podspec-hostipc and kubernetes.podspec-hostnetwork
	podSpec.HostIPC = networkingFields.HostIpc.Deref()
	podSpec.HostNetwork = networkingFields.HostNetwork.Deref()
	if podSpec.HostNetwork {
		podSpec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	// for validation purposes
	_, err = workloadbuilders.GetPorts(ctx, networkingFields, true)
	if err != nil {
		return nil, false, 0, err
	}

	servingPort, err = GetServingPort(ctx, networkingFields, workloadKind)
	if err != nil {
		return nil, false, 0, err
	}

	servingPortAuth, servingPortAuthUsers, servingPortAuthGroups, err := validateAndGetServingPortAuthFields(routingAndAuthorizationFields)
	if err != nil {
		return nil, false, 0, err
	}

	// the serving port should be declared on the user-container, unless it's declared in the auth container
	if !servingPortAuth {
		podSpec.Containers[0].Ports = []corev1.ContainerPort{
			*servingPort,
		}
		containerWithPortIndex = 0
	}

	// Proxy Authorization related

	var exposedUrls map[string]field_types.ExposedUrlItemAndRules
	if workloadconfig.ExternalAuthUrlEnabled() {
		exposedUrls = networkingFields.ExposedUrls.Deref()
	}

	featureGateUtil := featuregate.GetFeatureGateUtilInstance()

	if len(exposedUrls) != 0 || servingPortAuth {
		// add username loader container, if needed
		if err := createAuthUsernameLoaderInitContainer(networkingFields, servingPortAuthUsers, podSpec); err != nil {
			return nil, false, 0, err
		}

		noSuffixUrlGenerated := false

		// handle servingPort
		idx := 0
		if servingPortAuth {
			servingPortUrlStr, err := getServingPortClusterUrl(ctx, workloadName)
			if err != nil {
				return nil, false, 0, err
			}

			servingPortUrl, err := apis.ParseURL(servingPortUrlStr)
			if err != nil {
				return nil, false, 0, err
			}

			cookiePrefix := fmt.Sprintf("%s-%s-%s-%d", workloadKind, objectNaming.Namespace, objectNaming.Name, idx)
			authContainer, err := workloadbuilders.GetOauth2ProxyContainer(servingPortUrl, servingPort.ContainerPort,
				servingPortAuthUsers, servingPortAuthGroups, idx, "serving-port", true, true, cookiePrefix, requestTimeoutSeconds...)
			if err != nil {
				return nil, false, 0, err
			}
			authContainer.Ports[0].Name = servingPort.Name // keep the name, for proper grpc support
			addContainerToPodSpec(podSpec, *authContainer, featureGateUtil)
			// the auth container with the serving port is the last container added
			containerWithPortIndex = len(podSpec.Containers) - 1

			idx += 1
		}

		// handle exposedUrls
		for _, nick := range commonutils.SortedKeys(exposedUrls) {
			exposedUrl := exposedUrls[nick]
			if exposedUrl.Value == nil {
				continue
			}

			jobUrl, err := workloadbuilders.ComposeOauthIngressUrl(ctx, exposedUrl, nick, &noSuffixUrlGenerated, workloadName)
			if err != nil {
				return nil, false, 0, err
			}

			cookiePrefix := fmt.Sprintf("%s-%s-%s-%d", workloadKind, objectNaming.Namespace, objectNaming.Name, idx)
			authContainer, err := workloadbuilders.GetOauth2ProxyContainer(jobUrl, *exposedUrl.Value.ContainerPort,
				exposedUrl.Value.AuthorizedUsers, exposedUrl.Value.AuthorizedGroups, idx, nick, true, false, cookiePrefix)
			if err != nil {
				return nil, false, 0, err
			}
			addContainerToPodSpec(podSpec, *authContainer, featureGateUtil)

			idx += 1
		}
	}

	return servingPort, servingPortAuth, containerWithPortIndex, nil
}

// GetServingPort validates the serving port for inference workloads and returns the serving port
func GetServingPort(_ context.Context, networkingFields *field_groups.NetworkingFields, workloadType string) (*corev1.ContainerPort, error) {
	if workloadType != constants.KindInferenceWorkload {
		return nil, nil
	}

	ports := networkingFields.Ports.Deref()

	servingPort := &corev1.ContainerPort{
		Protocol: corev1.ProtocolTCP,
	}

	for _, nick := range field_types.SortedPortsMapKeys(ports) {
		port := ports[nick].Value
		if port == nil {
			continue
		}

		serviceType := ptr.Deref(ports[nick].Value.ServiceType, networkingFields.ServiceType.Deref())
		if serviceType == constants.ServiceTypeServingPort {
			if servingPort.ContainerPort != 0 {
				return nil, fmt.Errorf("inference workload should have exactly one serving port")
			}
			servingPort.ContainerPort = int32(ptr.Deref(port.ContainerPort, 0))
		} else {
			continue
		}
		// Knative supports grpc protocol by adding a port with name "h2c"
		// See the docs related: https://github.com/meteatamel/knative-tutorial/blob/master/docs/grpc.md
		if ports[nick].Value.Protocol != nil {
			protocol := *ports[nick].Value.Protocol
			if protocol == constants.ProtocolGRPC {
				servingPort.Name = knativeProtocolGRPC
			}
		}

	}

	if servingPort.ContainerPort == 0 {
		return nil, fmt.Errorf("inference workload should have exactly one serving port")
	}

	return servingPort, nil
}

func getConcurrencyHardLimit(concurrencyHardLimit *field_types.IntField) *int64 {
	if concurrencyHardLimit == nil {
		return nil
	}
	return ptr.To(int64(concurrencyHardLimit.Deref()))
}

// BuildNIMMetricsService creates a dedicated Kubernetes Service for NIM metrics scraping
// This service targets the NIM container port (8000) directly, bypassing OAuth2 Proxy
func BuildNIMMetricsService(
	ctx context.Context,
	workload client.Object,
	workloadName string,
	workloadNamespace string,
	servingPort *corev1.ContainerPort,
	k8sClient client.Client,
) (*corev1.Service, error) {
	if servingPort == nil {
		return nil, nil
	}

	serviceName := fmt.Sprintf("%s-metrics", workloadName)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: workloadNamespace,
			Labels: map[string]string{
				constants.LabelApp:                       serviceName,
				constants.LabelNIMWorkload:               constants.BoolTrueAsString,
				constants.LabelWorkloadKind:              constants.KindInferenceWorkload,
				constants.LabelNIMWorkloadMetricsService: constants.BoolTrueAsString,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       servingPort.ContainerPort,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: servingPort.ContainerPort},
					Protocol:   corev1.ProtocolTCP,
				},
			},
			// Selector intentionally targets all pods across all revisions of the workload
			Selector: map[string]string{
				constants.LabelWorkloadName: workloadName,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if workload != nil {
		if err := ctrl.SetControllerReference(workload, service, k8sClient.Scheme()); err != nil {
			return nil, err
		}
	}

	return service, nil
}
