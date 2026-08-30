// Package endpoints is the shared library for the run.ai/endpoints annotation contract.
package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"

	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// Annotation is the WorkloadRunner annotation that carries the JSON-encoded Config.
	Annotation = "run.ai/endpoints"

	// SchemaVersionV1 is the current payload schema version.
	SchemaVersionV1 = "v1"
)

var (
	ErrEmptyAnnotation          = errors.New("annotation is empty")
	ErrNoEndpoints              = errors.New("at least one endpoint is required")
	ErrUnsupportedSchemaVersion = errors.New("unsupported schemaVersion")
)

// Config is the JSON shape of the run.ai/endpoints annotation.
type Config struct {
	SchemaVersion string     `json:"schemaVersion"`
	Endpoints     []Endpoint `json:"endpoints"`
}

// Endpoint groups a Service with an optional ExternalAccess that fronts it.
type Endpoint struct {
	Services       []Service       `json:"services"`
	ExternalAccess *ExternalAccess `json:"externalAccess,omitempty"`
}

// Service is a single Service entry in an Endpoint. Exactly one of Create or Ref must be set.
type Service struct {
	Create   *corev1.Service `json:"create,omitempty"`
	Ref      *ServiceRef     `json:"ref,omitempty"`
	Auth     *AuthConfig     `json:"auth,omitempty"`
	AuthMode AuthMode        `json:"authMode,omitempty"`
}

// ServiceRef references an existing K8s Service by name and port.
type ServiceRef struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
}

// AuthConfig configures authentication and authorization for a Service.
type AuthConfig struct {
	AuthorizationType AuthorizationType `json:"authorizationType,omitempty"`
	AuthorizedUsers   string            `json:"authorizedUsers,omitempty"`
	AuthorizedGroups  string            `json:"authorizedGroups,omitempty"`
	ForceJsonResponse *bool             `json:"forceJsonResponse,omitempty"`
}

// ExternalAccess is the per-Endpoint external access target. Ingress and Route are mutually exclusive;
// HTTPRoute may be supplied alongside either one to support Gateway API migrations.
type ExternalAccess struct {
	Ingress   *networkingv1.Ingress `json:"ingress,omitempty"`
	Route     *openshiftv1.Route    `json:"route,omitempty"`
	HTTPRoute *gatewayv1.HTTPRoute  `json:"httpRoute,omitempty"`
}

// AuthMode selects the auth-proxy topology for a Service that has Auth configured.
type AuthMode string

const (
	AuthModeStandalone AuthMode = "Standalone"
)

// AuthorizationType selects the authorization policy for an AuthConfig.
type AuthorizationType string

const (
	AuthorizationTypeAuthenticatedUsers AuthorizationType = "authenticatedUsers"
	AuthorizationTypeAuthorizedUsers    AuthorizationType = "authorizedUsers"
	AuthorizationTypeAuthorizedGroups   AuthorizationType = "authorizedGroups"
)

// Has reports whether obj carries the run.ai/endpoints annotation.
func Has(obj metav1.Object) bool {
	_, ok := obj.GetAnnotations()[Annotation]
	return ok
}

// Get parses the Config from obj's annotations. Returns (nil, nil) when the annotation is absent.
func Get(obj metav1.Object) (*Config, error) {
	raw, ok := obj.GetAnnotations()[Annotation]
	if !ok {
		return nil, nil
	}
	if raw == "" {
		return nil, ErrEmptyAnnotation
	}
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("malformed %s annotation: %w", Annotation, err)
	}
	if cfg.SchemaVersion != SchemaVersionV1 {
		return nil, fmt.Errorf("%w %q (expected %q)",
			ErrUnsupportedSchemaVersion, cfg.SchemaVersion, SchemaVersionV1)
	}
	if len(cfg.Endpoints) == 0 {
		return nil, ErrNoEndpoints
	}
	return &cfg, nil
}

// Set serializes cfg to JSON and writes it to obj's annotations under Annotation.
func Set(obj metav1.Object, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("cannot set nil Config on %s annotation", Annotation)
	}
	if len(cfg.Endpoints) == 0 {
		return ErrNoEndpoints
	}
	cfg.SchemaVersion = SchemaVersionV1
	raw, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal Config: %w", err)
	}
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[Annotation] = string(raw)
	obj.SetAnnotations(ann)
	return nil
}

// Remove deletes the run.ai/endpoints annotation from obj. No-op when absent.
func Remove(obj metav1.Object) {
	ann := obj.GetAnnotations()
	if _, ok := ann[Annotation]; !ok {
		return
	}
	delete(ann, Annotation)
	obj.SetAnnotations(ann)
}
