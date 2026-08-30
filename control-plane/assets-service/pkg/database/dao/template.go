package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

// TemplateToAssetsMap is a mapping b/w template ID and its asset references
type TemplateToAssetsMap map[string]api.AssetsRef

// EnvironmentToolsMap is a mapping b/w environment ID to the tools defined by its connections
type EnvironmentToolsMap map[string][]api.ToolType

func JoinAssetsTemplateContentAndMetadata() string {
	assetsMetaJoin := fmt.Sprintf("left join %s on %s.contained_asset_id = %s.id", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsTemplateContent, dbtables.TableNameAssetsMeta)
	workloadSupportedTypesJoin := fmt.Sprintf("left join %s on %s.contained_asset_id = %s.asset_id", dbtables.TableNameWorkloadSupportedTypes,
		dbtables.TableNameAssetsTemplateContent, dbtables.TableNameWorkloadSupportedTypes)
	return fmt.Sprintf("%s %s", assetsMetaJoin, workloadSupportedTypesJoin)
}

func joinMetadataAndTemplatesToGetAssets() string {
	return fmt.Sprintf("join %s on %s.id = %s.contained_asset_id", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsTemplateContent)
}

func JoinMetadataAndTemplatesToGetTemplates() string {
	return fmt.Sprintf("join %s on %s.id = %s.template_asset_id", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsTemplateContent)
}

func CreateTemplateAsset(ctx context.Context, request api.WorkloadTemplateCreationRequest, containedAssetIds []string, nameValidationQuery *database.ParametrizedQuery) (*api.WorkloadTemplate, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, createErr := createAssetMeta(ctx, tx, id, api.ASSETKIND_WORKLOAD_TEMPLATE, request.Meta, nameValidationQuery)
	if createErr != nil {
		if strings.Contains(createErr.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_WORKLOAD_TEMPLATE, http.StatusConflict).WithDetails(createErr)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(createErr)
	}

	templateAssetRec := dbmodel.TemplateAssetRec{
		AssetId:     id,
		SpecificEnv: request.Spec.SpecificEnv,
	}

	_, crErr := tx.Create(&templateAssetRec)
	if crErr != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(crErr)
	}

	if err := writeTemplateContent(tx, id, containedAssetIds); err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(err)
	}

	template, _, getError := GetTemplateAssetByIdDuringTransaction(ctx, tx, assetMeta.Id)
	if getError != nil {
		return nil, getError
	}

	templateWorkloadSupportedTypes, gErr := GetTemplateWorkloadSupportedTypes(tx, template.Meta.Id, nil)
	if gErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to get template environment asset", gErr)
	}
	template.Meta.WorkloadSupportedTypes = templateWorkloadSupportedTypes

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(err)
	}

	committed = true

	return template, nil
}

