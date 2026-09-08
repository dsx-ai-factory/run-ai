/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"sort"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const dynamoConditionReady = "Ready"

type DynamoGraphDeploymentTopOwnerHandler struct {
	cachedClient   client.Client
	draParser      *dra.DRAParser
	getEmptyObject func() client.Object
}

func NewDynamoGraphDeploymentTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser,
	getEmptyObject func() client.Object) *DynamoGraphDeploymentTopOwnerHandler {
	return &DynamoGraphDeploymentTopOwnerHandler{
		cachedClient:   cachedClient,
		draParser:      draParser,
		getEmptyObject: getEmptyObject,
	}
}

// Parse handles DynamoGraphDeployment served as either nvidia.com/v1alpha1 or the
// preferred nvidia.com/v1beta1 version. The two versions differ structurally
// (v1alpha1 spec.services map vs v1beta1 spec.components list, custom Resources
// vs a standard pod template), so each version is parsed by its own routine
// rather than converting between them.
func (h *DynamoGraphDeploymentTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	dynamoGraphDeploymentAlpha1, ok := object.(*dynamov1alpha1.DynamoGraphDeployment)
	if !ok {
		dynamoGraphDeploymentBeta1, okBeta1 := object.(*dynamov1beta1.DynamoGraphDeployment)
		if !okBeta1 {
			return fmt.Errorf("failed converting object to dynamov1alpha1/v1beta1.DynamoGraphDeployment")
		}

		h.parseSpecV1beta1(ctx, dynamoGraphDeploymentBeta1, externalWorkload)
		fillObservedStatus(string(dynamoGraphDeploymentBeta1.Status.State), dynamoGraphDeploymentBeta1.Status.Conditions, &externalWorkload.Status.ObservedStatus)
		return nil
	}

	h.parseSpecV1alpha1(ctx, dynamoGraphDeploymentAlpha1, externalWorkload)
	fillObservedStatus(string(dynamoGraphDeploymentAlpha1.Status.State), dynamoGraphDeploymentAlpha1.Status.Conditions, &externalWorkload.Status.ObservedStatus)

	return nil
}

func (h *DynamoGraphDeploymentTopOwnerHandler) parseSpecV1alpha1(ctx context.Context, dynamoGraphDeployment *dynamov1alpha1.DynamoGraphDeployment,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	// Calculate total requested resources and replicas from all services
	totalReplicas := int32(0)
	totalRequests := corev1.ResourceList{}
	totalLimits := corev1.ResourceList{}

	for _, serviceSpec := range dynamoGraphDeployment.Spec.Services {
		if serviceSpec == nil {
			continue
		}

		// Get replicas for this service
		replicas := int32(1) // default
		if serviceSpec.Replicas != nil {
			replicas = *serviceSpec.Replicas
		}

		// Multiply by number of nodes if multinode
		numberOfNodes := serviceSpec.GetNumberOfNodes()
		replicas = replicas * numberOfNodes

		totalReplicas += replicas

		// Aggregate resources if available
		h.parseSpecResources(ctx, dynamoGraphDeployment.Namespace, serviceSpec, replicas, totalLimits, totalRequests)

		externalWorkload.Spec.Urls = h.getFrontendUrls(dynamoGraphDeployment)
		// Set image from the first service that has extra pod spec
		if externalWorkload.Spec.ImageName == "" && serviceSpec.ExtraPodSpec != nil {
			if serviceSpec.ExtraPodSpec.MainContainer != nil {
				externalWorkload.Spec.ImageName = serviceSpec.ExtraPodSpec.MainContainer.Image
			}
		}
	}

	// Set total replicas
	if totalReplicas > 0 {
		externalWorkload.Spec.RequestedPods.Replicas = ptr.To(totalReplicas)
	}

	// Set aggregated resources
	if len(totalRequests) > 0 || len(totalLimits) > 0 {
		externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{
			Requests: totalRequests,
			Limits:   totalLimits,
		}
	}

	// Set default node pools
	commonparsers.FillDefaultNodePoolIfNeeded(externalWorkload)

	h.parseTopology(dynamoGraphDeployment, externalWorkload)

	if externalWorkload.Spec.PriorityClassName == "" || externalWorkload.Spec.Preemptibility == "" {
		priorityClassName, preemptibility := commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, dynamoGraphDeployment.Labels,
			dynamoGraphDeployment.GroupVersionKind().GroupKind(), "")
		if externalWorkload.Spec.PriorityClassName == "" {
			externalWorkload.Spec.PriorityClassName = priorityClassName
		}
		if externalWorkload.Spec.Preemptibility == "" {
			externalWorkload.Spec.Preemptibility = preemptibility
		}
	}
}

