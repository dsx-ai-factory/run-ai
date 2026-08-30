package tests

import (
	"encoding/json"
	"fmt"
	"strconv"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kaiconsts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics/node_mig_exporter"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	TestProjectName  = "test"
	TestProjectName2 = "test2"

	TestDepartmentName  = "dep1"
	TestDepartmentName2 = "dep2"
	TestNodePoolName    = "default"
	TestNodePoolName2   = "np2"

	TestQueueName  = "test"
	TestQueueName2 = "test-np2"

	TestNamespaceName  = "runai-test"
	TestNamespaceName2 = "ns-test2"

	TestNodeName = "node1"
)

var (
	TestProject = &v2.Project{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2",
			Kind:       "Project",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestProjectName,
			Labels: map[string]string{
				v2alpha2.DepartmentIdLabel: "25",
			},
		},
		Spec: v2.ProjectSpec{
			Department: TestDepartmentName,
		},
		Status: v2.ProjectStatus{
			Namespace: TestNamespaceName,
		},
	}

	// kai.resources twin of TestProject, so the suite passes under both values of the compile-time
	// kai.resources project. Spec.Parent replaces Spec.Department.
	TestKaiProject = &kaiorg.Project{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kai.resources/v1alpha1",
			Kind:       "Project",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestProjectName,
			Labels: map[string]string{
				v2alpha2.DepartmentIdLabel: "25",
			},
		},
		Spec: kaiorg.ProjectSpec{
			Parent: TestDepartmentName,
		},
		Status: kaiorg.ProjectStatus{
			Namespace: TestNamespaceName,
		},
	}

	TestNodePool = &v1alpha1.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.run.ai/v2alpha2",
			Kind:       "NodePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNodePoolName,
		},
		Spec: v1alpha1.NodePoolSpec{},
	}

	TestNodePool2 = &v1alpha1.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha2",
			Kind:       "NodePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNodePoolName2,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   "key1",
			LabelValue: "val1",
		},
	}

	// kai.resources twins of the node pools above. workload-exporter resolves a pod's node
	// pools through node-pool-utils/converter, which reads whichever CRD UseKaiNodePools
	// selects, so both are seeded with identical content.
	TestKaiNodePool = &kaiorg.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kai.resources/v1alpha1",
			Kind:       "NodePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNodePoolName,
		},
		Spec: kaiorg.NodePoolSpec{},
	}

	TestKaiNodePool2 = &kaiorg.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kai.resources/v1alpha1",
			Kind:       "NodePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNodePoolName2,
		},
		Spec: kaiorg.NodePoolSpec{
			LabelKey:   "key1",
			LabelValue: "val1",
		},
	}

	TestQueue = &kaiv2.Queue{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.run.ai/v2",
			Kind:       "Queue",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestQueueName,
			Labels: map[string]string{
				clustercommonconstants.RunaiProjectLabel: TestProjectName,
				"runai/queue":                            TestProjectName,
			},
		},
	}

	TestQueue2 = &kaiv2.Queue{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.run.ai/v2",
			Kind:       "Queue",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestQueueName2,
			Labels: map[string]string{
				clustercommonconstants.RunaiProjectLabel: TestProjectName,
				"runai/queue":                            TestProjectName,
				v1alpha1.NodePoolNameLabel:               TestNodePoolName2,
			},
		},
	}

	TestNamespace = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNamespaceName,
			Labels: map[string]string{
				clustercommonconstants.RunaiQueueLabel: TestProjectName,
			},
		},
	}

	TestNamespace2 = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestNamespaceName2,
		},
	}

	TestPodExpectedPodMetricValues = &ExpectedPodMetricValues{
		requestedGpus:      1,
		requestedGPUMemory: 0,
		limitGpus:          1,
		limitGPUMemory:     0,
		allocatedGpus:      1,
		allocatedGPUMemory: 10000,
		allocatedCpus:      100,
		allocatedCpuMemory: 1000,
	}

	PriorityClassTrain = &v1.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.k8s.io/v1",
			Kind:       "PriorityClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "train",
		},
		Value: 50,
	}

	PriorityClassInteractivePreemtible = &v1.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.k8s.io/v1",
			Kind:       "PriorityClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "interactive-preemptible",
		},
		Value: 75,
	}

	PriorityClassBuild = &v1.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.k8s.io/v1",
			Kind:       "PriorityClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "build",
		},
		Value: 100,
	}

	PriorityClassInference = &v1.PriorityClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "scheduling.k8s.io/v1",
			Kind:       "PriorityClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "inference",
		},
		Value: 125,
	}
)

