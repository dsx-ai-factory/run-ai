package common

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BytesToMegaFactor = 1000000

	TotalGPUsAnnotationKey       = "totalGPUs"
	TotalGPUsMemoryAnnotationKey = "totalGPUsMemory"
)

var FixedDate = metav1.NewTime(time.Date(2022, 2, 22, 17, 0, 0, 0, time.UTC))

// DistributedBuilder - common builder for distributed jobs
type DistributedBuilder struct {
	factory.Builder
	common.OldSharedBuilder
	//  each job has its own container name (mpi, pytorch, etc).
	ContainerName   string
	ExisingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec
	//  the input to the builder is the merged workload
	MergedWorkload *v2alpha1.DistributedWorkload
	PodLabels      map[string]string
	PodAnnotations map[string]string
}

func (db *DistributedBuilder) GetChildName() string {
	return db.MergedWorkload.GetChildName()
}

func (db *DistributedBuilder) BuildObjectMeta(ctx context.Context, om *metav1.ObjectMeta) error {
	db.OldSharedBuilder.BuildObjectMeta(ctx, om, &db.MergedWorkload.TypeMeta)

	om.Name = db.GetChildName()

	om.Annotations[constants.AnnotationImage] = db.MergedWorkload.Spec.Image.Deref()
	om.Annotations[constants.AnnotationUserName] = db.MergedWorkload.Spec.UserName.Deref()
	om.Annotations[constants.AnnotationTemplate] = db.MergedWorkload.Name

	return nil
}

func (db *DistributedBuilder) GetObjectMetaLabels(ctx context.Context, labels map[string]string) error {
	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err != nil {
		return err
	}

	labels[constants.RunaiProjectLabel] = projectName
	labels[constants.LabelRelease] = db.GetChildName()

	priorityClassNameLabel := calculateDistributedPriorityClassNameLabel(db.MergedWorkload.Spec)
	err = workloadbuilders.SetPriorityClassLabelOnLabels(ctx, db.K8sClient, priorityClassNameLabel, labels)
	if err != nil {
		return err
	}

	calculatedPreemptibilityNameLabel := common.CalculatePreemptibilityNameLabel(db.MergedWorkload.Spec.Preemptibility, db.MergedWorkload.Kind)
	if calculatedPreemptibilityNameLabel != "" {
		labels[constants.LabelPreemptibilityName] = calculatedPreemptibilityNameLabel
	}

	labels[commonconstants.TopOwnerUidLabel] = topowner.GetRunaiWorkloadId(db.MergedWorkload.UID, db.MergedWorkload.Annotations)

	return nil
}

func TotalGpuPods(workload *v2alpha1.DistributedWorkload) (int, int) {
	masterPods := 1
	if workload.Spec.DistributedFields.NoMaster.Deref() {
		masterPods = 0
	}
	return masterPods, workload.Spec.DistributedFields.Workers.Deref()
}

func TotalGpus(workload *v2alpha1.DistributedWorkload) float64 {
	workerGpus := workload.Spec.Gpu.Deref()
	var masterGpus float64
	if workload.MasterSpec != nil {
		masterGpus = workload.MasterSpec.Gpu.Deref()
	}

	masterPods, workerPods := TotalGpuPods(workload)

	return (float64(masterPods) * float64(masterGpus)) + (float64(workerPods) * float64(workerGpus))
}

func TotalGpusMemory(workload *v2alpha1.DistributedWorkload) int64 {
	workerMemory := int64(0)
	workerMemoryQuantity, err := resource.ParseQuantity(workload.Spec.GpuMemory.Deref())
	if err == nil {
		workerMemory = workerMemoryQuantity.Value()
	}

	masterMemory := int64(0)
	if workload.MasterSpec != nil {
		masterMemoryQuantity, err := resource.ParseQuantity(workload.MasterSpec.GpuMemory.Deref())
		if err == nil {
			masterMemory = masterMemoryQuantity.Value()
		}
	}

	masterPods, workerPods := TotalGpuPods(workload)

	return ((int64(masterPods) * masterMemory) + (int64(workerPods) * workerMemory)) / BytesToMegaFactor
}

