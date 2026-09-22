/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ExternalWorkloadKind = "ExternalWorkload"

	// DefaultNodeAffinitySpecName is the SpecName used for ExternalWorkloads with a
	// single pod template (Deployment, StatefulSet, etc.). Multi-template workloads
	// (e.g. Ray head/worker, kubeflow master/worker) supply their own SpecName.
	DefaultNodeAffinitySpecName = "worker"
)

// ExternalWorkloadSpec defines the desired state of ExternalWorkload
type ExternalWorkloadSpec struct {
	OriginalWorkloadUID  types.UID `json:"originalWorkloadUID,omitempty" protobuf:"bytes,1,opt,name=originalWorkloadUID,casttype=k8s.io/kubernetes/pkg/types.UID"`
	OriginalWorkloadName string    `json:"originalWorkloadName,omitempty" protobuf:"bytes,2,opt,name=originalWorkloadName"`
	// the type of the workload - usually the k8s object's kind
	Type    string `json:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
	Group   string `json:"group,omitempty" protobuf:"bytes,4,opt,name=group"`
	Version string `json:"version,omitempty" protobuf:"bytes,20,opt,name=version"`

	ProjectName    string `json:"projectName,omitempty" protobuf:"bytes,5,opt,name=projectName"`
	ProjectId      string `json:"projectId,omitempty" protobuf:"bytes,6,opt,name=projectId"`
	DepartmentName string `json:"departmentName,omitempty" protobuf:"bytes,7,opt,name=departmentName"`
	DepartmentId   string `json:"departmentId,omitempty" protobuf:"bytes,8,opt,name=departmentId"`
	// the CreationTimestamp of the associated workload object
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty" protobuf:"bytes,9,opt,name=creationTimestamp"`

	// the following fields will be populated if the workload object contains the info
	RequestedNodePools []string      `json:"requestedNodePools,omitempty" protobuf:"bytes,10,opt,name=requestedNodePools"`
	RequestedPods      RequestedPods `json:"requestedPods,omitempty" protobuf:"bytes,11,opt,name=requestedPods"`
	ImageName          string        `json:"imageName,omitempty" protobuf:"bytes,12,opt,name=imageName"`
	PriorityClassName  string        `json:"priorityClassName,omitempty" protobuf:"bytes,13,opt,name=priorityClassName"`
	Preemptibility     string        `json:"preemptibility,omitempty" protobuf:"bytes,14,opt,name=preemptibility"`

	Topology Topology `json:"topology,omitempty" protobuf:"bytes,15,opt,name=topology"`

	// request and limit of each resource type we need
	// +optional
	RequestedResources corev1.ResourceRequirements `json:"requestedResources,omitempty" protobuf:"bytes,16,opt,name=requestedResources"`

	// urls of the workload, if exist
	// +optional
	Urls []string `json:"urls,omitempty" protobuf:"bytes,17,opt,name=additionalFields"`

	// if we have more fields we want to pass to the server, we can use this field
	// +optional
	AdditionalFields *string `json:"additionalFields,omitempty" protobuf:"bytes,18,opt,name=additionalFields"`

	// NodeAffinities captures the user-requested node affinity per pod template,
	// extracted from the top owner. Each entry's specName matches the pod-template
	// name in the owning workload (e.g. "worker", "master", "leader") and aligns
	// with the control-plane WorkloadNodeAffinity API shape.
	// +optional
	// +listType=map
	// +listMapKey=specName
	NodeAffinities []NodeAffinityEntry `json:"nodeAffinities,omitempty" protobuf:"bytes,19,rep,name=nodeAffinities"`

	// ComponentTopologies captures the requested network topology constraints per component,
	// for workloads whose framework declares them per component rather than only for the whole
	// workload. Each entry's specName matches the instanceKey of the corresponding
	// StructureElement, so the constraint can be attached to that element.
	// +optional
	// +listType=map
	// +listMapKey=specName
	ComponentTopologies []ComponentTopologyEntry `json:"componentTopologies,omitempty" protobuf:"bytes,22,rep,name=componentTopologies"`

	// StorageRefs lists the existing PVCs, Secrets, and ConfigMaps referenced by
	// this workload's pod template (volumes and imagePullSecrets).
	// +optional
	// +listType=atomic
	StorageRefs []StorageRef `json:"storageRefs,omitempty" protobuf:"bytes,21,rep,name=storageRefs"`
}

