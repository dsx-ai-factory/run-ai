package dao

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"k8s.io/utils/ptr"
)

const (
	NameField   = "name"
	GroupField  = "\"group\""
	IdField     = "id"
	SourceField = "source"

	WorkloadTypeConfigColumnCategoryId     = "category_id"
	WorkloadTypeConfigColumnPriorityId     = "priority_id"
	WorkloadTypeConfigColumnPreemptibility = "preemptibility"
	WorkloadTypeConfigColumnUpdatedBy      = "updated_by"
	WorkloadTypeConfigColumnUpdatedAt      = "updated_at"

	WorkloadTypeStatusColumnUpdatedAt = "updated_at"
	WorkloadTypeStatusColumnPhase     = "phase"
	WorkloadTypeStatusConditionsPhase = "conditions"

	NonPreemptibleFromPriorityValue = 100
)

type WorkloadTypeCreateUpdateFieldsParsed struct {
	WorkloadTypeName string
	WorkloadTypeId   uuid.UUID
	CategoryId       *uuid.UUID
	PriorityId       *uuid.UUID
	Preemptibility   *string

	Kartas *[]models.Karta
}

type WorkloadTypeCreateFieldsParsed struct {
	WorkloadTypeCreateUpdateFieldsParsed
	Group string
}

type WorkloadTypeUpdateFieldsParsed struct {
	WorkloadTypeCreateUpdateFieldsParsed
}

type fieldNameValue struct {
	name  string
	value string
}

func DbListWorkloadTypeConfigs(ctx context.Context, tenantId uint, permittedScopes *authorization_api.PermittedScopes,
	filterBy []validators.ParameterizedFilter, noStatus bool) ([]models.ExternalWorkloadTypeConfigView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list workload type configs", nil)

	workloadTypeConfigs := []models.ExternalWorkloadTypeConfigView{}

	query := dao_utils.DB().WithContext(ctx).
		Table(models.ExternalWorkloadTypeConfigView{}.TableName()).
		Select("DISTINCT ON (id) *").
		Where("tenant_id = ? OR tenant_id IS NULL", strconv.Itoa(int(tenantId)))

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	result := query.
		Order("id").
		Order(fmt.Sprintf("CASE WHEN tenant_id = '%d' THEN 0 ELSE 1 END", tenantId)).
		Scan(&workloadTypeConfigs)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workload type configs: %v", result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	if noStatus {
		return workloadTypeConfigs, nil
	}

	workloadTypeIds := make([]string, len(workloadTypeConfigs))
	for i, wt := range workloadTypeConfigs {
		workloadTypeIds[i] = wt.ID.String()
	}

	statuses, dbStatusesErr := dbListWorkloadTypeStatuses(ctx, workloadTypeIds, permittedScopes)
	if dbStatusesErr != nil {
		return nil, dbStatusesErr
	}

	// group statuses by workload_type_id
	statusesMap := make(map[string][]models.WorkloadTypeStatus)
	for _, status := range statuses {
		workloadTypeId := status.WorkloadTypeId.String()

		if _, ok := statusesMap[workloadTypeId]; !ok {
			statusesMap[workloadTypeId] = []models.WorkloadTypeStatus{}
		}

		statusesMap[workloadTypeId] = append(statusesMap[workloadTypeId], status)
	}

	// Add statuses to each workload type config
	for i, workloadType := range workloadTypeConfigs {
		workloadTypeId := workloadType.ID.String()
		if workloadTypeStatuses, ok := statusesMap[workloadTypeId]; ok {
			workloadTypeConfigs[i].Statuses = workloadTypeStatuses
		}
	}

	return workloadTypeConfigs, nil
}

func dbListWorkloadTypeStatuses(ctx context.Context, workloadTypeIds []string, permittedScopes *authorization_api.PermittedScopes) ([]models.WorkloadTypeStatus, *api.Error) {
	if len(workloadTypeIds) == 0 {
		return []models.WorkloadTypeStatus{}, nil // No workload types to list statuses for
	}

	errMessage := "Failed to list workload type statuses"
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	query := dao_utils.DB().WithContext(ctx).
		Table(models.WorkloadTypeStatus{}.TableName()).
		Where("workload_type_id IN ?", workloadTypeIds)

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if !isMgmtTenant && permittedScopes != nil {
		query = query.Where(getPermittedScopesWorkloadTypesDbFilter(*permittedScopes))
	}

	var statuses []models.WorkloadTypeStatus
	result := query.Find(&statuses)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return statuses, nil
}

