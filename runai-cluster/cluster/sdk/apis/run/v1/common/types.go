// +kubebuilder:object:generate:=true
package common

import (
	"fmt"
	"os"

	"k8s.io/utils/pointer"

	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	AgentSyncIntervalMillisDefault = 10000

	DefaultRegistryEnvVarName = "IMG_DEFAULT_REGISTRY"
	DefaultVersionEnvVarName  = "IMG_DEFAULT_VERSION"

	DefaultEnv         = "prod"
	DefaultRepositeory = "runai.jfrog.io/op-containers-"
)

// ImageSpec is a struct describing a container image
type Image struct {
	// DEPRECATED: Use Registry instead
	// Repository is the repository prefix for the image
	// +kubebuilder:validation:Optional
	Repository *string `json:"repository,omitempty"`

	// DEPRECATED: Use Registry instead
	// Env is the repository suffix for the image
	// +kubebuilder:validation:Optional
	Env *string `json:"env,omitempty"`

	// Registry is the registry of the image
	// +kubebuilder:validation:Optional
	Registry *string `json:"registry,omitempty"`

	// Name is the name of the image
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// Tag is the tag of the image
	// +kubebuilder:validation:Optional
	Tag *string `json:"tag,omitempty"`

	// PullPolicy is the pull policy of the image
	// +kubebuilder:validation:Optional
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (i *Image) SetDefaultsWhereNeeded(defaultImageValues *Image) {
	/*
		Image registry precedence (from highest to lowest):
		1. Specified registry
		2. Specified repository/env
		3. Global registry
		4. Default registry
	*/
	if i.Registry == nil {
		if i.Repository != nil || i.Env != nil {
			// Backward compatibility
			if i.Repository == nil {
				i.Repository = ptr.To[string](DefaultRepositeory)
			}
			if i.Env == nil {
				i.Env = ptr.To[string](DefaultEnv)
			}
			i.Registry = ptr.To[string](fmt.Sprintf("%s%s", *i.Repository, *i.Env))
		} else {
			if defaultImageValues.Registry != nil {
				i.Registry = defaultImageValues.Registry
			} else {
				i.Registry = ptr.To[string](os.Getenv(DefaultRegistryEnvVarName))
			}
		}
	}

	if i.Name == nil {
		i.Name = defaultImageValues.Name
	}

	if i.Tag == nil {
		if defaultImageValues.Tag != nil {
			i.Tag = defaultImageValues.Tag
		} else {
			i.Tag = ptr.To[string](os.Getenv(DefaultVersionEnvVarName))
		}
	}

	if i.PullPolicy == "" {
		if defaultImageValues.PullPolicy != "" {
			i.PullPolicy = defaultImageValues.PullPolicy
		} else {
			i.PullPolicy = v1.PullAlways
		}
	}
}

type Service struct {
	// Port specifies the service port
	Port *int `json:"port,omitempty"`

	// ServiceType specifies the service type
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`
	// TargetPort specifies the target port in the container
	TargetPort *int `json:"targetPort,omitempty"`
}

func (service *Service) SetDefaultsWhereNeeded(port, targetPort int, serviceType v1.ServiceType) {
	if service.Port == nil {
		service.Port = ptr.To[int](port)
	}
	if service.TargetPort == nil {
		service.TargetPort = ptr.To[int](targetPort)
	}
	if service.ServiceType == "" {
		service.ServiceType = serviceType
	}
}

// Ingress defines configuration of the researcher-service ingress
type Ingress struct {
	// TLSEnabled specifies whether the ingress should utilize TLS
	TLSEnabled *bool `json:"tlsEnabled,omitempty"`

	// TlsSecret specifies the name of the secret that holds the TLS info of the ingress, in case one exists in advance.
	// +kubebuilder:validation:Optional
	TlsSecret string `json:"tls-secret,omitempty"`

	// Annotations allows to configure the ingress according to the ingress-class specifications, defaults for nginx.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Class specifies the ingress-class to be used for this ingress, defaults to nginx.
	Class string `json:"ingressClass,omitempty"`
}

func (in *Ingress) SetDefaultsWhereNeeded() {
	if in.TLSEnabled == nil {
		in.TLSEnabled = pointer.Bool(true)
	}

	if in.TlsSecret == "" {
		in.TlsSecret = "runai-cluster-domain-tls-secret"
	}

	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
}

// Route defines configuration of the researcher-service route
type Route struct {
	// TlsSecret specifies the name of the secret that holds the TLS info of the route, in case one exists in advance.
	// +kubebuilder:validation:Optional
	TlsSecret string `json:"tlsSecret,omitempty"`
}

func (r *Route) SetDefaultsWhereNeeded() {
}

// GatewayRef specifies a reference to a Gateway resource for per-operand override.
// When provided, both Name and Namespace must be non-empty.
// +kubebuilder:validation:XValidation:rule="!has(self.name) && !has(self.__namespace__) || (has(self.name) && self.name != '' && has(self.__namespace__) && self.__namespace__ != '')",message="both name and namespace are required when gatewayAPI override is specified"
type GatewayRef struct {
	// Name is the name of the Gateway resource.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the Gateway resource.
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace,omitempty"`
}

type CustomCA struct {
	// Enabled specifies whether to use a custom CA
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Secret specifies the selector for the CA Secret
	// +kubebuilder:validation:Optional
	Secret *v1.SecretKeySelector `json:"secret,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (c *CustomCA) SetDefaultsWhereNeeded() {
	if c.Enabled == nil {
		c.Enabled = ptr.To[bool](false)
	}

	if c.Secret == nil {
		c.Secret = &v1.SecretKeySelector{}
	}

	if c.Secret.Name == "" {
		// Align with the name the platform actually creates/replicates the CA under
		// (chart default, installer, project-controller).
		c.Secret.Name = "runai-ca-cert"
	}

	if c.Secret.Key == "" {
		c.Secret.Key = "runai-ca.pem"
	}

	if c.Secret.Optional == nil {
		c.Secret.Optional = new(bool)
		*c.Secret.Optional = true
	}
}