// StorageRefKind enumerates the existing-resource kinds a workload can reference.
// +kubebuilder:validation:Enum=PersistentVolumeClaim;Secret;ConfigMap
type StorageRefKind string

const (
	StorageRefKindPVC       StorageRefKind = "PersistentVolumeClaim"
	StorageRefKindSecret    StorageRefKind = "Secret"
	StorageRefKindConfigMap StorageRefKind = "ConfigMap"
)

// StorageRef is a reference to an existing resource consumed by the workload's
// pod template. The external-workload-integrator resolves UID once per reconcile
// so cluster-sync passes it through instead of re-resolving on every sync.
type StorageRef struct {
	Kind StorageRefKind `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name string         `json:"name" protobuf:"bytes,2,opt,name=name"`
	// UID is the resolved Kubernetes UID of the referenced resource.
	// +optional
	UID string `json:"uid,omitempty" protobuf:"bytes,3,opt,name=uid"`
}

// NodeAffinityEntry pairs a node-affinity scheduling specification with the
// pod-template SpecName it applies to (e.g. "worker", "master", "leader").
type NodeAffinityEntry struct {
	// SpecName identifies the pod template within the owning workload (e.g. "worker", "master").
	SpecName string `json:"specName" protobuf:"bytes,1,opt,name=specName"`
	// NodeAffinity is the scheduling constraint for pods of the named spec.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,2,opt,name=nodeAffinity"`
}

// ComponentTopologyEntry pairs a network topology constraint with the component
// SpecName it applies to (e.g. "Frontend", "PrefillWorker").
type ComponentTopologyEntry struct {
	// SpecName identifies the component within the owning workload. It matches the
	// instanceKey of the corresponding StructureElement.
	SpecName string `json:"specName" protobuf:"bytes,1,opt,name=specName"`
	// Topology is the network topology constraint requested for the named component.
	// +optional
	Topology Topology `json:"topology,omitempty" protobuf:"bytes,2,opt,name=topology"`
}

// ExternalWorkloadStatus defines the observed state of ExternalWorkload
type ExternalWorkloadStatus struct {
	ObservedStatus    ExternalWorkloadObservedStatus `json:"observedStatus,omitempty" protobuf:"bytes,1,opt,name=observedStatus"`
	SyncStatus        WorkloadSyncStatus             `json:"syncStatus,omitempty" protobuf:"bytes,2,opt,name=syncStatus"`
	NetworkStatus     WorkloadNetworkStatus          `json:"networkStatus,omitempty" protobuf:"bytes,3,opt,name=networkStatus"`
	StructureElements *StructureElements             `json:"structureElements,omitempty" protobuf:"bytes,4,opt,name=structureElements"`
}

type StructureElements struct {
	// Flat list of intermediate child elements (levels 1-3).
	Elements []StructureElement `json:"elements"`

	// Last time structure was calculated.
	LastCalculated metav1.Time `json:"lastCalculated"`
}

type StructureElement struct {
	// Generated UUID of the element
	ID string `json:"id"`

	// Display name / deterministic key for stable matching across recalculations.
	Name string `json:"name"`

	// Reference to the backing K8s object. Present for object-backed elements, nil for logical elements.
	// +optional
	K8sRef *K8sObjectRef `json:"k8sRef,omitempty"`

	// Instance identifier from ComponentInstanceSelector (e.g., "Frontend", "PrefillWorker").
	// +optional
	InstanceKey *string `json:"instanceKey,omitempty"`

	// Replica identifier if applicable.
	// +optional
	ReplicaKey *string `json:"replicaKey,omitempty"`

	// Parent element ID. Empty for level 1 elements (parent is implicitly the workload).
	// +optional
	ParentID *string `json:"parentId,omitempty"`

	// Hierarchy level: 1-3 only (intermediate child elements).
	Level int `json:"level"`
}