func DbGetWorkloadTypeConfigByIdWithStatuses(ctx context.Context, tenantId uint, workloadTypeId string, permittedScopes *authorization_api.PermittedScopes) (*models.ExternalWorkloadTypeConfigView, *api.Error) {
	result, apiErr := DbGetWorkloadTypeConfigById(ctx, tenantId, workloadTypeId)
	if apiErr != nil {
		return nil, apiErr
	}

	statuses, dbStatusesErr := dbListWorkloadTypeStatuses(ctx, []string{workloadTypeId}, permittedScopes)
	if dbStatusesErr != nil {
		return nil, dbStatusesErr
	}

	result.Statuses = statuses
	return result, nil
}

func DbGetWorkloadTypeConfigById(ctx context.Context, tenantId uint, workloadTypeId string) (*models.ExternalWorkloadTypeConfigView, *api.Error) {
	byFields := []fieldNameValue{{name: IdField, value: workloadTypeId}}
	return dbGetWorkloadTypeConfigByFields(ctx, tenantId, byFields, []fieldNameValue{})
}

func DbGetWorkloadTypeConfigByNameGroup(ctx context.Context, tenantId uint, workloadTypeName, group string) (*models.ExternalWorkloadTypeConfigView, *api.Error) {
	byFields := []fieldNameValue{{name: NameField, value: workloadTypeName}}
	excludedByFields := []fieldNameValue{}
	if group != "" {
		byFields = append(byFields, fieldNameValue{name: GroupField, value: group})
	} else {
		// if group is empty, we should filter by source != "custom"
		// because "custom" workload types must have a group;
		// while for "native" and "external" workload types this func may be called with empty group
		excludedByFields = append(excludedByFields, fieldNameValue{name: SourceField, value: string(models.WorkloadTypeSourceCustom)})
	}

	return dbGetWorkloadTypeConfigByFields(ctx, tenantId, byFields, excludedByFields)
}

func DbCountWorkloadTypeConfigsByNameGroup(ctx context.Context, tenantId uint, name, group string) (int64, *api.Error) {
	errMessage := fmt.Sprintf("Failed to count workload type config by name = %s, group = %s", name, group)
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	var count int64

	result := dao_utils.DB().WithContext(ctx).
		Table(models.ExternalWorkloadTypeConfigView{}.TableName()).
		Where("name = ?", name).
		Where("\"group\" = ?", group).
		Where("(scope_type = ? OR tenant_id = ?)", api.SCOPETYPE_SYSTEM, strconv.Itoa(int(tenantId))).
		Count(&count)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return 0, dbErr.WithDetails(result.Error)
	}

	return count, nil
}

// dbGetWorkloadTypeConfigByFields - a helper function to retrieve workload type config by specific fields.
// Returns the first matching config for the given fields and values - either in tenant scope,
// or if it doesn't exist in tenant scope - returns the config in system scope.
// If excludeFieldName and excludeFieldValue are provided - will exclude the config with that field value.
func dbGetWorkloadTypeConfigByFields(ctx context.Context, tenantId uint, byFields []fieldNameValue,
	excludedByFields []fieldNameValue) (*models.ExternalWorkloadTypeConfigView, *api.Error) {
	if len(byFields) == 0 {
		errMessage := "Failed to get workload type config by fields - invalid parameters"
		apiErr := api.NewError(http.StatusInternalServerError, errMessage, nil)
		return nil, apiErr
	}

	errMessage := fmt.Sprintf("Failed to get workload type config by fields: %v", byFields)
	if len(excludedByFields) > 0 {
		errMessage += fmt.Sprintf(", excluded by fields: %v", excludedByFields)
	}
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	workloadTypeConfig := models.ExternalWorkloadTypeConfigView{}

	query := dao_utils.DB().WithContext(ctx).
		Table(models.ExternalWorkloadTypeConfigView{}.TableName())
	for _, byField := range byFields {
		query = query.Where(fmt.Sprintf("%s = ?", byField.name), byField.value)
	}
	for _, excludedByField := range excludedByFields {
		query = query.Where(fmt.Sprintf("%s != ?", excludedByField.name), excludedByField.value)
	}

	result := query.
		Where("(scope_type = ? OR tenant_id = ?)", api.SCOPETYPE_SYSTEM, strconv.Itoa(int(tenantId))).
		Order(fmt.Sprintf("CASE WHEN tenant_id = '%d' THEN 0 ELSE 1 END", tenantId)).
		First(&workloadTypeConfig)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &workloadTypeConfig, nil
}

