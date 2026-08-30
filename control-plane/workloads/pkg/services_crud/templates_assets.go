package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var (
	workloadCracker       = jsonutils.NewJsonCracker(api.SupersetSpec{})
	workloadComputeField  = workloadCracker.NameOf(&workloadCracker.Struct().Compute)
	workloadSecurityField = workloadCracker.NameOf(&workloadCracker.Struct().Security)

	envAssetCracker = jsonutils.NewJsonCracker(assetsApi.EnvironmentAssetSpec{})
)

// map the kind to the category of asset in the template
func kindToCategory(kind assetsApi.AssetKind) string {
	switch kind {
	case assetsApi.ASSETKIND_ENVIRONMENT:
		return models_crud.AssetCategoryEnvironment
	case assetsApi.ASSETKIND_COMPUTE:
		return models_crud.AssetCategoryCompute
	default:
		return models_crud.AssetCategoryDatasource
	}
}

// information that we keep per asset: its category, and if it appears in master/worker/both specs
type replicas struct {
	category string

	master bool
	worker bool
}

func (replicas replicas) withCategory(category string) replicas {
	replicas.category = category
	return replicas
}

var workerReplica = replicas{worker: true}
var masterReplica = replicas{master: true}
var workerAndMaster = replicas{worker: true, master: true}

// collect the assets that the template is linked with, into a map
func collectAssets(linkedAssets mapAssetIdToReplicas, assets *api.TemplateAssetsIdsOptional, replicas replicas) {
	if assets == nil {
		return
	}

	if assets.Environment != nil {
		linkedAssets.Upsert(*assets.Environment, replicas.withCategory(models_crud.AssetCategoryEnvironment))
	}

	if assets.Compute != nil {
		linkedAssets.Upsert(*assets.Compute, replicas.withCategory(models_crud.AssetCategoryCompute))
	}

	for _, assetId := range assets.Datasources {
		linkedAssets.Upsert(assetId, replicas.withCategory(models_crud.AssetCategoryDatasource))
	}
}

// mapAssetIdToReplicas holds the list of assets that template is linked with. for each
// asset it maintains an indication if this asset is linked to worker/master replica or both
type mapAssetIdToReplicas map[string]replicas

// Upsert an asset to the map. turn on the relevant bits
func (atr mapAssetIdToReplicas) Upsert(assetId string, replicas replicas) {
	if assetReplicas, ok := atr[assetId]; !ok {
		atr[assetId] = replicas
	} else {
		assetReplicas.master = assetReplicas.master || replicas.master
		assetReplicas.worker = assetReplicas.worker || replicas.worker
		atr[assetId] = assetReplicas
	}
}

func (atr mapAssetIdToReplicas) Ids() []string {
	result := make([]string, 0, len(atr))
	for assetId := range atr {
		result = append(result, assetId)
	}
	return result
}

func (atr mapAssetIdToReplicas) IdsOfReplica(filterReplica api.ReplicaType) []string {
	result := make([]string, 0, len(atr))
	for assetId, replicas := range atr {
		if filterReplica == api.REPLICATYPE_WORKER && replicas.worker {
			result = append(result, assetId)
		}
		if filterReplica == api.REPLICATYPE_MASTER && replicas.master {
			result = append(result, assetId)
		}
	}
	return result
}

func (atr mapAssetIdToReplicas) workerAndMasterHaveSameAssets() bool {
	for _, replicas := range atr {
		if replicas.worker != replicas.master {
			return false
		}
	}
	return true
}

