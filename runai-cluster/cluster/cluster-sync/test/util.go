/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"

	dinfp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"

	clustercommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	appsv1 "k8s.io/api/apps/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	schedv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	ipv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	runaijobv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"

	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

const (
	migValidNameOne = "nvidia.com/mig-1g.5gb"
	migValidNameTwo = "nvidia.com/mig-2g.10gb"
)

var (
	scheme *runtime.Scheme
	once   sync.Once
)

func BuildResourceList(cpu string, memory string) corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpu),
		corev1.ResourceMemory: resource.MustParse(memory),
	}
}

func BuildMigRunaiPod(namespace, name, imageName, nodeName string, phase corev1.PodPhase, req corev1.ResourceList, owner []metav1.OwnerReference, labels map[string]string, migName string, migCount int) *corev1.Pod {
	pod := BuildPod(namespace, name, imageName, nodeName, phase, req, owner, labels)
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{
			migName: strconv.Itoa(migCount),
		}
	} else {
		pod.Annotations[migName] = strconv.Itoa(migCount)
	}
	return pod
}

func BuildPVC(namespace string, name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(fmt.Sprintf("%v-%v", namespace, name)),
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				clustercommon.RunaiResourceLabel: "pvc",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: convutils.Ref(fmt.Sprintf("storage-class-for-%s", name)),
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": *resource.NewQuantity(1000, resource.DecimalSI),
				},
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
			Capacity: corev1.ResourceList{
				"storage": *resource.NewQuantity(1000, resource.DecimalSI),
			},
		},
	}
}

func BuildConfigMap(namespace string, name string, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(fmt.Sprintf("%v-%v", namespace, name)),
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func getWorkloadConfig() []workloadTypeConfig.WorkloadTypeConfig {
	return []workloadTypeConfig.WorkloadTypeConfig{
		{TypeName: "AmlJob", Group: "amlarc.azureml.com", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "CronJob", Group: "batch", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "Deployment", Group: "apps", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "DevWorkspace", Group: "workspace.devfile.io", CategoryName: "Build", PriorityName: "high"},
		{TypeName: "JAXJob", Group: "kubeflow.org", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "Job", Group: "batch", CategoryName: "Build", PriorityName: "high"},
		{TypeName: "MPIJob", Group: "kubeflow.org", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "Notebook", Group: "kubeflow.org", CategoryName: "Build", PriorityName: "high"},
		{TypeName: "PipelineRun", Group: "tekton.dev", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "Pod", Group: "", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "PyTorchJob", Group: "kubeflow.org", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "RayCluster", Group: "ray.io", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "RayJob", Group: "ray.io", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "RayService", Group: "ray.io", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "ReplicaSet", Group: "apps", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "ScheduledWorkflow", Group: "kubeflow.org", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "SeldonDeployment", Group: "machinelearning.seldon.io", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "Service", Group: "serving.knative.dev", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "StatefulSet", Group: "apps", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "TaskRun", Group: "tekton.dev", CategoryName: "Build", PriorityName: "high"},
		{TypeName: "TFJob", Group: "kubeflow.org", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "VirtualMachineInstance", Group: "kubevirt.io", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "Workflow", Group: "argoproj.io", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "XGBoostJob", Group: "kubeflow.org", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "DistributedWorkload", Group: "run.ai", CategoryName: "Train", PriorityName: "low"},
		{TypeName: "InteractiveWorkload", Group: "run.ai", CategoryName: "Build", PriorityName: "high"},
		{TypeName: "InferenceWorkload", Group: "run.ai", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "DistributedInferenceWorkload", Group: "run.ai", CategoryName: "Deploy", PriorityName: "very-high"},
		{TypeName: "TrainingWorkload", Group: "run.ai", CategoryName: "Train", PriorityName: "low"},
	}
}

func getCategories() []string {
	return []string{"Build", "Train", "Deploy"}
}

func BuildWorkloadTypeConfigMap() *corev1.ConfigMap {
	typesBytes, _ := json.Marshal(getWorkloadConfig())
	categoriesBytes, _ := json.Marshal(getCategories())
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(fmt.Sprintf("%v-%v", clustercommon.WorkloadsConfigMapNamespace, clustercommon.WorkloadsConfigMapName)),
			Name:      clustercommon.WorkloadsConfigMapName,
			Namespace: clustercommon.WorkloadsConfigMapNamespace,
		},
		Data: map[string]string{
			"categories": string(categoriesBytes),
			"types":      string(typesBytes),
		},
	}
}

