package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	nim "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nim/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NimCacheTopOwnerHandler struct {
	cachedClient client.Client
}

type NimServiceTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewNimCacheTopOwnerHandler(cachedClient client.Client) *NimCacheTopOwnerHandler {
	return &NimCacheTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func NewNimServiceTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *NimServiceTopOwnerHandler {
	return &NimServiceTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *NimServiceTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &nim.NIMService{} })
}

func (h *NimCacheTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &nim.NIMCache{} })
}

func (h *NimCacheTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	marshaled, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed marshaling apps.nvidia.com.NIMCache object: %v", err)
	}
	nimCache := &nim.NIMCache{}
	err = json.Unmarshal(marshaled, nimCache)
	if err != nil {
		return fmt.Errorf("failed unmarshaling apps.nvidia.com.NIMCache object: %v", err)
	}

	// PriorityClassName is not set in NIMCache spec, so we don't set it in ExternalWorkload.

	externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    nimCache.Spec.Resources.CPU,
			corev1.ResourceMemory: nimCache.Spec.Resources.Memory,
		},
	}

	externalWorkload.Status.ObservedStatus.Phase = translateNimCacheState(nimCache.Status.State, nimCache.Status.Conditions)
	externalWorkload.Status.ObservedStatus.K8sStatus = nimCache.Status.State

	commonparsers.SetObservedMetav1Conditions(nimCache.Status.Conditions, &externalWorkload.Status.ObservedStatus)

	return nil
}

func (h *NimServiceTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	marshaled, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed marshaling apps.nvidia.com.NIMService object: %v", err)
	}
	nimService := &nim.NIMService{}
	err = json.Unmarshal(marshaled, nimService)
	if err != nil {
		return fmt.Errorf("failed unmarshaling apps.nvidia.com.NIMService object: %v", err)
	}

	if nimService.Spec.Resources != nil {
		externalWorkload.Spec.RequestedResources = *(nimService.Spec.Resources.DeepCopy())
	} else {
		// reset so re-parsing the same ExternalWorkload stays idempotent across reconciles
		externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{}
	}

	// run:ai requests fractional GPU / GPU memory via annotations rather than container
	// limits. The SDK helper writes into the lists, so they must be non-nil.
	if externalWorkload.Spec.RequestedResources.Requests == nil {
		externalWorkload.Spec.RequestedResources.Requests = corev1.ResourceList{}
	}
	if externalWorkload.Spec.RequestedResources.Limits == nil {
		externalWorkload.Spec.RequestedResources.Limits = corev1.ResourceList{}
	}
	resources.AddRunaiAnnotationsRequestedResources(nimService.Spec.Annotations,
		externalWorkload.Spec.RequestedResources.Requests, externalWorkload.Spec.RequestedResources.Limits)

	for _, draResource := range nimService.Spec.DRAResources {
		gpuCount := h.draParser.ResolveNamedClaimGpuCount(ctx, nimService.Namespace,
			draResource.ResourceClaimName, draResource.ResourceClaimTemplateName)
		dra.AddDRAGpuToResources(gpuCount, &externalWorkload.Spec.RequestedResources)
	}

	// Calculate total pods: when multiNode is set, total = nodesPerReplica * replicas
	// nodesPerReplica is defined by pipeline parallelism (1 leader + N workers)
	replicas := int32(1)
	if nimService.Spec.Replicas != 0 {
		replicas = int32(nimService.Spec.Replicas)
	}
	nodesPerReplica := int32(nimService.GetPipelineParallelism())
	totalPods := replicas * nodesPerReplica

	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(totalPods)
	commonparsers.MultiplyRequestedResourcesPerNumberOfRequestedPods(totalPods, externalWorkload)

	networkingPods := h.countNetworkingPluginPods(ctx, nimService.Namespace, nimService.Name)
	if networkingPods > 0 {
		externalWorkload.Spec.RequestedPods.Replicas = ptr.To(totalPods + networkingPods)
	}

	externalWorkload.Spec.ImageName = strings.Join([]string{nimService.Spec.Image.Repository, nimService.Spec.Image.Tag}, ":")
	externalWorkload.Spec.Urls = h.getNimServiceUrls(ctx, nimService)

	externalWorkload.Status.ObservedStatus.Phase = translateNimServiceState(nimService.Status.State)
	externalWorkload.Status.ObservedStatus.K8sStatus = nimService.Status.State
	externalWorkload.Status.ObservedStatus.Message = calculateNimServicePhaseMessage(nimService.Status.State, nimService.Status.Conditions)

	commonparsers.SetObservedMetav1Conditions(nimService.Status.Conditions, &externalWorkload.Status.ObservedStatus)

	return nil
}

