/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SeverityInfo  = "Info"
	SeverityError = "Error"
)

// +k8s:deepcopy-gen=true
type PolicySyncStatus struct {
	// should be populated by the agent when syncing the object, last sync time which was successful
	LastSyncTimestamp metav1.Time `json:"lastSyncTimestamp,omitempty" protobuf:"bytes,1,opt,name=lastSyncTimestamp"`

	// last seq number that currently lives in the cluster
	LastSyncRevision int64 `json:"lastSyncRevision,omitempty" protobuf:"bytes,2,opt,name=lastSyncRevision"`

	// the last revision that was attempted to be synced - in case it failed and is different from the LastSyncRevision
	LastAttemptedSyncRevision int64 `json:"lastAttemptedSyncRevision,omitempty" protobuf:"bytes,3,opt,name=lastAttemptedSyncRevision"`

	// in case of an error - a string describing the error
	// +optional
	SyncErrorMsg string `json:"syncErrorMsg,omitempty" protobuf:"bytes,4,opt,name=syncErrorMsg"`
}

// PolicyStatus is holding the sync status, rather than that policies do not have any meaningful status they are only stored in etcd
// +k8s:deepcopy-gen=true
type PolicyStatus struct {
	SyncStatus PolicySyncStatus `json:"syncStatus,omitempty"`
}

// ResourceTypeAndName is a struct that holds the kind, apiVersion, namespace and name of a resource
// +k8s:deepcopy-gen=true
type ResourceTypeAndName struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
}

func (self *ResourceTypeAndName) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", self.APIVersion, self.Kind, self.Namespace, self.Name)
}

// WorkloadStatus hols the status of the workload resource
// +k8s:deepcopy-gen=true
type WorkloadStatus struct {
	// +kubebuilder:validation:Enum:=Created;Submitted;Failed
	CreationStatus string `json:"creationStatus,omitempty"`

	CreatedResources []ResourceTypeAndName `json:"createdResources,omitempty"`

	Events []Event `json:"events,omitempty"`
}

func (self *WorkloadStatus) Failed() *WorkloadStatus {
	// dont override existing status if it is already set
	// this is workaround for submit failed case for updating workload
	if self.CreationStatus != "" {
		return self
	}
	self.CreationStatus = StatusFailed
	return self
}

func (self *WorkloadStatus) LogEvent(event Event) *WorkloadStatus {
	if self.Events == nil {
		self.Events = make([]Event, 0, 5)
	}
	self.Events = append(self.Events, event)
	return self
}

func (self *WorkloadStatus) LogResource(resourceTypeAndName ResourceTypeAndName) *WorkloadStatus {
	if self.CreatedResources == nil {
		self.CreatedResources = make([]ResourceTypeAndName, 0, 5)
	}
	self.CreatedResources = append(self.CreatedResources, resourceTypeAndName)
	return self
}

func (self *WorkloadStatus) LogInfo(message string) *WorkloadStatus {
	self.LogEvent(NewEvent(SeverityInfo, message))

	return self
}

func (self *WorkloadStatus) LogError(err error) *WorkloadStatus {
	self.LogEvent(NewEvent(SeverityError, err.Error()))
	return self
}

//	list of events for status of the workload
//
// +k8s:deepcopy-gen=true
type Event struct {
	Time metav1.Time `json:"time,omitempty"`
	// +kubebuilder:validation:Enum:=Info;Warning;Error
	Severity string `json:"severity,omitempty"`
	Details  string `json:"details,omitempty"`
}

func NewEvent(severity string, details string) Event {
	return Event{
		Time:     metav1.NewTime(time.Now()),
		Severity: severity,
		Details:  details,
	}
}
