package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	nvcfv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/priority"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NvcfFunctionVersionIdLabel     = "function-version-id"
	NvcfFunctionType               = "NVCF Function"
	NvcfDefaultFunctionVersionName = "fv"

	NvcfEnvVarNameFunctionId              = "FUNCTION_ID"
	NvcfEnvVarNameFunctionVersionId       = "FUNCTION_VERSION_ID"
	NvcfEnvVarNameFunctionName            = "FUNCTION_NAME"
	NvcfEnvVarNameInferenceContainerImage = "INFERENCE_CONTAINER"
	NvcfEnvVarNameFqdn                    = "NVCF_FQDN"
	NvcfEnvVarNameFqdnGrpc                = "NVCF_FQDN_GRPC"
	NvcfEnvVarNameProtocol                = "INFERENCE_PROTOCOL"
	NvcfEnvVarNameProtocolHttp            = "REST"
)

type NVCFTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewNVCFTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *NVCFTopOwnerHandler {
	return &NVCFTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *NVCFTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	spotRequest, ok := object.(*nvcfv1.SPOTRequest)
	if !ok {
		return fmt.Errorf("failed converting object to nvcfv1.SPOTRequest")
	}

	functionVersionId := spotRequest.Spec.FunctionVersionID

	// get all the pods of this function version
	pods, err := getPodsOfFunctionVersion(ctx, functionVersionId, spotRequest.Namespace, h.cachedClient)
	if err != nil {
		return fmt.Errorf("failed getting pods of function version %s: %v", functionVersionId, err)
	}

	numReplicas := int32(len(pods))

	if numReplicas == 0 {
		return fmt.Errorf("no pods found for function version %s", functionVersionId)
	}

	// set the type
	externalWorkload.Spec.Type = NvcfFunctionType

	// fill the external-workload using the data from one of the pods
	pod := pods[0]
	podTemplate := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: pod.Annotations,
			Labels:      pod.Labels,
		},
		Spec: pod.Spec,
	}

	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, spotRequest.ObjectMeta, spotRequest.GroupVersionKind().GroupKind(),
		podTemplate, nil, h.cachedClient, h.draParser, externalWorkload)

	// the default priority class is inference (non-preemptible)
	if externalWorkload.Spec.PriorityClassName == "" {
		externalWorkload.Spec.PriorityClassName = priority.InferencePriorityClassName
	}

	commonparsers.MultiplyRequestedResourcesPerNumberOfRequestedPods(numReplicas, externalWorkload)

	// set the image name to the correct one. the above common functions don't extract the image name properly since
	// there are multiple containers. we take the image name from an environment variable of the init container,
	// instead of from the inference container's spec
	externalWorkload.Spec.ImageName = getEnvironmentVariableFromInitContainer(&pod, NvcfEnvVarNameInferenceContainerImage)

	// we only have the number of replicas, but not the min/max replicas configured
	// for the function version deployment
	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(numReplicas)

	// set the URL
	externalWorkload.Spec.Urls = getUrlsFromPod(&pod)

	// set the additional fields
	additionalFields := getAdditionalFieldsFromPod(&pod)
	externalWorkload.Spec.AdditionalFields = &additionalFields

	// calculate the external-workload status using the status of all the pods
	externalWorkload.Status.ObservedStatus.Phase = calculateFunctionPhaseFromPods(pods)

	for _, pod = range pods {
		err = h.addTopOwnerWorkloadIdLabel(ctx, &pod, functionVersionId)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetOriginalWorkloadNameUid - since we aggregate multiple SPOTRequest instances into one nvcf function, i.e.
// one external workload, we cannot use any of the individual SPOTRequests's uid/name for the external workload uid/name.
// instead, we want to use the function version id and name
func (h *NVCFTopOwnerHandler) GetOriginalWorkloadNameUid(object client.Object) (types.UID, string, error) {
	// this function is sometimes called with a metav1.PartialObjectMetadata, which is not a SPOTRequest.
	// in this case, we want to get the SPOTRequest object and use it.
	// so we first try to convert the given object to a SPOTRequest, and if it fails, we try to get the SPOTRequest
	var spotRequest nvcfv1.SPOTRequest
	spotRequestPtr, ok := object.(*nvcfv1.SPOTRequest)
	if !ok {
		spotRequestPtr = nil
		metadata, ok := object.(*metav1.PartialObjectMetadata)
		if ok {
			err := h.cachedClient.Get(context.Background(), types.NamespacedName{Namespace: metadata.GetNamespace(), Name: metadata.GetName()}, &spotRequest)
			if err == nil {
				spotRequestPtr = &spotRequest
			}
		}

		if spotRequestPtr == nil {
			return "", "", fmt.Errorf("failed converting object to nvcfv1.SPOTRequest")
		}
	}

	// get the function version name from the pod, from the env var FUNCTION_NAME (init container)
	pods, err := getPodsOfFunctionVersion(context.Background(), spotRequestPtr.Spec.FunctionVersionID, spotRequestPtr.Namespace, h.cachedClient)
	if err != nil {
		return "", "", fmt.Errorf("error getting pods of function version %s: %v", spotRequestPtr.Spec.FunctionVersionID, err)
	}
	if len(pods) == 0 {
		return "", "", fmt.Errorf("no pods found for function version %s", spotRequestPtr.Spec.FunctionVersionID)
	}

	functionVersionName := getEnvironmentVariableFromInitContainer(&pods[0], NvcfEnvVarNameFunctionName)
	if functionVersionName == "" { // not expected to happen
		functionVersionName = NvcfDefaultFunctionVersionName
	}

	// using a combination of the function version name and id, since the function version name is not unique
	workloadName := fmt.Sprintf("%s-%s", functionVersionName, spotRequestPtr.Spec.FunctionVersionID)

	return types.UID(spotRequestPtr.Spec.FunctionVersionID), workloadName, nil
}

// ShouldTriggerTopOwnerHandlerAsWell - since we look at pods to calculate status of the nvcf function,
// in case of a pod event, we need to trigger the nvcf handler as well
func (h *NVCFTopOwnerHandler) ShouldTriggerTopOwnerHandlerAsWell(triggeredOwnedObject client.Object) bool {
	switch triggeredOwnedObject.(type) {
	case *corev1.Pod:
		return true
	}
	return false
}

func (h *NVCFTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &nvcfv1.SPOTRequest{} })
}