// DbGetWorkloadTypeConfigByIdForScope retrieves the workload type config by ID for a specific scope (tenant or system) -
// from the table, and not from the view.
// Should be used before calling "Delete"/"Update" to know if the config exists for the given scope.
func DbGetWorkloadTypeConfigByIdForScope(ctx context.Context, tenantId uint, workloadTypeId string) (*models.ExternalWorkloadTypeConfig, *api.Error) {
	errMessage := fmt.Sprintf("Failed to get workload type config by id = <%s>", workloadTypeId)
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	workloadTypeConfig := models.ExternalWorkloadTypeConfig{}

	query := dao_utils.DB().WithContext(ctx).
		Table(models.ExternalWorkloadTypeConfig{}.TableName()).
		Where("workload_type_id = ?", workloadTypeId)

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if isMgmtTenant {
		query = query.Where("scope_type = ?", api.SCOPETYPE_SYSTEM)
	} else {
		query = query.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).
			Where("scope_type = ?", api.SCOPETYPE_TENANT)
	}

	result := query.
		First(&workloadTypeConfig)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &workloadTypeConfig, nil
}

func DbListWorkloadTypes(ctx context.Context, filterWorkloadSource *models.WorkloadTypeSource, filterBy []validators.ParameterizedFilter) ([]models.WorkloadType, *api.Error) {
	errMessage := "Failed to list workload types"
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	workloadTypes := []models.WorkloadType{}

	query := dao_utils.DB().WithContext(ctx)

	if filterWorkloadSource != nil {
		query = query.Where("source = ?", *filterWorkloadSource)
	}

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	result := query.
		Find(&workloadTypes)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return workloadTypes, nil
}

func DbGetWorkloadTypeById(ctx context.Context, workloadTypeId uuid.UUID) (*models.WorkloadType, *api.Error) {
	byFields := []fieldNameValue{{name: IdField, value: workloadTypeId.String()}}
	return dbGetWorkloadTypeByFields(ctx, byFields)
}

func DbGetWorkloadTypeByNameGroup(ctx context.Context, name, group string) (*models.WorkloadType, *api.Error) {
	byFields := []fieldNameValue{{name: NameField, value: name}, {name: GroupField, value: group}}
	return dbGetWorkloadTypeByFields(ctx, byFields)
}

// DbGetKartaByGVK returns the Karta for the exact workload GVK in the current scope.
// In tenant scope, tenant-scoped Kartas take precedence over system-scoped Kartas.
// Returns nil without error when no matching Karta exists.
func DbGetKartaByGVK(ctx context.Context, tenantId uint, kind, group, version string) (*models.WorkloadTypeKarta, *api.Error) {
	workloadTypeConfig, apiErr := DbGetWorkloadTypeConfigByNameGroup(ctx, tenantId, kind, group)
	if apiErr != nil {
		return nil, apiErr
	}

	for _, karta := range workloadTypeConfig.Kartas {
		if karta.Version == version {
			return &karta, nil
		}
	}
	return nil, nil
}