func BuildSecret(namespace string, name string, assetOrResource string, secretType string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(fmt.Sprintf("%v-%v", namespace, name)),
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				assetOrResource: secretType,
			},
		},
	}
}

const IsDefaultClassLabel = "storageclass.kubernetes.io/is-default-class"

func BuildStorageClass(name string, isDefault bool) *storagev1.StorageClass {
	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			UID:  types.UID(fmt.Sprintf("%v-%v", name, isDefault)),
			Name: name,
			Annotations: map[string]string{
				IsDefaultClassLabel: strconv.FormatBool(isDefault),
			},
		},
	}
}

func BuildPod(namespace, name, imageName, nodeName string, phase corev1.PodPhase, req corev1.ResourceList,
	owner []metav1.OwnerReference, labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:             types.UID(fmt.Sprintf("%v-%v", namespace, name)),
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: owner,
			Labels:          labels,
		},
		Status: corev1.PodStatus{
			Phase: phase,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,

			SchedulerName: constants.RunaiSchedulerName,
			Containers: []corev1.Container{
				{
					Image: imageName,
					Resources: corev1.ResourceRequirements{
						Requests: req,
					},
				},
			},
		},
	}
}

func BuildPodGeneric(namespace, name string) *corev1.Pod {
	return BuildPod(namespace,
		name,
		"imageName",
		"node1",
		corev1.PodRunning,
		BuildResourceList("1000m", "1G"),
		[]metav1.OwnerReference{{
			APIVersion:         "v1",
			Kind:               "RunaiJob",
			Name:               "job1",
			UID:                "123",
			Controller:         nil,
			BlockOwnerDeletion: nil,
		}},
		make(map[string]string))
}

func BuildPodGenericWithSchedulerName(namespace, name, uid, schedulerName string) *corev1.Pod {
	pod := BuildPod(namespace,
		name,
		"imageName",
		"node1",
		corev1.PodRunning,
		BuildResourceList("1000m", "1G"),
		[]metav1.OwnerReference{},
		make(map[string]string))
	pod.UID = types.UID(uid)
	pod.Spec.SchedulerName = schedulerName
	return pod
}

func BuildPodGroup(name, namespace, uid string, hasOwner bool) *kaiv2alpha2.PodGroup {
	return BuildPodGroupInner(name, namespace, uid, hasOwner, nil, nil, nil)
}

func BuildPodGroupInner(name, namespace, uid string, hasOwner bool, schedulingBackOff *int32, labels, annotations map[string]string) *kaiv2alpha2.PodGroup {
	return BuildPodGroupInnerWithOwnerInfo(name, namespace, uid, hasOwner, schedulingBackOff, labels, annotations, "ExampleKind", fmt.Sprintf("%v-owner", name), fmt.Sprintf("%v-owner-uid", name))
}

func BuildPodGroupInnerWithOwnerInfo(name, namespace, uid string, hasOwner bool, schedulingBackOff *int32, labels, annotations map[string]string,
	ownerKind, ownerName, ownerUid string) *kaiv2alpha2.PodGroup {
	pg := &kaiv2alpha2.PodGroup{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
		Spec:   kaiv2alpha2.PodGroupSpec{},
		Status: kaiv2alpha2.PodGroupStatus{},
	}
	if hasOwner {
		pg.OwnerReferences = []metav1.OwnerReference{
			{
				Kind: ownerKind,
				Name: ownerName,
				UID:  types.UID(ownerUid),
			}}
	}
	if schedulingBackOff != nil {
		pg.Spec.SchedulingBackoff = schedulingBackOff
	}
	if labels != nil {
		pg.Labels = labels
	}
	if annotations != nil {
		pg.Annotations = annotations
	}
	return pg
}

func BuildDistributed(namespace, jobName, userId, workloadId, workloadStatus string) *diwv2.DistributedWorkload {
	workloadName := jobName + time.Now().String()
	return &diwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: namespace,
			Annotations: map[string]string{
				commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadId,
			},
		},
		Spec: dp.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					UserId: &flds.StringField{
						Value: ptr.To(userId),
					},
				},
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To(jobName),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
			},
		},
		MasterSpec: &dp.DistributedMasterSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To(jobName),
					},
				},
			},
		},
		Status: diwv2.DistributedWorkloadStatus{
			WorkloadStatus: common.WorkloadStatus{CreationStatus: workloadStatus},
		},
	}
}

