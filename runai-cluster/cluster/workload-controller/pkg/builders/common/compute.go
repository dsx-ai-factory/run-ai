package common

import (
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComputeResources struct {
	Job           client.Object
	ComputeDomain *nvapi.ComputeDomain
}

func NewComputeResourcesFromJob(job client.Object) *ComputeResources {
	return &ComputeResources{
		Job: job,
	}
}

// GetAll returns all the non-nil resources stored in the ComputeResources object
func (cr *ComputeResources) GetAll() []client.Object {
	var all []client.Object
	if cr == nil {
		return all
	}
	if cr.Job != nil {
		all = append(all, cr.Job)
	}
	if cr.ComputeDomain != nil {
		all = append(all, cr.ComputeDomain)
	}

	return all
}