func dbGetWorkloadTypeByFields(ctx context.Context, byFields []fieldNameValue) (*models.WorkloadType, *api.Error) {
	if len(byFields) == 0 {
		errMessage := "Failed to get workload type by fields - invalid parameters"
		apiErr := api.NewError(http.StatusInternalServerError, errMessage, nil)
		return nil, apiErr
	}

	errMessage := fmt.Sprintf("Failed to get workload type config by fields: %v", byFields)
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	workloadType := models.WorkloadType{}

	query := dao_utils.DB().WithContext(ctx).
		Table(models.WorkloadType{}.TableName())
	for _, byField := range byFields {
		query = query.Where(fmt.Sprintf("%s = ?", byField.name), byField.value)
	}

	result := query.First(&workloadType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &workloadType, nil
}

func DbCreateWorkloadTypeWithConfig(ctx context.Context, tenantId uint,
	createFields WorkloadTypeCreateFieldsParsed) (*models.ExternalWorkloadTypeConfig, *api.Error) {
	errorMsg := fmt.Sprintf("Failed to create workload type with config with name <%s>, group <%s>",
		createFields.WorkloadTypeName, createFields.Group)
	dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return nil, dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	// first upsert to workload_types table
	workloadType, apiErr := DbUpsertWorkloadType(ctx, tx, createFields)
	if apiErr != nil {
		return nil, apiErr
	}
	createFields.WorkloadTypeId = workloadType.ID

	// then create the external workload type config
	workloadTypeConfig, apiErr := dbCreateExternalWorkloadTypeConfig(ctx, tx, tenantId, createFields)
	if apiErr != nil {
		return nil, apiErr
	}

	_, apiErr = dbCreateWorkloadKartas(ctx, tx, tenantId, createFields.WorkloadTypeCreateUpdateFieldsParsed)
	if apiErr != nil {
		return nil, apiErr
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the create workload type with config, error: %v", result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}
	committed = true

	return workloadTypeConfig, nil
}

func DbUpsertWorkloadType(ctx context.Context, tx *gorm.DB, createFields WorkloadTypeCreateFieldsParsed) (*models.WorkloadType, *api.Error) {
	errorMsg := fmt.Sprintf("Failed to upsert workload type with name <%s>, group <%s>",
		createFields.WorkloadTypeName, createFields.Group)
	dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)

	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	timeNow := time.Now()
	workloadType := models.WorkloadType{
		Name:      createFields.WorkloadTypeName,
		CreatedAt: timeNow,
		CreatedBy: subjectId,
		Source:    models.WorkloadTypeSourceCustom,
		Group:     createFields.Group,
	}

	result := tx.
		Table(workloadType.TableName()).
		Where("name = ?", createFields.WorkloadTypeName).
		Where("\"group\" = ?", createFields.Group).
		FirstOrCreate(&workloadType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}
	return &workloadType, nil
}

func dbDeleteWorkloadType(ctx context.Context, tx *gorm.DB, workloadTypeId string) *api.Error {
	errorMsg := fmt.Sprintf("Failed to delete workload type with id <%s>", workloadTypeId)
	dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)

	result := tx.
		Table(models.WorkloadType{}.TableName()).
		Where("id = ?", workloadTypeId).
		Delete(&models.WorkloadType{})
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		return dbErr.WithDetails(result.Error)
	}
	return nil
}

// dbCreateExternalWorkloadTypeConfig - separate function for create;
// it is called only for "Create" workload type, when the config does not exist yet.
// Will return an error if the config already exists in the required scope.
func dbCreateExternalWorkloadTypeConfig(ctx context.Context, tx *gorm.DB, tenantId uint, createFields WorkloadTypeCreateFieldsParsed) (*models.ExternalWorkloadTypeConfig, *api.Error) {
	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	timeNow := time.Now()

	preemptibility := createFields.Preemptibility
	if preemptibility == nil {
		var apiErr *api.Error
		preemptibility, apiErr = DerivePreemptibilityFromPriorityID(ctx, tenantId, createFields.PriorityId)
		if apiErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "Failed to create workload type config", apiErr)
		}
	}

	newConfig := models.ExternalWorkloadTypeConfig{
		ID:             uuid.New(),
		WorkloadTypeId: createFields.WorkloadTypeId,
		CreatedAt:      timeNow,
		CreatedBy:      subjectId,
		UpdatedAt:      timeNow,
		UpdatedBy:      subjectId,
		CategoryID:     createFields.CategoryId,
		PriorityID:     createFields.PriorityId,
		Preemptibility: preemptibility,
	}

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if isMgmtTenant {
		newConfig.TenantId = nil
		newConfig.ScopeType = string(api.SCOPETYPE_SYSTEM)
	} else {
		newConfig.TenantId = ptr.To(strconv.Itoa(int(tenantId)))
		newConfig.ScopeType = string(api.SCOPETYPE_TENANT)
	}

	result := tx.Create(&newConfig)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("Failed to create workload type config with WorkloadTypeId <%s> for %s scope",
			createFields.WorkloadTypeId, scopeStr(isMgmtTenant))
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)
		return nil, dbErr.WithDetails(result.Error)
	}
	return &newConfig, nil
}