// requestAssetInfo from assets service. we get back both the metadata and the specs in flat format
func requestAssetInfo(ctx context.Context, workloadType api.WorkloadType, linkedAssets mapAssetIdToReplicas) (*assetsApi.AssetFlatResponse, map[string]*assetsApi.AssetMeta, *api.Error) {
	asMap := make(map[string]*assetsApi.AssetMeta)

	if len(linkedAssets) == 0 {
		return nil, asMap, nil
	}

	var result assetsApi.AssetFlatResponse

	//
	//  it is possible that worker and master have different set of assets. in this case we have to request
	//  flat format for each set separately
	//
	replicaTypes := []api.ReplicaType{api.REPLICATYPE_WORKER}
	if workloadType == api.WORKLOADTYPE_DISTRIBUTED {
		replicaTypes = append(replicaTypes, api.REPLICATYPE_MASTER)
	}

	for _, replicaType := range replicaTypes {
		assetInfoRequest := assetsApi.AssetFlatRequest{
			AssetIds:     linkedAssets.IdsOfReplica(replicaType),
			WorkloadType: assetsApi.WorkloadType(workloadType),
		}

		if len(assetInfoRequest.AssetIds) == 0 {
			continue
		}

		response, err := assets.Get().RequestAssetFlatFormat(ctx, assetInfoRequest)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to obtain asset info %s: %v", err.Message, err.Details)
			return nil, nil, api.NewError(int(err.Code), err.Message, nil)
		}

		for _, assetMeta := range response.AssetsMeta {
			if asMap[assetMeta.Id] == nil {
				response.AssetsMeta = append(response.AssetsMeta, assetMeta)
				asMap[assetMeta.Id] = &assetMeta
			}
		}

		if replicaType == api.REPLICATYPE_WORKER {
			result.Worker = response.Worker

			//  if it happens that same assets are given for worker and master -> no need to call assets service
			//  again, because it returns both worker and master spec anyway.
			if linkedAssets.workerAndMasterHaveSameAssets() {
				result.Master = response.Master
				break
			}
		} else {
			result.Master = response.Master
		}
	}

	for _, assetMeta := range asMap {
		result.AssetsMeta = append(result.AssetsMeta, *assetMeta)
	}

	return &result, asMap, nil
}

// validate that spec of template does not contain any fields of compute/environment asset,
// in case template links to compute/environment asset. datasource have no restrictions because
// it is itemized so no problem to include assets and spec items.
func validateSpecAndAssetsDoNoConflict(spec interface{}, assets *api.TemplateAssetsIdsOptional) *api.Error {
	if assets == nil {
		return nil
	}

	specAsJson, jsonErr := gojsonq.LoadJson(spec)
	if jsonErr != nil {
		return api.NewValidationError(fmt.Sprintf("failed to load spec: %v", jsonErr))
	}

	if assets.Compute != nil {
		if specAsJson[workloadComputeField] != nil {
			return api.NewValidationError("compute fields are not allowed in the spec of a template that is linked to a compute asset")
		}
	}

	if assets.Environment == nil {
		return nil
	}

	securitySection := convutils.MapMe(specAsJson[workloadSecurityField])

	// if environment asset is linked, make sure no asset field in either the workload root or its security section
	for _, fldName := range envAssetCracker.AllNames() {
		if specAsJson[fldName] != nil {
			return api.NewValidationError(fmt.Sprintf("field %s is not allowed in the spec of a template that is linked to an environment asset", fldName))
		}

		if securitySection[fldName] != nil {
			return api.NewValidationError(fmt.Sprintf("field security.%s is not allowed in the spec of a template that is linked to an environment asset", fldName))
		}
	}

	return nil
}

