/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"k8s.io/utils/ptr"

	"go.uber.org/zap"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	sdkkaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NodeNotReadyConditionType     = "NodeNotReady"
	SchedulingDisabled            = "SchedulingDisabled"
	UndrainedMigratedNode         = "UndrainedMigratedNode"
	MissingNvidiaContainerToolkit = "MissingNvidiaContainerToolkit"
	MissingNvidiaDcgmExporter     = "MissingNvidiaDcgmExporter"
	MissingNrtHealthyPrerequisite = "MissingNrtHealthyPrerequisite"

	// The ConfigMap has no ResourceVersion when it cannot be read, so fail-open results are
	// cached under a version no real object can have.
	unreadableClusterInfoVersion = "\x00unreadable"

	gpuStackDependencyName      = "gpu-stack"
	gpuFractioningComponentName = "gpu-fractioning"

	// Operator-published readiness: True means healthy, the inverse of the faults above.
	GpuFractioningReadyConditionType  = "gpu-fractioning.nvidia.com/Ready"
	MissingGpuFractioningPrerequisite = "MissingGpuFractioningPrerequisite"
	GpuFractioningNotManaged          = "NotManaged"

	gpuFractioningNotManagedMessage = "GPU fractions are not supported on this node - " +
		"it is not covered by the GPU fractioning operator."

	ReadyPhase                = "Ready"
	NotReadyPhase             = "NotReady"
	MissingPrerequisitesPhase = "MissingPrerequisites"

	NvidiaGpuCountLabelKey         = "nvidia.com/gpu.count"
	NvidiaGpuTypeLabelKey          = "nvidia.com/gpu.product"
	NvidiaGpuMemoryLabelKey        = "nvidia.com/gpu.memory"
	AmdGpuTypeLabelPrefix          = "beta.amd.com/gpu.family."
	GPUNetworkAccelerationLabelKey = "gpuNetworkAccelerationLabelKey"
	NvidiaGpuCliqueLabel           = "nvidia.com/gpu.clique"

	bytesInMib = 1024 * 1024
	// Some gpu-feature-discovery versions report nvidia.com/gpu.memory in bytes rather than
	// MiB (https://github.com/NVIDIA/gpu-feature-discovery/issues/26). No single device has
	// more than 1 TiB of memory, so anything above that is really bytes. The boundary itself
	// counts as MiB: a 1 TiB device is conceivable, a 1 MiB one is not.
	maxPlausibleGpuMemoryMib = 1024 * 1024
)

var (
	nvidiaComponentNameToConditionType = map[string]string{
		"container-toolkit": MissingNvidiaContainerToolkit,
		"dcgm-exporter":     MissingNvidiaDcgmExporter,
	}

	// https://github.com/RadeonOpenCompute/k8s-device-plugin/tree/master/cmd/k8s-node-labeller#deployment
	amdGpuFamilies = map[string]string{
		"SI": "Southern Islands",
		"CI": "Sea Islands",
		"KV": "Kaveri",
		"VI": "Volcanic Islands",
		"CZ": "Carrizo",
		"AI": "Arctic Islands",
		"RV": "Raven",
		"NV": "Navi",
	}

	knownNotReadyConditionTypes = map[string]bool{
		NodeNotReadyConditionType:     true,
		"MemoryPressure":              true,
		"DiskPressure":                true,
		"PIDPressure":                 true,
		"NetworkUnavailable":          true,
		SchedulingDisabled:            true,
		UndrainedMigratedNode:         true,
		MissingNvidiaContainerToolkit: true,
		MissingNvidiaDcgmExporter:     true,
	}

	knownMissingPrerequisiteConditionTypes = map[string]bool{
		MissingNrtHealthyPrerequisite:     true,
		MissingGpuFractioningPrerequisite: true,
	}
)

type NodeParser struct {
	cachedReader client.Reader

	nodeLabelNvidiaContainerToolkit string
	nodeLabelNvidiaDcgmExporter     string

	gpuFractioningSupport lastGpuFractioningSupport
}

// Holds the verdict the parser last acted on. Not a read cache — the ConfigMap is re-read and
// re-parsed per call like every other parser here — this only exists so a flip can be detected,
// which is otherwise invisible because the Node objects do not change when the verdict does.
type lastGpuFractioningSupport struct {
	mutex     sync.RWMutex
	supported bool
	loaded    bool
}

func (c *lastGpuFractioningSupport) last() (bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.supported, c.loaded
}