func dbCreateWorkloadKartas(ctx context.Context, tx *gorm.DB, tenantId uint, fields WorkloadTypeCreateUpdateFieldsParsed) ([]uuid.UUID, *api.Error) {
	if fields.Kartas == nil || len(*fields.Kartas) == 0 {
		return []uuid.UUID{}, nil
	}

	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	timeNow := time.Now()

	var tenantForRd *string = nil
	scopeTypeForRd := string(api.SCOPETYPE_SYSTEM)
	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if !isMgmtTenant {
		tenantForRd = ptr.To(strconv.Itoa(int(tenantId)))
		scopeTypeForRd = string(api.SCOPETYPE_TENANT)
	}

	uuids := make([]uuid.UUID, len(*fields.Kartas))
	for i, karta := range *fields.Kartas {
		kartaToUpsert := models.WorkloadTypeKarta{
			TenantId:       tenantForRd,
			ScopeType:      scopeTypeForRd,
			WorkloadTypeId: fields.WorkloadTypeId,
			Version:        karta.GroupVersionKind().Version,
			Karta:          karta,
			CreatedAt:      timeNow,
			CreatedBy:      subjectId,
		}

		query := tx.Table(kartaToUpsert.TableName()).
			Where("workload_type_id = ?", kartaToUpsert.WorkloadTypeId).
			Where("version = ?", kartaToUpsert.Version)

		if isMgmtTenant {
			query = query.Where("tenant_id IS NULL").
				Where("scope_type = ?", api.SCOPETYPE_SYSTEM)
		} else {
			query = query.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).
				Where("scope_type = ?", api.SCOPETYPE_TENANT)
		}

		result := query.
			FirstOrCreate(&kartaToUpsert)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create workload type kartas for WorkloadTypeId <%s>: %v", fields.WorkloadTypeId, result.Error)
			return nil, api.NewError(http.StatusInternalServerError, "Failed to create workload kartas", result.Error)
		}

		uuids[i] = kartaToUpsert.ID
	}

	return uuids, nil
}

// deletes all workload kartas of the given workloadTypeId in the given scope (system or tenant),
// except for the ones with karta ids in excludedIdsFromDelete (if any).
func dbDeleteWorkloadKartasOfWorkloadTypeInScope(ctx context.Context, tx *gorm.DB, tenantId uint, workloadTypeId uuid.UUID, excludedIdsFromDelete []uuid.UUID) *api.Error {
	query := tx.
		Table(models.WorkloadTypeKarta{}.TableName()).
		Where("workload_type_id = ?", workloadTypeId.String())

	if len(excludedIdsFromDelete) > 0 {
		query = query.Where("id NOT IN ?", excludedIdsFromDelete)
	}

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if !isMgmtTenant {
		query = query.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).
			Where("scope_type = ?", api.SCOPETYPE_TENANT)
	}
	// else - delete all configs for the given workloadTypeId - for system scope and for all tenants

	result := query.Delete(&models.WorkloadTypeKarta{})
	if result.Error != nil {
		errorMsg := fmt.Sprintf("Failed to delete workload kartas of workloadId <%s> for %s scope",
			workloadTypeId.String(), scopeStr(isMgmtTenant))
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)
		return dbErr.WithDetails(result.Error)
	}
	return nil
}

func DbUpdateWorkloadTypeWithConfig(ctx context.Context, tenantId uint, updateFields WorkloadTypeUpdateFieldsParsed) *api.Error {
	errorMsg := fmt.Sprintf("Failed to update workload type with config for workloadTypeId <%s>", updateFields.WorkloadTypeId)
	dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	apiErr := dbAddOrUpdateExternalWorkloadTypeConfig(ctx, tx, tenantId, updateFields)
	if apiErr != nil {
		return apiErr
	}

	excludedIdsFromDelete, apiErr := dbCreateWorkloadKartas(ctx, tx, tenantId, updateFields.WorkloadTypeCreateUpdateFieldsParsed)
	if apiErr != nil {
		return apiErr
	}

	if updateFields.Kartas != nil {
		// only delete if the kartas were explicitly provided in the update request
		apiErr = dbDeleteWorkloadKartasOfWorkloadTypeInScope(ctx, tx, tenantId, updateFields.WorkloadTypeId, excludedIdsFromDelete)
		if apiErr != nil {
			return apiErr
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the create workload type with config, error: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}

	committed = true
	return nil
}

// dbAddOrUpdateExternalWorkloadTypeConfig - called for "Update" workload type:
// if the config didn't exist before for tenant scope (but only for system scope) -
// it will create a new config for tenant scope.
// And the opposite - if the config didn't exist before for system scope (but only for tenant scope) -
// it will create a new config for system scope.
func dbAddOrUpdateExternalWorkloadTypeConfig(ctx context.Context, tx *gorm.DB, tenantId uint, updateFields WorkloadTypeUpdateFieldsParsed) *api.Error {
	_, apiErr := DbGetWorkloadTypeConfigByIdForScope(ctx, tenantId, updateFields.WorkloadTypeId.String())
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			_, apiErr = dbCreateExternalWorkloadTypeConfig(ctx, tx, tenantId, WorkloadTypeCreateFieldsParsed{WorkloadTypeCreateUpdateFieldsParsed: updateFields.WorkloadTypeCreateUpdateFieldsParsed})
			return apiErr
		}
		return apiErr
	}

	return dbUpdateExternalWorkloadTypeConfig(ctx, tx, tenantId, updateFields)
}