type K8sObjectRef struct {
	UID  types.UID        `json:"uid"`
	Name string           `json:"name"`
	GVK  GroupVersionKind `json:"gvk"`
}

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// ExternalWorkloadObservedStatus defines the observed status of the referred ExternalWorkload object
type ExternalWorkloadObservedStatus struct {
	// mapping native k8s statuses to pre-defined list of phases
	Phase ExternalWorkloadPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`
	// a string describing the reason for the phase, if exists
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// native k8s status
	K8sStatus string `json:"k8sStatus,omitempty" protobuf:"bytes,3,opt,name=k8sStatus"`

	// Represents the latest observed conditions of the original workload, if any.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ObservedCondition `json:"conditions,omitempty" protobuf:"bytes,4,opt,name=conditions"`

	// the k8s UUIDs of all child objects of the workload
	// +patchMergeKey=uid
	// +patchStrategy=merge
	ChildIds []ChildId `json:"childIds,omitempty" protobuf:"bytes,5,opt,name=childIds"`

	// the timestamp of the last status update (except sync status)
	LastStatusChangeTime metav1.Time `json:"lastStatusChangeTime,omitempty" protobuf:"bytes,6,opt,name=lastStatusChangeTime"`
}

type ExternalEndpoint struct {
	// endpoint type
	SourceType ExternalEndpointSourceType `json:"sourceType,omitempty" protobuf:"bytes,1,opt,name=sourceType"`
	// endpoint host
	Host string `json:"host,omitempty" protobuf:"bytes,2,opt,name=host"`
	// endpoint path
	Path string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	// endpoint tls
	TlsEnabled *bool `json:"tlsEnabled,omitempty" protobuf:"bytes,4,opt,name=tlsEnabled"`
	// endpoint protocol
	Protocol string `json:"protocol,omitempty" protobuf:"bytes,5,opt,name=protocol"`
	// endpoint port
	Port string `json:"port,omitempty" protobuf:"bytes,6,opt,name=port"`
	// endpoint url
	Url string `json:"url,omitempty" protobuf:"bytes,7,opt,name=url"`
	// endpoint sourceRef
	SourceRef ExternalEndpointSourceRef `json:"SourceRef,omitempty" protobuf:"bytes,8,opt,name=SourceRef"`
}

type ExternalEndpointSourceRef struct {
	// endpoint sourceRef kind
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	// endpoint sourceRef name
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
}

// WorkloadNetworkStatus defines the status of the network detected
type WorkloadNetworkStatus struct {
	// External endpoint discovered on workload
	ExternalEndpoints []ExternalEndpoint `json:"externalEndpoints,omitempty" protobuf:"bytes,18,opt,name=externalEndpoints"`
}

// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
type WorkloadSyncStatus struct {
	// should be populated by the cluster-sync when syncing the object
	LastSyncTimestamp metav1.Time `json:"lastSyncTimestamp,omitempty" protobuf:"bytes,1,opt,name=lastSyncTimestamp"`
	// the last object Generation that was observed and synced to control-plane - "Generation" is only for spec fields
	LastSyncGeneration int64 `json:"lastSyncGeneration,omitempty" protobuf:"bytes,2,opt,name=lastSyncGeneration"`
	// the last status change time that was observed and synced to control-plane.
	LastSyncStatusChangeTime metav1.Time `json:"lastSyncStatusChangeTime,omitempty" protobuf:"bytes,3,opt,name=lastSyncStatusChangeTime"`
	// the last observed and synced status change time (LastStatusChangeTime)
	LastSyncStatusChecksum uint32 `json:"lastSyncStatusChecksum,omitempty" protobuf:"bytes,4,opt,name=lastSyncStatusChecksum"`
	// the last generation that was attempted to be synced - in case sync failed
	// +optional
	LastAttemptedSyncGeneration *int64 `json:"lastAttemptedSyncGeneration,omitempty" protobuf:"bytes,5,opt,name=lastAttemptedSyncGeneration"`
	// the last statusChangeTime that was attempted to be synced - in case sync failed
	// +optional
	LastAttemptedSyncStatusChangeTime *metav1.Time `json:"lastAttemptedSyncStatusChangeTime,omitempty" protobuf:"bytes,6,opt,name=lastAttemptedSyncStatusChangeTime"`
	// in case of an error - a string describing the error
	// +optional
	SyncErrorMsg *string `json:"syncErrorMsg,omitempty" protobuf:"bytes,7,opt,name=syncErrorMsg"`
}

// ObservedCondition describes the state of the original workload at a certain point.
type ObservedCondition struct {
	// Type of the condition.
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// Last time the condition transitioned from one status to another.
	// +optional
	// +nullable
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

type ChildId struct {
	UID  string `json:"uid" protobuf:"bytes,1,opt,name=uid"`
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`
}