// store reports whether this verdict differs from the one already held, so callers can act —
// or log — only on a real change.
func (c *lastGpuFractioningSupport) store(supported bool) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	changed := !c.loaded || c.supported != supported
	c.supported = supported
	c.loaded = true
	return changed
}

func NewNodeParser(cachedReader client.Reader) *NodeParser {
	return &NodeParser{
		cachedReader:                    cachedReader,
		nodeLabelNvidiaContainerToolkit: config.Config().NodeLabelNvidiaContainerToolkit,
		nodeLabelNvidiaDcgmExporter:     config.Config().NodeLabelNvidiaDcgmExporter,
	}
}

// Value of label GPUNetworkAccelerationLabelKey should contain another label name - the domainClique label.
// domainClique label value should be structured as following <DOMAIN_UID>.<CLIQUE_ID>
// if value of domainClique label is correct returns parsed values, otherwise empty values.
// if label key is different from NvidiaGpuCliqueLabel - set clique ID as empty value
func tryGetNVLinkValues(node *corev1.Node) (domainUid *string, cliqueId *string) {
	labelKey, exists := node.Annotations[GPUNetworkAccelerationLabelKey]
	if !exists || labelKey == "" {
		return nil, nil
	}
	value := node.Labels[labelKey]
	if value == "" {
		return
	}
	values := strings.Split(value, ".")
	isValid := len(values) == 2 && len(values[0]) > 0 && len(values[1]) > 0
	// if parsing value not valid - set only domainUid with full value
	if !isValid || NvidiaGpuCliqueLabel != labelKey {
		domainUid = ptr.To(value)
		cliqueId = nil
		return
	} else {
		domainUid = ptr.To(values[0])
		cliqueId = ptr.To(values[1])
	}
	return domainUid, cliqueId
}

func getTopologyManagerPolicy(node *corev1.Node) *swagger.TopologyManagerPolicy {
	value, ok := node.Annotations[sdkkaiv1alpha1.AnnotationTopologyManagerPolicy]
	if !ok || value == "" {
		return nil
	}
	policy := swagger.TopologyManagerPolicy(value)
	if !policy.IsValid() {
		zap.S().Warnf("Node <%s> has an unrecognized topology manager policy <%s>; skipping", node.Name, value)
		return nil
	}
	return &policy
}

func (nodeParser *NodeParser) ParseNodeToServerScheme(node *corev1.Node) swagger.NodeInfo {
	conditions := nodeParser.calculateNodeConditions(node)

	nodePhase := getReadinessPhase(conditions)

	domainUid, cliqueId := tryGetNVLinkValues(node)
	// Parse Node label <gpuNetworkAccelerationLabelKey> value and set relevant fields
	nodeInfo := swagger.NodeInfo{
		Status:     nodePhase,
		Conditions: conditions,
		Taints:     parseTaints(node.Spec.Taints),
		NodePool:   getObjNodePool(node.Labels),
		CreatedAt:  node.CreationTimestamp.Time,

		NvLinkCliqueId:  *swagger.NewNullableString(cliqueId),
		NvLinkDomainUid: *swagger.NewNullableString(domainUid),

		NetworkTopologyLabels: nodeParser.getNetworkTopologyLabels(node),
	}

	if policy := getTopologyManagerPolicy(node); policy != nil {
		numaInfo := swagger.NumaInfo{}
		numaInfo.TopologyManagerPolicy.Set(policy)
		nodeInfo.Numa.Set(&numaInfo)
	}

	gpuInfo := getGpuInfo(node.Labels, node.Name)
	if gpuInfo != nil {
		nodeInfo.GpuInfo = *swagger.NewNullableGpuInfo(gpuInfo)
	}

	return nodeInfo
}

// ParseNodesToServerScheme returns a map of all node names; and the list of swagger NodeForSync.
func (nodeParser *NodeParser) ParseNodesToServerScheme(nodes []corev1.Node) (map[string]bool, []swagger.NodeForSync) {
	nodeNamesMap := map[string]bool{}
	result := []swagger.NodeForSync{}
	for idx := range nodes {
		node := nodes[idx]
		nodeInfo := nodeParser.ParseNodeToServerScheme(&node)

		nodeNamesMap[node.Name] = true
		result = append(result, swagger.NodeForSync{
			Name:     node.Name,
			NodeInfo: nodeInfo,
		})
	}
	return nodeNamesMap, result
}