func dbUpdateExternalWorkloadTypeConfig(ctx context.Context, tx *gorm.DB, tenantId uint, updateFields WorkloadTypeUpdateFieldsParsed) *api.Error {
	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	timeNow := time.Now()

	preemptibility := updateFields.Preemptibility
	if preemptibility == nil && updateFields.PriorityId != nil {
		var apiErr *api.Error
		preemptibility, apiErr = DerivePreemptibilityFromPriorityID(ctx, tenantId, updateFields.PriorityId)
		if apiErr != nil {
			return api.NewError(http.StatusInternalServerError, "Failed to update workload type config", apiErr)
		}
	}

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if isMgmtTenant {
		tx = tx.Where("tenant_id IS NULL").Where("scope_type = ?", api.SCOPETYPE_SYSTEM)
	} else {
		tx = tx.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).Where("scope_type = ?", api.SCOPETYPE_TENANT)
	}

	updates := map[string]any{
		WorkloadTypeConfigColumnUpdatedBy: subjectId,
		WorkloadTypeConfigColumnUpdatedAt: timeNow,
	}
	if updateFields.CategoryId != nil {
		updates[WorkloadTypeConfigColumnCategoryId] = updateFields.CategoryId.String()
	}
	if updateFields.PriorityId != nil {
		updates[WorkloadTypeConfigColumnPriorityId] = updateFields.PriorityId.String()
	}
	if preemptibility != nil {
		updates[WorkloadTypeConfigColumnPreemptibility] = *preemptibility
	}

	result := tx.
		Table(models.ExternalWorkloadTypeConfig{}.TableName()).
		Where("workload_type_id = ?", updateFields.WorkloadTypeId).
		Updates(updates)

	if result.Error != nil {
		errorMsg := fmt.Sprintf("Failed to update workload type config with WorkloadTypeId <%s> for %s scope",
			updateFields.WorkloadTypeId, scopeStr(isMgmtTenant))
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)
		return dbErr.WithDetails(result.Error)
	}
	return nil
}

func DbDeleteWorkloadTypeWithConfig(ctx context.Context, tenantId uint, workloadTypeId uuid.UUID) *api.Error {
	errorMsg := fmt.Sprintf("Failed to delete workload type config with id <%s> for %s scope",
		workloadTypeId.String(), scopeStr(runaictx.IsManagementTenant(ctx)))
	dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	apiErr := dbDeleteWorkloadKartasOfWorkloadTypeInScope(ctx, tx, tenantId, workloadTypeId, []uuid.UUID{})
	if apiErr != nil {
		return apiErr
	}

	apiErr = dbDeleteExternalWorkloadTypeConfig(ctx, tx, tenantId, workloadTypeId)
	if apiErr != nil {
		return apiErr
	}

	count, apiErr := dbCountExternalWorkloadTypeConfigsForIdForAllScopes(ctx, tx, workloadTypeId.String())
	if apiErr != nil {
		return apiErr
	}

	if count == 0 {
		// if there are no configs for the given workloadTypeId - delete the workload type itself
		apiErr = dbDeleteWorkloadType(ctx, tx, workloadTypeId.String())
		if apiErr != nil {
			return apiErr
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the delete workload type with config with id <%s> for %s scope, error: %v",
			workloadTypeId.String(), scopeStr(runaictx.IsManagementTenant(ctx)), result.Error)
		return dbErr.WithDetails(result.Error)
	}

	committed = true
	return nil
}

func DbDoesExternalWorkloadTypeConfigExistForIdForSystemScope(ctx context.Context, workloadTypeId string) (bool, *api.Error) {
	tx := dao_utils.DB().WithContext(ctx)
	count, dbErr := dbCountExternalWorkloadTypeConfigsForIdForScopesInner(ctx, tx, workloadTypeId, true)
	if dbErr != nil {
		return false, dbErr
	}

	return count > 0, nil
}

