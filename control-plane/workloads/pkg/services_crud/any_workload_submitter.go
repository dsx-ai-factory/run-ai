package services_crud

import (
	"context"
	"fmt"
	"net/http"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/notifications"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	workloadsMgr "github.com/run-ai/runai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/settings"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Submission is the output of an AnyWorkloadSubmitter for create operations.
// When non-nil, createWorkload submits via workloads-manager.
type Submission struct {
	Request workloadMgrApi.WorkloadV2CreateRequest

	Requirements composer.WorkloadRequirements

	// Connection metadata is known only to workloads-service's native request
	// translator. Keep the creation fields beside it so workloads-service can
	// create a fallback row if workloads-manager's async notification is delayed.
	ConnectionMetadata       *WorkloadsServiceConnectionMetadata
	ConnectionCreationFields *workloadsApi.WorkloadCreationFields
}

// WorkloadsServiceConnectionMetadata contains only the workloads-service fields that are
// known by the native request translator and are not part of workloads-manager's
// generic creation notification.
type WorkloadsServiceConnectionMetadata struct {
	Environments        []workloadsApi.Environment `json:"environments,omitempty"`
	ExternalConnections []workloadsApi.Connection  `json:"externalConnections,omitempty"`
}

func newWorkloadsServiceConnectionMetadata(creation *workloadsApi.WorkloadCreationFields) *WorkloadsServiceConnectionMetadata {
	if creation == nil || (len(creation.Environments) == 0 && len(creation.ExternalConnections) == 0) {
		return nil
	}

	return &WorkloadsServiceConnectionMetadata{
		Environments:        creation.Environments,
		ExternalConnections: creation.ExternalConnections,
	}
}

// Update is the output of an AnyWorkloadSubmitter for update operations.
// When non-nil, updateWorkload submits via workloads-manager.
type Update struct {
	Request      workloadMgrApi.WorkloadV2UpdateRequest
	Requirements composer.WorkloadRequirements
}

// workloadCreateContext contains fields specific to workload creation.
type workloadCreateContext struct {
	createParams           *createWorkloadParams
	project                *orgUnitApi.Project
	namespace              string
	assetsMapping          []workloadsApi.AssetsMapping
	complimentaryAssetInfo *composer.ComplimentaryAssetInfo
	notificationsComposer  *notifications.NotificationsComposer
	actualNodePools        []clusterApi.Nodepool
}

// workloadContext bundles the preprocessed inputs that strategies need.
// It contains common fields and a pointer to create-specific context.
type workloadContext struct {
	// Common fields for both create and update
	workloadType             string
	workloadName             string
	createdBy                string
	tenantId                 int32
	spec                     *workloadsApi.SupersetSpec
	inferenceSpec            *workloadsApi.InferenceSpecSpec
	distributedInferenceSpec *workloadsApi.DistributedInferenceSpecSpec
	options                  composer.FieldsComposerOptions
	clusterCapabilities      *checker.ClusterCapabilitiesChecker
	segmentRouting           *segmentedDistributedRoutingDecision
	storageClassConfigs      map[string]k8sObjectsApi.StorageClassConfiguration
	dataVolumeClaimNames     map[string]string // maps data volume ID to PVC claim name

	// For distributed we also need the master spec
	masterSpec         *workloadsApi.MasterSpec
	masterSameAsWorker bool

	// Create-specific context (nil for update)
	create *workloadCreateContext
}

type SubmissionMethods struct {
	Native      bool
	AnyWorkload bool
}

var SubmissionMethodNative = SubmissionMethods{Native: true}
var SubmissionMethodAnyWorkload = SubmissionMethods{AnyWorkload: true}

// AnyWorkloadSubmitter decides between workload-runner submission and native submission.
// Each implementation handles a specific workload type (e.g., interactive/training batch jobs).
type AnyWorkloadSubmitter interface {
	// DetermineSubmissionMethods determines b/w native or any-workload submission.
	DetermineSubmissionMethods(ctx context.Context, wc *workloadContext) SubmissionMethods

	// CanUpdate checks if this submitter can handle workload updates for a workload
	// that has been submitted in a given submission mode.
	CanUpdate(ctx context.Context, origSubmissionMethod SubmissionMethods, wc *workloadContext, workloadId string) *workloadsApi.SubmissionErrorV2

	// Create handles workload creation and returns the submission request.
	// Should only be called if CanCreate returns true.
	// Returns:
	//   (*Submission, nil)              -> success, submit via workloads-manager
	//   (nil, nil)                      -> spec not yet supported, fall back to native
	//   (nil, *SubmissionErrorV2)       -> hard error, abort the request
	Create(ctx context.Context, wc *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2)

	// Update recomposes the orchestrator's already-merged spec (mergedWorkload) into the workload's
	// manifest. Only call if CanUpdate returned true; unsupported submitters return nil,nil.
	// Returns:
	//   (*Update, nil)                  -> success, submit via workloads-manager
	//   (nil, nil)                      -> not supported, fall back to native
	//   (nil, *SubmissionErrorV2)       -> hard error, abort the request
	Update(ctx context.Context, wc *workloadContext, workloadId string, existingWorkload *workloadMgrApi.WorkloadV2GetResponse, mergedWorkload *models_crud.Workload) (*Update, *workloadsApi.SubmissionErrorV2)
}

// anyWorkloadSubmitters maps workload type to its submitter.
// Each workload type has at most one submitter.
var anyWorkloadSubmitters = map[string]AnyWorkloadSubmitter{
	Workspace:            &AnyWorkloadSubmitterInterTrain{},
	Training:             &AnyWorkloadSubmitterInterTrain{},
	Distributed:          &AnyWorkloadSubmitterDistributed{},
	Inference:            &AnyWorkloadSubmitterInference{},
	DistributedInference: &AnyWorkloadSubmitterDinf{},
}

// getAnyWorkloadSubmitter returns the submitter for the given workload context,
// or nil if no submitter is registered for this workload type.
func getAnyWorkloadSubmitter(wc *workloadContext) AnyWorkloadSubmitter {
	submitter, exists := anyWorkloadSubmitters[wc.workloadType]
	if !exists {
		return nil
	}
	return submitter
}

// persistUpdateAndRespondNative handles the common persist + get-phase + return-response pattern
// used by update functions.
func persistUpdateAndRespondNative(
	ctx context.Context,
	tenantSettingsClient clients.TenantSettingsClient,
	nativeWorkload *models_crud.Workload,
	workloadId string,
	requirements *composer.WorkloadRequirements,
) (*workloadsApi.SupersetWorkload, *workloadsApi.SubmissionErrorV2) {
	// Validate requirements
	if requirementsErr := validateRequirements(ctx, tenantSettingsClient, requirements); requirementsErr != nil {
		return nil, NewSubmissionErrorWithDetails(requirementsErr.Code, requirementsErr.Message, requirementsErr.Details)
	}

	// Update the workload in the db
	if dbErr := dao_crud.DbUpsertWorkload(ctx, nativeWorkload); dbErr != nil {
		return nil, ToSubmissionError(dbErr)
	}

	// Finalize the response by getting the actual phase of the workload
	actualPhase, phaseErr := services_integration.GetPhaseOfWorkload(ctx, &nativeWorkload.ClusterId, workloadId)
	if phaseErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get phase of workload id %s, err: %v", workloadId, phaseErr.Message)
		actualPhase = "Unknown"
	}

	response := WorkloadFromDbRecord(nativeWorkload, workloadsApi.Phase(actualPhase))

	return &response, nil
}