func (nodeParser *NodeParser) getNetworkTopologyLabels(node *corev1.Node) []swagger.NetworkTopologyLabel {
	nodePoolName := getObjNodePool(node.Labels)

	nodePool := &sdkkaiv1alpha1.NodePool{}
	if err := nodeParser.cachedReader.Get(context.Background(), client.ObjectKey{Name: nodePoolName}, nodePool); err != nil {
		zap.S().Warnf("Failed to get nodepool <%s> for node <%s>: %s", nodePoolName, node.Name, err.Error())
		return nil
	}
	topologyName := nodePool.Spec.PreferredNetworkTopologyName

	if topologyName == "" {
		return nil
	}

	topology := &kaiv1alpha1.Topology{}
	err := nodeParser.cachedReader.Get(context.Background(), client.ObjectKey{Name: topologyName}, topology)
	if err != nil {
		zap.S().Warnf("Failed to get topology <%s> for nodepool <%s>: %s", topologyName, nodePoolName, err.Error())
		return nil
	}

	var labels []swagger.NetworkTopologyLabel
	for _, level := range topology.Spec.Levels {
		labelValue := node.Labels[level.NodeLabel]
		labels = append(labels, swagger.NetworkTopologyLabel{
			LabelKey:   level.NodeLabel,
			LabelValue: labelValue,
		})
	}

	return labels
}

func (nodeParser *NodeParser) calculateNodeConditions(node *corev1.Node) []swagger.NodeStatusConditionDetails {
	result := []swagger.NodeStatusConditionDetails{}

	gpuFractioningEnabled := config.Config().EnableNodeGpuFractioningReadiness

	ignoredConditionTypes := map[corev1.NodeConditionType]bool{}
	if gpuFractioningEnabled {
		ignoredConditionTypes[GpuFractioningReadyConditionType] = true
	}

	result = append(result, calculateStatusConditions(node.Status.Conditions, ignoredConditionTypes)...)

	result = append(result, calculateUnschedulableCondition(node)...)

	if gpuFractioningEnabled && nodeParser.clusterSupportsGpuFractioning() {
		result = append(result, calculateGpuFractioningConditions(node)...)
	}

	if config.Config().EnableNodeNvidiaComponentsReadiness {
		result = append(result, nodeParser.calculateMissingNvidiaComponentsConditions(node)...)
	}

	return result
}

// The fractioning operator stamps a cluster-wide reason onto every GPU node, so without this
// every node would repeat what the cluster status already says once.
//
// Unreadable or absent cluster state is treated as supported, matching what the cluster status
// does with the same input, so the two never disagree.
func (nodeParser *NodeParser) clusterSupportsGpuFractioning() bool {
	supported, reason := nodeParser.readClusterGpuFractioningSupport()
	if nodeParser.gpuFractioningSupport.store(supported) && reason != "" {
		zap.S().Warnf("Assuming the cluster supports GPU fractioning: %s", reason)
	}
	return supported
}

// A non-empty reason means the verdict could not be read and support was assumed.
func (nodeParser *NodeParser) readClusterGpuFractioningSupport() (bool, string) {
	clusterInfoCm := &corev1.ConfigMap{}
	err := nodeParser.cachedReader.Get(context.Background(), client.ObjectKey{
		Name:      constants.ClusterInfoConfigMapName,
		Namespace: runaiv1.DefaultNamespace,
	}, clusterInfoCm)
	if err != nil {
		return true, fmt.Sprintf("failed to read %s: %v", constants.ClusterInfoConfigMapName, err)
	}

	var runaiPublic runaiv1.RunaiPublic
	if err := yaml.Unmarshal([]byte(clusterInfoCm.Data[constants.ClusterInfoConfigMapName]), &runaiPublic); err != nil {
		return true, fmt.Sprintf("failed to unmarshal %s: %v", constants.ClusterInfoConfigMapName, err)
	}

	if deps := runaiPublic.Dependencies; deps != nil {
		if gpuStack, found := deps.Required[gpuStackDependencyName]; found {
			if fractioning, found := gpuStack.Components[gpuFractioningComponentName]; found {
				return ptr.Deref(fractioning.Available, true), ""
			}
		}
	}
	return true, ""
}

// GpuFractioningSupportChanged reports whether the currently published verdict differs from the
// one the parser last used. The cluster-info ConfigMap carries unrelated fields that churn on
// every operator reconcile, so node resyncs are driven off this rather than off the update event.
func (nodeParser *NodeParser) GpuFractioningSupportChanged() bool {
	previous, hadPrevious := nodeParser.gpuFractioningSupport.last()
	return hadPrevious && previous != nodeParser.clusterSupportsGpuFractioning()
}

