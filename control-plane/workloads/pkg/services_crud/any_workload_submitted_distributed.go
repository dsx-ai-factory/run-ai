package services_crud

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/composer_v2/distributed_to_kubeflow"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// AnyWorkloadSubmitterDistributed handles submission decisions for distributed workloads.
// It composes workloads into one of the supported kubeflow targets and decides between workload-runner and native submission.
type AnyWorkloadSubmitterDistributed struct{}

// DetermineSubmissionMethods determines b/w native or any-workload submission.
func (s *AnyWorkloadSubmitterDistributed) DetermineSubmissionMethods(ctx context.Context, wc *workloadContext) SubmissionMethods {
	if wc.segmentRouting != nil && wc.segmentRouting.hasIntent {
		// Segment intent does not override the ordinary submission-method
		// decision. Topology enrichment already ran during preprocessing using
		// this same cached decision, so we must reuse it here rather than
		// recompute and risk diverging. A segmented request that could not be
		// validated fails closed rather than silently choosing a path.
		if !wc.segmentRouting.validated || wc.segmentRouting.err != nil {
			return SubmissionMethods{}
		}
		return wc.segmentRouting.method
	}
	// Segmented routing must be evaluated during preprocessing so topology
	// enrichment and submission consume the same capability/tenant decision.
	// Fail closed if a caller constructs a segmented context without it.
	if hasDistributedSegmentIntent(wc.spec) {
		return SubmissionMethods{}
	}
	return determineSubmissionMethods(ctx, wc.tenantId, wc.workloadType, wc.clusterCapabilities)
}

// CanUpdate checks if this submitter can handle workload updates for a specific workload.
// Distributed workloads do not support updates.
func (s *AnyWorkloadSubmitterDistributed) CanUpdate(_ context.Context, _ SubmissionMethods, _ *workloadContext, _ string) *workloadsApi.SubmissionErrorV2 {
	return nil
}

// Create handles workload creation for distributed workloads.
func (s *AnyWorkloadSubmitterDistributed) Create(ctx context.Context, wc *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2) {
	projectName := wc.create.project.Name
	namespace := wc.create.namespace

	options := wc.options
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, wc.clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}
	// v1 parity: distributed workloads delete the PVCs created for them on terminal phase.
	options.CleanupCreatedPvcsOnCompletion = convutils.Ref(true)

	distComposer := distributed_to_kubeflow.NewDistributedToKubeflowComposer(
		wc.workloadType,
		wc.workloadName,
		projectName,
		namespace,
		wc.createdBy,
		wc.create.assetsMapping,
		wc.create.assetsMapping,
		options,
		wc.create.actualNodePools,
	)

	clusterStorageConfig, cfgErr := buildClusterStorageConfig(ctx, wc.create.createParams.Response.ClusterId)
	if cfgErr != nil {
		return nil, cfgErr
	}

	distComposer.WithStorageClassConfiguration(wc.storageClassConfigs)
	distComposer.WithDataVolumeClaimNames(wc.dataVolumeClaimNames)
	distComposer.WithClusterStorageConfig(clusterStorageConfig)

	if !convutils.HasValue(wc.spec.DistributedFramework) {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "distributed framework is not specified")
	}

	framework := *wc.spec.DistributedFramework.Get()
	if framework == "" || !framework.IsValid() {
		return nil, workloadsApi.NewSubmissionErrorV2(
			http.StatusBadRequest,
			fmt.Sprintf("unsupported distributed framework %q: valid values are %v",
				framework, workloadsApi.AllowedDistributedFrameworkEnumValues),
		)
	}

	isMaster := wc.masterSameAsWorker || wc.masterSpec != nil
	target := distributed_to_kubeflow.NewMinimalDistributedTarget(framework, isMaster)
	if target == nil || target.Manifest() == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(
			http.StatusBadRequest,
			fmt.Sprintf("unsupported distributed framework %q: failed to build manifest target", framework),
		)
	}

	var distributedSpec workloadsApi.DistributedSpecSpec
	err := convutils.CopySpecs(&distributedSpec, wc.spec)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}

	masterAndWorker := workloadsApi.DistributedWorkload{
		Spec:                   &distributedSpec,
		MasterSpec:             *workloadsApi.NewNullableMasterSpec(wc.masterSpec),
		MasterSpecSameAsWorker: *workloadsApi.NewNullableBool(&wc.masterSameAsWorker),
	}

	if composerErr := distComposer.Compose(ctx, &masterAndWorker, target)(); composerErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	composedManifest := target.Manifest()
	if composedManifest == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(
			http.StatusBadRequest,
			fmt.Sprintf("failed to compose distributed workload manifest for framework %q", framework),
		)
	}

	manifest, err := gojsonq.LoadJson(composedManifest)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if len(manifest) == 0 {
		return nil, workloadsApi.NewSubmissionErrorV2(
			http.StatusBadRequest,
			fmt.Sprintf("composed distributed workload manifest is empty for framework %q", framework),
		)
	}

	if metadataErr := composeDistributedConnectionMetadata(ctx, wc, projectName); metadataErr != nil {
		return nil, metadataErr
	}

	return &Submission{
		Request: workloadMgrApi.WorkloadV2CreateRequest{
			Metadata: workloadMgrApi.WorkloadV2MetadataCreateParams{
				Name:          &wc.create.createParams.Response.Name,
				ProjectId:     &wc.create.createParams.Response.ProjectId,
				Configuration: &target.Config,
			},
			Manifest: manifest,
		},
		Requirements:             distComposer.WorkloadRequirements,
		ConnectionMetadata:       newWorkloadsServiceConnectionMetadata(wc.create.notificationsComposer.Output),
		ConnectionCreationFields: wc.create.notificationsComposer.Output,
	}, nil
}