// mutateTemplate takes template + assets and create a combined specs. it is assumed that asset has been validated
// and the spec does not have field conflicts b/w template and assets
func mutateTemplate(template *api.SupersetTemplate, workloadType api.WorkloadType, assetsAsFlat *assetsApi.AssetFlatResponse) *api.Error {
	var jsonErr error

	//
	//   load the template to a json map with two sections worker and optional master
	//
	templateSpecs := make(map[string]interface{})

	if workloadType == Distributed {
		templateSpecs[workerSection], jsonErr = gojsonq.LoadJson(template.Spec)
		if jsonErr != nil {
			return api.NewError(http.StatusInternalServerError, "failed to load template worker spec", jsonErr)
		}

		if template.Master != nil {
			templateSpecs[masterSection], jsonErr = gojsonq.LoadJson(template.Master.Spec)
			if jsonErr != nil {
				return api.NewError(http.StatusInternalServerError, "failed to load template master spec", jsonErr)
			}
		}
	} else {
		templateSpecs, jsonErr = gojsonq.LoadJson(template.Spec)
		if jsonErr != nil {
			return api.NewError(http.StatusInternalServerError, "failed to load template spec", jsonErr)
		}
	}

	//
	//   load the asset specs to a json map with two sections worker and optional master
	//
	assetSpecs := make(map[string]interface{})

	if workloadType == Distributed {
		if assetsAsFlat.Worker != nil {
			assetSpecs[workerSection], jsonErr = gojsonq.LoadJson(assetsAsFlat.Worker)
			if jsonErr != nil {
				return api.NewError(http.StatusInternalServerError, "failed to load assets worker spec", jsonErr)
			}
		}

		if convutils.HasValue(assetsAsFlat.Master) {
			assetSpecs[masterSection], jsonErr = gojsonq.LoadJson(assetsAsFlat.Master)
			if jsonErr != nil {
				return api.NewError(http.StatusInternalServerError, "failed to load assets master spec", jsonErr)
			}
		}
	} else {
		assetSpecs, jsonErr = gojsonq.LoadJson(assetsAsFlat.Worker)
		if jsonErr != nil {
			return api.NewError(http.StatusInternalServerError, "failed to load assets spec", jsonErr)
		}
	}

	masterSameAsWorker := convutils.BoolIsTrue(template.MasterSpecSameAsWorker)

	mutatedSpec, mutateErr := MutateWorkload(templateSpecs, nil, masterSameAsWorker, assetSpecs, nil)
	if mutateErr != nil {
		return api.NewError(int(mutateErr.Code), mutateErr.Message, nil)
	}

	combinedSpec, combinedMasterSpec, buildErr := buildResponseSpecFromMap[api.SupersetSpec, api.MasterSpec](string(workloadType), mutatedSpec)
	if buildErr != nil {
		return api.NewError(int(buildErr.Code), buildErr.Message, nil)
	}

	if combinedSpec != nil {
		template.CombinedSpec = *combinedSpec
	} else {
		template.CombinedSpec = api.SupersetSpec{}
	}

	if combinedMasterSpec != nil {
		template.CombinedMasterSpec = &api.MasterTemplateSpec{}

		copyErr := convutils.CopySpecs(template.CombinedMasterSpec, combinedMasterSpec)
		if copyErr != nil {
			return api.NewError(http.StatusInternalServerError, "failed to copy specs", copyErr)
		}
	}

	return nil
}

func assetsGiven(template *api.SupersetTemplate) bool {
	return assetsSpecified(template.Assets) || (template.Master != nil && assetsSpecified(template.Master.Assets))
}

func assetsSpecified(assets *api.TemplateAssetsIdsOptional) bool {
	if assets == nil {
		return false
	}

	return assets.Environment != nil || assets.Compute != nil || (len(assets.Datasources) > 0)
}

func linkedAssetsToTemplateAssetsRec(tenantId uint, templateId string, linkedAssets mapAssetIdToReplicas) []models_crud.TemplateAssets {
	result := make([]models_crud.TemplateAssets, 0, len(linkedAssets)*2)

	for assetId, replicas := range linkedAssets {
		if replicas.worker {
			result = append(result, models_crud.TemplateAssets{
				ID:          uuid.New(),
				AssetId:     assetId,
				TemplateId:  templateId,
				TenantId:    tenantId,
				ReplicaType: string(api.REPLICATYPE_WORKER),
				Category:    replicas.category,
			})
		}

		if replicas.master {
			result = append(result, models_crud.TemplateAssets{
				ID:          uuid.New(),
				AssetId:     assetId,
				TemplateId:  templateId,
				TenantId:    tenantId,
				ReplicaType: string(api.REPLICATYPE_MASTER),
				Category:    replicas.category,
			})
		}
	}

	return result
}

