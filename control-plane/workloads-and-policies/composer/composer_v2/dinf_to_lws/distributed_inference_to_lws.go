package dinf_to_lws

import (
	"context"
	"net/http"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	trackerApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	LeaderWorkerSetAPIVersion = constants.LwsGroup + "/" + constants.LwsVersion
	LeaderWorkerSetKind       = constants.LwsKind
)

// LeaderWorkerSetTarget is submitted to workloads-manager, which wraps the
// manifest and internal plugin configuration in a WorkloadRunner.
type LeaderWorkerSetTarget struct {
	LeaderWorkerSet lwsv1.LeaderWorkerSet
	Config          workloadsManagerApi.WorkloadV2Configuration
}

func NewMinimalLeaderWorkerSetTarget() *LeaderWorkerSetTarget {
	target := &LeaderWorkerSetTarget{}
	target.LeaderWorkerSet.TypeMeta = metav1.TypeMeta{
		APIVersion: LeaderWorkerSetAPIVersion,
		Kind:       LeaderWorkerSetKind,
	}
	target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate = &corev1.PodTemplateSpec{}
	target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers = []corev1.Container{{}}
	target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers = []corev1.Container{{}}
	return target
}

type DistributedInferenceToLWSComposer struct {
	fc *composer.FieldsComposer

	WorkloadID string
	LWSName    string

	LeaderAssetsMapping []workloadsApi.AssetsMapping
	WorkerAssetsMapping []workloadsApi.AssetsMapping
	ActualNodePools     []clusterApi.Nodepool

	WorkloadRequirements composer.WorkloadRequirements
}

func (dinfComposer *DistributedInferenceToLWSComposer) WithWorkloadID(workloadID string) *DistributedInferenceToLWSComposer {
	dinfComposer.WorkloadID = workloadID
	return dinfComposer
}

func (dinfComposer *DistributedInferenceToLWSComposer) WithLeaderWorkerSetName(name string) *DistributedInferenceToLWSComposer {
	dinfComposer.LWSName = name
	return dinfComposer
}

func (dinfComposer *DistributedInferenceToLWSComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *DistributedInferenceToLWSComposer {
	dinfComposer.fc.WithStorageClassConfiguration(config)
	return dinfComposer
}

func (dinfComposer *DistributedInferenceToLWSComposer) WithClusterStorageConfig(config composer.ClusterStorageConfig) *DistributedInferenceToLWSComposer {
	dinfComposer.fc.ClusterStorageConfig = config
	return dinfComposer
}

func NewDistributedInferenceToLWSComposer(
	workloadName string,
	projectName string,
	namespace string,
	requester string,
	leaderAssetsMapping []workloadsApi.AssetsMapping,
	workerAssetsMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
	actualNodePools []clusterApi.Nodepool,
) *DistributedInferenceToLWSComposer {
	return &DistributedInferenceToLWSComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      composer.DistributedInference,
			WorkloadName:      workloadName,
			ProjectName:       projectName,
			Namespace:         namespace,
			SubmittedBy:       requester,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
		},
		LeaderAssetsMapping: leaderAssetsMapping,
		WorkerAssetsMapping: workerAssetsMapping,
		ActualNodePools:     actualNodePools,
	}
}

