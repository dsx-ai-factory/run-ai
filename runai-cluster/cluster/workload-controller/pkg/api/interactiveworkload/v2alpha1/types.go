/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InteractiveWorkload is a schema for the interactive workload API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="iw"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Based On",type=string,JSONPath=`.spec.baseWorkload`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type InteractiveWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//  The specifications of this InteractiveWorkload
	Spec ipv1.InteractiveSpec `json:"spec,omitempty"`

	//  The observed state of this InteractiveWorkload
	Status InteractiveWorkloadStatus `json:"status,omitempty"`
}

func (self *InteractiveWorkloadStatus) DeepCopyToJobStatus() jobs.JobStatus {
	syncStatus := &common.WorkloadSyncStatus{}
	syncStatus = self.SyncStatus.DeepCopy()
	return jobs.JobStatus{
		WorkloadCommonStatus: self.WorkloadCommonStatus,
		WorkloadDebugStatus:  self.WorkloadDebugStatus,
		WorkloadStatus:       self.WorkloadStatus,
		NetworkStatus:        self.NetworkStatus,
		SyncStatus:           syncStatus,
	}
}

func (self *InteractiveWorkloadStatus) DeepCopyToInteractiveStatus(status jobs.JobStatus) {
	// Deep copy SyncStatus, handling potential nil value
	if status.SyncStatus != nil {
		self.SyncStatus = &common.WorkloadSyncStatus{
			LastSyncTimestamp:                 status.SyncStatus.LastSyncTimestamp,
			LastSyncGeneration:                status.SyncStatus.LastSyncGeneration,
			LastSyncStatusChangeTime:          status.SyncStatus.LastSyncStatusChangeTime,
			LastSyncStatusChecksum:            status.SyncStatus.LastSyncStatusChecksum,
			LastAttemptedSyncGeneration:       status.SyncStatus.LastAttemptedSyncGeneration,
			LastAttemptedSyncStatusChangeTime: status.SyncStatus.LastAttemptedSyncStatusChangeTime,
			SyncErrorMsg:                      status.SyncStatus.SyncErrorMsg,
		}
	} else {
		self.SyncStatus = nil // Or set a default value if appropriate
	}

	// Copy other fields directly
	self.WorkloadCommonStatus = status.WorkloadCommonStatus
	self.WorkloadDebugStatus = status.WorkloadDebugStatus
	self.WorkloadStatus = status.WorkloadStatus
	self.NetworkStatus = status.NetworkStatus
	self.MinRequiredReadyPods = status.MinRequiredReadyPods
}

// InteractiveWorkloadStatus unifies interactive and training status
// +k8s:deepcopy-gen=true
type InteractiveWorkloadStatus struct {
	common.WorkloadCommonStatus `json:""`
	common.WorkloadDebugStatus  `json:""`
	common.WorkloadStatus       `json:""`
	common.NetworkStatus        `json:",inline"`
	MinRequiredReadyPods        int32 `json:"minRequiredReadyPods,omitempty"`

	// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
	SyncStatus *common.WorkloadSyncStatus `json:"syncStatus,omitempty"`
}

// NewFromJobWorkload copy generic job (interactive/training) to InteractiveWorkload
func NewFromJobWorkload(jobWorkload *jobs.JobWorkload) *InteractiveWorkload {
	if jobWorkload != nil {
		result := &InteractiveWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       constants.KindInteractiveWorkload,
				APIVersion: GroupVersion.String(),
			},
		}

		jobWorkload.ObjectMeta.DeepCopyInto(&result.ObjectMeta)
		ipv1.LoadFromJobSpec(&jobWorkload.Spec, &result.Spec)

		// copy the status
		result.Status.SyncStatus = jobWorkload.Status.SyncStatus
		result.Status.WorkloadCommonStatus = jobWorkload.Status.WorkloadCommonStatus
		result.Status.WorkloadDebugStatus = jobWorkload.Status.WorkloadDebugStatus
		result.Status.WorkloadStatus = jobWorkload.Status.WorkloadStatus
		result.Status.NetworkStatus = jobWorkload.Status.NetworkStatus
		result.Status.MinRequiredReadyPods = jobWorkload.Status.MinRequiredReadyPods
		return result
	}

	return nil
}

func (self *InteractiveWorkload) CopyTo(jobWorkload *jobs.JobWorkload) {

	if jobWorkload != nil {
		jobWorkload.TypeMeta = metav1.TypeMeta{
			Kind:       constants.KindInteractiveWorkload,
			APIVersion: GroupVersion.String(),
		}
		self.ObjectMeta.DeepCopyInto(&jobWorkload.ObjectMeta)
		self.Spec.CopyTo(&jobWorkload.Spec)

		// copy the status
		jobWorkload.Status.SyncStatus = self.Status.SyncStatus
		jobWorkload.Status.WorkloadCommonStatus = self.Status.WorkloadCommonStatus
		jobWorkload.Status.WorkloadDebugStatus = self.Status.WorkloadDebugStatus
		jobWorkload.Status.WorkloadStatus = self.Status.WorkloadStatus
		jobWorkload.Status.NetworkStatus = self.Status.NetworkStatus

	}
}

// InteractiveWorkloadList contains a list of InteractiveWorkload(s)
// +kubebuilder:object:root:=true
type InteractiveWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InteractiveWorkload `json:"items"`
}

func (self *InteractiveWorkloadList) AsJobWorkLoad() []jobs.JobWorkload {

	result := make([]jobs.JobWorkload, len(self.Items))
	for idx := 0; idx < len(self.Items); idx++ {
		self.Items[idx].CopyTo(&result[idx])
	}
	return result
}

func (self *InteractiveWorkloadList) GetItems() []client.Object {
	objects := make([]client.Object, len(self.Items))
	for idx := 0; idx < len(self.Items); idx++ {
		objects[idx] = &self.Items[idx]
	}
	return objects
}

func (self *InteractiveWorkloadList) GetWorkloadKind() string {
	return constants.KindInteractiveWorkload
}

func init() {
	SchemeBuilder.Register(&InteractiveWorkload{}, &InteractiveWorkloadList{})
}

func GetNewInteractiveWorkloadFn() client.Object {
	return &InteractiveWorkload{}
}