func AssetModified(ctx context.Context, event *events.AssetModified) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("AssetModified event for asset %s tenant %s", event.AssetId, event.AssetId)

	tenantId, convErr := strconv.ParseUint(event.TenantId, 10, 32)
	if convErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("AssetModified event for asset %s contains invalid tenant id %s", event.AssetId, event.TenantId)
		return nil
	}

	relevantTemplates, dbErr := dao_crud.ListTemplatesByAssetId(ctx, uint(tenantId), event.AssetId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("AssetModified failed to find templates relevant for asset %s: %s: %v", event.AssetId, dbErr.Message, dbErr.Details)
		return nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Update of asset %s affects %d templates", event.AssetId, len(relevantTemplates))

	for _, template := range relevantTemplates {
		emptyRequest := UnifiedRequest{
			tenantId:     template.TenantId,
			workloadType: api.WorkloadType(template.WorkloadType),
			template:     &api.SupersetTemplate{},
		}

		copier.Copy(&emptyRequest.template.Meta, template.TemplateMeta)

		_, err := updateTemplateInt(ctx, &template, &emptyRequest, updateOptions{patchMode: true})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update template %s with update of asset %s: %s: %v",
				template.ID.String(), event.AssetId, err.Message, err.Details)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully updated template %s with update of asset %s", template.ID.String(), event.AssetId)
		}
	}

	return nil
}

func AssetDeleted(ctx context.Context, event *events.AssetDeleted) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("AssetDeleted event for asset %s tenant %s", event.AssetId, event.TenantId)

	tenantId, convErr := strconv.ParseUint(event.TenantId, 10, 32)
	if convErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("AssetDeleted event for asset %s contains invalid tenant id %s", event.AssetId, event.TenantId)
		return nil
	}

	relevantTemplates, dbErr := dao_crud.ListTemplatesByAssetId(ctx, uint(tenantId), event.AssetId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("AssetDeleted failed cont find templates relevant for asset %s: %s: %v", event.AssetId, dbErr.Message, dbErr.Details)
		return nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Deletion of asset %s affects %d templates", event.AssetId, len(relevantTemplates))

	for _, template := range relevantTemplates {
		//
		//   remove the asset from the list of assets that the template is linked with
		//
		updatedAssets := make([]models_crud.TemplateAssets, 0, len(template.Assets))

		for _, asset := range template.Assets {
			if asset.AssetId != event.AssetId {
				updatedAssets = append(updatedAssets, asset)
			}
		}

		//
		//  build request with the updated set of assets
		//
		_, workerAssets := models_crud.ContainedAssetsOfReplica(updatedAssets, api.REPLICATYPE_WORKER)

		updatedAssetsRequests := UnifiedRequest{
			tenantId:     template.TenantId,
			workloadType: api.WorkloadType(template.WorkloadType),
			template: &api.SupersetTemplate{
				Assets: workerAssets,
			},
		}

		_, masterAssets := models_crud.ContainedAssetsOfReplica(updatedAssets, api.REPLICATYPE_MASTER)

		if masterAssets != nil {
			updatedAssetsRequests.template.Master = &api.TemplateMasterSection{
				Assets: masterAssets,
			}
		}

		copier.Copy(&updatedAssetsRequests.template.Meta, template.TemplateMeta)

		_, err := updateTemplateInt(ctx, &template, &updatedAssetsRequests, updateOptions{patchMode: true})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update template %s with update of asset %s: %s: %v",
				template.ID.String(), event.AssetId, err.Message, err.Details)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully updated template %s with update of asset %s", template.ID.String(), event.AssetId)
		}
	}

	return nil
}
