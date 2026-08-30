package spec

import (
	"context"
	"encoding/json"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/submit"
	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	mnnvltypes "github.com/run-ai/runai/control-plane/runai-cli/internal/utils/mnnvl"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type InferenceSpec struct {
	ServingPort          openapi.NullableServingPort          `json:"servingPort,omitempty"`
	AutoScaling          openapi.NullableAutoScaling          `json:"autoScaling,omitempty"`
	ServingConfiguration openapi.NullableServingConfiguration `json:"servingConfiguration,omitempty"`
}

func NewInferenceSpec() *GeneralSpec {
	return &GeneralSpec{
		workloadType: spectype.Inference,
		framework:    api.Knative,
	}
}

func (gs *GeneralSpec) GenerateInferenceUpdateSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.InferenceUpdateSpecSpec, error) {
	var (
		inferenceSpec = openapi.InferenceUpdateSpecSpec{}
		err           error
		bytes         []byte
	)
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &inferenceSpec)
	if err != nil {
		return nil, err
	}
	return &inferenceSpec, nil
}

func (gs *GeneralSpec) GenerateInferenceSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.InferenceSpecSpec, error) {
	var (
		inferenceSpec = openapi.InferenceSpecSpec{}
		err           error
		bytes         []byte
	)
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &inferenceSpec)
	if err != nil {
		return nil, err
	}
	return &inferenceSpec, nil
}

