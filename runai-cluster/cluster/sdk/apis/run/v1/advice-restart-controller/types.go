// +kubebuilder:object:generate:=true
package advice_restart_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// AdviceRestartController defines configurations for the advice-restart-controller
type AdviceRestartController struct {
	// Enabled specifies if the advice-restart-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the advice-restart-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the advice-restart-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the controller's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// Service describes the service for the advice-restart-controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Mode specifies the controller operation mode: "log" or "action"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=log;action
	Mode *string `json:"mode,omitempty"`

	// RestartLimit is the cluster-wide restart cap per workload
	// +kubebuilder:validation:Optional
	RestartLimit *int `json:"restartLimit,omitempty"`

	// MaxRetriesPerAdvice is the max retry count for creating restart requests per advice
	// +kubebuilder:validation:Optional
	MaxRetriesPerAdvice *int `json:"maxRetriesPerAdvice,omitempty"`

	// Metric specifies the service configuration for the metrics endpoint
	// +kubebuilder:validation:Optional
	Metric *common.Service `json:"metrics,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *AdviceRestartController) SetDefaultsWhereNeeded(globalRegistry string) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(false)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("advice-restart-controller"),
	})

	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()

	if a.K8sClientConfig == nil {
		a.K8sClientConfig = &K8sClientConfig{}
	}
	a.K8sClientConfig.SetDefaultsWhereNeeded()

	if a.UseDebugLogLevel == nil {
		a.UseDebugLogLevel = ptr.To(false)
	}

	if a.Service == nil {
		a.Service = &Service{}
	}
	a.Service.SetDefaultsWhereNeeded()

	if a.Mode == nil {
		a.Mode = ptr.To("action")
	}

	if a.RestartLimit == nil {
		a.RestartLimit = ptr.To(3)
	}

	if a.MaxRetriesPerAdvice == nil {
		a.MaxRetriesPerAdvice = ptr.To(3)
	}

	if a.Metric == nil {
		a.Metric = &common.Service{}
	}
	a.Metric.SetDefaultsWhereNeeded(8383, 8383, v1.ServiceTypeClusterIP)
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
		r.Limits[v1.ResourceMemory] = resource.MustParse("256Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("100m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("128Mi")
	}
}

// Service defines configuration for the advice-restart-controller service
type Service struct {
	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook Webhook `json:"webhook,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	s.Webhook.SetDefaultsWhereNeeded()
}

// Webhook defines configuration for the advice-restart-controller webhook
type Webhook struct {
	// Port specifies the webhook service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the webhook service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (w *Webhook) SetDefaultsWhereNeeded() {
	if w.Port == nil {
		w.Port = ptr.To(443)
	}

	if w.TargetPort == nil {
		w.TargetPort = ptr.To(8443)
	}
}

// K8sClientConfig specifies the configuration of k8s client
type K8sClientConfig struct {
	// Burst specifies the burst rate for the k8s client
	// +kubebuilder:validation:Optional
	Burst *int `json:"burst,omitempty"`

	// QPS specifies the QPS rate for the k8s client
	// +kubebuilder:validation:Optional
	QPS *int `json:"qps,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (c *K8sClientConfig) SetDefaultsWhereNeeded() {
	if c.QPS == nil {
		c.QPS = ptr.To(50)
	}

	if c.Burst == nil {
		c.Burst = ptr.To(300)
	}
}