func BuildRunaiJob(name, namespace, uid string, parallelism, completions int32, labels, annotations map[string]string) *runaijobv1.RunaiJob {
	runaijob := &runaijobv1.RunaiJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
		Spec: runaijobv1.JobSpec{
			Parallelism: ptr.To(parallelism),
			Completions: ptr.To(completions),
		},
	}

	if labels != nil {
		runaijob.Labels = labels
	}
	if annotations != nil {
		runaijob.Annotations = annotations
	}
	return runaijob
}

func BuildPyTorchJob(name, namespace, uid string, labels, annotations map[string]string, spec kubeflowv1.PyTorchJobSpec) *kubeflowv1.PyTorchJob {
	pyTorchJob := &kubeflowv1.PyTorchJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
		Spec: spec,
	}

	if labels != nil {
		pyTorchJob.Labels = labels
	}
	if annotations != nil {
		pyTorchJob.Annotations = annotations
	}
	return pyTorchJob
}

func BuildDeployment(name, namespace, uid string) *appsv1.Deployment {
	var projectName string
	if strings.HasPrefix(namespace, constants.RunaiNamespacePrefix) {
		projectName = strings.TrimPrefix(namespace, constants.RunaiNamespacePrefix)
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
			Annotations: map[string]string{
				clustercommon.AnnotationJobUrl: "http://test.com/url1, http://test.com/url2",
			},
			Labels: map[string]string{
				"project": projectName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "ExampleKind",
					Name: fmt.Sprintf("%v-owner", name),
					UID:  types.UID(fmt.Sprintf("%v-owner", name)),
				},
			},
			CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: Int32Ptr(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "example-container-name",
							Image: "example-container-image",
						},
					},
					SchedulerName: "example-scheduler-name",
				},
			},
		},
	}
}
func BuildService(name, namespace, uid string) *knv1.Service {
	var projectName string
	if strings.HasPrefix(namespace, constants.RunaiNamespacePrefix) {
		projectName = strings.TrimPrefix(namespace, constants.RunaiNamespacePrefix)
	}

	return &knv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			UID:               types.UID(uid),
			CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
			Labels: map[string]string{
				"project": projectName,
			},
		},
		Status: knv1.ServiceStatus{
			RouteStatusFields: knv1.RouteStatusFields{
				URL: &apis.URL{
					Host: "testhost.com",
				},
			},
		},
		// Spec: knv1.ServiceSpec{

		// 	},
	}
}

func BuildKnativeConfiguration(name, namespace, uid string) *knv1.Configuration {
	var projectName string
	if strings.HasPrefix(namespace, constants.RunaiNamespacePrefix) {
		projectName = strings.TrimPrefix(namespace, constants.RunaiNamespacePrefix)
	}

	return &knv1.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			UID:               types.UID(uid),
			CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
			Labels: map[string]string{
				"project": projectName,
				"release": name,
			},
		},
	}
}

func BuildKnativeRevision(name, namespace, revId string, uid string) *knv1.Revision {
	var projectName string
	if strings.HasPrefix(namespace, constants.RunaiNamespacePrefix) {
		projectName = strings.TrimPrefix(namespace, constants.RunaiNamespacePrefix)
	}

	revisionName := name + "-" + revId

	return &knv1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:              revisionName,
			Namespace:         namespace,
			UID:               types.UID(uid),
			CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
			Labels: map[string]string{
				"project": projectName,
				"release": name,
			},
		},
	}
}

func BuildNodePool(
	name string, labelKey string, labelValue string, phase kaiorg.NodePoolPhase, statusMessage string,
) *kaiorg.NodePool {
	return &kaiorg.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				kaiorg.AnnotationGPUNetworkAccelerationDetected: "true",
				kaiorg.AnnotationGPUNetworkAccelerationLabelKey: "nvidia/network",
			},
		},
		Spec: kaiorg.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
		},
		Status: kaiorg.NodePoolStatus{
			Phase:   phase,
			Message: statusMessage,
			Nodes: []kaiorg.NodeInNodePool{
				{Name: "node1", TopologyMismatch: false},
				{Name: "node2", TopologyMismatch: false},
			},
		},
	}
}