// determineSubmissionMethods determines b/w native or any-workload submission
// The logic is this
//  1. if we do not know which submission the cluster supports - return empty list of methods
//  2. if the cluster supports only native (release < 2.6) - returns native
//  3. if the cluster supports only any-workload (release >= 3.0) - returns any-workload
//  4. for cluster of version 2.6, there is an option for the system or tenant to choose. for
//     any-workload submission we need:
//     a. Feature must be enabled at deployment level: either the global AnyWorkloadSubmissionEnabled
//     config flag OR the per-workload-type config flag must be true. Unknown types are rejected.
//     b. If enabled at deployment level, check per-tenant setting - tenants can disable the feature
//     even if it's enabled at deployment level.
func determineSubmissionMethods(ctx context.Context, tenantId int32, workloadType string, clusterCapabilities *checker.ClusterCapabilitiesChecker) SubmissionMethods {
	result := SubmissionMethods{}

	// Check if the cluster supports any-workload submission for V1 workloads.
	// Fail-closed: if clusterCapabilities is nil, we cannot verify cluster support,
	// so we return false to prevent unsafe submission.
	if clusterCapabilities == nil {
		return result
	}

	anyWorkloadSupported, err := clusterCapabilities.ValidateClusterCapability(ctx, checker.MinVersionOfAnyWorkloadForV1Workloads)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check cluster capability for any-workload submission: %v", err)
	} else {
		result.AnyWorkload = anyWorkloadSupported
	}

	nativeNoLongerSupported, err := clusterCapabilities.ValidateClusterCapability(ctx, checker.MinVersionNoNativeSubmission)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check cluster capability for no-native submission: %v", err)
	} else {
		result.Native = !nativeNoLongerSupported
	}

	// for clusters of version 2.6, both are supported, with preference to AnyWorkload, but there is an
	// option for the system or tenant to disable any-workload submission
	if result.AnyWorkload && result.Native {
		cfg := runai_config.Get[*config.ServicesConfig]()
		if !isDeploymentEnabledForWorkloadType(cfg, workloadType) || !isAnyWorkloadTenantSettingEnabled(tenantId) {
			result.AnyWorkload = false
		}
	}

	return result
}

