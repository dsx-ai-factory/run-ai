/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	RemoteConfigSyncKind  = "RemoteConfigSync"
	DefaultSyncIntervalMs = 10000
	RemoteConfigSyncName  = "runai"
)

const (
	ConditionTypeConfigPulled     ConditionType = "Pulled"
	ConditionTypeConfigPushed     ConditionType = "Pushed"
	ConditionTypeConfigTerminated ConditionType = "Terminated"
)

const (
	PullSuccessful ConditionReason = "PullSuccessful"
	PullFailed     ConditionReason = "PullFailed"

	PushSuccessful ConditionReason = "PushSuccessful"
	PushFailed     ConditionReason = "PushFailed"
	PushNotNeeded  ConditionReason = "PushNotNeeded"

	TerminatedDueToDeactivation ConditionReason = "TerminatedDueToDeactivation"
)

// RemoteConfigSyncSpec defines the desired state of RemoteConfigSync
type RemoteConfigSyncSpec struct {
	// SyncIntervalInMs is the interval in milliseconds in which the remote config should be synced, If set to 0 the sync will be disabled
	SyncIntervalInMs *int64 `json:"syncIntervalInMS,omitempty"`
}

// RemoteConfigSyncStatus defines the observed state of RemoteConfigSync
type RemoteConfigSyncStatus struct {
	RunaiCustomTypeStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// RemoteConfigSync is the Schema for the remoteconfigsyncs API
type RemoteConfigSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteConfigSyncSpec   `json:"spec,omitempty"`
	Status RemoteConfigSyncStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteConfigSyncList contains a list of RemoteConfigSync
type RemoteConfigSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteConfigSync `json:"items"`
}

func (remoteConfigSync *RemoteConfigSync) SetDefaultsWhereNeeded() {
	remoteConfigSync.TypeMeta.APIVersion = GroupVersion.Group + "/" + GroupVersion.Version
	remoteConfigSync.TypeMeta.Kind = RemoteConfigSyncKind

	spec := &remoteConfigSync.Spec
	if spec.SyncIntervalInMs == nil {
		spec.SyncIntervalInMs = ptr.To[int64](DefaultSyncIntervalMs)
	}
}

func init() {
	SchemeBuilder.Register(&RemoteConfigSync{}, &RemoteConfigSyncList{})
}

func (remoteConfigSync *RemoteConfigSync) GetRunaiCustomStatus() *RunaiCustomTypeStatus {
	return &remoteConfigSync.Status.RunaiCustomTypeStatus
}