func BuildNodePoolWithCondition(
	name string, labelKey string, labelValue string, phase kaiorg.NodePoolPhase,
	statusMessage string, conditionType string, conditionReason string, conditionMessage string,
) *kaiorg.NodePool {
	nodePool := BuildNodePool(name, labelKey, labelValue, phase, statusMessage)
	nodePool.Status.Conditions = []kaiorg.NodePoolCondition{
		{
			Type:    kaiorg.NodePoolConditionType(conditionType),
			Status:  "True",
			Reason:  conditionReason,
			Message: conditionMessage,
		},
	}
	return nodePool
}

func BuildLegacyNodePool(name string, labelKey string, labelValue string, phase v1alpha1.NodePoolPhase, statusMessage string) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:                        labelKey,
			LabelValue:                      labelValue,
			GPUNetworkAccelerationDetection: ptr.To(v1alpha1.UseGPUNetworkAcceleration),
			GPUNetworkAccelerationLabelKey:  "nvidia/network",
		},

		Status: v1alpha1.NodePoolStatus{
			Phase:                          phase,
			Message:                        statusMessage,
			Nodes:                          []v1alpha1.NodeInNodePool{{Name: "node1", TopologyMismatch: false}, {Name: "node2", TopologyMismatch: false}},
			GPUNetworkAccelerationDetected: true,
		},
	}
}

func BuildLegacyNodePoolWithCondition(
	name string, labelKey string, labelValue string, phase v1alpha1.NodePoolPhase,
	statusMessage string, conditionType string, conditionReason string, conditionMessage string,
) *v1alpha1.NodePool {
	nodePool := BuildLegacyNodePool(name, labelKey, labelValue, phase, statusMessage)
	nodePool.Status.Conditions = []v1alpha1.NodePoolCondition{
		{
			Type:    v1alpha1.NodePoolConditionType(conditionType),
			Status:  "True",
			Reason:  conditionReason,
			Message: conditionMessage,
		},
	}
	return nodePool
}

func BuildDatavolume(name string, dvUid string, phase v1alpha1.DataVolumePhase, statusMessage string,
	dvPvcName, dvPvName string) *v1alpha1.DataVolume {
	return &v1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  types.UID(name),
		},
		Spec: v1alpha1.DataVolumeSpec{
			DatavolumeId: dvUid,
		},
		Status: v1alpha1.DataVolumeStatus{
			Phase:             phase,
			Message:           statusMessage,
			DatavolumePvcName: dvPvcName,
			DatavolumePvName:  dvPvName,
		},
	}

}

func BuildAccessRule(name string, id int32, phase v1alpha1.AccessRulePhase, statusMessage string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v1alpha1.AccessRuleIdLabel: strconv.Itoa(int(id)),
			},
		},
		Status: v1alpha1.AccessRuleStatus{
			Phase:   phase,
			Message: statusMessage,
		},
	}
}

func BuildExternalWorkload(name, namespace, originalWorkloadUID string,
	phase v1alpha1.ExternalWorkloadPhase) (*v1alpha1.ExternalWorkload, *corev1.Pod, *kaiv2alpha2.PodGroup) {
	if phase == "" {
		phase = v1alpha1.ExWRunningPhase
	}
	uid := originalWorkloadUID + "different-iddddddd"
	workloadUid := types.UID(uid)
	podUid := uuid.New().String()

	workload := &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       workloadUid,
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       "Deployment",
				Name:       name,
				UID:        types.UID(originalWorkloadUID),
				Controller: ptr.To(true),
			}},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID: types.UID(originalWorkloadUID),
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:                phase,
				LastStatusChangeTime: metav1.Now(),
				ChildIds:             []v1alpha1.ChildId{{UID: podUid, Type: "Pod"}},
			},
		},
	}
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-0",
			Namespace: namespace,
			UID:       types.UID(podUid),
			Labels: map[string]string{
				commonconstants.TopOwnerWorkloadIdLabel: originalWorkloadUID,
				commonconstants.TopOwnerUidLabel:        originalWorkloadUID,
			},
			Annotations: map[string]string{commonconstants.PodGroupAnnotationForPod: name},
		},
		Spec: corev1.PodSpec{},
	}

	podGroup := BuildPodGroup(name, namespace, name, false)
	podGroup.Labels = map[string]string{
		"some-key1": "some-value1",
		"some-key2": "some-value2",
	}
	podGroup.Annotations = map[string]string{
		"some-key11": "some-value11",
		"some-key22": "some-value22",
	}

	return workload, pod, podGroup
}

