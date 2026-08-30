// +kubebuilder:object:generate:=true
package metrics_exporter

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

type MetricsExporter struct {
	// Enabled specifies if the metrics-exporter should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the metrics-exporter image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the metrics exporter pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Service specifies the service configuration for the metrics-exporter
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Replicas specifies the number of replicas of the metrics exporter controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the metrics-exporter
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the metrics-exporter
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *MetricsExporter) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if p.Enabled == nil {
		p.Enabled = ptr.To[bool](true)
	}

	if p.Image == nil {
		p.Image = &common.Image{}
	}
	p.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("metrics-exporter"),
	})
	if p.Replicas == nil {
		if globalReplicaCount != nil {
			p.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			p.Replicas = ptr.To[int32](1)
		}
	}
	if p.Resources == nil {
		p.Resources = &Resources{}
	}
	p.Resources.SetDefaultsWhereNeeded()

	if p.Service == nil {
		p.Service = &Service{}
	}
	p.Service.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("1000m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("300m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}

// Service defines configuration for the metrics-exporter service
type Service struct {
	// Type specifies the service type
	// +kubebuilder:validation:Optional
	Type v1.ServiceType `json:"type,omitempty"`

	// Port specifies the service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the target port in the container
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

func (s *Service) SetDefaultsWhereNeeded() {
	if s.Type == "" {
		s.Type = v1.ServiceTypeClusterIP
	}
	if s.Port == nil {
		s.Port = ptr.To[int](8080)
	}
	if s.TargetPort == nil {
		s.TargetPort = ptr.To[int](8080)
	}
}
