// +kubebuilder:object:generate:=true
package pod_group_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"

	v1 "k8s.io/api/core/v1"
)

// PodGroupController defines the configuration of the pod-group-controller deployment
type PodGroupController struct {
	// Enabled specifies if pod-group-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the pod-group-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the pod-group-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the pod-group-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the pod-group-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements
