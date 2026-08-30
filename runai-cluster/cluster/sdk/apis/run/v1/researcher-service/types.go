// +kubebuilder:object:generate:=true
package researcher_service

import (
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
)

// ResaercherService defines configuration for the researcher-service
type ResearcherService struct {
	// Enabled specifies if the researcher-service should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the researcher-service image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Replicas specifies the number of replicas of the researcher-service
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Args specify the researcher-service cli arguments
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// RunAISharedDirectoryHostPath specifies the shared directory
	// of RunAI pods on the host node
	// +kubebuilder:validation:Optional
	RunAISharedDirectoryHostPath string `json:"runaiSharedDirectoryHostPath,omitempty"`

	// Resources specifies the resource requirements for researcher-service containers
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Service specifies the researcher-service service configuration
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Ingress specifies the ingress configuration of the researcher-service
	// +kubebuilder:validation:Optional
	Ingress *common.Ingress `json:"ingress,omitempty"`

	// GatewayAPI specifies a per-operand Gateway reference override.
	// When set, this takes precedence over the global gatewayAPI name/namespace.
	// +kubebuilder:validation:Optional
	GatewayAPI *common.GatewayRef `json:"gatewayAPI,omitempty"`

	// Route specifies the route configuration of the researcher-service
	// This is only relevant for OpenShift
	// +kubebuilder:validation:Optional
	Route *common.Route `json:"route,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the researcher-service
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the researcher-service
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *ResearcherService) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if r.Enabled == nil {
		r.Enabled = ptr.To[bool](false)
	}

	if r.Image == nil {
		r.Image = &common.Image{}
	}
	r.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("researcher-service"),
	})

	if r.Replicas == nil {
		if globalReplicaCount != nil {
			r.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			r.Replicas = ptr.To[int32](1)
		}
	}

	if r.Args == nil {
		r.Args = &Args{}
	}
	r.Args.SetDefaultsWhereNeeded()

	if r.RunAISharedDirectoryHostPath == "" {
		r.RunAISharedDirectoryHostPath = "/runai/shared"
	}

	if r.Resources == nil {
		r.Resources = &Resources{}
	}
	r.Resources.SetDefaultsWhereNeeded()

	if r.Service == nil {
		r.Service = &Service{}
	}
	r.Service.SetDefaultsWhereNeeded()

	if r.Ingress == nil {
		r.Ingress = &common.Ingress{}
	}
	r.Ingress.SetDefaultsWhereNeeded()

	if r.Route == nil {
		r.Route = &common.Route{}
	}
	r.Route.SetDefaultsWhereNeeded()
}

// Args defines command line arguments for researcher-service
type Args struct {
	// Verbosity specifies the verbosity level of the researcher-service cli
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`

	// AuthEnabled specifies wheather auth is enabled
	// +kubebuilder:validation:Optional
	AuthEnabled *bool `json:"authEnabled,omitempty"`

	// GkeOidcEnvoyHost
	// +kubebuilder:validation:Optional
	GkeOidcEnvoyHost string `json:"gkeOidcEnvoyHost,omitempty"`

	// CacheExpirationPeriod specifies the expiration priod of auth and authz cache values
	CacheExpirationPeriod *time.Duration `json:"cacheExpirationPeriod,omitempty"`
}

func (a *Args) SetDefaultsWhereNeeded() {
	if a.Verbosity == nil {
		a.Verbosity = ptr.To[int](4)
	}

	if a.AuthEnabled == nil {
		a.AuthEnabled = ptr.To[bool](true)
	}

	if a.CacheExpirationPeriod == nil {
		a.CacheExpirationPeriod = pointer.Duration(3 * time.Minute)
	}
}

// Resources define configuration for the researcher-service containers
type Resources struct {
	// ResearcherService specifies the researcher-service container resource requirements
	// +kubebuilder:validation:Optional
	ResearcherService ResearcherServiceRequirements `json:"researcherService,omitempty"`
}

func (r *Resources) SetDefaultsWhereNeeded() {
	if r.ResearcherService.Limits == nil {
		r.ResearcherService.Limits = v1.ResourceList{}
	}

	if r.ResearcherService.Requests == nil {
		r.ResearcherService.Requests = v1.ResourceList{}
	}

	if _, exists := r.ResearcherService.Limits[v1.ResourceCPU]; !exists {
		r.ResearcherService.Limits[v1.ResourceCPU] = resource.MustParse("1000m")
	}

	if _, exists := r.ResearcherService.Limits[v1.ResourceMemory]; !exists {
		r.ResearcherService.Limits[v1.ResourceMemory] = resource.MustParse("1024Mi")
	}

	if _, exists := r.ResearcherService.Requests[v1.ResourceCPU]; !exists {
		r.ResearcherService.Requests[v1.ResourceCPU] = resource.MustParse("500m")
	}

	if _, exists := r.ResearcherService.Requests[v1.ResourceMemory]; !exists {
		r.ResearcherService.Requests[v1.ResourceMemory] = resource.MustParse("512Mi")
	}
}

// ResourcesServiceRequirements specifies container resource
// requirements for the researcher-service
type ResearcherServiceRequirements v1.ResourceRequirements

// Service defines configuration of the researcher-service service
type Service struct {
	// NodePort specifies the node port of the researcher-service service
	// +kubebuilder:validation:Optional
	NodePort int `json:"nodePort,omitempty"`
}

func (s *Service) SetDefaultsWhereNeeded() {
	if s.NodePort == 0 {
		s.NodePort = 32282
	}
}
