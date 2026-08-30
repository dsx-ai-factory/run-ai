package types

import "time"

type ProvisioningMode string

const (
	ProvisioningModeControlPlaneOnly ProvisioningMode = "control-plane-only"
	ProvisioningModeFull             ProvisioningMode = "full"
)

type CreateEnvironmentRequest struct {
	ProvisioningMode       ProvisioningMode `json:"provisioningMode"`
	Domain                 string           `json:"domain"`
	WildcardDomain         string           `json:"wildcardDomain"`
	KubernetesDistribution *string          `json:"kubernetesDistribution,omitempty"`
	UserEmail              *string          `json:"userEmail,omitempty"`
}

type EnvironmentID string

type EnvironmentPhase string

const (
	EnvironmentPhasePending             EnvironmentPhase = "pending"
	EnvironmentPhaseInProgress          EnvironmentPhase = "in_progress"
	EnvironmentPhaseCreatingRunaiAssets EnvironmentPhase = "creating_runai_metadata"
	EnvironmentPhaseSucceeded           EnvironmentPhase = "succeeded"
	EnvironmentPhaseFailed              EnvironmentPhase = "failed"
)

// RunaiUserCredentials holds the created L1 researcher user credentials
type RunaiUserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EnvironmentStatus struct {
	ID              EnvironmentID         `json:"id"`
	Phase           EnvironmentPhase      `json:"phase"`
	Message         string                `json:"message,omitempty"`
	UpdatedAt       time.Time             `json:"updatedAt"`
	Mode            ProvisioningMode      `json:"mode,omitempty"`
	UserCredentials *RunaiUserCredentials `json:"userCredentials,omitempty"` // only for full mode
}