func translateNimCacheState(nimCacheState string, statusConditions []metav1.Condition) v1alpha1.ExternalWorkloadPhase {
	trueConditionsFound := map[string]bool{}
	for _, c := range statusConditions {
		if c.Status == metav1.ConditionTrue &&
			(c.Type == nim.NimCacheConditionJobCompleted || c.Type == nim.NimCacheConditionJobPending || c.Type == nim.NimCacheConditionJobCreated) {
			trueConditionsFound[c.Type] = true
		}
	}

	if _, found := trueConditionsFound[nim.NimCacheConditionJobCompleted]; found {
		return v1alpha1.ExWCompletedPhase
	} else if _, found = trueConditionsFound[nim.NimCacheConditionJobPending]; found {
		return v1alpha1.ExWPendingPhase
	} else if _, found = trueConditionsFound[nim.NimCacheConditionJobCreated]; found {
		return v1alpha1.ExWInitializingPhase
	}

	switch nimCacheState {
	case nim.NimCacheStatusNotReady:
		return v1alpha1.ExWDegradedPhase
	case nim.NimCacheStatusPVCCreated:
		return v1alpha1.ExWInitializingPhase
	case nim.NimCacheStatusStarted:
		return v1alpha1.ExWInitializingPhase
	case nim.NimCacheStatusReady:
		return v1alpha1.ExWCompletedPhase
	case nim.NimCacheStatusInProgress:
		return v1alpha1.ExWRunningPhase
	case nim.NimCacheStatusPending:
		return v1alpha1.ExWPendingPhase
	case nim.NimCacheStatusFailed:
		return v1alpha1.ExWFailedPhase
	default:
		return v1alpha1.ExWUnknownPhase
	}
}

func translateNimServiceState(nimServiceState string) v1alpha1.ExternalWorkloadPhase {
	switch nimServiceState {
	case nim.NIMServiceStatusPending:
		return v1alpha1.ExWPendingPhase
	case nim.NIMServiceStatusNotReady:
		return v1alpha1.ExWInitializingPhase
	case nim.NIMServiceStatusReady:
		return v1alpha1.ExWRunningPhase
	case nim.NIMServiceStatusFailed:
		return v1alpha1.ExWFailedPhase
	default:
		return v1alpha1.ExWUnknownPhase
	}
}

// calculateNimServicePhaseMessage calculates the phase message based on NIMService status and conditions
// It checks for update-in-progress state and extracts replica information from conditions
// It also extracts error messages when the NIMService state is Failed
func calculateNimServicePhaseMessage(nimServiceState string, statusConditions []metav1.Condition) string {
	var readyCondition *metav1.Condition
	var failedConditionMessage string
	var failedConditionReason string

	// Find the Ready condition and any Failed condition message
	for i, condition := range statusConditions {
		if condition.Type == nim.NIMServiceStatusReady {
			readyCondition = &statusConditions[i]
		}
		// If state is Failed, look for conditions with error messages
		// Check both ConditionFalse and any condition with a non-empty message
		if nimServiceState == nim.NIMServiceStatusFailed {
			if condition.Message != "" {
				failedConditionMessage = condition.Message
			}
			// Also capture the reason which may contain useful error info
			if condition.Reason != "" && failedConditionReason == "" {
				failedConditionReason = condition.Reason
			}
		}
	}

	// For Failed state, prioritize returning the error message from conditions
	if nimServiceState == nim.NIMServiceStatusFailed {
		if failedConditionMessage != "" {
			return failedConditionMessage
		}
		// If no message but we have a reason, use it
		if failedConditionReason != "" {
			return fmt.Sprintf("Failed: %s", failedConditionReason)
		}
		// If Ready condition exists, use its message even if state is Failed
		if readyCondition != nil && readyCondition.Message != "" {
			return readyCondition.Message
		}
		// Fallback message for Failed state with no details
		return "NIMService failed - check cluster events for details"
	}

	// If no Ready condition found, return empty message
	if readyCondition == nil || readyCondition.Message == "" {
		return ""
	}

	msg := readyCondition.Message

	// Check if update is in progress by looking for specific patterns in Ready condition message
	// We differentiate between creation and update:
	// - Update: "X out of Y new replicas have been updated..."
	// - Creation: "X of Y updated replicas are available..."
	if readyCondition.Status == metav1.ConditionFalse && strings.Contains(strings.ToLower(msg), "rollout to finish") {
		if strings.Contains(strings.ToLower(msg), "new replicas have been updated") {
			return fmt.Sprintf("Update in progress: %s", msg)
		}
	}

	return msg
}

