/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// Source of the application (e.g., "helm", "dynamo")
	Source string `json:"source,omitempty"`
	// The type/name of the application (e.g., chart name)
	Type string `json:"type,omitempty"`
	// Helm contains optional Helm chart deployment configuration.
	// When specified, the controller will deploy the Helm chart.
	// When not specified, the controller will only track externally deployed resources.
	// +optional
	Helm *HelmSpec `json:"helm,omitempty"`
}

// HelmSpec defines Helm chart deployment configuration
// +kubebuilder:validation:XValidation:rule="(has(self.chartRef) && !has(self.chartUrl)) || (!has(self.chartRef) && has(self.chartUrl))",message="exactly one of chartRef or chartUrl must be set"
type HelmSpec struct {
	// Reference to the chart to be installed by means of Repo, Version and Name.
	// ChartRef and ChartUrl are mutually exclusive.
	// +optional
	ChartRef *ChartRef `json:"chartRef"`
	// URL pointing to the chart to be installed.
	// ChartUrl and ChartRef are mutually exclusive.
	// +optional
	ChartUrl string `json:"chartUrl,omitempty"`
	// RequiresControlPlaneAuth indicates the chart is hosted by the control plane (an uploaded
	// chart) and must be downloaded with the cluster's control-plane bearer token. Set by the
	// control plane at sync time for uploaded charts; never set for user-provided ChartUrl charts.
	// +optional
	RequiresControlPlaneAuth bool `json:"requiresControlPlaneAuth,omitempty"`
	// Values are the Helm values to override chart defaults
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Values map[string]apiextensionsv1.JSON `json:"values,omitempty"`
	// CredentialBindings override effective chart values with resolved Secret names after chart loading.
	// +optional
	// +kubebuilder:validation:MaxItems=128
	CredentialBindings []HelmCredentialBinding `json:"credentialBindings,omitempty"`
	// Auth contains authentication configuration for private repositories
	// +optional
	Auth *HelmAuthSpec `json:"auth,omitempty"`
}

// HelmCredentialBinding references a runtime Secret in the application's namespace.
type HelmCredentialBinding struct {
	// Variable is a dotted or indexed path in the chart's values.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Variable string `json:"variable"`
	// SecretName is a reference only; credential contents never enter Helm values.
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`
}

type ChartRef struct {
	// Repository is the Helm chart repository URL (HTTP/HTTPS or OCI)
	// Examples: "https://charts.bitnami.com/bitnami", "oci://registry.example.com/charts"
	//+kubebuilder:validation:Required
	Repository string `json:"repository"`
	// Chart name to deploy
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	// Version is the chart version to deploy
	//+kubebuilder:validation:Required
	Version string `json:"version"`
}

// HelmAuthSpec defines authentication for Helm repositories
type HelmAuthSpec struct {
	// SecretRef references a Secret containing authentication credentials
	// The Secret should contain keys based on the auth type:
	// - For basic auth: "username" and "password"
	// - For TLS: "ca.crt", "tls.crt", "tls.key"
	// - For OCI registries: "dockerconfigjson" (same format as imagePullSecrets)
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`
	// InsecureSkipTLSVerify skips TLS certificate verification (not recommended for production)
	// +optional
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify,omitempty"`
	// CAFile is the path to a CA certificate file for custom CA verification
	// Typically loaded from the secret's "ca.crt" key
	// +optional
	CAFile string `json:"caFile,omitempty"`
}

// SecretReference references a Secret in the same namespace
type SecretReference struct {
	// Name is the name of the Secret
	Name string `json:"name"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// k8sPhase is the recently observed phase of the application.
	K8sPhase string `json:"k8sPhase,omitempty"`
	// List of Resources that are related to this application
	Resources []ApplicationResource `json:"resources,omitempty"`
	// Information about the action that is currently
	// performed on the application, or the last action that
	// has been performed on the application.
	// +optional
	ActionStatus *ActionStatus `json:"action,omitempty"`
	// Status of the helm installer (only set when deployed via spec.helm)
	// +optional
	HelmStatus *HelmStatus `json:"helmStatus,omitempty"`
}

// Action provides status information about the activity performed on the application
type ActionStatus struct {
	// Action type (Install, Upgrade, Uninstall)
	Type string `json:"type,omitempty"`
	// Phase
	Phase string `json:"phase,omitempty"`
	// Operation start time
	StartedAt metav1.Time `json:"startedAt,omitempty"`
	// Operation completion time
	// +optional
	CompletedAt metav1.Time `json:"completedAt,omitempty"`
	// Operation info message
	// +optional
	Message string `json:"message,omitempty"`
	// Operation error
	// +optional
	Error string `json:"error,omitempty"`
}

// HelmStatus defines the state of the Helm installer
type HelmStatus struct {
	// Phase defines the phase of the Helm installer
	// (queued, started, downloading, running, completed, failed)
	Phase string `json:"helmPhase,omitempty"`
	// URL of the chart
	ChartUrl string `json:"chartUrl,omitempty"`
	// ReleaseVersion is the Helm release version numbering
	// +optional
	ReleaseVersion *int `json:"releaseVersion,omitempty"`
	// Any error that occurred while preparing or executing
	// install/upgrade/delete operation
	Error string `json:"helmError,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Action",type=string,JSONPath=`.status.action.type`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.action.phase`
// +kubebuilder:printcolumn:name="Helm Status",type=string,JSONPath=`.status.helmStatus.helmPhase`
// +kubebuilder:printcolumn:name="Started",type=date,JSONPath=`.status.action.startedAt`
// +kubebuilder:printcolumn:name="Completed",type=date,JSONPath=`.status.action.completedAt`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.action.message`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

type ApplicationResource struct {
	Name             string `json:"name,omitempty"`
	Kind             string `json:"kind,omitempty"`
	UID              string `json:"uid,omitempty"`
	ControlPlaneId   string `json:"controlPlaneId,omitempty"`
	ControlPlaneName string `json:"controlPlaneName,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