func (h *NVCFTopOwnerHandler) SecondaryFilter(object client.Object) bool {
	// filter out SPOTRequests that don't have any pods. we do not want to create external workloads for them,
	// as it's a temporary state and there's essential info we're using for the external workload that only
	// exists in the pods

	spotRequest, ok := object.(*nvcfv1.SPOTRequest)
	if !ok {
		return false
	}

	pods, err := getPodsOfFunctionVersion(context.Background(), spotRequest.Spec.FunctionVersionID, spotRequest.Namespace, h.cachedClient)
	if err != nil {
		return false
	}

	return len(pods) > 0
}

func getPodsOfFunctionVersionLabelSelector(functionVersionId, namespace string) *client.ListOptions {
	labelSelector := client.MatchingLabels{
		NvcfFunctionVersionIdLabel: functionVersionId,
	}
	listOpts := &client.ListOptions{Namespace: namespace}
	labelSelector.ApplyToList(listOpts)
	return listOpts
}

func getPodsOfFunctionVersion(ctx context.Context, functionVersionId, namespace string, readerClient client.Client) ([]corev1.Pod, error) {
	if functionVersionId == "" {
		return []corev1.Pod{}, nil
	}

	selector := getPodsOfFunctionVersionLabelSelector(functionVersionId, namespace)
	podList := &corev1.PodList{}
	err := readerClient.List(ctx, podList, selector)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return []corev1.Pod{}, nil
		}
		return nil, err
	}

	return podList.Items, nil
}

func getEnvironmentVariableFromInitContainer(pod *corev1.Pod, envVarName string) string {
	if pod == nil || len(pod.Spec.InitContainers) == 0 {
		return ""
	}

	for _, e := range pod.Spec.InitContainers[0].Env {
		if e.Name == envVarName {
			return e.Value
		}
	}

	return ""
}

func getUrlsFromPod(pod *corev1.Pod) []string {
	protocol := getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameProtocol)
	if protocol == "" {
		return nil
	}

	fqdnEnvVarName := NvcfEnvVarNameFqdn
	if protocol != NvcfEnvVarNameProtocolHttp {
		fqdnEnvVarName = NvcfEnvVarNameFqdnGrpc
	}

	baseUrl := getEnvironmentVariableFromInitContainer(pod, fqdnEnvVarName)
	functionId := getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameFunctionId)
	functionVersionId := getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameFunctionVersionId)

	url := fmt.Sprintf("%s/v2/nvcf/pexec/functions/%s/versions/%s", baseUrl, functionId, functionVersionId)

	return []string{url}
}

