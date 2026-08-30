package validator

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DistributedInferenceParamsValidator struct {
	k8sClient       client.Client
	mnnvlEnabled    bool
	topologyEnabled bool
	elasticEnabled  bool
}

func NewDistributedInferenceParamsValidator(k8sClient client.Client, mnnvlEnabled, topologyEnabled, elasticEnabled bool) *DistributedInferenceParamsValidator {
	return &DistributedInferenceParamsValidator{
		k8sClient:       k8sClient,
		mnnvlEnabled:    mnnvlEnabled,
		topologyEnabled: topologyEnabled,
		elasticEnabled:  elasticEnabled,
	}
}

func (pv *DistributedInferenceParamsValidator) Validate(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	// Validate spec fields
	err := pv.validateSpecFields(workload)
	if err != nil {
		return err
	}

	// Validate exactly one serving port is provided
	err = pv.validateServingPorts(workload.Spec.DistributedInferenceFields.ServingPorts)
	if err != nil {
		return err
	}

	// Validate leader spec using generic validation functions
	err = pv.validateEntitySpec(&workload.LeaderSpec)
	if err != nil {
		return fmt.Errorf("failed to validate leaderSpec: %v", err)
	}

	// Validate worker spec using generic validation functions (if not nil)
	if workload.WorkerSpec != nil {
		err = pv.validateEntitySpec(workload.WorkerSpec)
		if err != nil {
			return fmt.Errorf("failed to validate workerSpec: %v", err)
		}
	}

	// validate MNNVL
	err = pv.validateMnnvl(ctx, workload)
	if err != nil {
		return fmt.Errorf("failed to validate %s: %v", mnnvltypes.UserFacingFeatureName, err)
	}

	// validate topology
	if pv.topologyEnabled {
		err = topology.ValidateTopologyAnnotation(ctx, workload.Annotations)
		if err != nil {
			return fmt.Errorf("failed to validate topology: %v", err)
		}
	}

	return nil
}

func (pv *DistributedInferenceParamsValidator) validateSpecFields(workload *dinfw.DistributedInferenceWorkload) error {
	distributedFields := workload.Spec.DistributedInferenceFields
	// Validate workers field
	workers := 0
	if distributedFields.Workers != nil {
		workers = ptr.Deref(distributedFields.Workers.Value, 0)
	}

	// workers >= 0
	if workers < 0 {
		return fmt.Errorf("workers must be non-negative, got: %d", workers)
	}

	// workers can be zero if and only if workerSpec is nil
	if workers == 0 && workload.WorkerSpec != nil {
		return fmt.Errorf("workers cannot be zero when workerSpec is provided")
	}
	if workers > 0 && workload.WorkerSpec == nil {
		return fmt.Errorf("workerSpec must be provided when workers is greater than zero")
	}

	// Validate replicas field
	replicas := distributedFields.Replicas
	if replicas != nil {
		replicasValue := ptr.Deref(replicas.Value, 1)

		// replicas >= 0
		if replicasValue < 0 {
			return fmt.Errorf("replicas must be non-negative, got: %d", replicasValue)
		}
	}

	// Validate labels
	err := workloads.ValidateLabels(workload.Labels)
	if err != nil {
		return fmt.Errorf("failed to validate labels, error: %v", err)
	}

	return nil
}

func (pv *DistributedInferenceParamsValidator) validateServingPorts(servingPorts *field_types.ServingPortsField) error {
	if servingPorts == nil {
		return fmt.Errorf("servingPorts must be provided")
	}

	servingPortBuilder := builders.NewServingPortBuilder(nil, false)
	_, err := servingPortBuilder.GetPorts(servingPorts)
	if err != nil {
		return fmt.Errorf("failed to validate serving port: %v", err)
	}

	servingPort, err := servingPortBuilder.GetServingPort(servingPorts)
	if err != nil {
		return fmt.Errorf("failed to validate serving port: %v", err)
	}
	_, _, _, err = servingPortBuilder.ValidateAndGetServingPortAuthFields(servingPort)
	if err != nil {
		return fmt.Errorf("failed to validate serving port auth fields: %v", err)
	}

	return nil
}

func (pv *DistributedInferenceParamsValidator) validateEntitySpec(spec *dp.DistributedInferenceEntitySpec) error {
	// Validate image
	err := workloads.ValidateImage(spec.ContainerFields.Image.Deref())
	if err != nil {
		return fmt.Errorf("failed to validate image: %v", err)
	}

	// Validate PVCs
	err = workloads.ValidatePVCs(spec.StorageFields)
	if err != nil {
		return fmt.Errorf("failed to validate PVCs: %v", err)
	}

	// Validate EmptyDirs
	err = workloads.ValidateEmptyDirs(spec.StorageFields.EmptyDirVolumes)
	if err != nil {
		return fmt.Errorf("failed to validate EmptyDir volumes: %v", err)
	}

	// Validate pod affinity
	err = workloads.ValidatePodAffinity(spec.DistributedInferenceEntitySchedulerFields.PodAffinity, spec.DistributedInferenceEntitySchedulerFields.PodAffinityTopology)
	if err != nil {
		return fmt.Errorf("failed to validate pod affinity: %v", err)
	}

	// Validate probes
	err = workloads.ValidateProbes(spec.ContainerFields)
	if err != nil {
		return fmt.Errorf("failed to validate probes: %v", err)
	}

	// Validate resources
	err = workloads.ValidateWorkloadContainerResourceValues(&spec.ResourceAllocationFields, "")
	if err != nil {
		return fmt.Errorf("failed to validate resource values: %v", err)
	}

	return nil
}

func (pv *DistributedInferenceParamsValidator) validateMnnvl(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	mnnvlMode, ok := workload.Labels[mnnvltypes.FeatureLabelKey]
	if !ok || mnnvlMode == mnnvltypes.None {
		// MNNVL is not set, nothing to validate
		return nil
	}

	// MNNVL is set, validate its value
	if mnnvlMode != mnnvltypes.Required {
		return fmt.Errorf("invalid mode: %s, only %s and %s are allowed", mnnvlMode, mnnvltypes.None, mnnvltypes.Required)
	}

	// MNNVL is set to Required
	// validate the cluster support
	if !pv.mnnvlEnabled {
		return fmt.Errorf("the capability is not enabled in the cluster")
	}

	// validate the workload support
	// validate the node pools support MNNVL
	mnnvlBuilder := builders.NewMnnvlBuilder(pv.k8sClient, pv.elasticEnabled)
	isNodePoolsSupportMnnvl, err := mnnvlBuilder.IsNodePoolsSupportMnnvl(ctx, workload)
	if err != nil {
		return fmt.Errorf("failed to validate node pools support: %v", err)
	}
	if !isNodePoolsSupportMnnvl {
		return fmt.Errorf("found at least one node pool that does not support the capability")
	}

	return nil
}
