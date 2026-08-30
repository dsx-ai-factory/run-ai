package mpi

import (
	"errors"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
)

const (
	ErrorMpiInvalidLCPValue = "LauncherCreationPolicy invalid value \"%s\""
)

var (
	ErrorMpiMasterNoGPUAllowed = errors.New("GPU resources are not allowed on MPI master pod. Please resubmit without master GPU fields")
)

// ValidateWorkload validates the configuration of a JAX job for unsupported fields.
func ValidateWorkload(workload *v2alpha1.DistributedWorkload) error {
	err := masterResourceGPUCheck(workload)
	if err != nil {
		return err
	}
	err = launcherCreationPolicyCheck(workload.Spec.DistributedFields.MpiJob)
	return err
}

func masterResourceGPUCheck(workload *v2alpha1.DistributedWorkload) error {
	if workload.MasterSpec == nil {
		return nil
	}

	if workload.MasterSpec.ResourceAllocationFields.GpuDevices != nil && workload.MasterSpec.ResourceAllocationFields.GpuDevices.Deref() > 0 {
		return ErrorMpiMasterNoGPUAllowed
	}
	if workload.MasterSpec.ResourceAllocationFields.Gpu != nil && workload.MasterSpec.ResourceAllocationFields.Gpu.Deref() > 0 {
		return ErrorMpiMasterNoGPUAllowed
	}
	if workload.MasterSpec.ResourceAllocationFields.GpuLimit != nil && workload.MasterSpec.ResourceAllocationFields.GpuLimit.Deref() > 0 {
		return ErrorMpiMasterNoGPUAllowed
	}
	return nil
}

func launcherCreationPolicyCheck(mpiField *field_types.MpiField) error {
	if mpiField == nil || mpiField.Values == nil || mpiField.Values.LauncherCreationPolicy == nil {
		return nil
	}

	switch v2beta1.LauncherCreationPolicy(mpiField.Values.LauncherCreationPolicy.Deref()) {
	case v2beta1.LauncherCreationPolicyAtStartup, v2beta1.LauncherCreationPolicyWaitForWorkersReady:
		return nil
	}
	return fmt.Errorf(ErrorMpiInvalidLCPValue, mpiField.Values.LauncherCreationPolicy.Deref())
}
