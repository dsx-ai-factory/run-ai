// Package v1alpha1 contains local NIMCache types
package v1alpha1

import (
	nim "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Re-export NIMCache status constants from upstream
const (
	NimCacheStatusReady           = nim.NimCacheStatusReady
	NimCacheStatusNotReady        = nim.NimCacheStatusNotReady
	NimCacheStatusPending         = nim.NimCacheStatusPending
	NimCacheStatusPVCCreated      = nim.NimCacheStatusPVCCreated
	NimCacheStatusStarted         = nim.NimCacheStatusStarted
	NimCacheStatusFailed          = nim.NimCacheStatusFailed
	NimCacheStatusInProgress      = nim.NimCacheStatusInProgress
	NimCacheConditionPVCCreated   = nim.NimCacheConditionPVCCreated
	NimCacheConditionJobCreated   = nim.NimCacheConditionJobCreated
	NimCacheConditionJobPending   = nim.NimCacheConditionJobPending
	NimCacheConditionJobCompleted = nim.NimCacheConditionJobCompleted
)

// NIMCache is the Schema for the nimcaches API
// This wraps the upstream NIMCache
type NIMCache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   nim.NIMCacheSpec   `json:"spec,omitempty"`
	Status nim.NIMCacheStatus `json:"status,omitempty"`
}

// NIMCacheList contains a list of NIMCache
type NIMCacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NIMCache `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NIMCache{}, &NIMCacheList{})
}