type RequestedPods struct {
	// Replicas is the number of desired pods.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"bytes,1,opt,name=replicas"`
	// MinReplicas denotes the minimum number of desired pods.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"bytes,2,opt,name=minReplicas"`
	// MaxReplicas denotes the maximum number of desired pods.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty" protobuf:"bytes,3,opt,name=maxReplicas"`
	// Parallelism - used with Hyperparameter Optimization. Specifies the number of successful pods
	// the job should reach to be completed. The 'parallelism' flag should be smaller or equal to 'completions'
	// +optional
	Parallelism *int32 `json:"parallelism,omitempty" protobuf:"bytes,4,opt,name=parallelism"`
	// Specifies the maximum desired number of pods the workload should run at any given time.
	// +optional
	Completions *int32 `json:"completions,omitempty" protobuf:"bytes,5,opt,name=completions"`
}

type Topology struct {
	Name               string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	PreferredPlacement string `json:"preferredPlacement,omitempty" protobuf:"bytes,2,opt,name=preferredPlacement"`
	RequiredPlacement  string `json:"requiredPlacement,omitempty" protobuf:"bytes,3,opt,name=requiredPlacement"`
}

// +enum
type ExternalWorkloadPhase string

const (
	ExWCreatingPhase     ExternalWorkloadPhase = "Creating"
	ExWPendingPhase      ExternalWorkloadPhase = "Pending"
	ExWInitializingPhase ExternalWorkloadPhase = "Initializing"
	ExWRunningPhase      ExternalWorkloadPhase = "Running"
	ExWStoppingPhase     ExternalWorkloadPhase = "Stopping"
	ExWStoppedPhase      ExternalWorkloadPhase = "Stopped"
	ExWDegradedPhase     ExternalWorkloadPhase = "Degraded"
	ExWFailedPhase       ExternalWorkloadPhase = "Failed"
	ExWCompletedPhase    ExternalWorkloadPhase = "Completed"
	ExWDeletingPhase     ExternalWorkloadPhase = "Deleting"
	ExWUnknownPhase      ExternalWorkloadPhase = "Unknown"
)

type ExternalEndpointSourceType string

const (
	ExternalEndpointIngressSource     ExternalEndpointSourceType = "Ingress"
	ExternalEndpointNodePortSource    ExternalEndpointSourceType = "NodePort"
	ExternalEndpointLoadBalanceSource ExternalEndpointSourceType = "LoadBalance"
	ExternalEndpointGatewayAPISource  ExternalEndpointSourceType = "GatewayAPI"
	ExternalEndpointOtherSource       ExternalEndpointSourceType = "Other"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced
//+kubebuilder:storageversion
//+kubebuilder:resource:shortName=ew

// ExternalWorkload is the Schema for the externalworkloads API
type ExternalWorkload struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// +optional
	Spec ExternalWorkloadSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// +optional
	Status ExternalWorkloadStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// ExternalWorkloadList contains a list of ExternalWorkload
type ExternalWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ExternalWorkload `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&ExternalWorkload{}, &ExternalWorkloadList{})
}
