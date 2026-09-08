/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	WorkloadRunnerKind = "WorkloadRunner"
)

// WorkloadRunnerSpec defines the desired state of WorkloadRunner
type WorkloadRunnerSpec struct {
	// Template is the raw workload manifest the user wants to run.
	// This can be any Kubernetes object that has a Karta configured in the cluster.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	Template runtime.RawExtension `json:"template"`

	// Instructions is a collection of instructions for the controller to apply.
	// This section is designed to be extensible with plugins.
	// +optional
	Instructions *Instructions `json:"instructions,omitempty"`

	// ManagedBy, when non-empty, causes dromos to skip reconciliation of this
	// runner; the named external controller is expected to manage it. There is
	// no enforcement that such a controller exists.
	// Empty means dromos manages it.
	// +kubebuilder:default=""
	// +optional
	ManagedBy string `json:"managedBy,omitempty"`
}

// Instructions defines the instructions for the controller to apply.
// This struct is designed to be extensible with plugins through unstructured fields.
// +kubebuilder:pruning:PreserveUnknownFields
type Instructions struct {
	// +optional
	GpuOptimizations *GpuOptimizations `json:"gpuOptimizations,omitempty"`
	// +optional
	Networking *NetworkingConfig `json:"networking,omitempty"`
	// +optional
	Storage *StorageConfig `json:"storage,omitempty"`
	// +optional
	NodeAffinity *NodeAffinityConfig `json:"nodeAffinity,omitempty"`
}

// NodeAffinityConfig defines the node-affinity plugin configuration: scheduling
// constraints applied on top of whatever the submitted template already asks for.
// The plugin merges them into every component's pod affinity, it never replaces
// the workload's own affinity.
type NodeAffinityConfig struct {
	// NodeAffinity is standard Kubernetes node affinity, merged into every
	// component's pod affinity.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
}

// StorageConfig defines the storage plugin configuration.
// Only lists PVCs to materialize. Mounting (volumes/volumeMounts) is the
// responsibility of the submitted workload manifest.
type StorageConfig struct {
	// NewPvcs is the list of PersistentVolumeClaims to create, owned by the WorkloadRunner.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	NewPvcs []corev1.PersistentVolumeClaim `json:"newPvcs,omitempty"`

	// CleanupOnCompletion controls whether owned PVCs are deleted when the WorkloadRunner
	// reaches a terminal phase. When unset, the plugin-level default applies.
	// +optional
	CleanupOnCompletion *bool `json:"cleanupOnCompletion,omitempty"`
}

type GpuOptimizations struct {
	// +optional
	Mnnvl *MnnvlConfig `json:"mnnvl,omitempty"`
}

const (
	MnnvlModeNone     MnnvlMode = "None"
	MnnvlModeRequired MnnvlMode = "Required"
)

type MnnvlMode string

type MnnvlConfig struct {
	// +kubebuilder:validation:Enum=None;Required
	Mode MnnvlMode `json:"mode"`
	// GpuNetworkAccelerationLabelKeys is a list of GPU network acceleration label keys
	// This key is used for preferring to place pods on the same GPU rack using pod affinity.
	// If empty, pod affinity will not be added.
	// +optional
	GpuNetworkAccelerationLabelKeys []string `json:"gpuNetworkAccelerationLabelKeys,omitempty"`
}

// AuthorizationType defines the type of authorization for the workload
// +kubebuilder:validation:Enum=authenticatedUsers;authorizedUsers;authorizedGroups
type AuthorizationType string

const (
	// AuthorizationTypeAuthenticatedUsers allows any authenticated user
	AuthorizationTypeAuthenticatedUsers AuthorizationType = "authenticatedUsers"
	// AuthorizationTypeAuthorizedUsers allows only specific users
	AuthorizationTypeAuthorizedUsers AuthorizationType = "authorizedUsers"
	// AuthorizationTypeAuthorizedGroups allows only specific groups
	AuthorizationTypeAuthorizedGroups AuthorizationType = "authorizedGroups"
)

