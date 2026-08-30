package services_crud

import (
	"context"
	"fmt"
	"net/http"

	workloadsComposer "github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/composer_v2/inference_workloads"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

const (
	knativeGroup   = "serving.knative.dev"
	knativeVersion = "v1"
	knativeKind    = "Service"
)

type AnyWorkloadSubmitterInference struct{}

func (s *AnyWorkloadSubmitterInference) DetermineSubmissionMethods(ctx context.Context, wc *workloadContext) SubmissionMethods {
	return determineSubmissionMethods(ctx, wc.tenantId, wc.workloadType, wc.clusterCapabilities)
}

func (s *AnyWorkloadSubmitterInference) CanUpdate(ctx context.Context, origSubmissionMethod SubmissionMethods, wc *workloadContext, _ string) *workloadsApi.SubmissionErrorV2 {
	return canUpdateForSubmissionMethod(ctx, origSubmissionMethod, wc)
}

// Create returns (nil, nil) on NotImplemented so the submission falls through to the legacy InferenceWorkload CRD.
func (s *AnyWorkloadSubmitterInference) Create(ctx context.Context, wc *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2) {
	projectName := wc.create.project.Name
	namespace := wc.create.namespace

	options := wc.options
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, wc.clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}

	composer := inference_workloads.NewInferenceToKnativeComposer(
		wc.workloadName,
		projectName,
		namespace,
		wc.createdBy,
		wc.create.assetsMapping,
		options,
	)
	clusterStorageConfig, cfgErr := buildClusterStorageConfig(ctx, wc.create.createParams.Response.ClusterId)
	if cfgErr != nil {
		return nil, cfgErr
	}

	composer.WithStorageClassConfiguration(wc.storageClassConfigs)
	composer.WithClusterStorageConfig(clusterStorageConfig)
	composer.WithDataVolumeClaimNames(wc.dataVolumeClaimNames)

	target := inference_workloads.NewMinimalKnativeServiceTarget()

	if composerErr := composer.Compose(ctx, wc.inferenceSpec, target)(); composerErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	manifest, err := gojsonq.LoadJson(target.Service)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if metadataErr := composeInferenceConnectionMetadata(ctx, wc, projectName); metadataErr != nil {
		return nil, metadataErr
	}

	return &Submission{
		Request: workloadMgrApi.WorkloadV2CreateRequest{
			Metadata: buildInferenceWorkloadV2Metadata(
				&wc.create.createParams.Response.Name,
				&wc.create.createParams.Response.ProjectId,
				&target.Config,
				wc.inferenceSpec,
			),
			Manifest: manifest,
		},
		Requirements:             composer.WorkloadRequirements,
		ConnectionMetadata:       newWorkloadsServiceConnectionMetadata(wc.create.notificationsComposer.Output),
		ConnectionCreationFields: wc.create.notificationsComposer.Output,
	}, nil
}

func composeInferenceConnectionMetadata(ctx context.Context, wc *workloadContext, projectName string) *workloadsApi.SubmissionErrorV2 {
	if wc.create.notificationsComposer == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "workload notification composer was not initialized")
	}

	notificationsComposer := wc.create.notificationsComposer.WithConnectionsPreProcessor(projectName, workloadsComposer.NetworkingResourcePrefixInference)
	if composerErr := notificationsComposer.Compose(ctx, nil, wc.create.complimentaryAssetInfo, wc.spec)(); composerErr != nil {
		return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return nil
}

func buildInferenceWorkloadV2Metadata(
	name, projectID *string,
	configuration *workloadMgrApi.WorkloadV2Configuration,
	spec *workloadsApi.InferenceSpecSpec,
) workloadMgrApi.WorkloadV2MetadataCreateParams {
	metadata := workloadMgrApi.WorkloadV2MetadataCreateParams{
		Name:          name,
		ProjectId:     projectID,
		Configuration: configuration,
	}
	if spec == nil {
		return metadata
	}
	if category := spec.Category.Get(); category != nil {
		metadata.Category.Set(category)
	}
	if priority := spec.PriorityClass.Get(); priority != nil {
		metadata.Priority.Set(priority)
	}
	if preemptibility := spec.Preemptibility.Get(); preemptibility != nil {
		converted := workloadMgrApi.Preemptibility(*preemptibility)
		metadata.Preemptibility.Set(&converted)
	}
	return metadata
}

// Update recomposes the Knative manifest from the orchestrator's already-merged spec and wraps it in
// an update request (mirrors Create's compose step).
func (s *AnyWorkloadSubmitterInference) Update(ctx context.Context, wc *workloadContext, workloadId string,
	existingWorkload *workloadMgrApi.WorkloadV2GetResponse, mergedWorkload *models_crud.Workload,
) (*Update, *workloadsApi.SubmissionErrorV2) {
	// Project the merged SupersetSpec into the inference spec the composer consumes (as Create does).
	inferenceSpec := &workloadsApi.InferenceSpecSpec{}
	if err := convutils.CopySpecs(inferenceSpec, mergedWorkload.Spec); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to project inference spec: %v", err))
	}

	clusterId := mergedWorkload.ClusterId
	projectName := existingWorkload.Workload.Metadata.ProjectName

	options := wc.options
	clusterCapabilities := wc.clusterCapabilities
	if clusterCapabilities == nil {
		clusterCapabilities = checker.NewClusterCapabilitiesChecker(clusterId)
	}
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}

	composer := inference_workloads.NewInferenceToKnativeComposer(
		mergedWorkload.Name,
		projectName,
		mergedWorkload.Namespace,
		mergedWorkload.CreatedBy,
		mergedWorkload.AssetsMapping,
		options,
	)

	clusterStorageConfig, cfgErr := buildClusterStorageConfig(ctx, clusterId)
	if cfgErr != nil {
		return nil, cfgErr
	}
	composer.WithStorageClassConfiguration(wc.storageClassConfigs)
	composer.WithClusterStorageConfig(clusterStorageConfig)
	composer.WithDataVolumeClaimNames(wc.dataVolumeClaimNames)

	target := inference_workloads.NewMinimalKnativeServiceTarget()
	if composerErr := composer.Compose(ctx, inferenceSpec, target)(); composerErr != nil {
		// No native fallback on update — surface composer errors (e.g. 501 for unsupported features).
		return nil, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	manifest, err := gojsonq.LoadJson(target.Service)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}

	return &Update{
		Request: workloadMgrApi.WorkloadV2UpdateRequest{
			Metadata: &workloadMgrApi.WorkloadV2MetadataUpdateParams{
				Configuration: *workloadMgrApi.NewNullableWorkloadV2Configuration(&target.Config),
			},
			Manifest: manifest,
		},
		Requirements: composer.WorkloadRequirements,
	}, nil
}