func DbSyncWorkloadTypeStatuses(
	ctx context.Context, tenantId string, clusterId uuid.UUID,
	workloadTypeStatuses, workloadTypeStatusesToDelete []models.WorkloadTypeStatus, syncDeleted bool) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync workload type statuses", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	synced := []string{}
	failed := []string{}
	excludedIdsFromDelete := []string{}

	for _, workloadTypeStatus := range workloadTypeStatuses {
		workloadTypeId := workloadTypeStatus.WorkloadTypeId.String()

		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "workload_type_id"}, {Name: "version"}, {Name: "cluster_id"}, {Name: "tenant_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				WorkloadTypeStatusColumnUpdatedAt: workloadTypeStatus.UpdatedAt,
				WorkloadTypeStatusColumnPhase:     workloadTypeStatus.Phase,
				WorkloadTypeStatusConditionsPhase: workloadTypeStatus.Conditions,
			}),
		}).Create(&workloadTypeStatus)

		if result.Error != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workload type status for clusterId <%s>, workloadTypeId <%s>, error: %v",
				clusterId.String(), workloadTypeId, result.Error)
			failed = append(failed, workloadTypeId)
		} else {
			tx.SavePoint("sp")
			synced = append(synced, workloadTypeId)
		}

		excludedIdsFromDelete = append(excludedIdsFromDelete, workloadTypeId)
	}

	for _, workloadTypeIdVersionToDelete := range workloadTypeStatusesToDelete {
		result := tx.Table(models.WorkloadTypeStatus{}.TableName()).
			Where("tenant_id = ?", tenantId).
			Where("cluster_id = ?", clusterId.String()).
			Where("workload_type_id = ?", workloadTypeIdVersionToDelete.WorkloadTypeId).
			Where("version = ?", workloadTypeIdVersionToDelete.Version).
			Delete(&models.WorkloadTypeStatus{})
		if result.Error != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload type status for clusterId <%s>, workloadTypeId <%s>, version <%s>, error: %v",
				clusterId.String(), workloadTypeIdVersionToDelete.WorkloadTypeId, workloadTypeIdVersionToDelete.Version, result.Error)
			failed = append(failed, workloadTypeIdVersionToDelete.WorkloadTypeId.String())
		} else {
			tx.SavePoint("sp")
			synced = append(synced, workloadTypeIdVersionToDelete.WorkloadTypeId.String())
		}
	}

	if syncDeleted {
		// delete all workload type statuses that were not synced and not in the excludedIdsFromDelete
		query := tx.Table(models.WorkloadTypeStatus{}.TableName()).
			Where("tenant_id = ?", tenantId).
			Where("cluster_id = ?", clusterId.String())
		if len(excludedIdsFromDelete) > 0 {
			query = query.Where("workload_type_id NOT IN ?", excludedIdsFromDelete)
		}

		result := query.
			Delete(&models.WorkloadTypeStatus{})
		if result.Error != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete workload type statuses during full sync for clusterId <%s>, error: %v",
				clusterId.String(), result.Error)
		} else {
			tx.SavePoint("sp")
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the sync workload type statuses, error: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Synced workload type statuses in clusterId <%s> for: %v",
		clusterId.String(), synced)
	if len(failed) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync workload type statuses in clusterId <%s> for: %v",
			clusterId.String(), failed)
	}

	return nil
}

// dbCountExternalWorkloadTypeConfigsForIdForAllScopes
// Counts the number of external workload type configs across all scopes (tenant and system).
// should only be used internally, as it does not filter by tenantId.
func dbCountExternalWorkloadTypeConfigsForIdForAllScopes(ctx context.Context, tx *gorm.DB, workloadTypeId string) (int64, *api.Error) {
	return dbCountExternalWorkloadTypeConfigsForIdForScopesInner(ctx, tx, workloadTypeId, false)
}