// NetworkingConfig defines the networking plugin configuration
type NetworkingConfig struct {
	// Optional service for the workload. If empty, creates Ingress/Route only without managing services.
	// +kubebuilder:validation:MaxItems=1
	// +optional
	Services []NetworkingServiceItem `json:"services,omitempty"`

	// ExternalAccess configures external access (Ingress or OpenShift Route)
	// +optional
	ExternalAccess *NetworkingExternalAccessConfig `json:"externalAccess,omitempty"`
}

// NetworkingServiceItem defines a single service configuration with optional authentication
type NetworkingServiceItem struct {
	// Create specifies a new K8s Service to create. Mutually exclusive with Ref.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Create *corev1.Service `json:"create,omitempty"`

	// Ref references an existing service. Mutually exclusive with Create.
	// +optional
	Ref *NetworkingServiceRef `json:"ref,omitempty"`

	// Auth configures authentication and authorization. Required when using Ref.
	// +optional
	Auth *NetworkingAuthConfig `json:"auth,omitempty"`
}

// NetworkingAuthConfig defines authentication and authorization settings.
// Authentication verifies user identity via OAuth2/OIDC (all authenticated users can log in).
// Authorization determines which authenticated users are allowed to access the workload.
type NetworkingAuthConfig struct {
	// AuthorizationType defines the authorization policy for accessing the workload:
	//   - authenticatedUsers: Any authenticated user can access (no additional authorization)
	//   - authorizedUsers: Only specific users in the authorizedUsers list can access
	//   - authorizedGroups: Only users belonging to groups in the authorizedGroups list can access
	// +kubebuilder:default=authenticatedUsers
	// +optional
	AuthorizationType AuthorizationType `json:"authorizationType,omitempty"`

	// AuthorizedUsers is a comma-separated list of authorized user emails
	// Required when authorizationType is "authorizedUsers"
	// +optional
	AuthorizedUsers string `json:"authorizedUsers,omitempty"`

	// AuthorizedGroups is a comma-separated list of authorized groups
	// Required when authorizationType is "authorizedGroups"
	// +optional
	AuthorizedGroups string `json:"authorizedGroups,omitempty"`

	// ForceJsonResponse configures OAuth2 proxy to return JSON error responses instead of HTML.
	// +optional
	ForceJsonResponse *bool `json:"forceJsonResponse,omitempty"`
}

// NetworkingServiceRef references an existing service
type NetworkingServiceRef struct {
	// Name is the service name
	Name string `json:"name"`

	// Port is the service port
	Port int32 `json:"port"`
}

// NetworkingExternalAccessConfig defines external access configuration
type NetworkingExternalAccessConfig struct {
	// Ingress for external access (standard Kubernetes).
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Ingress *networkingv1.Ingress `json:"ingress,omitempty"`

	// Route for external access (OpenShift).
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Route *openshiftv1.Route `json:"route,omitempty"`

	// HTTPRoute for external access (Gateway API).
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	HTTPRoute *gatewayv1.HTTPRoute `json:"httpRoute,omitempty"`
}

type Issue struct {
	// Kind is the kind of the resource the issue related to.
	Kind string `json:"kind"`
	// Name is the name of the issue.
	Name string `json:"name"`
	// Description is the description of the issue.
	Description string `json:"description"`
	// Group is the group of the resource the issue related to.
	Group string `json:"group"`
	// IsCritical reports whether the issue affects the workload's phase.
	IsCritical bool `json:"isCritical"`
}

// +kubebuilder:validation:Enum=Creating;Pending;Initializing;Ready;Disabled;Failed;Unknown
type PluginPhase string

const (
	PluginPhaseCreating     PluginPhase = "Creating"
	PluginPhasePending      PluginPhase = "Pending"
	PluginPhaseInitializing PluginPhase = "Initializing"
	PluginPhaseReady        PluginPhase = "Ready"
	PluginPhaseFailed       PluginPhase = "Failed"
	PluginPhaseDisabled     PluginPhase = "Disabled"
	PluginPhaseUnknown      PluginPhase = "Unknown"
)

