/*
  Copyright (c) 2025, NVIDIA CORPORATION.  All rights reserved.

  NVIDIA CORPORATION and its licensors retain all intellectual property
  and proprietary rights in and to this software, related documentation
  and any modifications thereto.  Any use, reproduction, disclosure or
  distribution of this software and related documentation without an express
  license agreement from NVIDIA CORPORATION is strictly prohibited.
*/

// TODO: Remove this file and import the Heimdall CRD Go types directly when
// they are available as a Go module.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdviceJobReference contains information about the job this advice applies to
type AdviceJobReference struct {
	// HeimdallID is the unique identifier assigned by Heimdall for the job
	// +kubebuilder:validation:MinLength=1
	// +required
	HeimdallID string `json:"heimdallId"`

	// APIVersion is the apiVersion of the job resource
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind is the kind of the job resource
	// +optional
	Kind string `json:"kind,omitempty"`

	// Name is the Kubernetes resource name for the job
	// +optional
	Name string `json:"name,omitempty"`

	// JobID is the scheduler-specific identifier for the job
	// +kubebuilder:validation:MinLength=1
	// +required
	JobID string `json:"jobId"`

	// AttemptIndex is the attempt/restart index this advice applies to
	// +kubebuilder:validation:Minimum=0
	// +required
	AttemptIndex int32 `json:"attemptIndex"`
}

// JobState represents the recommended state of the job
// +kubebuilder:validation:Enum=ADVICE_STATE_TERMINATED;ADVICE_STATE_RESTARTED
type JobState string

const (
	JobStateTerminated JobState = "ADVICE_STATE_TERMINATED"
	JobStateRestarted  JobState = "ADVICE_STATE_RESTARTED"
)

// NodeConstraints contains scheduling constraints for the job
type NodeConstraints struct {
	// Exclusions is a list of nodes the job should be excluded from
	// +optional
	Exclusions []string `json:"exclusions,omitempty"`

	// Inclusions is a list of nodes the job should prefer for scheduling
	// +optional
	Inclusions []string `json:"inclusions,omitempty"`
}

// AdviceRecommendation contains the recommended actions for the job
type AdviceRecommendation struct {
	// State is the recommended state of the job (null, Terminated, or Restarted)
	// +optional
	State *JobState `json:"state,omitempty"`

	// NodeConstraints contains scheduling constraints for the job
	// +optional
	NodeConstraints *NodeConstraints `json:"nodeConstraints,omitempty"`
}

// ResourceDetail describes a specific resource involved in a remediation action
type ResourceDetail struct {
	// ResourceType is the type of resource (Node, Network, GPU, etc)
	// +kubebuilder:validation:MinLength=1
	// +required
	ResourceType string `json:"resourceType"`

	// ResourceId is the identifier for the resource
	// +kubebuilder:validation:MinLength=1
	// +required
	ResourceId string `json:"resourceId"`

	// Details provides optional additional information about the resource
	// +optional
	Details string `json:"details,omitempty"`
}

// RemediationAction describes a specific action to take to remediate an issue
type RemediationAction struct {
	// ActionType describes the type of action (Restart, Restart and Include Node, Restart and Exclude Node, Stop)
	// +kubebuilder:validation:MinLength=1
	// +required
	ActionType string `json:"actionType"`

	// Timestamp is when this action was recommended (ISO8601 format)
	// +kubebuilder:validation:Format="date-time"
	// +required
	Timestamp string `json:"timestamp"`

	// Reason explains why this action is being recommended
	// +kubebuilder:validation:MinLength=1
	// +required
	Reason string `json:"reason"`

	// ResourceDetails contains information about resources involved in this action
	// +optional
	ResourceDetails []ResourceDetail `json:"resourceDetails,omitempty"`
}

// Issue describes a detected issue for the job
type Issue struct {
	// IssueType categorizes the issue (CRASH, HANG, STRAGGLER, etc)
	// +kubebuilder:validation:MinLength=1
	// +required
	IssueType string `json:"issueType"`

	// Timestamp is when this issue was detected (ISO8601 format)
	// +kubebuilder:validation:Format="date-time"
	// +required
	Timestamp string `json:"timestamp"`

	// DetectedBy indicates what system detected this issue
	// +kubebuilder:validation:MinLength=1
	// +required
	DetectedBy string `json:"detectedBy"`

	// RelevantLogLines contains log lines that are relevant to this issue
	// +optional
	RelevantLogLines []string `json:"relevantLogLines,omitempty"`

	// RemediationActions contains the list of recommended remediation actions
	// +optional
	RemediationActions []RemediationAction `json:"remediationActions,omitempty"`
}

// AdviceState represents the state of the advice
// +kubebuilder:validation:Enum=Pending;Ready;Applied;Invalid
type AdviceState string

const (
	AdviceStatePending AdviceState = "Pending"
	AdviceStateReady   AdviceState = "Ready"
	AdviceStateApplied AdviceState = "Applied"
	AdviceStateInvalid AdviceState = "Invalid"
)

// AdviceSpec defines the desired state of Advice
type AdviceSpec struct {
	// Job contains information about the job this advice applies to
	// +required
	Job AdviceJobReference `json:"job"`

	// Advice contains the recommended actions for the job
	// +required
	Advice AdviceRecommendation `json:"advice"`

	// Issues is the list of issues detected for the job
	// +optional
	Issues []Issue `json:"issues,omitempty"`
}

// AdviceStatus defines the observed state of Advice.
type AdviceStatus struct {
	// State represents the current state of the advice
	// +optional
	State AdviceState `json:"state,omitempty"`

	// Conditions represent the latest available observations of the Advice's state
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="HeimdallID",type=string,JSONPath=`.spec.job.heimdallId`
// +kubebuilder:printcolumn:name="APIVersion",type=string,JSONPath=`.spec.job.apiVersion`
// +kubebuilder:printcolumn:name="Kind",type=string,JSONPath=`.spec.job.kind`
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.job.name`
// +kubebuilder:printcolumn:name="JobID",type=string,JSONPath=`.spec.job.jobId`
// +kubebuilder:printcolumn:name="AttemptIndex",type=integer,JSONPath=`.spec.job.attemptIndex`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Advice is the Schema for the advices API
type Advice struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// +required
	Spec AdviceSpec `json:"spec"`

	// +optional
	Status AdviceStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// AdviceList contains a list of Advice
type AdviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Advice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Advice{}, &AdviceList{})
}