func (nodeParser *NodeParser) calculateMissingNvidiaComponentsConditions(node *corev1.Node) []swagger.NodeStatusConditionDetails {
	if !shouldNodeHaveNvidiaComponents(node.Labels) {
		return []swagger.NodeStatusConditionDetails{}
	}

	result := []swagger.NodeStatusConditionDetails{}
	for nvidiaComponentName, conditionType := range nvidiaComponentNameToConditionType {
		if nodeParser.isNodeLabeledWithNvidiaComponent(node, nvidiaComponentName) {
			continue
		}

		foundOnNode, err := nodeParser.isNvidiaComponentPodRunningOnNode(node.Name, nvidiaComponentName)
		if err != nil {
			result = append(result, swagger.NodeStatusConditionDetails{
				Type:    conditionType,
				Reason:  "ErrorSearchingForNvidiaComponent",
				Message: utils.Ref(err.Error()),
			})
			continue
		}

		if !foundOnNode {
			result = append(result, swagger.NodeStatusConditionDetails{
				Type:    conditionType,
				Reason:  conditionType,
				Message: utils.Ref(fmt.Sprintf("nvidia-%s pod is missing from node", nvidiaComponentName)),
			})
		}
	}

	return result
}

func (nodeParser *NodeParser) isNodeLabeledWithNvidiaComponent(node *corev1.Node, nvidiaComponentName string) bool {
	dgxKey := fmt.Sprintf("nvidia.com/gpu.deploy.%s", nvidiaComponentName)
	dgxValue := "true"

	var labelKeyFromConfig string
	switch nvidiaComponentName {
	case "container-toolkit":
		labelKeyFromConfig = nodeParser.nodeLabelNvidiaContainerToolkit
	case "dcgm-exporter":
		labelKeyFromConfig = nodeParser.nodeLabelNvidiaDcgmExporter
	}

	if labelKeyFromConfig == dgxKey {
		labelKeyFromConfig = ""
	}

	for key, value := range node.Labels {
		if value != dgxValue {
			continue
		}
		if key == dgxKey {
			return true
		}
		if labelKeyFromConfig != "" && key == labelKeyFromConfig {
			return true
		}
	}

	return false
}

func (nodeParser *NodeParser) isNvidiaComponentPodRunningOnNode(nodeName, nvidiaComponentName string) (bool, error) {
	nvidiaComponentLabelOptions := getLabelSelectorsForNvidiaComponent(nvidiaComponentName)
	podsByNodeNameFieldSelector := fields.OneTermEqualSelector(constants.NonRunaiPodByNodeNameIndexer, nodeName)

	for _, nvidiaComponentLabelOption := range nvidiaComponentLabelOptions {
		pods, err := nodeParser.listPodsWithFieldSelectorAndLabelOption(podsByNodeNameFieldSelector, nvidiaComponentLabelOption)
		if err != nil {
			return false, err
		}

		if len(pods) > 0 {
			return true, nil
		}
	}

	zap.S().Warnf("Couldn't find nvidia-%s pod running on node %s", nvidiaComponentName, nodeName)
	return false, nil
}

func (nodeParser *NodeParser) listPodsWithFieldSelectorAndLabelOption(
	fieldSelector fields.Selector, listOption client.ListOption,
) ([]corev1.Pod, error) {
	pods := &corev1.PodList{}
	err := nodeParser.cachedReader.List(context.Background(), pods,
		listOption, &client.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		zap.S().Errorf("Failed to list pods with fields selector <%s> and label option, error: %s",
			fieldSelector.String(), err.Error())
		return []corev1.Pod{}, err
	}
	return pods.Items, nil
}

func getReadinessPhase(conditions []swagger.NodeStatusConditionDetails) string {
	if len(conditions) == 0 {
		return ReadyPhase
	}

	// Precedence: NotReady > MissingPrerequisites > Ready.
	hasMissingPrerequisite := false
	for _, condition := range conditions {
		if _, ok := knownNotReadyConditionTypes[condition.Type]; ok {
			return NotReadyPhase
		}
		if _, ok := knownMissingPrerequisiteConditionTypes[condition.Type]; ok {
			hasMissingPrerequisite = true
		}
	}

	if hasMissingPrerequisite {
		return MissingPrerequisitesPhase
	}
	return ReadyPhase
}

