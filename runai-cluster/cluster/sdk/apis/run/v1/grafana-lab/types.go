// +kubebuilder:object:generate:=true
package grafana_lab

import v1 "k8s.io/api/core/v1"

const defaultCredentialsSecretName = "runai-grafana-secret"

// GrafanaLab contains configuration for grafana labs
type GrafanaLab struct {
	// UsernameSecret is the reference to a secret containing the username for the grafana labs instance
	// +kubebuilder:validation:Optional
	UsernameSecret *v1.SecretKeySelector `json:"usernameSecret,omitempty"`

	// PasswordSecret is the reference to a secret containing the password for the grafana labs instance
	// +kubebuilder:validation:Optional
	PasswordSecret *v1.SecretKeySelector `json:"passwordSecret,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (g *GrafanaLab) SetDefaultsWhereNeeded() {
	if g.UsernameSecret == nil {
		g.UsernameSecret = &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{Name: defaultCredentialsSecretName},
			Key:                  "grafanaLabUsername",
		}
	}
	if g.PasswordSecret == nil {
		g.PasswordSecret = &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{Name: defaultCredentialsSecretName},
			Key:                  "grafanaLabPassword",
		}
	}
}
