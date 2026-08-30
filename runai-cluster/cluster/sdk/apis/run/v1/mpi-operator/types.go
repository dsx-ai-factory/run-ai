// +kubebuilder:object:generate:=true
package mpi_operator

import (
	"k8s.io/utils/ptr"
)

type MpiOperator struct {
	// MonitorPortNumber specifies the port number of monitoring port in mpi deployment
	// +kubebuilder:validation:Optional
	MonitorPortNumber *int32 `json:"monitorPortNumber,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *MpiOperator) SetDefaultsWhereNeeded() {
	if r.MonitorPortNumber == nil {
		r.MonitorPortNumber = ptr.To[int32](8081)
	}
}
