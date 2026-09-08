/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	cpclusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RemoteConfigKind = "RemoteConfig"
	RemoteConfigName = "runai"
)
const (
	ConditionTypeRemoteConfigApplying ConditionType   = "Applying"
	ConditionTypeRemoteConfigApplied  ConditionType   = "Applied"
	ConfigActionSuccess               ConditionReason = "completed_with_success"
	ConfigActionFailed                ConditionReason = "completed_with_failure"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RemoteConfigSpec defines the desired state of RemoteConfig
type RemoteConfigSpec struct {
	internal.ClusterFullConfig `json:",inline"`
}

// RemoteConfigStatus defines the observed state of RemoteConfig
type RemoteConfigStatus struct {
	RunaiCustomTypeStatus `json:",inline"`
	ActualConfig          internal.ClusterFullConfig `json:"actualConfig,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// RemoteConfig is the Schema for the remoteconfigs API
type RemoteConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteConfigSpec   `json:"spec,omitempty"`
	Status RemoteConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteConfigList contains a list of RemoteConfig
type RemoteConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteConfig{}, &RemoteConfigList{})
}

func (remoteConfig *RemoteConfig) GetRunaiCustomStatus() *RunaiCustomTypeStatus {
	return &remoteConfig.Status.RunaiCustomTypeStatus
}

func (remoteConfig *RemoteConfig) SetDefaultsWhereNeeded() {
	remoteConfig.TypeMeta.APIVersion = GroupVersion.Group + "/" + GroupVersion.Version
	remoteConfig.TypeMeta.Kind = RemoteConfigKind

	spec := &remoteConfig.Spec
	if spec.SystemConfig == nil {
		spec.SystemConfig = &cpclusterclient.ClusterSystemConfigData{}
	}
	if spec.UserConfig == nil {
		spec.UserConfig = &cpclusterclient.ClusterUserConfigData{}
	}
}