// parseSpecV1beta1 parses a v1beta1 DGD, whose spec.components is a list and
// whose per-component pod spec is a standard corev1.PodTemplateSpec (unlike
// v1alpha1's spec.services map with custom Resources/ExtraPodSpec). Resources,
// image, and node affinity are derived from each component's pod template, which
// lets us reuse the shared pod-template parsers. v1beta1 components have no
// ingress field, so no frontend URL is discovered here (tracked separately).
func (h *DynamoGraphDeploymentTopOwnerHandler) parseSpecV1beta1(ctx context.Context, dynamoGraphDeployment *dynamov1beta1.DynamoGraphDeployment,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	totalReplicas := int32(0)
	// Accumulate into a local value and assign once, so re-parsing an existing
	// ExternalWorkload (on reconcile) recomputes resources rather than summing
	// onto the previous result.
	totalResources := corev1.ResourceRequirements{}

	for i := range dynamoGraphDeployment.Spec.Components {
		component := &dynamoGraphDeployment.Spec.Components[i]

		replicas := int32(1) // default
		if component.Replicas != nil {
			replicas = *component.Replicas
		}
		// Multiply by number of nodes if multinode
		replicas = replicas * component.GetNumberOfNodes()
		totalReplicas += replicas

		if component.PodTemplate == nil {
			continue
		}

		// Aggregate requested resources from the pod template, scaled by replicas.
		podResources := commonparsers.GetPodRequestedResources(component.PodTemplate)
		totalResources = resources.SumResourceRequirements(
			totalResources, commonparsers.MultiplyResources(replicas, podResources))

		// DRA GPUs declared via the pod template's resource claims.
		if h.draParser != nil && len(component.PodTemplate.Spec.ResourceClaims) > 0 {
			gpuCount := h.draParser.ResolveDRAGpuCount(ctx, dynamoGraphDeployment.Namespace,
				component.PodTemplate.Spec.ResourceClaims)
			if gpuCount > 0 {
				dra.AddDRAGpuToResources(gpuCount*int64(replicas), &totalResources)
			}
		}

		if externalWorkload.Spec.ImageName == "" {
			commonparsers.CalculateImageInfo(component.PodTemplate, externalWorkload)
		}

		// Capture each component's nodeAffinity under its component name so each
		// component is distinguishable in the EW spec and downstream API.
		commonparsers.UpsertNodeAffinityFromPodTemplate(component.ComponentName, component.PodTemplate, externalWorkload)

		commonparsers.FillRequestedNodePools(ctx, dynamoGraphDeployment.ObjectMeta, dynamoGraphDeployment.GroupVersionKind().GroupKind().Kind, component.PodTemplate.Spec.Affinity,
			h.cachedClient, externalWorkload)
	}

	if len(totalResources.Requests) > 0 || len(totalResources.Limits) > 0 {
		externalWorkload.Spec.RequestedResources = totalResources
	}

	// Set total replicas
	if totalReplicas > 0 {
		externalWorkload.Spec.RequestedPods.Replicas = ptr.To(totalReplicas)
	}

	// Set default node pools
	commonparsers.FillDefaultNodePoolIfNeeded(externalWorkload)

	h.parseTopology(dynamoGraphDeployment, externalWorkload)

	if externalWorkload.Spec.PriorityClassName == "" || externalWorkload.Spec.Preemptibility == "" {
		priorityClassName, preemptibility := commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, dynamoGraphDeployment.Labels,
			dynamoGraphDeployment.GroupVersionKind().GroupKind(), "")
		if externalWorkload.Spec.PriorityClassName == "" {
			externalWorkload.Spec.PriorityClassName = priorityClassName
		}
		if externalWorkload.Spec.Preemptibility == "" {
			externalWorkload.Spec.Preemptibility = preemptibility
		}
	}
}

