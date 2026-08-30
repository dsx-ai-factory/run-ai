/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ip "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:object:root:=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName="infw"
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="url",type="string",JSONPath=".status.externalUrls[-1]"
// +kubebuilder:printcolumn:name="latest created",type="string",JSONPath=".status.latestCreatedRevisionName"
// +kubebuilder:printcolumn:name="latest ready",type="string",JSONPath=".status.latestReadyRevisionName"
// Schema for inference workload API
type InferenceWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//  The specifications of this workload
	Spec ip.InferenceSpec `json:"spec,omitempty"`

	//  The observed state of this workload
	Status InferenceWorkloadStatus `json:"status,omitempty"`
}

type InferenceWorkloadStatus struct {
	common.WorkloadDebugStatus `json:""`
	common.NetworkStatus       `json:",inline"`
	common.InferenceStatus     `json:""`
	// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
	SyncStatus *common.WorkloadSyncStatus `json:"syncStatus,omitempty"`
}

// +kubebuilder:object:root:=true
// InferenceWorkloadList contains a list of InferenceWorkload(s)
type InferenceWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceWorkload `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceWorkload{}, &InferenceWorkloadList{})
}

// Builder return the factory of the inference builder
func (iw *InferenceWorkload) Builder(options factory.BuilderOptions, k8sClient client.Client,
	podLabels map[string]string, podAnnotations map[string]string) factory.Builder {
	builderConstructor := factory.BuildersRegistry[factory.InferenceBuilder]
	return builderConstructor(iw, &servingv1.Service{}, k8sClient, options, podLabels, podAnnotations)
}

func GetNewInferenceWorkloadFn() client.Object {
	return &InferenceWorkload{}
}

// GetChildName returns the name of the workload's children resources (e.g. run-ai jobs, services)
func (iw *InferenceWorkload) GetChildName() string {
	name := iw.ObjectMeta.Name
	// The job name should be metadata.Name, however for backward compatibility old workload uses sepc.Name
	// Therefor need to check if spec.Name isn't empty and use it instead of metadata.Name
	if iw.Spec.Name != nil && iw.Spec.Name.Deref() != "" {
		name = iw.Spec.Name.Deref()
	}
	return name
}

// NewFromJobWorkload copy generic job (interactive/training) to InferenceWorkload
func NewFromJobWorkload(jobWorkload *jobs.JobWorkload) *InferenceWorkload {
	if jobWorkload != nil {
		result := &InferenceWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       clusterconstants.KindInferenceWorkload,
				APIVersion: GroupVersion.String(),
			},
		}

		jobWorkload.ObjectMeta.DeepCopyInto(&result.ObjectMeta)
		ip.LoadFromJobSpec(&jobWorkload.Spec, &result.Spec)
		result.Status = InferenceWorkloadStatus{
			WorkloadDebugStatus: jobWorkload.Status.WorkloadDebugStatus,
			NetworkStatus:       jobWorkload.Status.NetworkStatus,
			InferenceStatus:     jobWorkload.Status.InferenceStatus,
			SyncStatus:          jobWorkload.Status.SyncStatus,
		}

		return result
	}

	return nil
}

func (self *InferenceWorkload) CopyTo(jobWorkload *jobs.JobWorkload) {

	if jobWorkload != nil {
		jobWorkload.TypeMeta = metav1.TypeMeta{
			Kind:       clusterconstants.KindInferenceWorkload,
			APIVersion: GroupVersion.String(),
		}
		self.ObjectMeta.DeepCopyInto(&jobWorkload.ObjectMeta)
		self.Spec.CopyTo(&jobWorkload.Spec)

		// Copy status
		jobWorkload.Status.WorkloadDebugStatus = self.Status.WorkloadDebugStatus
		jobWorkload.Status.NetworkStatus = self.Status.NetworkStatus
		jobWorkload.Status.InferenceStatus = self.Status.InferenceStatus
		jobWorkload.Status.SyncStatus = self.Status.SyncStatus

	}
}
