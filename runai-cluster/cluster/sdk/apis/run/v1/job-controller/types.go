// +kubebuilder:object:generate:=true
package job_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// JobController defines the configuration for the job-controller
type JobController struct {
	// Enabled specifies if the job-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the job controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments passed to the job controller CLI
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Resources describes the resource requirements for the job controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// K8sClientBurst specifies the burst rate for the k8s client
	// +kubebuilder:validation:Optional
	K8sClientBurst *string `json:"k8sClientBurst,omitempty"`

	// K8sClientBurst specifies the QPS rate for the k8s client
	// +kubebuilder:validation:Optional
	K8sClientQPS *string `json:"k8sClientQPS,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the job-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the job-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (j *JobController) SetDefaultsWhereNeeded(globalRegistry string) {
	if j.Enabled == nil {
		j.Enabled = ptr.To[bool](true)
	}

	if j.Image == nil {
		j.Image = &common.Image{}
	}
	j.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("runaijob-controller"),
	})
	if j.Args == nil {
		j.Args = &Args{}
	}
	j.Args.SetDefaultsWhereNeeded()

	if j.Resources == nil {
		j.Resources = &Resources{}
	}
	j.Resources.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("400m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("200m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}

// Args defines command line arguments for the job-controller
type Args struct {
	// Verbosity specifies the logging verbosity of the job controller
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {
	if a.Verbosity == nil {
		a.Verbosity = ptr.To[int](4)
	}
}