func GetTemplateAssetById(ctx context.Context, id string) (*api.WorkloadTemplate, []string, *api.Error) {
	var combinedRec dbmodel.TemplateCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_WORKLOAD_TEMPLATE,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsTemplate)
	if err != nil {
		return nil, nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, nil, database.Error(database.GET, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.WorkloadTemplate{
		Meta: combinedRec.AssetMeta,
		Spec: api.SpecificRunInfoFields{
			SpecificEnv: combinedRec.SpecificEnv,
		},
	}

	containedIds, rows, assetError, scanErr := getTemplateContent(runaidb.DB().WithContext(ctx), id, &response.Spec.Assets)
	if assetError != nil {
		return nil, nil, assetError
	}
	if scanErr != nil || rows == 0 {
		return nil, nil, database.Error(database.GET, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(rows, scanErr)
	}

	return &response, containedIds, nil
}

func GetTemplateAssetByIdDuringTransaction(ctx context.Context, tx orm.DB, id string) (*api.WorkloadTemplate, []string, *api.Error) {
	var combinedRec dbmodel.TemplateCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_WORKLOAD_TEMPLATE,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, tx, query, nil, dbtables.TableNameAssetsTemplate)
	if err != nil {
		return nil, nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, nil, database.Error(database.GET, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.WorkloadTemplate{
		Meta: combinedRec.AssetMeta,
		Spec: api.SpecificRunInfoFields{
			SpecificEnv: combinedRec.SpecificEnv,
		},
	}

	containedIds, rows, assetError, scanErr := getTemplateContent(tx, id, &response.Spec.Assets)
	if assetError != nil {
		return nil, nil, assetError
	}
	if scanErr != nil || rows == 0 {
		return nil, nil, database.Error(database.GET, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(rows, scanErr)
	}

	return &response, containedIds, nil
}

func UpdateTemplateAsset(ctx context.Context, id string, request api.WorkloadTemplateUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.WorkloadTemplate, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	containedAssetIds, assetsError := ValidateContainedAssets(ctx, nil, assetMeta.Scope, assetMeta.ClusterId, assetMeta.ProjectId, nil, &request.Spec.Assets)
	if assetsError != nil {
		return nil, assetsError
	}

	recId := dbmodel.TemplateAssetRec{AssetId: id}
	templateAssetUpdate := dbmodel.TemplateAssetRec{}

	if err := tx.Model(recId).Where(recId).First(&templateAssetUpdate); err != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(err)
	}

	templateAssetUpdate.SpecificEnv = request.Spec.SpecificEnv

	saveRows, saveErr := tx.Save(&templateAssetUpdate)
	if saveErr != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if contentErr := updateTemplateContent(tx, id, containedAssetIds); contentErr != nil {
		return nil, contentErr
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(err)
	}
	committed = true

	template, _, getError := GetTemplateAssetById(ctx, assetMeta.Id)
	if getError != nil {
		return nil, getError
	}

	templateWorkloadSupportedTypes, gErr := GetTemplateWorkloadSupportedTypes(runaidb.DB().WithContext(ctx), template.Meta.Id, nil)
	if gErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to get template environment asset", gErr)
	}
	template.Meta.WorkloadSupportedTypes = templateWorkloadSupportedTypes

	return template, nil
}

func GetTemplateAssetIds(tx orm.DB, templateId string) (int64, []string, error) {
	containedIds := make([]string, 0)
	record := dbmodel.TemplateContentRec{}
	rowsAffeced, scanErr := tx.Model(record).
		Select("contained_asset_id").
		Where(dbmodel.TemplateContentRec{TemplateAssetId: templateId}).
		Order(&record.ContainedAssetId).
		Scan(&containedIds)
	return rowsAffeced, containedIds, scanErr
}

type TemplateEnvironmentAsset struct {
	api.AssetMeta                  `gorm:"embedded"`
	dbmodel.WorkloadSupportedTypes `gorm:"embedded"`
	dbmodel.TemplateContentRec     `gorm:"embedded"`
}

func GetTemplateWorkloadSupportedTypes(tx orm.DB, templateId string, queryParams *query_types.QueryParams) (*api.WorkloadSupportedTypes, error) {
	templateEnvironmentAssetIds := []string{templateId}
	_, TemplateEnvironmentAsset, envErr := GetEnvironmentAssetOfTemplates(tx, templateEnvironmentAssetIds, queryParams)
	if envErr != nil {
		return nil, envErr
	}

	templatesEnvironmentWST := TemplateEnvironmentAsset[0].ToApiWorkloadSupportedTypes()
	return templatesEnvironmentWST, nil
}

func GetEnvironmentAssetOfTemplates(tx orm.DB, templateIds []string, queryParams *query_types.QueryParams) (int64, []TemplateEnvironmentAsset, error) {
	combinedList := make([]TemplateEnvironmentAsset, 0)
	record := dbmodel.TemplateContentRec{}
	where := fmt.Sprintf("%s.template_asset_id IN (?) and assets_metadata.kind = ?", dbtables.TableNameAssetsTemplateContent)
	result := tx.Model(record).
		Select("assets_template_content.*, assets_metadata.*, assets_supported_workload_types.*").
		Where(where, templateIds, "environment").
		Joins(JoinAssetsTemplateContentAndMetadata())

	if queryParams != nil {
		HandleDistributedQueryParams(result, queryParams)
	}

	rows, scanErr := result.
		Order(&record.ContainedAssetId).
		Scan(&combinedList)

	return rows, combinedList, scanErr
}

func ListTemplates(ctx context.Context, queryParams *query_types.QueryParams) (*api.WorkloadTemplateListResponse, *api.Error) {
	combinedList := make([]dbmodel.TemplateCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_WORKLOAD_TEMPLATE,
		},
	}

	// we will not filter by distributed query params here (GetAssetMeta)
	// it relevant only for assets that are not environment
	distributedQueryParams := queryParams.DistributedQueryParams
	queryParams.DistributedQueryParams = nil

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsTemplate)
	if err != nil {
		return nil, err
	}
	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list workspace templates", scanErr)
	}

	// add template supported types by template environment supported types
	templateAssetIds := make([]string, 0)
	for _, template := range combinedList {
		templateAssetIds = append(templateAssetIds, template.Id)
	}

	_, templatesEnvironmentAsset, envErr := GetEnvironmentAssetOfTemplates(runaidb.DB().WithContext(ctx), templateAssetIds, &query_types.QueryParams{
		DistributedQueryParams: distributedQueryParams,
	})
	if envErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to template environment asset", envErr)
	}

	assetIdWithWST := make(map[string]dbmodel.WorkloadSupportedTypes)
	for _, templateEnvironmentAsset := range templatesEnvironmentAsset {
		assetIdWithWST[templateEnvironmentAsset.TemplateAssetId] = templateEnvironmentAsset.WorkloadSupportedTypes
	}

	envIdToToolsMap, err := getEnvironmentToToolsMap(ctx)
	if err != nil {
		return nil, err
	}

	templateToAssetsMap, err := getTemplateToAssetsMap(ctx, envIdToToolsMap)
	if err != nil {
		return nil, err
	}

	response := api.WorkloadTemplateListResponse{
		Entries: make([]api.WorkloadTemplate, 0),
	}

	// combinedList (list of all template metadata that we will add to each the WST)
	for _, template := range combinedList {
		if assetIdWithWST[template.Id].AssetId != "" {
			// here per each template we have one environment supported types
			assetIdWithWSTByTemp := assetIdWithWST[template.Id]
			template.WorkloadSupportedTypes = assetIdWithWSTByTemp.ToApiWorkloadSupportedTypes()

			entry := api.WorkloadTemplate{
				Meta: template.AssetMeta,
				Spec: api.SpecificRunInfoFields{
					SpecificEnv: template.SpecificEnv,
				},
			}
			if assetsRef, ok := templateToAssetsMap[template.Id]; ok {
				entry.Spec.Assets = assetsRef
			}
			response.Entries = append(response.Entries, entry)
		}
	}

	return &response, nil
}