func (gs *GeneralSpec) GenerateDistributedInferenceSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.DistributedInferenceSpecSpec, error) {
	spec := openapi.NewDistributedInferenceSpecSpec()

	if err := gs.generateSpec(ctx, clusterId, nodesService); err != nil {
		return nil, err
	}

	leaderSpec, err := convertToLeaderSpec(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	spec.Leader.Set(leaderSpec)

	if err := gs.setDistributedInferenceFields(spec, &gs.flags.DistributedInference, &gs.flags.Lifecycle); err != nil {
		return nil, err
	}

	if err := gs.setWorkerSpec(spec, leaderSpec, &gs.flags.DistributedInference, ctx, clusterId, nodesService); err != nil {
		return nil, err
	}

	// Add MNNVL label to leader spec only (after worker spec is created to avoid copying it)
	gs.setMnnvlLabelOnLeader(spec.Leader.Get(), &gs.flags.DistributedInference)

	return spec, nil
}

func convertToLeaderSpec(fullSpec *GeneralFullSpec) (*openapi.DistributedInferenceLeaderWorkerSpec, error) {
	return convertSpec[openapi.DistributedInferenceLeaderWorkerSpec](fullSpec)
}

func (gs *GeneralSpec) setMnnvlLabelOnLeader(leaderSpec *openapi.DistributedInferenceLeaderWorkerSpec, distributed *submit.DistributedInference) {
	if distributed.Mnnvl == "" || leaderSpec == nil {
		return
	}

	// Add the MNNVL label to the leader spec labels using the cluster SDK constant
	labelName := mnnvltypes.FeatureLabelKey
	labelValue := distributed.Mnnvl
	leaderSpec.Labels = append(leaderSpec.Labels, openapi.Label{
		Name:  *openapi.NewNullableString(&labelName),
		Value: *openapi.NewNullableString(&labelValue),
	})
}

func (gs *GeneralSpec) setAutoScaling(inference *submit.Inference, spec *GeneralFullSpec) {
	autoScaling := openapi.NewAutoScaling()
	fieldSet := false

	if inference.MinReplicas.IsSet() {
		autoScaling.MinReplicas.Set(inference.MinReplicas.Get())
		fieldSet = true
	}

	if inference.MaxReplicas.IsSet() {
		autoScaling.MaxReplicas.Set(inference.MaxReplicas.Get())
		fieldSet = true
	}

	if inference.Metric != "" {
		autoScaling.Metric.Set(&inference.Metric)
		fieldSet = true
	}

	if inference.MetricThreshold.IsSet() {
		autoScaling.MetricThreshold.Set(inference.MetricThreshold.Get())
		fieldSet = true
	}

	if inference.InitialReplicas.IsSet() {
		autoScaling.InitialReplicas.Set(inference.InitialReplicas.Get())
		fieldSet = true
	}

	if inference.ActivationReplicas.IsSet() {
		autoScaling.ActivationReplicas.Set(inference.ActivationReplicas.Get())
		fieldSet = true
	}

	if inference.ConcurrencyHardLimit.IsSet() {
		autoScaling.ConcurrencyHardLimit.Set(inference.ConcurrencyHardLimit.Get())
		fieldSet = true
	}

	if inference.ScaleToZeroRetentionSeconds.IsSet() {
		autoScaling.ScaleToZeroRetentionSeconds.Set(inference.ScaleToZeroRetentionSeconds.Get())
		fieldSet = true
	}

	if inference.ScaleDownDelaySeconds.IsSet() {
		autoScaling.ScaleDownDelaySeconds.Set(inference.ScaleDownDelaySeconds.Get())
		fieldSet = true
	}

	if inference.MetricThresholdPercentage.IsSet() {
		autoScaling.MetricThresholdPercentage.Set(inference.MetricThresholdPercentage.Get())
		fieldSet = true
	}

	if fieldSet {
		spec.AutoScaling.Set(autoScaling)
	}
}

func (gs *GeneralSpec) setServingConfiguration(inference *submit.Inference, spec *GeneralFullSpec) {
	servingConfiguration := openapi.NewServingConfiguration()
	fieldSet := false

	if inference.InitializationTimeoutSeconds.IsSet() {
		servingConfiguration.InitializationTimeoutSeconds.Set(inference.InitializationTimeoutSeconds.Get())
		fieldSet = true
	}

	if inference.RequestTimeoutSeconds.IsSet() {
		servingConfiguration.RequestTimeoutSeconds.Set(inference.RequestTimeoutSeconds.Get())
		fieldSet = true
	}

	if fieldSet {
		spec.ServingConfiguration.Set(servingConfiguration)
	}
}

func (gs *GeneralSpec) setDistributedInferenceFields(
	spec *openapi.DistributedInferenceSpecSpec,
	distributed *submit.DistributedInference,
	lifecycle *submit.Lifecycle,
) error {
	if distributed.Replicas > 0 {
		spec.Replicas.Set(&distributed.Replicas)
	}
	if distributed.Workers.IsSet() {
		spec.Workers.Set(distributed.Workers.Get())
	}
	if distributed.StartupPolicy.IsSet() {
		spec.StartupPolicy = distributed.StartupPolicy
	}
	if lifecycle.DistInferenceRestartPolicy.IsSet() {
		spec.RestartPolicy = lifecycle.DistInferenceRestartPolicy
	}
	if lifecycle.Preemptibility.IsSet() {
		spec.Preemptibility = lifecycle.Preemptibility
	}
	if gs.fullSpec.PriorityClass.IsSet() {
		spec.PriorityClass = gs.fullSpec.PriorityClass
	}
	if len(gs.fullSpec.NodePools) > 0 {
		spec.SetNodePools(gs.fullSpec.NodePools)
	}
	if err := gs.setDistributedInferenceServingPort(&spec.ServingPort); err != nil {
		return err
	}
	return nil
}

func (gs *GeneralSpec) setDistributedInferenceServingPort(servingPort *openapi.NullableDistributedInferenceServingPort) error {
	servingPortInput := gs.flags.DistributedInference.ServingPort
	if len(servingPortInput) == 0 {
		return nil
	}

	// Parse using distributed inference specific validation - returns DistributedInferenceServingPort directly
	parsed, err := submit.ParseDistributedInferenceServingPort(servingPortInput)
	if err != nil {
		return err
	}

	servingPort.Set(parsed)
	return nil
}

func (gs *GeneralSpec) setWorkerSpec(
	spec *openapi.DistributedInferenceSpecSpec,
	leaderSpec *openapi.DistributedInferenceLeaderWorkerSpec,
	distributed *submit.DistributedInference,
	ctx context.Context,
	clusterId string,
	nodesService *nodes.BackendNodeService,
) error {
	if !shouldCreateWorkerSpec(distributed, &gs.flags) {
		return nil
	}

	workerSpec, err := gs.generateWorkerSpec(leaderSpec, distributed, ctx, clusterId, nodesService)
	if err != nil {
		return err
	}

	spec.Worker.Set(workerSpec)
	return nil
}

func shouldCreateWorkerSpec(distributed *submit.DistributedInference, flags *GeneralFlags) bool {
	return (distributed.Workers.IsSet() && *distributed.Workers.Get() > 0) || uniqueWorkerSpecRequired(flags)
}

func (gs *GeneralSpec) generateWorkerSpec(
	leaderSpec *openapi.DistributedInferenceLeaderWorkerSpec,
	distributed *submit.DistributedInference,
	ctx context.Context,
	clusterId string,
	nodesService *nodes.BackendNodeService,
) (*openapi.DistributedInferenceLeaderWorkerSpec, error) {
	if !distributed.WorkerAsLeader || uniqueWorkerSpecRequired(&gs.flags) {
		return gs.createCustomWorkerSpec(ctx, clusterId, nodesService)
	}
	return cloneLeaderSpec(leaderSpec)
}

func (gs *GeneralSpec) createCustomWorkerSpec(
	ctx context.Context,
	clusterId string,
	nodesService *nodes.BackendNodeService,
) (*openapi.DistributedInferenceLeaderWorkerSpec, error) {
	workerSpec, err := GenerateWorkerSpec(gs, ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}
	return convertSpec[openapi.DistributedInferenceLeaderWorkerSpec](workerSpec)
}

func cloneLeaderSpec(leaderSpec *openapi.DistributedInferenceLeaderWorkerSpec) (*openapi.DistributedInferenceLeaderWorkerSpec, error) {
	return convertSpec[openapi.DistributedInferenceLeaderWorkerSpec](leaderSpec)
}

func convertSpec[T any](source any) (*T, error) {
	bytes, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	target := new(T)
	if err := json.Unmarshal(bytes, target); err != nil {
		return nil, err
	}

	return target, nil
}