func BuildTrainingWorkload(name, namespace, uid string) (*twv2.TrainingWorkload, *corev1.Pod) {
	return buildTrainingWorkload(name, namespace, uid, "")
}

func buildTrainingWorkload(name, namespace, uid, backendWorkloadId string) (*twv2.TrainingWorkload, *corev1.Pod) {
	tw := &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: constants.RunaiWorkloadRelevantUsage,
				},
			},
		},
	}

	workloadUid := uid
	if backendWorkloadId != "" {
		tw.Annotations = map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: backendWorkloadId,
		}
		workloadUid = backendWorkloadId
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-0",
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels: map[string]string{
				commonconstants.TopOwnerUidLabel:        workloadUid,
				commonconstants.TopOwnerWorkloadIdLabel: workloadUid,
			},
		},
		Spec: corev1.PodSpec{},
	}

	return tw, pod
}

func BuildTrainingWorkloadBackendSourced(name, namespace, uid, backendWorkloadId string) (*twv2.TrainingWorkload, *corev1.Pod) {
	return buildTrainingWorkload(name, namespace, uid, backendWorkloadId)
}

func BuildInteractiveWorkload(name, namespace, uid string) (*iwv2.InteractiveWorkload, *corev1.Pod) {
	workloadUid := types.UID(uid)
	workload := &iwv2.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       workloadUid,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: constants.RunaiWorkloadRelevantUsage,
				},
			},
		},
	}
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-0",
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels:    map[string]string{commonconstants.TopOwnerUidLabel: uid},
		},
		Spec: corev1.PodSpec{},
	}

	return workload, pod
}

func BuildInferenceWorkload(name, namespace, uid string) (*infwv2.InferenceWorkload, *corev1.Pod) {
	workloadUid := types.UID(uid)
	workload := &infwv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       workloadUid,
		},
		Spec: infpv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: constants.RunaiWorkloadRelevantUsage,
				},
			},
		},
	}
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-0",
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels: map[string]string{
				commonconstants.TopOwnerUidLabel:        uid,
				commonconstants.TopOwnerWorkloadIdLabel: uid,
			},
		},
		Spec: corev1.PodSpec{},
	}
	return workload, pod
}

func BuildDistributedWorkload(name, namespace, uid string) (*diwv2.DistributedWorkload, *corev1.Pod, *corev1.Pod) {
	workload := &diwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
		Spec: dp.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: constants.RunaiWorkloadRelevantUsage,
				},
			},
		},
	}
	pod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-0",
			Namespace:   namespace,
			UID:         types.UID(uuid.New().String()),
			Labels:      map[string]string{commonconstants.TopOwnerUidLabel: uid},
			Annotations: map[string]string{"runai-job-id": uid},
		},
		Spec: corev1.PodSpec{},
	}
	pod2 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-1",
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels: map[string]string{
				commonconstants.TopOwnerUidLabel:        uid,
				commonconstants.TopOwnerWorkloadIdLabel: uid,
			},
		},
		Spec: corev1.PodSpec{},
	}
	return workload, pod1, pod2
}

func BuildDistributedInferenceWorkload(name, namespace, uid string) (*dinfwv1.DistributedInferenceWorkload, *corev1.Pod, *corev1.Pod) {
	workloadUid := types.UID(uid)
	workload := &dinfwv1.DistributedInferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       workloadUid,
		},
		Spec: dinfp.DistributedInferenceSpec{
			IdentityFields: field_groups.IdentityFields{
				Usage: constants.RunaiWorkloadRelevantUsage,
			},
		},
	}
	pod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-leader",
			Namespace:   namespace,
			UID:         types.UID(uuid.New().String()),
			Labels:      map[string]string{commonconstants.TopOwnerUidLabel: uid},
			Annotations: map[string]string{"runai-job-id": uid},
		},
		Spec: corev1.PodSpec{},
	}
	pod2 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-worker-0",
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels: map[string]string{
				commonconstants.TopOwnerUidLabel:        uid,
				commonconstants.TopOwnerWorkloadIdLabel: uid,
			},
		},
		Spec: corev1.PodSpec{},
	}
	return workload, pod1, pod2
}

func BuildNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
func BuildExcludedNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v1alpha1.NodePoolNameLabel: clustercommon.ExcludedNodepoolName,
			},
		},
	}
}