// PluginStatus defines the status of the workload runner.
type PluginStatus struct {
	// Phase is the current phase of the plugin.
	// +optional
	Phase PluginPhase `json:"phase,omitempty"`

	// Conditions is a list of conditions that occurred during the reconciliation of the plugin.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Issues is a list of issues that occurred during the reconciliation of the plugin.
	// Each issue contains the kind, name, group, and description of the resource the issue related to.
	// +optional
	Issues []Issue `json:"issues,omitempty"`

	// AdditionalInfo contains plugin-owned structured status.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	AdditionalInfo *runtime.RawExtension `json:"additionalInfo,omitempty"`
}

// SetCondition sets a condition in the plugin status.
func (ps *PluginStatus) SetCondition(condition metav1.Condition) bool {
	return meta.SetStatusCondition(&ps.Conditions, condition)
}

// SetPhase sets the phase of the plugin.
func (ps *PluginStatus) SetPhase(phase PluginPhase) {
	ps.Phase = phase
}

// SetAdditionalInfo sets plugin-owned structured status information.
func (ps *PluginStatus) SetAdditionalInfo(additionalInfo *runtime.RawExtension) {
	ps.AdditionalInfo = additionalInfo
}

// WorkloadRunnerStatus defines the observed state of WorkloadRunner
type WorkloadRunnerStatus struct {
	// Phase is the current phase of the WorkloadRunner.
	// +optional
	Phase WorkloadRunnerPhase `json:"phase,omitempty"`

	// ResourcePhase is the current phase of the underlying main resource.
	// +optional
	ResourcePhase WorkloadRunnerPhase `json:"resourcePhase,omitempty"`

	// Conditions represent the latest available observations of the WorkloadRunner's state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Issues is a list of issues that occurred during the reconciliation of the WorkloadRunner.
	// +optional
	Issues []Issue `json:"issues,omitempty"`

	// Plugins is a map of plugin statuses, keyed by plugin name.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxProperties=25
	Plugins map[string]*PluginStatus `json:"plugins,omitempty"`
}

// +kubebuilder:validation:Enum=Creating;Pending;Initializing;Running;Stopped;Degraded;Failed;Completed;Suspended;Suspending;Resuming;Restarting;Unknown
type WorkloadRunnerPhase string

const (
	WorkloadRunnerPhaseCreating     WorkloadRunnerPhase = "Creating"
	WorkloadRunnerPhasePending      WorkloadRunnerPhase = "Pending"
	WorkloadRunnerPhaseInitializing WorkloadRunnerPhase = "Initializing"
	WorkloadRunnerPhaseRunning      WorkloadRunnerPhase = "Running"
	WorkloadRunnerPhaseStopped      WorkloadRunnerPhase = "Stopped"
	WorkloadRunnerPhaseDegraded     WorkloadRunnerPhase = "Degraded"
	WorkloadRunnerPhaseFailed       WorkloadRunnerPhase = "Failed"
	WorkloadRunnerPhaseCompleted    WorkloadRunnerPhase = "Completed"
	WorkloadRunnerPhaseSuspended    WorkloadRunnerPhase = "Suspended"
	WorkloadRunnerPhaseSuspending   WorkloadRunnerPhase = "Suspending"
	WorkloadRunnerPhaseResuming     WorkloadRunnerPhase = "Resuming"
	WorkloadRunnerPhaseRestarting   WorkloadRunnerPhase = "Restarting"
	WorkloadRunnerPhaseUnknown      WorkloadRunnerPhase = "Unknown"

	LastReconcileSucceeded           = "LastReconcileSucceeded"
	ConditionTypePrerequisitesPassed = "PrerequisitesPassed"

	ReasonCRDNotFound     = "CRDNotFound"
	ReasonKartaNotFound   = "KartaNotFound"
	ReasonKartaNotReady   = "KartaNotReady"
	ReasonSucceeded       = "Succeeded"
	ReasonReconcileFailed = "ReconcileFailed"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,shortName=wlr
//+kubebuilder:storageversion
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// WorkloadRunner is the Schema for the workloadrunners API
type WorkloadRunner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadRunnerSpec   `json:"spec,omitempty"`
	Status WorkloadRunnerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkloadRunnerList contains a list of WorkloadRunner
type WorkloadRunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkloadRunner `json:"items"`
}