func buildPod(name, namespace, workloadId, node, phase string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(name),
			Labels: map[string]string{
				commonconstants.TopOwnerWorkloadIdLabel: workloadId,
				commonconstants.TopOwnerUidLabel:        workloadId,
			},
			Annotations: map[string]string{
				clustercommonconstants.PodStatusAnnotation:             phase,
				clustercommonconstants.PodAllocatedGpusAnnotation:      "1",
				clustercommonconstants.PodAllocatedGpuMemoryAnnotation: "10000",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:      node,
			SchedulerName: clustercommonconstants.RunaiScheduler,
			Containers: []corev1.Container{
				{
					Name:  "container1",
					Image: "image1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:                     resource.MustParse("100m"),
							corev1.ResourceMemory:                  resource.MustParse("1000"),
							clustercommonconstants.GpuResourceName: resource.MustParse("1"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:                     resource.MustParse("1000m"),
							corev1.ResourceMemory:                  resource.MustParse("5000"),
							clustercommonconstants.GpuResourceName: resource.MustParse("1"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPhase(phase),
		},
	}
}

func buildPodWithNodePool(name, namespace, workloadId, node, phase string) *corev1.Pod {
	pod := buildPod(name, namespace, workloadId, node, phase)

	pod.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      TestNodePool2.Spec.LabelKey,
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{TestNodePool2.Spec.LabelValue},
							},
						},
					},
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      v1alpha1.NodePoolNameLabel,
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
	}

	if phase == "Running" {
		pod.Labels[v1alpha1.NodePoolNameLabel] = TestNodePool2.Name
	}

	return pod
}

func buildPodWithElementIDs(name, namespace, workloadId, node, phase, element1ID, element2ID, element3ID string) *corev1.Pod {
	pod := buildPod(name, namespace, workloadId, node, phase)
	for i, id := range []string{element1ID, element2ID, element3ID} {
		if id != "" {
			pod.Labels[fmt.Sprintf("%s%d-id", clustercommonconstants.ElementLabelPrefix, i+1)] = id
		}
	}
	return pod
}

func buildPodOfWorkload(name, namespace, workloadId, node, phase, podGroupName string) *corev1.Pod {
	pod := buildPod(name, namespace, workloadId, node, phase)

	pod.Annotations[commonconstants.PodGroupAnnotationForPod] = podGroupName
	return pod
}

func buildPodOfWorkloadInNodePool(name, namespace, workloadId, node, phase, podGroupName string) *corev1.Pod {
	pod := buildPodWithNodePool(name, namespace, workloadId, node, phase)

	pod.Annotations[commonconstants.PodGroupAnnotationForPod] = podGroupName
	return pod
}

func buildExternalWorkload(name, namespace, originalWorkloadName, workloadId, workloadAnnotationId, workloadKind, workloadGroup string,
	phase v1alpha1.ExternalWorkloadPhase,
	requestedNodePools []string) *v1alpha1.ExternalWorkload {
	w := &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(name),
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID:  types.UID(workloadId),
			OriginalWorkloadName: originalWorkloadName,
			Type:                 workloadKind,
			Group:                workloadGroup,
			RequestedNodePools:   requestedNodePools,
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase: phase,
			},
		},
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}
	return w
}

func buildTrainingWorkload(name, namespace, workloadId, workloadAnnotationId, phase, user string) *twv2a1.TrainingWorkload {
	w := &twv2a1.TrainingWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha1",
			Kind:       "TrainingWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(workloadId),
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					UserName: &flds.StringField{Value: ptr.To(user)},
				},
			},
		},
		Status: twv2a1.TrainingWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				Phase: phase,
			},
		},
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}
	return w
}

func buildInteractiveWorkload(name, namespace, workloadId, workloadAnnotationId, phase, user string, preemptible bool) *iwv2a1.InteractiveWorkload {
	w := &iwv2a1.InteractiveWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha1",
			Kind:       "InteractiveWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(workloadId),
		},
		Status: iwv2a1.InteractiveWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				Phase: phase,
			},
		},
	}

	if user != "" {
		w.Spec.WorkloadFields.UserName = &flds.StringField{Value: ptr.To(user)}
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}

	if preemptible {
		w.Spec.InteractiveFields.Preemptible = &flds.BoolField{Value: ptr.To(true)}
	}

	return w
}

func buildDistributedWorkload(name, namespace, workloadId, workloadAnnotationId, phase, user string) *diwv2.DistributedWorkload {
	w := &diwv2.DistributedWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha1",
			Kind:       "DistributedWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(workloadId),
		},
		Spec: dp.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					UserName: &flds.StringField{Value: ptr.To(user)},
				},
			},
		},
		Status: diwv2.DistributedWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				Phase: phase,
			},
		},
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}
	return w
}