func BuildProject(name, id, namespace string) *v2.Project {
	return &v2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v2.ProjectIdLabel: id,
			},
		},
		Status: v2.ProjectStatus{
			Namespace: namespace,
		},
	}
}

func BuildProjectWithDepId(name, id, namespace, depId string) *v2.Project {
	return &v2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v2.ProjectIdLabel:         id,
				schedv2.DepartmentIdLabel: depId,
			},
		},
		Status: v2.ProjectStatus{
			Namespace: namespace,
		},
	}
}

// BuildKaiProject builds a kai.resources project fixture.
func BuildKaiProject(name, id, namespace string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v2.ProjectIdLabel: id,
			},
		},
		Status: kaiorg.ProjectStatus{
			Namespace: namespace,
		},
	}
}

// BuildKaiProjectWithDepId mirrors BuildProjectWithDepId as a kai.resources project.
func BuildKaiProjectWithDepId(name, id, namespace, depId string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				v2.ProjectIdLabel:         id,
				schedv2.DepartmentIdLabel: depId,
			},
		},
		Status: kaiorg.ProjectStatus{
			Namespace: namespace,
		},
	}
}

func BuildDepartment(name string, id string) *schedv2.Department {
	return &schedv2.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				schedv2.DepartmentIdLabel: id,
			},
		},
		Status: schedv2.DepartmentStatus{
			SyncStatus: schedv2.DepartmentSyncStatus{LastSyncTimestamp: metav1.Time{}},
		},
	}
}

// BuildKaiDepartment builds a kai.resources department fixture.
func BuildKaiDepartment(name string, id string) *kaiorg.Department {
	return &kaiorg.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				clustercommon.SyncedDepartmentIdLabel: id,
			},
		},
	}
}

func BuildPriorityClass(name string, id string, value int32, runaiSource bool) *schedulingv1.PriorityClass {
	if runaiSource {
		return &schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					clustercommon.PriorityClassIdLabel:     id,
					clustercommon.PriorityClassSourceLabel: clustercommon.PriorityClassRunaiSource,
				},
			},
			Value: value,
		}
	}
	return &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Value: value,
	}
}

func BuildApplication(name string, id types.UID, source string, appType string, departmentId string, departmentName string, projectId string, projectName string, phase string, resourceId string) *v1alpha1.Application {
	appResource := v1alpha1.ApplicationResource{
		Name:             "pvc-1",
		Kind:             "pvc",
		UID:              resourceId,
		ControlPlaneName: "pvc-1-cp",
		ControlPlaneId:   fmt.Sprintf("%s%s", resourceId, resourceId),
	}
	return &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			UID:  id,
			Name: name,
			Labels: map[string]string{
				clustercommon.DepartmentIdLabel:   departmentId,
				clustercommon.DepartmentNameLabel: departmentName,
				clustercommon.ProjectIdLabel:      projectId,
				clustercommon.ProjectNameLabel:    projectName,
			},
		},
		Spec: v1alpha1.ApplicationSpec{
			Source: source,
			Type:   appType,
		},
		Status: v1alpha1.ApplicationStatus{
			K8sPhase:  phase,
			Resources: []v1alpha1.ApplicationResource{appResource},
		},
	}
}

func BuildEvent(namespace, name string) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(name),
			Name:      name,
			Namespace: namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			UID: types.UID(uuid.New().String()),
		},
	}
}

func BuildMigPods() []*corev1.Pod {
	pod1 := BuildMigRunaiPod(
		"c1", "p1", "i1", "n1", corev1.PodRunning, BuildResourceList("1000m", "1G"),
		[]metav1.OwnerReference{}, make(map[string]string), migValidNameOne, 2)
	pod2 := BuildMigRunaiPod(
		"c1", "p2", "i1", "n1", corev1.PodRunning, BuildResourceList("2000m", "2G"),
		[]metav1.OwnerReference{}, make(map[string]string), migValidNameOne, 1)
	pod3 := BuildMigRunaiPod(
		"c1", "p3", "i1", "n1", corev1.PodRunning, BuildResourceList("2000m", "2G"),
		[]metav1.OwnerReference{}, make(map[string]string), migValidNameTwo, 1)
	return []*corev1.Pod{pod1, pod2, pod3}
}

func BuildNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func ReadFromChannel[T any](ch <-chan T, numOfTimesToRead int) []T {
	result := make([]T, 0, numOfTimesToRead)
	for i := 0; i < numOfTimesToRead; i++ {
		value, ok := <-ch
		if !ok {
			break
		}
		result = append(result, value)
	}
	return result
}

const ClusterIdForTests = "7daba6ee-2ea9-4af3-b735-a364bcdff72f"
const ClusterDomainForTests = "some-url.run.ai"
const ProjectIdForTests = "project"
const NamespaceIdForTests = "runai-namespace-project"
const ClientNameForTests = "clientName"
const ClientSecretForTests = "clientSecret"
const syncEventsIntervalSeconds = 2
const numOfPodsPerRequest = 3
const numOfWorkloadsPerRequest = 4
const numOfEventsPerRequest = 5
const namespaceForTest = "namespace"
const namespaceForTest2 = "runai-project2"
const projectForTest = "project"
const projectForTest2 = "project2"

// MockEnv Notice that after test is done env variables are cleaned up
func MockEnv(t *testing.T) {
	t.Setenv("CLUSTER_DOMAIN", ClusterDomainForTests)
	t.Setenv("ENDPOINT", "127.0.0.1")
	t.Setenv("LIVENESS_INTERVAL_SECONDS", "1")
	t.Setenv("MAX_NUM_OF_EVENTS_PER_REQUEST", fmt.Sprintf("%d", numOfEventsPerRequest))
	t.Setenv("SYNC_EVENTS_INTERVAL_SECONDS", fmt.Sprintf("%d", syncEventsIntervalSeconds))
	t.Setenv("SSL_CERTIFICATE_VERIFICATION", "false")
	t.Setenv("MAX_NUM_OF_PODS_PER_REQUEST", fmt.Sprintf("%d", numOfPodsPerRequest))
	t.Setenv("MAX_NUM_OF_WORKLOADS_PER_REQUEST", fmt.Sprintf("%d", numOfWorkloadsPerRequest))
	t.Setenv("ENABLE_NODE_NVIDIA_COMPONENTS_READINESS", "true")

	t.Setenv("RATE_LIMITER_WORKLOADS_DELAY_SECONDS", "0")
	t.Setenv("RATE_LIMITER_WORKLOADS_NUM_OF_WORKERS", "1")
	t.Setenv("RATE_LIMITER_REGULAR_DELAY_SECONDS", "0")
	t.Setenv("RATE_LIMITER_REGULAR_NUM_OF_WORKERS", "1")
	t.Setenv("SYNC_PODS_WITH_WORKLOADS", "true")

	c, err := rcpconfig.GetR[*config.ClusterSyncConfig]()
	if err != nil {
		t.Fatalf("Failed to get agent config: %v", err)
	}

	c.ClusterUuid = ClusterIdForTests
	c.ClientName = ClientNameForTests
	c.ClientSecret = ClientSecretForTests

	err = rcpconfig.Set(c)
	if err != nil {
		t.Fatalf("Failed to set agent config: %v", err)
	}

}

func Int32Ptr(i int32) *int32 { return &i }

func GetTestScheme() *runtime.Scheme {
	once.Do(func() {
		scheme = scheme_init.Scheme()
		_ = knv1.AddToScheme(scheme)
		_ = diwv2.AddToScheme(scheme)
		_ = twv2.AddToScheme(scheme)
		_ = iwv2.AddToScheme(scheme)
		_ = infwv2.AddToScheme(scheme)
		_ = dinfwv1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		_ = kartav1alpha1.AddToScheme(scheme)
		_ = kaiorg.AddToScheme(scheme)
		_ = nrtv1alpha2.AddToScheme(scheme)
	})
	return scheme
}

// SetIssuesOnNamespace set clone issues on the object's annotations and return expected issues list
func SetIssuesOnNamespace(obj *metav1.ObjectMeta, namespaces ...string) []k8sobjectstracker.Issue {
	expectedIssues := make([]k8sobjectstracker.Issue, 0, len(namespaces))
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string, len(namespaces))
	}
	for _, ns := range namespaces {
		annotations[fmt.Sprintf("clone.run.ai/%s", ns)] = string(k8sobjectstracker.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR)
		expectedIssues = append(expectedIssues, k8sobjectstracker.Issue{
			Namespace: ns,
			Issue:     string(k8sobjectstracker.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR),
		})
	}
	obj.SetAnnotations(annotations)
	return expectedIssues
}