func BuildAnnotationsAndLabels(ctx context.Context, k8sClient client.Client, objectMeta *metav1.ObjectMeta, workload *v2alpha1.DistributedWorkload, gpu *flds.NumberField, memory *flds.QuantityField, gpuDevices *flds.IntField, podLabels map[string]string, podAnnotations map[string]string, schedulerFields field_groups.RunaiSchedulerFields) error {
	var memoryQuantity resource.Quantity

	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err != nil {
		return err
	}

	if objectMeta.Labels == nil {
		objectMeta.Labels = map[string]string{}
	}

	for key, value := range podLabels {
		objectMeta.Labels[key] = value
	}

	objectMeta.Labels[constants.RunaiProjectLabel] = projectName

	// workload name and kind labels are expected also in the pod labels, not only in the workload labels
	objectMeta.Labels[constants.LabelWorkloadName] = workload.ObjectMeta.Name
	objectMeta.Labels[constants.LabelWorkloadKind] = constants.KindDistributedWorkload
	objectMeta.Labels[commonconstants.TopOwnerUidLabel] = topowner.GetRunaiWorkloadId(workload.GetUID(), workload.GetAnnotations())

	priorityClassNameLabel := calculateDistributedPriorityClassNameLabel(workload.Spec)
	err = workloadbuilders.SetPriorityClassLabel(ctx, k8sClient, priorityClassNameLabel, objectMeta)
	if err != nil {
		return err
	}

	calculatedPreemptibilityNameLabel := common.CalculatePreemptibilityNameLabel(workload.Spec.Preemptibility, workload.Kind)
	workloadbuilders.SetPreemptibilityLabel(calculatedPreemptibilityNameLabel, objectMeta)

	// set annotations

	if objectMeta.Annotations == nil {
		objectMeta.Annotations = map[string]string{}
	}

	for key, value := range podAnnotations {
		objectMeta.Annotations[key] = value
	}

	objectMeta.Annotations[constants.AnnotationUserName] = workload.Spec.UserName.Deref()

	if memory.IsValueInitialized() {
		memoryQuantity, err = memory.DerefQuantity(0)
		if err != nil {
			return fmt.Errorf("invalid gpuMemory: %s", err.Error())
		}
	}

	if workloadbuilders.IsFractionalGpuOrMemroy(ctx, gpu.Deref(), memoryQuantity) {
		workloadbuilders.AddPartialGpuAnnotations(objectMeta.Annotations, gpu, memoryQuantity, gpuDevices.Deref())
	}

	totalGpus := TotalGpus(workload)
	totalGpus = math.Round(totalGpus*100) / 100 // round to two decimal places, for pretty printing
	objectMeta.Annotations[TotalGPUsAnnotationKey] = fmt.Sprintf("%v", totalGpus)
	objectMeta.Annotations[TotalGPUsMemoryAnnotationKey] = strconv.Itoa(int(TotalGpusMemory(workload)))

	// set the used nodepools in an annotations
	nodepoolsValue := field_groups.GetNodepoolsValue(schedulerFields)
	if nodepoolsValue != "" {
		objectMeta.Annotations[commonconstants.AnnotationNodepools] = nodepoolsValue
	}

	return nil
}

func (db *DistributedBuilder) BuildRunPolicy(result *kubeflowcommon.RunPolicy) error {
	runPolicy := db.MergedWorkload.Spec.RunPolicy.Deref()

	cleanPolicy := db.GetDefaultCleanPodPolicy()
	if runPolicy.CleanPodPolicy != nil && runPolicy.CleanPodPolicy.Value != nil {
		cleanPolicy = kubeflowcommon.CleanPodPolicy(*runPolicy.CleanPodPolicy.Value)
	}

	result.CleanPodPolicy = &cleanPolicy

	autoDeletionTimeAfterCompletionSeconds := db.MergedWorkload.Spec.RunaiSchedulerFields.AutoDeletionTimeAfterCompletionSeconds
	if autoDeletionTimeAfterCompletionSeconds != nil && autoDeletionTimeAfterCompletionSeconds.Value != nil {
		// handle int overflow
		ttlSeconds, err := Int32PositiveCast(*autoDeletionTimeAfterCompletionSeconds.Value)
		if err != nil {
			return fmt.Errorf("autoDeletionTimeAfterCompletionSeconds value is non valid: %v", err)
		}
		result.TTLSecondsAfterFinished = ptr.To(ttlSeconds)
	}

	var backoff int32
	var err error

	if runPolicy.BackoffLimit != nil && runPolicy.BackoffLimit.Value != nil {
		backoff, err = Int32PositiveCast(*runPolicy.BackoffLimit.Value)
		if err != nil {
			return fmt.Errorf("backoffLimit value is non valid: %v", err)
		}
	} else {
		numOfWorkers := db.MergedWorkload.Spec.DistributedFields.Workers.Deref32()
		backoff = numOfWorkers * constants.DefaultBackoffLimit
	}
	result.BackoffLimit = ptr.To(backoff)

	if !db.IsWorkloadActive(db.MergedWorkload.Spec.StateFields) {
		// suspend option is false by default
		result.Suspend = ptr.To(true)
	}

	return nil
}

func (db *DistributedBuilder) NumberOfWorkerReplicas() *int32 {
	result := int32(constants.DefaultWorkers)
	if db.MergedWorkload.Spec.Workers.Deref() != 0 {
		result = int32(db.MergedWorkload.Spec.Workers.Deref())
	}
	return &result
}

// GetDefaultCleanPodPolicy by default, do not delete pods of completed job, except for MPI, TF (parameters server)
func (db *DistributedBuilder) GetDefaultCleanPodPolicy() kubeflowcommon.CleanPodPolicy {
	if db.MergedWorkload.Spec.JobType == constants.KindTFJob || db.MergedWorkload.Spec.JobType == constants.KindMpiJob {
		return kubeflowcommon.CleanPodPolicyRunning
	}

	return kubeflowcommon.CleanPodPolicyNone
}

func calculateDistributedPriorityClassNameLabel(workloadSpec dp.DistributedSpec) string {
	priorityClassName := workloadSpec.PriorityClassName.Deref()
	if priorityClassName != "" {
		return priorityClassName
	}

	if workloadSpec.NonPreemptible.Deref() {
		return constants.InteractivePriorityClass
	}

	return ""
}

func (db *DistributedBuilder) GetTopologyState(ctx context.Context) *topology.State {
	if db.Workload == nil {
		return topology.DisabledState()
	}

	return topology.GetState(ctx, db.Workload.Annotations)
}
