/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AccessRuleKind    = "AccessRule"
	AccessRuleIdLabel = "runai/accessrule-id"
)

// AccessRuleSpec defines the desired state of AccessRule
type AccessRuleSpec struct {
	// the id of the user or group the access rule refers to
	SubjectId string `json:"subjectId,omitempty"`

	// the type of the subject - user or group
	SubjectType AccessRuleSubjectType `json:"subjectType,omitempty"`

	// the name of the scope the access rule refers to
	// +optional: if cluster scoped; required otherwise
	ScopeName string `json:"scopeName,omitempty"`

	// the type of the scope
	ScopeType ScopeType `json:"scopeType,omitempty"`

	// the role name the access rule refers to
	RoleName string `json:"roleName,omitempty"`
}

// AccessRuleStatus defines the observed state of AccessRule
type AccessRuleStatus struct {
	// Phase is the recently observed phase of the access rule.
	Phase AccessRulePhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ProjectPhase"`

	// A human-readable message indicating details about why the access rule is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// Current conditions of the access rule
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// AccessRule is the Schema for the accessrules API
type AccessRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessRuleSpec   `json:"spec,omitempty"`
	Status AccessRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AccessRuleList contains a list of AccessRule
type AccessRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessRule `json:"items"`
}

// +enum
type AccessRuleSubjectType string

const (
	UserSubject                  AccessRuleSubjectType = "user"
	GroupSubject                 AccessRuleSubjectType = "group"
	AppSubject                   AccessRuleSubjectType = "app"
	ServiceAccountSubject        AccessRuleSubjectType = "service-account"
	AllAuthenticatedUsersSubject AccessRuleSubjectType = "all-authenticated-users"
)

// +enum
type ScopeType string

const (
	ProjectScope    ScopeType = "project"
	DepartmentScope ScopeType = "department"
	ClusterScope    ScopeType = "cluster"
)

// +enum
type AccessRulePhase string

const (
	AccessRuleReady    AccessRulePhase = "Ready"
	AccessRuleNotReady AccessRulePhase = "NotReady"
)

func init() {
	SchemeBuilder.Register(&AccessRule{}, &AccessRuleList{})
}

// IsOwner returns the index of this AccessRule in the given object's OwnerReferences array if it exists.
func (accessRule *AccessRule) IsOwner(obj client.Object) (indexOfOwnerInRefArray int) {
	indexOfOwnerInRefArray = -1
	for index, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.Kind == AccessRuleKind &&
			ownerRef.UID == accessRule.UID &&
			ownerRef.Name == accessRule.Name {
			indexOfOwnerInRefArray = index
		}
	}
	return indexOfOwnerInRefArray
}