func (s *AnyWorkloadSubmitterDistributed) Update(ctx context.Context, wc *workloadContext, workloadId string,
	existingWorkload *workloadMgrApi.WorkloadV2GetResponse, mergedWorkload *models_crud.Workload,
) (*Update, *workloadsApi.SubmissionErrorV2) {
	return nil, nil
}

func composeDistributedConnectionMetadata(ctx context.Context, wc *workloadContext, projectName string) *workloadsApi.SubmissionErrorV2 {
	if wc.create.notificationsComposer == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "workload notification composer was not initialized")
	}

	notificationsComposer := wc.create.notificationsComposer.WithConnectionsPreProcessor(projectName, composer.NetworkingResourcePrefixKubeflow)
	workerReplicaType := workloadsApi.REPLICATYPE_WORKER
	if composerErr := notificationsComposer.Compose(ctx, &workerReplicaType, wc.create.complimentaryAssetInfo, wc.spec)(); composerErr != nil {
		return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	if wc.masterSpec != nil {
		var masterSpecAsSuperset workloadsApi.SupersetSpec
		if err := convutils.CopySpecs(&masterSpecAsSuperset, wc.masterSpec); err != nil {
			return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
		}
		masterSpecAsSuperset.DistributedFramework.Set(wc.spec.DistributedFramework.Get())

		masterReplicaType := workloadsApi.REPLICATYPE_MASTER
		if composerErr := notificationsComposer.Compose(ctx, &masterReplicaType, loadDistributedMasterComplimentaryAssetInfo(ctx, wc), &masterSpecAsSuperset)(); composerErr != nil {
			return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
		}
	}

	// NotificationsComposer uses the concrete Kubeflow kind (for example, PyTorchJob)
	// for its legacy output. WorkloadCreationFields expects the public framework enum
	// (for example, PyTorch), so normalize the fallback notification at this boundary.
	if framework := wc.spec.DistributedFramework.Get(); framework != nil {
		normalizedFramework := string(*framework)
		notificationsComposer.Output.DistributedFramework = &normalizedFramework
	}

	return nil
}

func loadDistributedMasterComplimentaryAssetInfo(ctx context.Context, wc *workloadContext) *composer.ComplimentaryAssetInfo {
	if wc.create == nil || wc.create.createParams == nil || !wc.create.createParams.Response.MasterSpec.IsSet() ||
		wc.create.createParams.Response.MasterSpec.Get() == nil {
		return nil
	}

	complimentaryAssetInfoMaster, err := composer.LoadComplimentaryAssetInfo(wc.create.createParams.Response.MasterSpec.Get().Annotations, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to load asset info of master spec: %v", err)
	}

	return complimentaryAssetInfoMaster
}