func calculateStatusConditions(
	conditions []corev1.NodeCondition, ignoredConditionTypes map[corev1.NodeConditionType]bool,
) []swagger.NodeStatusConditionDetails {
	result := []swagger.NodeStatusConditionDetails{}
	for _, condition := range conditions {
		if ignoredConditionTypes[condition.Type] {
			continue
		}

		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionFalse || condition.Status == corev1.ConditionUnknown {
				result = append(result, swagger.NodeStatusConditionDetails{
					Type:    NodeNotReadyConditionType,
					Reason:  condition.Reason,
					Message: &condition.Message,
				})
			}

			continue
		}

		if condition.Status == corev1.ConditionTrue || condition.Status == corev1.ConditionUnknown {
			result = append(result, parseCondition(condition))
			continue
		}
	}

	return result
}

func calculateUnschedulableCondition(node *corev1.Node) []swagger.NodeStatusConditionDetails {
	if !node.Spec.Unschedulable {
		return []swagger.NodeStatusConditionDetails{}
	}

	_, found := node.Labels[clusterconstants.LabelMarkedUnschedulableByNodePoolController]
	if found {
		return []swagger.NodeStatusConditionDetails{{
			Type:    UndrainedMigratedNode,
			Reason:  UndrainedMigratedNode,
			Message: utils.Ref("node is Unschedualble due to nodepool migration while node is undrained"),
		}}
	}

	return []swagger.NodeStatusConditionDetails{{
		Type:    SchedulingDisabled,
		Reason:  SchedulingDisabled,
		Message: utils.Ref("node was marked Unschedulable"),
	}}
}

func parseCondition(condition corev1.NodeCondition) swagger.NodeStatusConditionDetails {
	return swagger.NodeStatusConditionDetails{
		Type:    string(condition.Type),
		Reason:  condition.Reason,
		Message: &condition.Message,
	}
}

// Only an explicit True confirms support: a stale True is what an uninstalled operator leaves
// behind, so anything else fails closed. GPU nodes only, since absent counts as unsupported and
// would otherwise fire on every CPU node.
func calculateGpuFractioningConditions(node *corev1.Node) []swagger.NodeStatusConditionDetails {
	if !shouldNodeHaveNvidiaComponents(node.Labels) {
		return []swagger.NodeStatusConditionDetails{}
	}

	condition := findNodeCondition(node, GpuFractioningReadyConditionType)
	switch {
	case condition == nil:
		return gpuFractioningConditionDetails(GpuFractioningNotManaged, gpuFractioningNotManagedMessage)
	case condition.Status == corev1.ConditionTrue:
		return []swagger.NodeStatusConditionDetails{}
	case condition.Status == corev1.ConditionFalse:
		// Reason is optional in the Kubernetes API but required by the Nodes API, so a producer
		// that omits it must not leave us emitting a blank one.
		if condition.Reason == "" {
			return gpuFractioningConditionDetails(MissingGpuFractioningPrerequisite, condition.Message)
		}
		return gpuFractioningConditionDetails(condition.Reason, condition.Message)
	default:
		zap.S().Warnf("Node <%s> has an unrecognized %s status <%s>; treating fractions as unsupported",
			node.Name, GpuFractioningReadyConditionType, condition.Status)
		return gpuFractioningConditionDetails(GpuFractioningNotManaged, gpuFractioningNotManagedMessage)
	}
}

func gpuFractioningConditionDetails(reason, message string) []swagger.NodeStatusConditionDetails {
	return []swagger.NodeStatusConditionDetails{{
		Type:    MissingGpuFractioningPrerequisite,
		Reason:  reason,
		Message: utils.Ref(message),
	}}
}

func findNodeCondition(node *corev1.Node, conditionType corev1.NodeConditionType) *corev1.NodeCondition {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			return &node.Status.Conditions[i]
		}
	}
	return nil
}

func shouldNodeHaveNvidiaComponents(nodeLabels map[string]string) bool {
	val, found := nodeLabels[config.Config().NvidiaGPULabel]
	return found && val == "true"
}