func buildUrlFromIngress(ingress *dynamov1alpha1.IngressSpec) string {
	if ingress == nil {
		return ""
	}
	prefix := ptr.Deref(ingress.HostPrefix, "")
	suffix := ptr.Deref(ingress.HostSuffix, "")
	return fmt.Sprintf("https://%s%s.%s", prefix, ingress.Host, suffix)
}

func (h *DynamoGraphDeploymentTopOwnerHandler) getFrontendUrls(dynamoGraphDeployment *dynamov1alpha1.DynamoGraphDeployment) []string {
	url := ""
	for _, service := range dynamoGraphDeployment.Spec.Services {
		if service.ComponentType == constants.ComponentTypeFrontend {
			url = buildUrlFromIngress(service.Ingress)
			break
		}
	}
	if url == "" {
		return nil
	}
	return []string{url}
}

func (h *DynamoGraphDeploymentTopOwnerHandler) parseSpecResources(ctx context.Context, namespace string, serviceSpec *dynamov1alpha1.DynamoComponentDeploymentSharedSpec, replicas int32, totalLimits, totalRequests corev1.ResourceList) {
	if serviceSpec.Resources != nil {
		// Handle GPU requests
		if serviceSpec.Resources.Requests != nil && serviceSpec.Resources.Requests.GPU != "" {
			quantityValue := serviceSpec.Resources.Requests.GPU
			var resourceName corev1.ResourceName = constants.GpuResourceName
			err := tryAddQuantity(&totalLimits, quantityValue, resourceName, replicas)
			if err != nil {
				log.Warn().Msgf("failed to add quantity <%s> to total requests for resource <%s>: %v", quantityValue, resourceName, err)
			}
		}

		// Handle GPU limits
		if serviceSpec.Resources.Limits != nil && serviceSpec.Resources.Limits.GPU != "" {
			quantityValue := serviceSpec.Resources.Limits.GPU
			var resourceName corev1.ResourceName = constants.GpuResourceName
			err := tryAddQuantity(&totalLimits, quantityValue, resourceName, replicas)
			if err != nil {
				log.Warn().Msgf("failed to add quantity <%s> to total requests for resource <%s>: %v", quantityValue, resourceName, err)
			}
		}

		// Handle CPU requests
		if serviceSpec.Resources.Requests != nil && serviceSpec.Resources.Requests.CPU != "" {
			quantityValue := serviceSpec.Resources.Requests.CPU
			resourceName := corev1.ResourceCPU
			err := tryAddQuantity(&totalRequests, quantityValue, resourceName, replicas)
			if err != nil {
				log.Warn().Msgf("failed to add quantity <%s> to total requests for resource <%s>: %v", quantityValue, resourceName, err)
			}
		}

		// Handle Memory requests
		if serviceSpec.Resources.Requests != nil && serviceSpec.Resources.Requests.Memory != "" {
			quantityValue := serviceSpec.Resources.Requests.Memory
			resourceName := corev1.ResourceMemory
			err := tryAddQuantity(&totalRequests, quantityValue, resourceName, replicas)
			if err != nil {
				log.Warn().Msgf("failed to add quantity <%s> to total requests for resource <%s>: %v", quantityValue, resourceName, err)
			}
		}
	}

	// Handle DRA GPU resources from ExtraPodSpec
	if serviceSpec.ExtraPodSpec != nil && serviceSpec.ExtraPodSpec.PodSpec != nil &&
		len(serviceSpec.ExtraPodSpec.PodSpec.ResourceClaims) > 0 {
		gpuCount := h.draParser.ResolveDRAGpuCount(ctx, namespace,
			serviceSpec.ExtraPodSpec.PodSpec.ResourceClaims)
		if gpuCount > 0 {
			draGpuQty := resource.NewQuantity(gpuCount*int64(replicas), resource.DecimalSI)
			utils.AddToResourceList(totalLimits, constants.GpuResourceName, *draGpuQty)
			utils.AddToResourceList(totalRequests, constants.GpuResourceName, *draGpuQty)
		}
	}
}