func (h *NimServiceTopOwnerHandler) getNimServiceUrls(ctx context.Context, nimService *nim.NIMService) []string {
	if nimService == nil {
		return nil
	}

	// we return only one url, and we prioritize the external one, if such
	externalUrl := h.getNimServiceExternalUrl(ctx, nimService)
	if externalUrl != "" {
		return []string{externalUrl}
	}

	internalUrl := h.getNimServiceInternalUrl(ctx, nimService)
	if internalUrl != "" {
		return []string{internalUrl}
	}

	return nil
}

func (h *NimServiceTopOwnerHandler) getNimServiceExternalUrl(ctx context.Context, nimService *nim.NIMService) string {
	ingress := &networkingv1.Ingress{}

	err := h.cachedClient.Get(ctx, client.ObjectKey{
		Name:      nimService.Name,
		Namespace: nimService.Namespace,
	}, ingress)

	if err != nil {
		log.Debug().Msgf("failed to get ingress for NIMService %s/%s: %s", nimService.Namespace, nimService.Name, err.Error())
		return ""
	}

	// Validate ingress spec - must have exactly one rule
	if len(ingress.Spec.Rules) != 1 {
		log.Debug().Msgf("invalid ingress spec for NIMService %s/%s: expected exactly 1 rule, got %d", nimService.Namespace, nimService.Name, len(ingress.Spec.Rules))
		return ""
	}

	rule := ingress.Spec.Rules[0]
	if rule.HTTP == nil || len(rule.HTTP.Paths) == 0 {
		log.Debug().Msgf("invalid ingress spec for NIMService %s/%s: rule has no HTTP paths", nimService.Namespace, nimService.Name)
		return ""
	}

	// Get the first path
	path := rule.HTTP.Paths[0]
	if path.Backend.Service == nil {
		log.Debug().Msgf("invalid ingress spec for NIMService %s/%s: path has no backend service", nimService.Namespace, nimService.Name)
		return ""
	}

	// Strip the regex pattern used for path rewriting before constructing the URL
	// The pattern (/|$)(.*) is added by ingress builders for Nginx path rewriting
	ingressPath := StripPathRewritePattern(path.Path)

	// Construct URL: https://<host>/<path>
	url := fmt.Sprintf("https://%s%s", rule.Host, ingressPath)
	return url
}

func (h *NimServiceTopOwnerHandler) getNimServiceInternalUrl(ctx context.Context, nimService *nim.NIMService) string {
	service := &corev1.Service{}

	err := h.cachedClient.Get(ctx, client.ObjectKey{
		Name:      nimService.Name,
		Namespace: nimService.Namespace,
	}, service)

	if err != nil {
		log.Debug().Msgf("failed to get service for NIMService %s/%s: %s", nimService.Namespace, nimService.Name, err.Error())
		return ""
	}

	// Get the first port from the service
	if len(service.Spec.Ports) == 0 {
		log.Debug().Msgf("service for NIMService %s/%s has no ports defined", nimService.Namespace, nimService.Name)
		return ""
	}

	port := service.Spec.Ports[0].Port

	// Construct URL: http://<svc-name>.<namespace>.svc.cluster.local:<port>
	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		service.Name, service.Namespace, port)
	return url
}

// StripPathRewritePattern removes the Nginx path rewrite regex pattern from a path
// This is used when displaying URLs to users
func StripPathRewritePattern(path string) string {
	if strings.Contains(path, "(/|$)(.*)") {
		i := strings.Index(path, "(/|$)(.*)")
		return path[:i]
	}

	return path
}

func (h *NimServiceTopOwnerHandler) countNetworkingPluginPods(ctx context.Context, namespace, name string) int32 {
	deploymentList := &appsv1.DeploymentList{}
	err := h.cachedClient.List(ctx, deploymentList,
		client.InNamespace(namespace),
		client.MatchingLabels{"run.ai/workload-runner": name})
	if err != nil {
		log.Warn().Msgf("Failed listing networking plugin deployments for %s/%s: %v", namespace, name, err)
		return 0
	}

	var count int32
	for _, dep := range deploymentList.Items {
		if dep.Spec.Replicas != nil {
			count += *dep.Spec.Replicas
		} else {
			count++
		}
	}
	return count
}

func (h *NimServiceTopOwnerHandler) ShouldUpdateNetworkStatus() bool {
	return true
}
