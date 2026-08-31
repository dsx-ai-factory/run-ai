/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SPOTRequest represents Request Object as obtained from AWS SQS to be applied
// by BART and the progress of the request.
//
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SPOTRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SPOTRequestSpec   `json:"spec,omitempty"`
	Status            SPOTRequestStatus `json:"status,omitempty"`
}

type SPOTMessageAction string

const (
	SPOTCreationAction    SPOTMessageAction = "RequestSpotInstances"
	SPOTTerminationAction SPOTMessageAction = "TerminateInstances"
)

// SPOTRequestSpec defines the desired request to be handled by BART.
//
// +k8s:openapi-gen=true
type SPOTRequestSpec struct {
	RequestID          string                     `json:"requestId,omitempty"`
	MessageReceipt     string                     `json:"sqsMessageReceipt,omitempty"`
	NCAId              string                     `json:"ncaId,omitempty"`
	FunctionID         string                     `json:"functionId,omitempty"`
	FunctionVersionID  string                     `json:"functionVersionId,omitempty"`
	Action             SPOTMessageAction          `json:"action,omitempty"`
	CreationMsgInfo    SPOTCreationMessageInfo    `json:"creationMsgInfo"`
	TerminationMsgInfo SPOTTerminationMessageInfo `json:"terminationMsgInfo"`
	MessageBatchID     string                     `json:"messageBatchId,omitempty"`
}

type RequestStatus string

const (
	// NVCA is waiting to create the instances for the request
	SPOTRequestStatusPending RequestStatus = "RequestPending"
	// NVCA is creating the requested number of Instances for the request
	SPOTRequestStatusInProgress RequestStatus = "RequestInProgress"
	// NVCA has ack'ed the request and attempting Caching setup as requested
	SPOTRequestStatusCachingInProgress RequestStatus = "RequestCachingInProgress"
	// NVCA has created requested number of instances
	SPOTRequestStatusInstancesInProgress RequestStatus = "RequestInstancesInProgress"
	// All requested instances are in Running State
	SPOTRequestStatusCompleted              RequestStatus = "RequestCompleted"
	SPOTRequestStatusCompletionAcknowledged RequestStatus = "RequestCompletionAcknowledged"
	SPOTRequestStatusFailed                 RequestStatus = "RequestFailed"
	SPOTRequestStatusFailureAcknowledged    RequestStatus = "RequestFailureAcknowledged"
)

type InstanceType string

const (
	InstanceTypePod       InstanceType = "Pod"
	InstanceTypeNGCJob    InstanceType = "NGCJob"
	InstanceTypeHelmChart InstanceType = "HelmChart"
)

type InstanceStatus struct {
	ID                    string            `json:"id,omitempty"`
	Type                  InstanceType      `json:"instanceType,omitempty"`
	Status                string            `json:"status,omitempty"`
	LastReportedStatus    string            `json:"lastReportedStatus,omitempty"`
	LastReportedTimestamp *metav1.Time      `json:"lastReportedTimestamp,omitempty"`
	Attributes            map[string]string `json:"attributes,omitempty"`
}

// SPOTRequestStatus is the NVCA managed status of all requests in the backend.
//
// +k8s:openapi-gen=true
type SPOTRequestStatus struct {
	RequestStatus                      RequestStatus             `json:"requestStatus,omitempty"`
	LastStatusUpdated                  *metav1.Time              `json:"lastStatusUpdated,omitempty"`
	LastObservedAllTerminatedInstances *metav1.Time              `json:"lastObservedAllTerminatedInstances,omitempty"`
	CacheReferenceName                 string                    `json:"cacheReferenceName,omitempty"`
	Instances                          map[string]InstanceStatus `json:"instances,omitempty"`
	ReconcileErrors                    uint64                    `json:"reconcileErrors,omitempty"`
	LastReconcileError                 string                    `json:"lastReconcileError,omitempty"`
}

// GetInstanceIDs returns the Function Instance IDs
// for the SPOTRequestStatus
func (s SPOTRequestStatus) GetInstanceIDs() []string {
	if len(s.Instances) == 0 {
		return nil
	}

	instanceIDs := make([]string, len(s.Instances))
	i := 0
	for k := range s.Instances {
		instanceIDs[i] = k
		i++
	}
	return instanceIDs
}

// SPOTRequestList contains a list of SPOTRequest.
//
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SPOTRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SPOTRequest `json:"items"`
}

type ArtifactType string

const (
	ArtifactTypePod          ArtifactType = "POD_SPEC"
	ArtifactTypeSecret       ArtifactType = "SECRET_SPEC"
	ArtifactTypeConfigmap    ArtifactType = "CONFIGMAP_SPEC"
	ArtifactTypeHelmChart    ArtifactType = "HELM_CHART"
	ArtifactTypeHelmCreds    ArtifactType = "HELM_CREDS"
	ArtifactTypeInitCacheJob ArtifactType = "INIT_CACHE_JOB_SPEC"
	ArtifactTypeBlockDevice  ArtifactType = "BD_CREATE_SPEC"
)

type Artifact struct {
	Type          ArtifactType `json:"artifactType,omitempty"`
	Specification string       `json:"artifactSpec,omitempty"`
}

// SPOTCreationMessageInfo contains information about spot creation messages.
//
// +k8s:openapi-gen=true
type SPOTCreationMessageInfo struct {
	ClusterGroup    string     `json:"clusterGroup,omitempty"`
	InstanceCount   uint64     `json:"instanceCount,omitempty"`
	InstanceType    string     `json:"instanceType,omitempty"`
	GPUName         string     `json:"gpuName"`
	QueueURL        string     `json:"queueURL,omitempty"`
	LaunchArtifacts []Artifact `json:"launchArtifacts,omitempty"`
}

// SPOTTerminationMessageInfo contains information about spot termination messages.
//
// +k8s:openapi-gen=true
type SPOTTerminationMessageInfo struct {
	ClusterName string `json:"availabilityZone,omitempty"`
	//nolint:revive
	InstanceIds []string `json:"instanceIds,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SPOTRequest{}, &SPOTRequestList{})
}