func getTemplateToAssetsMap(ctx context.Context, envIdToTools EnvironmentToolsMap) (TemplateToAssetsMap, *api.Error) {
	containment := make([]dbmodel.TemplateAssetsCombinedRec, 0)

	record := dbmodel.TemplateContentRec{}

	_, scanErr := runaidb.DB().WithContext(ctx).Model(record).
		Select("*").
		Joins(joinMetadataAndTemplatesToGetAssets()).
		Order(&record.ContainedAssetId).
		Scan(&containment)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to load template containment information", scanErr)
	}

	return templateAssetsRecToMap(containment, envIdToTools)
}

func getEnvironmentToToolsMap(ctx context.Context) (EnvironmentToolsMap, *api.Error) {
	envAssets, err := ListEnvironmentAssets(ctx, nil)
	if err != nil {
		return nil, err
	}

	environmentToolsMap := make(EnvironmentToolsMap)

	for _, envAsset := range envAssets.Entries {
		if len(envAsset.Spec.Connections) == 0 {
			continue
		}

		toolsWeEncountered := datastructs.Set[api.ToolType]{} // make sure each tool will appear once
		toolsOfEnv := make([]api.ToolType, 0, len(envAsset.Spec.Connections))

		for _, connection := range envAsset.Spec.Connections {
			if !toolsWeEncountered.Has(getToolType(connection)) {
				toolsOfEnv = append(toolsOfEnv, getToolType(connection))
				toolsWeEncountered.Add(getToolType(connection))
			}
		}

		environmentToolsMap[envAsset.Meta.Id] = toolsOfEnv
	}

	return environmentToolsMap, nil
}

func DeleteTemplateAsset(ctx context.Context, id string, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	specId := dbmodel.TemplateAssetRec{AssetId: id}
	specRows, specErr := tx.Where(&specId).Delete(&specId)
	if specErr != nil || specRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(specRows, specErr)
	}

	contentId := dbmodel.TemplateContentRec{TemplateAssetId: id}
	contentRows, contentErr := tx.Where(&contentId).Delete(&contentId)
	if contentErr != nil || contentRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetailsFromDbResult(contentRows, contentErr)
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_WORKLOAD_TEMPLATE).WithDetails(err)
	}
	committed = true

	return nil
}

func ListTemplatesUsingAssetId(ctx context.Context, id string) ([]dbmodel.AssetMeta, *api.Error) {
	containment, err := ListTemplatesUsingAssets(ctx, &id)
	if err != nil {
		return nil, err
	}

	templates := make([]dbmodel.AssetMeta, 0, len(containment))
	for _, template := range containment {
		templates = append(templates, template.AssetMeta.AssetMeta)
	}
	return templates, nil
}

func ListTemplatesUsingAssets(ctx context.Context, id *string) ([]dbmodel.TemplateAssetsCombinedRec, *api.Error) {
	result := runaidb.DB().WithContext(ctx).Model(dbmodel.TemplateContentRec{}).Select("*")

	if id != nil {
		result.Where(dbmodel.TemplateContentRec{ContainedAssetId: *id})
	}

	containment := make([]dbmodel.TemplateAssetsCombinedRec, 0)

	result.Joins(JoinMetadataAndTemplatesToGetTemplates())

	result.Where("(scope = ? or tenant_id = ?)", api.SCOPE_SYSTEM, auth.TenantIdOf(ctx))

	result.Order(fmt.Sprintf("%s.%s", dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnName))

	_, scanErr := result.Scan(&containment)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to validate the request", scanErr)
	}

	return containment, nil
}

func getToolType(connection api.Connection) api.ToolType {
	if connection.IsExternal {
		if connection.ExternalToolInfo != nil {
			return api.ToolType(connection.ExternalToolInfo.ToolType)
		}
	} else {
		if connection.InternalToolInfo != nil {
			return api.ToolType(connection.InternalToolInfo.ToolType)
		}
	}
	return api.TOOLTYPE_CUSTOM
}