// parseTopology surfaces the Dynamo topology constraints at both levels: the
// deployment-level constraint onto the workload, and each component's constraint
// onto Spec.ComponentTopologies, keyed by the component name so the control plane
// can attach it to the matching workload element.
//
// The field names are version-specific (v1alpha1 topologyProfile + a services map
// vs v1beta1 clusterTopologyName + a components list), so each version is read
// separately. A component without its own packDomain inherits the deployment-level
// one. packDomain maps to RequiredPlacement: Grove treats it as the deprecated
// spelling of pack.required, a hard placement requirement.
func (h *DynamoGraphDeploymentTopOwnerHandler) parseTopology(
	object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	switch dgd := object.(type) {
	case *dynamov1alpha1.DynamoGraphDeployment:
		if dgd.Spec.TopologyConstraint != nil && dgd.Spec.TopologyConstraint.TopologyProfile != "" {
			topologyName := dgd.Spec.TopologyConstraint.TopologyProfile
			defaultPackDomain := string(dgd.Spec.TopologyConstraint.PackDomain)
			externalWorkload.Spec.Topology.Name = topologyName
			externalWorkload.Spec.Topology.RequiredPlacement = defaultPackDomain

			packDomainByComponent := make(map[string]string, len(dgd.Spec.Services))
			for serviceName, serviceSpec := range dgd.Spec.Services {
				packDomain := defaultPackDomain
				if serviceSpec != nil && serviceSpec.TopologyConstraint != nil {
					packDomain = string(serviceSpec.TopologyConstraint.PackDomain)
				}
				packDomainByComponent[serviceName] = packDomain
			}
			externalWorkload.Spec.ComponentTopologies = buildComponentTopologies(topologyName, packDomainByComponent)
			return
		}
		// Legacy path: KAI topology annotations on the per-service specs
		serviceAnnotations := make([]map[string]string, 0, len(dgd.Spec.Services))
		for _, serviceSpec := range dgd.Spec.Services {
			if serviceSpec != nil {
				serviceAnnotations = append(serviceAnnotations, serviceSpec.Annotations)
			}
		}
		parseTopologyFromServiceAnnotations(serviceAnnotations, externalWorkload)
	case *dynamov1beta1.DynamoGraphDeployment:
		if dgd.Spec.TopologyConstraint != nil && dgd.Spec.TopologyConstraint.ClusterTopologyName != "" {
			topologyName := dgd.Spec.TopologyConstraint.ClusterTopologyName
			defaultPackDomain := string(dgd.Spec.TopologyConstraint.PackDomain)
			externalWorkload.Spec.Topology.Name = topologyName
			externalWorkload.Spec.Topology.RequiredPlacement = defaultPackDomain

			packDomainByComponent := make(map[string]string, len(dgd.Spec.Components))
			for _, component := range dgd.Spec.Components {
				packDomain := defaultPackDomain
				if component.TopologyConstraint != nil {
					packDomain = string(component.TopologyConstraint.PackDomain)
				}
				packDomainByComponent[component.ComponentName] = packDomain
			}
			externalWorkload.Spec.ComponentTopologies = buildComponentTopologies(topologyName, packDomainByComponent)
			return
		}
		// Legacy path: KAI topology annotations. v1beta1 has no per-component
		// annotations, so the deployment-level annotations are the only source.
		parseTopologyFromServiceAnnotations([]map[string]string{dgd.Spec.Annotations}, externalWorkload)
	}
}

