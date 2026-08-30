// +kubebuilder:object:generate:=true
package status_updater

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// StatusUpdater defines configurations for the status-updater
type StatusUpdater struct {
	// Enabled specifies if the status-updater should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the project controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the project controller
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Resources describes the resource requirements for the pod grouper pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Replicas specifies the number of replicas of the status updater controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the status-updater
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the status-updater
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *StatusUpdater) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if p.Enabled == nil {
		p.Enabled = ptr.To[bool](true)
	}

	if p.Image == nil {
		p.Image = &common.Image{}
	}
	p.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("status-updater"),
	})
	if p.Replicas == nil {
		if globalReplicaCount != nil {
			p.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			p.Replicas = ptr.To[int32](1)
		}
	}
	if p.Args == nil {
		p.Args = &Args{}
	}
	p.Args.SetDefaultsWhereNeeded()

	if p.Resources == nil {
		p.Resources = &Resources{}
	}
	p.Resources.SetDefaultsWhereNeeded()
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

// SetDefaultsWhereNeeded sets default fields for unset fields
func (r *Resources) SetDefaultsWhereNeeded() {
	if r.Limits == nil {
		r.Limits = v1.ResourceList{}
	}

	if r.Requests == nil {
		r.Requests = v1.ResourceList{}
	}

	if _, exists := r.Limits[v1.ResourceCPU]; !exists {
		r.Limits[v1.ResourceCPU] = resource.MustParse("200m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("600Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("100m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("300Mi")
	}
}

// Args defines command line arguments for the status-updater
type Args struct {
	// Verbosity specifies the logging verbosity of the pod grouper
	// use 0 for info, -1 for debug
	// ToDo: verify
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {
	if a.Verbosity == nil {
		a.Verbosity = ptr.To[int](0)
	}
}
