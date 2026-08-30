/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InteractiveWorkload is a schema for the interactive workload API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="iw"
// +kubebuilder:subresource:status
// +kubebuilder:unservedversion
// +kubebuilder:printcolumn:name="Based On",type=string,JSONPath=`.spec.baseWorkload`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.creationStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type InteractiveWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//  The specifications of this InteractiveWorkload
	Spec ipv1.InteractiveSpec `json:"spec,omitempty"`

	//  The observed state of this InteractiveWorkload
	Status v2alpha1.InteractiveWorkloadStatus `json:"status,omitempty"`
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

func init() {
	SchemeBuilder.Register(&InteractiveWorkload{}, &InteractiveWorkloadList{})
}