func buildDistributedInferenceWorkload(name, namespace, workloadId, workloadAnnotationId, phase, user string) *dinfw.DistributedInferenceWorkload {
	w := &dinfw.DistributedInferenceWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha1",
			Kind:       "DistributedInferenceWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(workloadId),
		},
		Status: dinfw.DistributedInferenceWorkloadStatus{
			DistributedInferenceStatus: common.DistributedInferenceStatus{
				Phase: phase,
			},
		},
	}

	if user != "" {
		w.Spec.UserName = &flds.StringField{Value: ptr.To(user)}
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}

	return w
}

func buildInferenceWorkload(name, namespace, workloadId, workloadAnnotationId, phase, user string) *infwv2a1.InferenceWorkload {
	w := &infwv2a1.InferenceWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "run.ai/v2alpha1",
			Kind:       "InferenceWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(workloadId),
		},
		Status: infwv2a1.InferenceWorkloadStatus{
			InferenceStatus: common.InferenceStatus{
				Phase: phase,
			},
		},
	}

	if user != "" {
		w.Spec.WorkloadFields.UserName = &flds.StringField{Value: ptr.To(user)}
	}

	if workloadAnnotationId != "" {
		w.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadAnnotationId,
		}
	}
	return w
}

func buildPodGroup(name, namespace, priorityClass, phase string) *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(name),
			Annotations: map[string]string{
				clustercommonconstants.PodGroupStatusAnnotation: phase,
			},
		},
		Spec: kaiv2alpha2.PodGroupSpec{
			PriorityClassName: priorityClass,
		},
	}
}

func getReconcileRequest(name, namespace string) ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func buildPodGroupWithMigDevices(name, namespace, queueName, workloadName, workloadUID string, migDevices map[string]int64) *kaiv2alpha2.PodGroup {
	pg := buildBasePodGroup(name, namespace, queueName, workloadName, workloadUID)

	// Add MIG devices annotation
	migDevicesJSON, _ := json.Marshal(migDevices)
	pg.Annotations[clustercommonconstants.PodGroupRequestedMigDevicesAnnotation] = string(migDevicesJSON)

	return pg
}

func buildPodGroupWithoutMigDevices(name, namespace, queueName, workloadName, workloadUID string) *kaiv2alpha2.PodGroup {
	return buildBasePodGroup(name, namespace, queueName, workloadName, workloadUID)
}

func buildBasePodGroup(name, namespace, queueName, workloadName, workloadUID string) *kaiv2alpha2.PodGroup {
	topOwnerMetadata := map[string]string{
		"name":    workloadName,
		"uid":     workloadUID,
		"group":   "run.ai",
		"version": "v1",
		"kind":    "TrainingWorkload",
	}
	topOwnerJSON, _ := json.Marshal(topOwnerMetadata)

	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(name),
			Annotations: map[string]string{
				kaiconsts.TopOwnerMetadataKey: string(topOwnerJSON),
			},
		},
		Spec: kaiv2alpha2.PodGroupSpec{
			Queue:             queueName,
			PriorityClassName: "train",
		},
	}
}

func buildNode(name string, gpuCount int, isMigNode bool, migStrategy string) *corev1.Node {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{},
		},
		Spec: corev1.NodeSpec{},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeMemoryPressure,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodePIDPressure,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   corev1.NodeNetworkUnavailable,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}

	if gpuCount > 0 {
		node.Labels[clustercommonconstants.GpuCountLabel] = strconv.Itoa(gpuCount)
	}

	if isMigNode {
		node.Labels[node_mig_exporter.GpuProductLabel] = node_mig_exporter.A100 + "-40GB"
		node.Labels[node_mig_exporter.MigEnabledLabel] = node_mig_exporter.TrueLabelValue
		node.Labels[node_mig_exporter.MigStrategyLabel] = migStrategy
	}

	return node
}

func buildNonMigNode(name string, gpuCount int) *corev1.Node {
	return buildNode(name, gpuCount, false, "")
}

func buildMigNode(name string, gpuCount int, migStrategy string) *corev1.Node {
	return buildNode(name, gpuCount, true, migStrategy)
}

func buildUnreadyNode(name string, gpuCount int, isMigNode bool, migStrategy string) *corev1.Node {
	node := buildNode(name, gpuCount, isMigNode, migStrategy)
	for i, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			node.Status.Conditions[i].Status = corev1.ConditionFalse
			break
		}
	}
	return node
}