func getAdditionalFieldsFromPod(pod *corev1.Pod) string {
	data := map[string]interface{}{}

	data["function-id"] = getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameFunctionId)
	data["function-version-id"] = getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameFunctionVersionId)
	data["function-version-name"] = getEnvironmentVariableFromInitContainer(pod, NvcfEnvVarNameFunctionName)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

func calculateFunctionPhaseFromPods(pods []corev1.Pod) v1alpha1.ExternalWorkloadPhase {
	// if at least one pod is ready, the function is considered running
	// otherwise, if at least one pod is running, the function is considered initializing
	// otherwise, if at least one pod is pending, the function is considered pending
	// otherwise, if at least one pod is failed, the function is considered failed
	// otherwise, the function is considered unknown

	// look for a ready pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return v1alpha1.ExWRunningPhase
				}
			}
		}
	}

	// look for a running pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			return v1alpha1.ExWInitializingPhase
		}
	}

	// look for a pending pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			return v1alpha1.ExWPendingPhase
		}
	}

	// look for a failed pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodFailed {
			return v1alpha1.ExWFailedPhase
		}
	}

	return v1alpha1.ExWUnknownPhase
}

// addTopOwnerWorkloadIdLabel - add the top owner workload id label to the pod if it's not already set.
// this is required for cluster-sync to know which external workload the pod belongs to.
// in our case, the default top owner id is not good, we need to set the function version id as
// the top owner id (see GetOriginalWorkloadNameUid)
func (h *NVCFTopOwnerHandler) addTopOwnerWorkloadIdLabel(ctx context.Context, pod *corev1.Pod,
	topOwnerWorkloadId string) error {
	// skip only if the label is already set, with the correct value
	topOwnerWorkloadIdLabelValue, topOwnerWorkloadIdFound := pod.Labels[commonconstants.TopOwnerWorkloadIdLabel]
	topOwnerUid, topOwnerUidFound := pod.Labels[commonconstants.TopOwnerUidLabel]
	if topOwnerWorkloadIdFound && topOwnerWorkloadIdLabelValue == topOwnerWorkloadId &&
		topOwnerUidFound && topOwnerUid == topOwnerWorkloadId {
		return nil
	}

	// NOTE: set deprecated as well, until we fully move to the new label
	patchBytes, err := utils.AddLabelsPatchBytes(map[string]string{
		commonconstants.TopOwnerWorkloadIdLabel: topOwnerWorkloadId,
		commonconstants.TopOwnerUidLabel:        topOwnerWorkloadId,
	})
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - add top owner label <%s> on pod <%s/%s>, error: %s",
			topOwnerWorkloadId, pod.Namespace, pod.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Patch(ctx, pod, patch)
	if err != nil {
		log.Error().Msgf("Failed to patch pod <%s/%s> with top owner label <%s>, error: %s",
			pod.Namespace, pod.Name, topOwnerWorkloadId, err.Error())
		return err
	}
	log.Info().Msgf("Patched pod <%s/%s> with top owner label <%s>",
		pod.Namespace, pod.Name, topOwnerWorkloadId)

	return nil
}

// GetOwnerReferences - since we aggregate multiple SPOTRequest instances into one nvcf function, i.e.
// one external workload, we want all the SPOTRequests to be the owner of the external workload,
// such that only once all of them are deleted, the external workload will be deleted
func (h *NVCFTopOwnerHandler) GetOwnerReferences(object client.Object) []metav1.OwnerReference {
	spotRequest, ok := object.(*nvcfv1.SPOTRequest)
	if !ok {
		return nil
	}

	// list all pods with the same function version id
	pods, err := getPodsOfFunctionVersion(context.Background(), spotRequest.Spec.FunctionVersionID, spotRequest.Namespace, h.cachedClient)
	if err != nil {
		return nil
	}

	// use the owner of each pod (SPOTRequest) as the owner of the external workload
	var ownerReferences []metav1.OwnerReference
	for _, pod := range pods {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == spotRequest.Kind { // just in case
				owner.Controller = nil
				owner.BlockOwnerDeletion = nil
				ownerReferences = append(ownerReferences, owner)
			}
		}
	}

	return ownerReferences
}

func (h *NVCFTopOwnerHandler) ShouldUpdateNetworkStatus() bool {
	return true
}
