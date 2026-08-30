package common

import (
	"fmt"
	"time"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadConditionType string

const (
	RunaiCurrentAllocatedGpuMemoryAnnotation = "runai-current-allocated-gpus-memory"
	RunaiCurrentAllocatedGpusAnnotation      = "runai-current-allocated-gpus"
)

const (
	// LastReconcileSucceeded means the last reconciling was successful or not
	LastReconcileSucceeded WorkloadConditionType = "LastReconcileSucceeded"
	// Suspended means the workload is suspended
	Suspended = "Suspended"
	// Succeeded means all the jobs of this workload finished successfully
	Succeeded = "Succeeded"
	// Failed means one or more sub-resource (job/services/ingress) of this workload reached phase failed
	Failed = "Failed"
	// Ready means all sub-resources (services/jobs/etc) of this workload reached ready phase
	Ready = "Ready"
	// Initialized means the workload has the minimum required ready replicas to consider it initialized and running
	Initialized = "Initialized"
	// Degraded means workload is running but with issues or reduced resources.
	Degraded = "Degraded"
	// Pending means one or more pods of this workload is in pending
	Pending = "Pending"
	// GPUNetworkAccelerated means the workload use GPU Network Acceleration (MNNVL) feature
	GPUNetworkAccelerated = "GPUNetworkAccelerated"
)

const (
	WorkloadBuildFailedReason = "FailedBuildingWorkloadResources"
	WorkloadFailedReason      = "WorkloadFailed"
	CleanupFailedReason       = "WorkloadCleanupFailed"
	FailedReconcileReason     = "FailedReconcile"
	PodsStillRunningReason    = "RunningPodsExist"
)

// +k8s:deepcopy-gen=true
type WorkloadCondition struct {
	// Type of workload condition.
	Type WorkloadConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=WorkloadConditionType"`
	//These are valid condition statuses.
	//"ConditionTrue" means a resource is in the condition.
	//"ConditionFalse" means a resource is not in the condition.
	//"ConditionUnknown" means kubernetes can't decide if a resource is in the condition or not.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// Issue resource specific details kind/name/description
// +k8s:deepcopy-gen=true
type Issue struct {
	Kind        string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name        string `json:"name" protobuf:"bytes,2,opt,name=name"`
	Description string `json:"description" protobuf:"bytes,3,opt,name=description"`
}

// Issues struct holding the different groups of resource and there issues
// +k8s:deepcopy-gen=true
type Issues struct {
	// Network represents issues found for network resources.
	// +optional
	Network []Issue `json:"network" protobuf:"bytes,1,opt,name=network"`
	// Storage represents issues found for storage resources.
	// +optional
	Storage []Issue `json:"storage" protobuf:"bytes,2,opt,name=storage"`
	// Compute represents issues found for compute resources.
	// +optional
	Compute []Issue `json:"compute"  protobuf:"bytes,3,opt,name=compute"`
	// Internal represents issues found for the workload resource.
	// +optional
	Internal []Issue `json:"internal" protobuf:"bytes,4,opt,name=internal"`
}

// WorkloadDebugStatus is the status for internal use by the workload controller.
// It includes information about the last reconcile time, the first acknowledged time and the owned resources.
// +k8s:deepcopy-gen=true
type WorkloadDebugStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// FirstAcknowledgedTime Represents time when the workload was acknowledged by the workload controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	FirstAcknowledgedTime *metav1.Time       `json:"firstAcknowledgedTime,omitempty" protobuf:"bytes,2,opt,name=firstAcknowledgedTime"`
	OwnedResources        []ObservedResource `json:"ownedResources,omitempty" protobuf:"bytes,3,opt,name=ownedResources"`
	// LastStatusChangeTime is the last time the status of the workload changed.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastStatusChangeTime *metav1.Time `json:"lastStatusChangeTime,omitempty" protobuf:"bytes,4,opt,name=lastStatusChangeTime"`
}

func (self *WorkloadDebugStatus) GetFirstAcknowledgedTime() *metav1.Time {
	if self.FirstAcknowledgedTime != nil {
		return self.FirstAcknowledgedTime.DeepCopy()
	}
	return &metav1.Time{}
}
func (self *WorkloadDebugStatus) SetFirstAcknowledgedTime(time time.Time) {
	ackTime := &metav1.Time{Time: time}
	self.FirstAcknowledgedTime = ackTime

}

// WorkloadCommonStatus
// +k8s:deepcopy-gen=true
type WorkloadCommonStatus struct {
	// Conditions represents the latest available observations of a workload's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []WorkloadCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	// FirstAcknowledgedTime Represents the time when the workload was acknowledged by the workload controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	CompletionTime *metav1.Time `json:"completionTime,omitempty" protobuf:"bytes,2,opt,name=completionTime"`
	// SucceededPods represents the number of pods complete successfully
	SucceededPods int32 `json:"succeededPods,omitempty" protobuf:"varint,3,opt,name=succeededPods"`
	// RunningPods represents the number of pods running, which are pods in running phase, consuming resources
	RunningPods int32 `json:"runningPods,omitempty" protobuf:"varint,4,opt,name=runningPods"`
	// ReadyPods represents the number of ready pods, which are pods with ready condition set to true
	ReadyPods int32 `json:"readyPods,omitempty" protobuf:"varint,5,opt,name=readyPods"`
	// PendingPods represents the number of pods pending
	PendingPods int32 `json:"pendingPods,omitempty" protobuf:"varint,15,opt,name=pendingPods"`
	// FailedPods represents the number of failed pods
	FailedPods int32 `json:"failedPods,omitempty" protobuf:"varint,6,opt,name=failedPods"`
	// TotalAllocatedGPU represents the total number of allocated GPU
	TotalAllocatedGPU string `json:"totalAllocatedGPU,omitempty" protobuf:"bytes,7,opt,name=totalAllocatedGPU"`
	// TotalAllocatedGPUMemory represents the total amount of allocated GPU memory
	TotalAllocatedGPUMemory int64 `json:"totalAllocatedGPUMemory,omitempty" protobuf:"varint,8,opt,name=totalAllocatedGPUMemory"`
	// StartTime represents the time when the workloads job first pod started running
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty" protobuf:"bytes,9,opt,name=startTime"`
	// ExternalUrls represents the external urls of the workload
	ExternalUrls []string `json:"externalUrls,omitempty" protobuf:"bytes,10,opt,name=externalUrls"`
	// +kubebuilder:validation:Enum:=Creating;Initializing;Running;Degraded;Pending;Stopped;Completed;Failed;Terminating;Suspending;Resuming;Suspended;Restarting
	Phase string `json:"phase,omitempty" protobuf:"bytes,11,opt,name=phase"`
	// PhaseMessage A human-readable message indicating details about why the workload is in this phase.
	// +optional
	PhaseMessage string `json:"phaseMessage,omitempty" protobuf:"bytes,12,opt,name=phaseMessage"`
	// RestartCount is the number of restarts performed on this workload (incremented when phase is set to Restarting).
	// +optional
	RestartCount int32 `json:"restartCount,omitempty" protobuf:"varint,14,opt,name=restartCount"`
	// Issues represents the issues found in workload.
	// +optional
	Issues *Issues `json:"issues,omitempty" protobuf:"bytes,13,opt,name=issues"`
}

type Resource struct {
	Kind      metav1.GroupVersionKind `json:"kind,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Namespace string                  `json:"namespace,omitempty"`
}

func (self Resource) String() string {
	return fmt.Sprintf("%s/%s", self.Kind.String(), self.Name)
}

func CommonResourceToResource(resources []workloadbuilders.Resource) []Resource {
	resourcesResult := make([]Resource, len(resources))
	for i, resource := range resources {
		resourcesResult[i] = Resource{
			Kind:      resource.Kind,
			Name:      resource.Name,
			Namespace: resource.Namespace,
		}
	}

	return resourcesResult
}

type ObservedResource struct {
	Resource           `json:",inline"`
	ObservedGeneration int64     `json:"observedGeneration,omitempty"`
	Uid                types.UID `json:"uid,omitempty"`
}

func ObservedResourceFromObject(obj client.Object) ObservedResource {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return ObservedResource{
		Resource: Resource{
			Kind:      metav1.GroupVersionKind(gvk),
			Name:      obj.(metav1.Object).GetName(),
			Namespace: obj.GetNamespace(),
		},
		ObservedGeneration: obj.(metav1.Object).GetGeneration(),
		Uid:                obj.GetUID(),
	}
}

func (self ObservedResource) String() string {
	return fmt.Sprintf("%s/%s/%d", self.Kind.String(), self.Name, self.ObservedGeneration)
}

func (self ObservedResource) IsChanged(generation int64) bool {
	return self.ObservedGeneration != generation
}

func (self *WorkloadCommonStatus) GetCompletionTime() metav1.Time {
	if self.CompletionTime != nil {
		return self.CompletionTime.Rfc3339Copy()
	}
	return metav1.Time{}
}

func (self *WorkloadCommonStatus) SetCompletionTime(time metav1.Time) {
	self.CompletionTime = time.DeepCopy()
}

// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
// +k8s:deepcopy-gen=true
type WorkloadSyncStatus struct {
	// cluster-sync will set this value when syncing the object to control-plane
	LastSyncTimestamp metav1.Time `json:"lastSyncTimestamp,omitempty" protobuf:"bytes,1,opt,name=lastSyncTimestamp"`
	// the last object Generation that was observed and synced to control-plane - "Generation" is only for spec fields
	LastSyncGeneration int64 `json:"lastSyncGeneration,omitempty" protobuf:"bytes,2,opt,name=lastSyncGeneration"`
	// the last status change time that was observed and synced to control-plane.
	LastSyncStatusChangeTime metav1.Time `json:"lastSyncStatusChangeTime,omitempty" protobuf:"bytes,3,opt,name=lastSyncStatusChangeTime"`
	// the last observed and synced status change time (LastStatusChangeTime)
	LastSyncStatusChecksum uint32 `json:"lastSyncStatusChecksum,omitempty" protobuf:"bytes,4,opt,name=lastSyncStatusChecksum"`
	// the last generation that was attempted to be synced - in case sync failed
	// +optional
	LastAttemptedSyncGeneration *int64 `json:"lastAttemptedSyncGeneration,omitempty" protobuf:"bytes,5,opt,name=lastAttemptedSyncGeneration"`
	// the last statusChangeTime that was attempted to be synced - in case sync failed
	// +optional
	LastAttemptedSyncStatusChangeTime *metav1.Time `json:"lastAttemptedSyncStatusChangeTime,omitempty" protobuf:"bytes,6,opt,name=lastAttemptedSyncStatusChangeTime"`
	// in case of an error - a string describing the error
	// +optional
	SyncErrorMsg *string `json:"syncErrorMsg,omitempty" protobuf:"bytes,7,opt,name=syncErrorMsg"`
}

type Reason string

const (
	ReplicaNotReadyReason         Reason = "ReplicaNotReady"
	MinimumReplicasNotReadyReason Reason = "MinimumReplicasNotReady"
	PendingReason                 Reason = "InPendingPhase"
	CreatingReason                Reason = "InCreatingPhase"
	SuspendedReason               Reason = "InSuspendedPhase"
	DegradedReason                Reason = "HasIssues"
)

func (ir Reason) String() string {
	return string(ir)
}
