package runai_workloads

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	trackerApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type DistributedWorkloadComposer struct {
	// we have one composer for each spec, both sharing the same fields composer (fc)
	fc                *composer.FieldsComposer
	workerJobComposer JobSpecComposer
	masterJobComposer JobSpecComposer

	// we also have a notifications composer with its own fields composer
	notificationsComposer *notifications.NotificationsComposer

	workerAssetsInfo *composer.ComplimentaryAssetInfo
	masterAssetsInfo *composer.ComplimentaryAssetInfo
}

func NewDistributedWorkloadComposer(
	workloadName string,
	requester string,
	notificationsComposer *notifications.NotificationsComposer,
	options composer.FieldsComposerOptions,
) *DistributedWorkloadComposer {
	fc := composer.FieldsComposer{
		WorkloadType:      string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		WorkloadName:      workloadName,
		SubmittedBy:       requester,
		Options:           options,
		ItemizedNumbering: make(itemized.ItemizedNumbering),
		OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
	}

	result := DistributedWorkloadComposer{
		fc: &fc,

		workerJobComposer: JobSpecComposer{
			fc:          &fc,
			ReplicaType: convutils.Ref(workloadsApi.REPLICATYPE_WORKER),
		},

		masterJobComposer: JobSpecComposer{
			fc:          &fc,
			ReplicaType: convutils.Ref(workloadsApi.REPLICATYPE_MASTER),
		},

		notificationsComposer: notificationsComposer,
	}

	return &result
}

func (dwc *DistributedWorkloadComposer) WithComplimentaryAssetInfo(replicaType workloadsApi.ReplicaType,
	assetInfo *composer.ComplimentaryAssetInfo,
) *DistributedWorkloadComposer {
	if replicaType == workloadsApi.REPLICATYPE_WORKER {
		dwc.workerAssetsInfo = assetInfo
	} else {
		dwc.masterAssetsInfo = assetInfo
	}

	return dwc
}

func (dwc *DistributedWorkloadComposer) WithAssetMapping(assetsMapping []workloadsApi.AssetsMapping) *DistributedWorkloadComposer {
	dwc.workerJobComposer.AssetsMapping = assetsMapping
	dwc.masterJobComposer.AssetsMapping = assetsMapping

	return dwc
}

func (dwc *DistributedWorkloadComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *DistributedWorkloadComposer {
	dwc.fc.WithStorageClassConfiguration(config)
	return dwc
}

func (dwc *DistributedWorkloadComposer) Compose(ctx context.Context,
	src *workloadsApi.SupersetWorkload, target *jobs.JobWorkload,
) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	// backward compatibility with old composer - with noMaster the replica type is set to nil
	noMaster := !src.MasterSpec.IsSet() && !convutils.BoolIsTrue(src.MasterSpecSameAsWorker.Get())
	if noMaster {
		dwc.workerJobComposer.ReplicaType = nil
	}

	runner := &composer.StepRunner{}

	runner.Batch(
		dwc.workerJobComposer.Compose(ctx, src.Spec, &target.Spec),
		dwc.ComposeDistributedFields(ctx, &dwc.workerJobComposer, src.Spec, &target.Spec),
	)

	if runner.Error() != nil {
		return runner.Error()
	}

	if dwc.notificationsComposer != nil {
		err := dwc.notificationsComposer.Compose(ctx, dwc.workerJobComposer.ReplicaType, dwc.workerAssetsInfo, src.Spec)()
		if err != nil {
			return err
		}
	}

	if src.MasterSpec.IsSet() {
		var masterSpecAsSuperset workloadsApi.SupersetSpec
		err := convutils.CopySpecs(&masterSpecAsSuperset, src.MasterSpec.Get())
		if err != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, err.Error())
		}

		target.MasterSpec = &jobs.JobSpec{}

		runner.Batch(
			dwc.masterJobComposer.Compose(ctx, &masterSpecAsSuperset, target.MasterSpec),
			dwc.ComposeDistributedFields(ctx, &dwc.masterJobComposer, &masterSpecAsSuperset, target.MasterSpec),
		)

		if runner.Error() != nil {
			return runner.Error()
		}

		if dwc.notificationsComposer != nil {
			err := dwc.notificationsComposer.Compose(ctx, dwc.masterJobComposer.ReplicaType, dwc.masterAssetsInfo, &masterSpecAsSuperset)()
			if err != nil {
				return err
			}
		}

	} else if noMaster {
		target.Spec.NoMaster = field_types.NewOptionalBoolField(convutils.Ref(true))
	}

	return runner.Error()
}

func (dwc *DistributedWorkloadComposer) ComposeDistributedFields(_ context.Context, jsw *JobSpecComposer, src *workloadsApi.SupersetSpec,
	target *jobs.JobSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		df := NewDistributedFieldsComposer(jsw.fc, src.DistributedFramework.Get(), jsw.ReplicaType)

		runner := &composer.StepRunner{}

		runner.Batch(
			df.ComposeJobType(composer.EnumToString(src.DistributedFramework.Get()), &target.JobType, composer.StringToString, nil),
			df.ComposeNumWorkers(src.NumWorkers.Get(), &target.Workers, intToIntField, nil),
			df.ComposeBackoffLimit(src.BackoffLimit.Get(), &target.RunPolicy, intToRunPolicyBackoffLimit, nil),
			df.ComposeCleanPodPolicy(composer.EnumToString(src.CleanPodPolicy.Get()), &target.RunPolicy, stringToRunPolicyCleanPod, nil),
			df.ComposeMpiSlotsPerWorker(src.SlotsPerWorker.Get(), &target.MpiJob, intToSlotsPerWorker, nil),
			df.ComposeMpiLauncherCreationPolicy(src.MpiLauncherCreationPolicy.Get(), &target.MpiJob, stringToLauncherCreationPolicy, nil),
			df.ComposeMpiSshAuthMountPath(src.SshAuthMountPath.Get(), &target.MpiJob, stringToSshAuthMountPath, nil),
			df.ComposePytorchMinReplicas(src.MinReplicas.Get(), &target.PyTorchJob, intToMinReplicas, nil),
			df.ComposePytorchMaxReplicas(src.MaxReplicas.Get(), src.MinReplicas.Get(), &target.PyTorchJob, intToMaxReplicas, nil),
		)

		return runner.Error()
	}
}

func (dwc *DistributedWorkloadComposer) Requirements() composer.WorkloadRequirements {
	return dwc.workerJobComposer.WorkloadRequirements.Or(dwc.masterJobComposer.WorkloadRequirements)
}
