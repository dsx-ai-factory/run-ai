package dao

import (
	"fmt"
	"net/http"
	"strings"

	"context"

	"github.com/google/uuid"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

func ListStorageClassConfigurations(ctx context.Context, filterBy []queryapi.Filter, sortBy string, sortOrder string, offset int32,
	limit int32, permittedScopes *openapi.PermittedScopes) ([]models.StorageClassConfiguration, *api.Error) {
	storageClassConfigurations := &[]models.StorageClassConfiguration{}

	query := runaidb.DB().WithContext(ctx).Model(&models.StorageClassConfiguration{})
	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if permittedScopes != nil {
		query = query.Where(getPermittedScopesQuery(*permittedScopes))
	}

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if err := query.Find(&storageClassConfigurations); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list storage class configurations, error: %v", err)
		return nil, api.NewError(http.StatusServiceUnavailable, "Failed to list storage class configurations", err)
	}

	return *storageClassConfigurations, nil
}

func GetStorageClassConfigurationById(ctx context.Context, id string, tenantId int32) (*models.StorageClassConfiguration, *api.Error) {
	record := &models.StorageClassConfiguration{}
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("get"), nil)

	db := runaidb.DB()

	uuidVal, err := uuid.Parse(id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Invalid UUID: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	keys := models.StorageClassConfiguration{
		ID: uuidVal,
		StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{
			TenantId: tenantId,
		},
	}

	if err := db.Where(&keys).First(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return record, nil
}

func GetStorageClassConfigurationByName(ctx context.Context, clusterId string, name string, tenantId int32) (*models.StorageClassConfiguration, *api.Error) {
	record := &models.StorageClassConfiguration{}
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("get"), nil)

	filter := models.StorageClassConfigurationFilters{
		TenantId:  tenantId,
		ScopeType: convutils.Ref(string(openapi.SCOPETYPE_CLUSTER)),
		ScopeId:   convutils.Ref(clusterId),
		Name:      convutils.Ref(name),
	}

	db := runaidb.DB()

	if err := db.Model(&record).Where(&filter).First(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return record, nil
}

func CreateStorageClassConfiguration(ctx context.Context, storageClassConfiguration *models.StorageClassConfiguration) (*models.StorageClassConfiguration, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, storageClassConfiguration.ErrorMessage("create"), nil)

	db := runaidb.DB()

	if _, err := db.Create(storageClassConfiguration); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return storageClassConfiguration, nil
}

func DeleteStorageClassConfiguration(ctx context.Context, id string, tenantId int32) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete storage class configuration", nil)

	db := runaidb.DB()

	uuidVal, err := uuid.Parse(id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Invalid UUID: %v", err)
		return dbErr.WithDetails(err)
	}

	keys := models.StorageClassConfiguration{
		ID: uuidVal,
		StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{
			TenantId: tenantId,
		},
	}

	rowsAffected, err := db.Where(&keys).Delete(&models.StorageClassConfiguration{})

	if err != nil || rowsAffected == 0 {
		dbErr = dbErr.WithDetailsFromDbResult(rowsAffected, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}

// DeleteProjectStorageClasses deletes storage class of a project
func DeleteProjectStorageClasses(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, openapi.SCOPETYPE_PROJECT, tenantId, projectId, &models.StorageClassConfiguration{})
}

// DeleteClusterStorageClasses deletes storage class within a cluster
func DeleteClusterStorageClasses(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	// StorageClasses doesn't hold cluster id, but uses scope type + id, therefore deleting by scope
	return deleteEntitiesByScope(ctx, openapi.SCOPETYPE_CLUSTER, tenantId, clusterId, &models.StorageClassConfiguration{})
}

// DeleteTenantStorageClasses deletes storage class within a cluster
func DeleteTenantStorageClasses(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.StorageClassConfiguration{})
}

func UpdateStorageClassConfiguration(ctx context.Context, id string, storageClassConfiguration *models.StorageClassConfiguration, tenantId int32) (*models.StorageClassConfiguration, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, storageClassConfiguration.ErrorMessage("update"), nil)

	db := runaidb.DB()

	uuidVal, err := uuid.Parse(id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Invalid UUID: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	keys := models.StorageClassConfiguration{
		ID: uuidVal,
		StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{
			TenantId: tenantId,
		},
	}

	if _, err := db.Model(&keys).Updates(storageClassConfiguration); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return storageClassConfiguration, nil
}

func getPermittedScopesQuery(permittedScopes openapi.PermittedScopes) string {
	if permittedScopes.System != nil {
		return "(true)"
	}
	orClause := "(false "
	if len(permittedScopes.Tenants) > 0 {
		orClause += fmt.Sprintf("OR %s IN %s", models.StorageClassConfigurationTenantIdColumn, listItemsForQuery(permittedScopes.Tenants))
	}

	if len(permittedScopes.Clusters) > 0 {
		orClause += fmt.Sprintf("OR (%s = 'cluster' AND %s IN %s)", models.StorageClassConfigurationScopeTypeColumn, models.StorageClassConfigurationScopeIdColumn, listItemsForQuery(permittedScopes.Clusters))
	}

	orClause += ")"
	return orClause
}

// listItemsForQuery formats a slice of items into an inline SQL list, e.g. ( 'a', 'b').
func listItemsForQuery(items []string) string {
	base := "("
	for _, item := range items {
		base += fmt.Sprintf(" '%s',", item)
	}
	return strings.TrimSuffix(base, ",") + ")"
}
