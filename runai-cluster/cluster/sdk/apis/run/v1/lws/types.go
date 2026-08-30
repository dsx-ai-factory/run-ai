// +kubebuilder:object:generate:=true
package lws

import (
	"k8s.io/utils/ptr"
)

type LWS struct {
	// Namespace specifies the namespace where the LWS operator is deployed
	// +kubebuilder:validation:Optional
	Namespace *string `json:"namespace,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *LWS) SetDefaultsWhereNeeded() {
	if r.Namespace == nil {
		r.Namespace = ptr.To("lws-system")
	}
}