func (dinfComposer *DistributedInferenceToLWSComposer) Compose(
	ctx context.Context,
	src *workloadsApi.DistributedInferenceSpecSpec,
	target *LeaderWorkerSetTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if target == nil {
			return workloadsApi.NewError(http.StatusInternalServerError, "Compose: target is nil")
		}
		if src == nil {
			return nil
		}
		if src.Leader.Get() == nil {
			return workloadsApi.NewError(http.StatusBadRequest, "leader spec is required for distributed inference workload")
		}

		lws := &target.LeaderWorkerSet
		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.CopyString(convutils.Ref(dinfComposer.leaderWorkerSetName()), &lws.Name, composer.StringToString, nil),
			dinfComposer.fc.CopyString(convutils.Ref(dinfComposer.fc.Namespace), &lws.Namespace, composer.StringToString, nil),
			dinfComposer.composeTopology(src.Workers.Get(), src.Worker.Get(), &lws.Spec.LeaderWorkerTemplate.Size),
			dinfComposer.fc.CopyInt(src.Replicas.Get(), &lws.Spec.Replicas, composer.IntToInt[int64], nil),
			dinfComposer.composeStartupPolicy(src.StartupPolicy.Get(), &lws.Spec.StartupPolicy),
			dinfComposer.composeRestartPolicy(src.RestartPolicy.Get(), &lws.Spec.LeaderWorkerTemplate.RestartPolicy),
			dinfComposer.fc.CopyString(convutils.Ref(string(lwsv1.RollingUpdateStrategyType)), &lws.Spec.RolloutStrategy.Type, rolloutStrategyToLWS, nil),
			dinfComposer.fc.ComposeFullySaasInd(&lws.Labels, composer.BoolToMapString, composer.ControlPlaneManagedLabel),
			dinfComposer.fc.CopyBoolean(convutils.Ref(true), &lws.Labels, composer.BoolToMapString, composer.ConvertedFromNativeLabel),
			dinfComposer.fc.CopyString(convutils.Ref(dinfComposer.fc.SubmittedBy), &lws.Annotations, composer.StringToMapString, composer.SubmittedByAnnotation),
			dinfComposer.fc.ComposeNameAndTypeLabels(dinfComposer.fc.WorkloadName, dinfComposer.fc.WorkloadType, LeaderWorkerSetKind, dinfComposer.fc.ProjectName, &lws.Labels, composer.MapToMapString, nil),
			dinfComposer.fc.CopyString(src.Category.Get(), &lws.Labels, composer.StringToMapString, composer.WorkloadCategoryLabel),
		)

		leader := src.Leader.Get()
		runner.Run(dinfComposer.processReplica(ctx, src, leader, lws.Spec.LeaderWorkerTemplate.LeaderTemplate, "leader", dinfComposer.LeaderAssetsMapping, target))

		workers := convutils.Deref(src.Workers.Get(), int32(0))
		worker := src.Worker.Get()
		if worker == nil || (workers == 0 && composer.IsEmpty(*worker)) {
			lws.Spec.LeaderWorkerTemplate.WorkerTemplate = *lws.Spec.LeaderWorkerTemplate.LeaderTemplate.DeepCopy()
			lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Name = "worker"
		} else {
			runner.Run(dinfComposer.processReplica(ctx, src, worker, &lws.Spec.LeaderWorkerTemplate.WorkerTemplate, "worker", dinfComposer.WorkerAssetsMapping, target))
		}

		runner.Run(dinfComposer.processNetworkingFields(src.ServingPort.Get(), target))

		runner.Run(dinfComposer.fc.ComposeMNNVL(leader.Labels, dinfComposer.defaultMNNVLMode(), &target.Config, composer.MNNVLToWorkloadRunner, nil))

		return runner.Error()
	}
}

// defaultMNNVLMode derives the MNNVL default for distributed-inference workloads
// from the cluster's inference-workload-controller GPU network acceleration flag.
// The dinf composer preserves the pre-existing plugin contract of always emitting
// a mode, so an unset cluster flag defaults to None (never left empty).
func (dinfComposer *DistributedInferenceToLWSComposer) defaultMNNVLMode() string {
	config := dinfComposer.fc.ClusterStorageConfig
	if config == nil {
		config = composer.DefaultClusterStorageConfig()
	}
	if config.DinfGPUNetworkAcceleration() {
		return string(workloadsManagerApi.MNNVLMODE_REQUIRED)
	}
	return string(workloadsManagerApi.MNNVLMODE_NONE)
}

func (dinfComposer *DistributedInferenceToLWSComposer) leaderWorkerSetName() string {
	const suffixLength = 4
	if dinfComposer.LWSName != "" {
		return dinfComposer.LWSName
	}
	if len(dinfComposer.WorkloadID) < suffixLength {
		return dinfComposer.fc.WorkloadName
	}
	return dinfComposer.fc.WorkloadName + "-" + dinfComposer.WorkloadID[len(dinfComposer.WorkloadID)-suffixLength:]
}