// canUpdateForSubmissionMethod validates that a workload originally submitted with
// origSubmissionMethod can be updated in the same mode on the current cluster.
// It first rejects an ambiguous/unset submission method as an internal error, then
// verifies that the cluster still supports the mode the workload was submitted with.
func canUpdateForSubmissionMethod(ctx context.Context, origSubmissionMethod SubmissionMethods, wc *workloadContext) *workloadsApi.SubmissionErrorV2 {
	if origSubmissionMethod.AnyWorkload && origSubmissionMethod.Native {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "conflicting submission methods: both any-workload and native submission derived for the workload")
	}

	if !origSubmissionMethod.AnyWorkload && !origSubmissionMethod.Native {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "could not determine the submission method of the original workload")
	}

	// this validation should not really happen because it has been validated in the update method,
	// but we're doing this for extra safety
	if wc.clusterCapabilities == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "cannot determine submission method for the given cluster")
	}

	clusterSupportedMethod := determineSubmissionMethods(ctx, wc.tenantId, wc.workloadType, wc.clusterCapabilities)

	if origSubmissionMethod.Native && !clusterSupportedMethod.Native {
		return workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, "the workload has been submitted with native submission, but the cluster does not support native submission")
	}

	if origSubmissionMethod.AnyWorkload && !clusterSupportedMethod.AnyWorkload {
		return workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, "the workload has been submitted with any-workload submission, but the cluster does not support any-workload submission")
	}

	return nil
}

// isDeploymentEnabledForWorkloadType checks if the workload type is enabled at deployment level.
// Returns true if either the global flag or the per-type flag is enabled for known workload types.
// Returns false for unknown workload types.
func isDeploymentEnabledForWorkloadType(cfg *config.ServicesConfig, workloadType string) bool {
	switch workloadType {
	case Workspace:
		return cfg.AnyWorkloadSubmissionEnabled || cfg.AnyWorkloadWorkspaceSubmission
	case Training:
		return cfg.AnyWorkloadSubmissionEnabled || cfg.AnyWorkloadTrainingSubmission
	case Distributed:
		return cfg.AnyWorkloadSubmissionEnabled || cfg.AnyWorkloadDistributedSubmission
	case Inference:
		return false // eventually, inference is not supported on 2.6.0 clusters, only from 3.0.0
	case DistributedInference:
		return false // eventually, inference is not supported on 2.6.0 clusters, only from 3.0.0
	default:
		return false // Unknown workload type
	}
}

// isAnyWorkloadTenantSettingEnabled checks if the per-tenant AnyWorkloadSubmissionEnabled
// setting is enabled for the given tenant.
func isAnyWorkloadTenantSettingEnabled(tenantId int32) bool {
	tenantSetting, err := setting_pkg.Get[*settings.AnyWorkloadSubmissionEnabled](fmt.Sprint(tenantId))
	if err != nil {
		logging.Errorf("failed to get AnyWorkloadSubmissionEnabled setting for tenant %d: %v", tenantId, err)
		return false
	}
	return tenantSetting.IsEnabled()
}

// getWorkloadFromWorkloadManager attempts to retrieve a workload from the workload-manager service.
// Returns the workload if found, nil if not found (404), or an error for other failures.
func getWorkloadFromWorkloadManager(ctx context.Context, workloadId string) (*workloadMgrApi.WorkloadV2GetResponse, *workloadMgrApi.Error) {
	response, err := workloadsMgr.Get().GetWorkloadV2ById(ctx, workloadId)
	if err != nil {
		if err.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return response, nil
}

// validateWorkloadGVK checks if the workload manifest has the expected Group, Version, and Kind.
func validateWorkloadGVK(workload *workloadMgrApi.WorkloadV2GetResponse, expectedGroup, expectedVersion, expectedKind string) bool {
	manifest := workload.Workload.WorkloadManifest
	if manifest == nil {
		return false
	}

	// Extract apiVersion (format: "group/version" or just "version" for core API)
	apiVersion, ok := manifest["apiVersion"].(string)
	if !ok {
		return false
	}

	// Extract kind
	kind, ok := manifest["kind"].(string)
	if !ok {
		return false
	}

	// Parse group and version from apiVersion using k8s schema package
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return false
	}

	return gv.Group == expectedGroup && gv.Version == expectedVersion && kind == expectedKind
}