func dbCountExternalWorkloadTypeConfigsForIdForScopesInner(ctx context.Context, tx *gorm.DB, workloadTypeId string, countSystemOnly bool) (int64, *api.Error) {
	errMessage := fmt.Sprintf("Failed to count external workload type configs for workload id <%s> all scopes", workloadTypeId)
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	var count int64

	query := tx.
		Table(models.ExternalWorkloadTypeConfig{}.TableName()).
		Where("workload_type_id = ?", workloadTypeId)

	if countSystemOnly {
		query = query.Where("scope_type = ?", api.SCOPETYPE_SYSTEM)
	}

	result := query.
		Count(&count)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, result.Error)
		return 0, dbErr.WithDetails(result.Error)
	}

	return count, nil
}

func dbDeleteExternalWorkloadTypeConfig(ctx context.Context, tx *gorm.DB, tenantId uint, workloadTypeId uuid.UUID) *api.Error {
	query := tx.
		Table(models.ExternalWorkloadTypeConfig{}.TableName()).
		Where("workload_type_id = ?", workloadTypeId.String())

	isMgmtTenant := runaictx.IsManagementTenant(ctx)
	if !isMgmtTenant {
		query = query.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).
			Where("scope_type = ?", api.SCOPETYPE_TENANT)
	}
	// else - delete all configs for the given workloadTypeId - for system scope and for all tenants

	result := query.Delete(&models.ExternalWorkloadTypeConfig{})
	if result.Error != nil {
		errorMsg := fmt.Sprintf("Failed to delete external workload type configs for workloadId <%s> for %s scope",
			workloadTypeId.String(), scopeStr(isMgmtTenant))
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, result.Error)
		dbErr := api.NewError(http.StatusInternalServerError, errorMsg, nil)
		return dbErr.WithDetails(result.Error)
	}
	return nil
}

// DeleteExternalWorkloadTypeConfigsOfTenantId deletes all external workload type configs for a specific tenant
func DeleteExternalWorkloadTypeConfigsOfTenantId(ctx context.Context, tenantId string) error {
	result := dao_utils.DB().WithContext(ctx).
		Table(models.ExternalWorkloadTypeConfig{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Delete(&models.ExternalWorkloadTypeConfig{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete external workload type configs for tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}

// DeleteExternalWorkloadTypeKartasOfTenantId deletes all workload type kartas for a specific tenant
func DeleteExternalWorkloadTypeKartasOfTenantId(ctx context.Context, tenantId string) error {
	result := dao_utils.DB().WithContext(ctx).
		Table(models.WorkloadTypeKarta{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Delete(&models.WorkloadTypeKarta{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete workload type kartas for tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}

// DeleteExternalWorkloadTypeResourceInterfacesOfTenantId is an alias for backward compatibility.
var DeleteExternalWorkloadTypeResourceInterfacesOfTenantId = DeleteExternalWorkloadTypeKartasOfTenantId

// DeleteWorkloadTypeStatusesOfTenantId deletes all workload type statuses for a specific tenant
func DeleteWorkloadTypeStatusesOfTenantId(ctx context.Context, tenantId string) error {
	result := dao_utils.DB().WithContext(ctx).Unscoped().
		Table(models.WorkloadTypeStatus{}.TableName()).
		Where("tenant_id = ?", tenantId).
		Delete(&models.WorkloadTypeStatus{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete workload type statuses for tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}

func DerivePreemptibilityFromPriority(priority *models.Priority) string {
	if priority.Value >= NonPreemptibleFromPriorityValue {
		return string(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)
	}
	return string(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
}

// DerivePreemptibilityFromPriorityID returns:
// - nil if priorityId is nil
// - "non-preemptible" if priority value >= 100
// - "preemptible" otherwise
func DerivePreemptibilityFromPriorityID(ctx context.Context, tenantId uint, priorityId *uuid.UUID) (*string, *api.Error) {
	if priorityId == nil {
		return nil, nil
	}

	priority, apiErr := GetPriorityById(ctx, strconv.Itoa(int(tenantId)), priorityId.String())
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to derive preemptibility from priority id <%s> for create/update workload type",
			priorityId.String())
		return nil, apiErr
	}

	return ptr.To(DerivePreemptibilityFromPriority(priority)), nil
}

func scopeStr(isMgmtTenant bool) string {
	if isMgmtTenant {
		return string(api.SCOPETYPE_SYSTEM)
	}
	return string(api.SCOPETYPE_TENANT)
}

func getPermittedScopesWorkloadTypesDbFilter(permittedScopes authorization_api.PermittedScopes) string {
	return authorization_utils.GetPermittedScopesDbFilterWithColumnNames(
		permittedScopes,
		"tenant_id",
		"cluster_id",
		"",
		"",
	)
}