// buildComponentTopologies turns a component name to pack domain mapping into
// ComponentTopologyEntry values, skipping components that resolved to no pack
// domain. Entries are sorted by name so that a re-parse of an unchanged
// DynamoGraphDeployment produces an identical spec.
func buildComponentTopologies(topologyName string, packDomainByComponent map[string]string) []v1alpha1.ComponentTopologyEntry {
	componentNames := make([]string, 0, len(packDomainByComponent))
	for componentName, packDomain := range packDomainByComponent {
		if componentName != "" && packDomain != "" {
			componentNames = append(componentNames, componentName)
		}
	}
	if len(componentNames) == 0 {
		return nil
	}
	sort.Strings(componentNames)

	entries := make([]v1alpha1.ComponentTopologyEntry, 0, len(componentNames))
	for _, componentName := range componentNames {
		entries = append(entries, v1alpha1.ComponentTopologyEntry{
			SpecName: componentName,
			Topology: v1alpha1.Topology{
				Name:              topologyName,
				RequiredPlacement: packDomainByComponent[componentName],
			},
		})
	}
	return entries
}

// parseTopologyFromServiceAnnotations reads the legacy KAI topology annotations
// from the given annotation maps, using the first map that carries any of them.
func parseTopologyFromServiceAnnotations(annotationMaps []map[string]string,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	for _, annotations := range annotationMaps {
		if annotations == nil {
			continue
		}
		name := annotations[clustercommonconstants.TopologyAnnotation]
		preferred := annotations[clustercommonconstants.TopologyPreferredPlacementAnnotation]
		required := annotations[clustercommonconstants.TopologyRequiredPlacementAnnotation]
		if name != "" || preferred != "" || required != "" {
			externalWorkload.Spec.Topology.Name = name
			externalWorkload.Spec.Topology.PreferredPlacement = preferred
			externalWorkload.Spec.Topology.RequiredPlacement = required
			return
		}
	}
}

// fillObservedStatus maps a DGD's high-level state and conditions onto the
// ExternalWorkload observed status. The DGD state strings ("pending",
// "successful", "failed") and the conditions slice ([]metav1.Condition) are
// identical across v1alpha1 and v1beta1, so a single routine serves both.
func fillObservedStatus(state string, conditions []metav1.Condition,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus,
) {
	// Set default status
	phase := v1alpha1.ExWUnknownPhase
	observedStatus.K8sStatus = state
	observedStatus.Message = ""

	if state == "" && len(conditions) == 0 {
		phase = v1alpha1.ExWCreatingPhase
	}
	// Convert conditions
	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               condition.Type,
			Status:             corev1.ConditionStatus(condition.Status),
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})

		if state == string(dynamov1alpha1.DGDStatePending) &&
			condition.Type == dynamoConditionReady && condition.Status == metav1.ConditionFalse &&
			condition.Reason == string(constants.SomeResourcesNotReady) {
			phase = v1alpha1.ExWPendingPhase
			observedStatus.Message = condition.Message
		} else if state == string(dynamov1alpha1.DGDStateSuccessful) &&
			condition.Type == dynamoConditionReady && condition.Status == metav1.ConditionTrue &&
			condition.Reason == string(constants.AllResourcesAreReady) {
			phase = v1alpha1.ExWRunningPhase
			observedStatus.Message = condition.Message
		} else if state == string(dynamov1alpha1.DGDStateFailed) &&
			condition.Type == dynamoConditionReady && condition.Status == metav1.ConditionFalse &&
			(condition.Reason == string(constants.ValidationFailed)) {
			phase = v1alpha1.ExWFailedPhase
			observedStatus.Message = condition.Message
		}
	}

	observedStatus.Phase = phase
	observedStatus.Conditions = observedConditions
}

func (h *DynamoGraphDeploymentTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, h.getEmptyObject)
}

func tryAddQuantity(resourcesList *corev1.ResourceList, quantityValue string, resourceName corev1.ResourceName, replicas int32) error {
	quantity, err := resource.ParseQuantity(quantityValue)
	if err != nil {
		return err
	}
	// Multiply by replicas
	list := *resourcesList
	quantity.Mul(int64(replicas))

	if existing, found := list[resourceName]; found {
		existing.Add(quantity)
		list[resourceName] = existing
	} else {
		list[resourceName] = quantity
	}

	return nil
}