func getLabelSelectorsForNvidiaComponent(nvidiaComponentName string) []client.ListOption {
	// different gpu-operator versions have different labels on the components
	appKeys := []string{"app", "app.kubernetes.io/name"}
	appValues := []string{
		fmt.Sprintf("nvidia-%s", nvidiaComponentName), fmt.Sprintf("nvidia-%s-daemonset", nvidiaComponentName),
		fmt.Sprintf("%s-daemonset", nvidiaComponentName), nvidiaComponentName,
	}

	result := []client.ListOption{}
	for _, appValue := range appValues {
		for _, appKey := range appKeys {
			result = append(result, client.MatchingLabels{appKey: appValue})
		}
	}

	return result
}

func parseTaints(taints []corev1.Taint) []swagger.NodeTaint {
	result := []swagger.NodeTaint{}
	for _, taint := range taints {
		result = append(result, swagger.NodeTaint{
			Key:    taint.Key,
			Value:  &taint.Value,
			Effect: string(taint.Effect),
		})
	}
	return result
}

func getGpuInfo(nodeLabels map[string]string, nodeName string) *swagger.GpuInfo {
	nvidiaGpuInfo := getNvidiaGpuInfo(nodeLabels, nodeName)
	if nvidiaGpuInfo != nil {
		return nvidiaGpuInfo
	}

	amdGpuInfo := getAmdGpuInfo(nodeLabels, nodeName)
	if amdGpuInfo != nil {
		return amdGpuInfo
	}

	return nil
}

func getNvidiaGpuInfo(nodeLabels map[string]string, nodeName string) *swagger.GpuInfo {
	nvidiaGpuType, found := nodeLabels[NvidiaGpuTypeLabelKey]
	if !found {
		return nil
	}

	result := &swagger.GpuInfo{
		GpuType:        nvidiaGpuType,
		GpuMemoryBytes: getNvidiaGpuMemoryBytes(nodeLabels, nodeName),
	}

	nvidiaGpuCountStr, countFound := nodeLabels[NvidiaGpuCountLabelKey]
	if !countFound {
		zap.S().Warnf("Found nvidia gpu type <%s> on node <%s>, but can't find gpu count",
			nvidiaGpuType, nodeName)
		return result
	}

	nvidiaGpuCount, err := strconv.Atoi(nvidiaGpuCountStr)
	if err != nil {
		zap.S().Errorf("Failed to parse nvidia gpu count label value: <%s> for node <%s>, error: %s",
			nvidiaGpuCountStr, nodeName, err.Error())
		return result
	}

	result.GpuCount = int32(nvidiaGpuCount)
	return result
}

// Returns nil when the label is missing or unusable, so a node still reports its GPU type
// and count when only the memory is unknown.
func getNvidiaGpuMemoryBytes(nodeLabels map[string]string, nodeName string) *int64 {
	gpuMemoryStr, found := nodeLabels[NvidiaGpuMemoryLabelKey]
	if !found {
		return nil
	}

	gpuMemory, err := strconv.Atoi(gpuMemoryStr)
	if err != nil {
		zap.S().Errorf("Failed to parse nvidia gpu memory label value: <%s> for node <%s>, error: %s",
			gpuMemoryStr, nodeName, err.Error())
		return nil
	}

	if gpuMemory <= 0 {
		zap.S().Warnf("Found non-positive nvidia gpu memory <%d> on node <%s>", gpuMemory, nodeName)
		return nil
	}

	if gpuMemory > maxPlausibleGpuMemoryMib {
		return ptr.To(int64(gpuMemory))
	}

	return ptr.To(int64(gpuMemory) * bytesInMib)
}

func getAmdGpuInfo(nodeLabels map[string]string, nodeName string) *swagger.GpuInfo {
	for key, value := range nodeLabels {
		if !strings.HasPrefix(key, AmdGpuTypeLabelPrefix) {
			continue
		}

		parts := strings.Split(key, AmdGpuTypeLabelPrefix)
		if len(parts) != 2 {
			zap.S().Errorf("Failed to parse amd gpu family label key: <%s> for node <%s>",
				key, nodeName)
			return nil
		}
		amdGpuFamilyAcronym := parts[1]

		amdGpuFamily, familyFound := amdGpuFamilies[amdGpuFamilyAcronym]
		if !familyFound {
			amdGpuFamily = amdGpuFamilyAcronym
		}

		result := &swagger.GpuInfo{
			GpuType: amdGpuFamily,
		}

		amdGpuCount, err := strconv.Atoi(value)
		if err != nil {
			zap.S().Errorf("Failed to parse amd gpu count label value: <%s> for node <%s>, error: %s",
				value, nodeName, err.Error())
			return result
		}

		result.GpuCount = int32(amdGpuCount)
		return result
	}

	return nil
}
