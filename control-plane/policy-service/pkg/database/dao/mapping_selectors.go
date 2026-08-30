package dao

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
	"github.com/run-ai/runai/runai-common-packages/logging"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

// the user is allowed to see only system scope or tenant scoped from his own tenant
func filterMappingSelectorsByTenant(res *gorm.DB, ctx context.Context) *gorm.DB {
	tenantID := auth.TenantIdOf(ctx)
	whereSQL := filterMappingSelectorsByTenantQuery()
	return res.Where(whereSQL, string(api.MAPPINGSCOPE_SYSTEM), string(api.MAPPINGSCOPE_TENANT), tenantID)
}

func filterMappingSelectorsByTenantQuery() string {
	return fmt.Sprintf("((%s = ?) OR (%s = ? AND %s = ?))",
		dbtables.MappingSelectorsColumnScopeType,
		dbtables.MappingSelectorsColumnScopeType,
		dbtables.MappingSelectorsColumnTenantId)
}

func CreateMapping(ctx context.Context, row *dbmodel.MappingSelectors) *api.Error {
	if row.SpecSelectors == nil {
		row.SpecSelectors = map[string]api.SpecSelectorsWithScopeValue{}
	}

	res := dao_utils.DB().WithContext(ctx).Table(dbtables.MappingSelectorsTable).Create(row)
	if res.Error != nil {
		dbErr := database.Error(database.CREATE, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to create mapping: %s: %v", dbErr.Message, res.Error)
		return dbErr
	}

	return nil
}

// MappingByScope holds system and/or tenant rows for the same workload GVK.
type MappingByScope map[api.MappingScope]dbmodel.MappingSelectors

// gvkKey is a string-based key for map lookups, avoiding pointer comparison issues with api.Gvk.
type gvkKey string

func makeGvkKey(group, kind, version string) gvkKey {
	return gvkKey(fmt.Sprintf("%s/%s/%s", group, kind, version))
}

// MappingsByGvkAndScope groups visible rows by workload GVK then by scope (parallel to FieldGroupsByNameAndScope).
// Uses string-based keys internally to avoid pointer comparison issues.
type MappingsByGvkAndScope map[gvkKey]MappingByScope

// GvkFor returns the api.Gvk for a given key.
func (m MappingsByGvkAndScope) GvkFor(key gvkKey) api.Gvk {
	if byScope, ok := m[key]; ok {
		for _, row := range byScope {
			return api.Gvk{
				Group:   convutils.Ref(row.GvkGroup),
				Kind:    row.GvkKind,
				Version: row.GvkVersion,
			}
		}
	}
	return api.Gvk{}
}

// ListMappings returns all visible mapping rows for the tenant, optionally narrowed by API filters.
func ListMappings(ctx context.Context, filters []validators.ParameterizedFilter) (MappingsByGvkAndScope, *api.Error) {
	var rows []dbmodel.MappingSelectors

	res := dao_utils.DB().WithContext(ctx).Model(&dbmodel.MappingSelectors{})

	res = filterMappingSelectorsByTenant(res, ctx)

	res = dao_utils.ApplyParameterizedFilters(res, filters)

	res = res.Find(&rows)
	if res.Error != nil {
		dbErr := database.Error(database.LIST, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to list mappings: %s: %v", dbErr.Message, res.Error)
		return nil, dbErr
	}

	out := make(MappingsByGvkAndScope, len(rows))

	for _, row := range rows {
		key := makeGvkKey(row.GvkGroup, row.GvkKind, row.GvkVersion)

		if out[key] == nil {
			out[key] = make(MappingByScope)
		}

		r := row
		out[key][r.ScopeType] = r
	}

	return out, nil
}

// ListMappingsReferencingFieldGroup returns mapping rows visible to ctx whose spec_selectors JSON references
// the given field group name in at least one selector's fieldGroups array.
func ListMappingsReferencingFieldGroup(ctx context.Context, tx *gorm.DB, fieldGroupName string) ([]dbmodel.MappingSelectors, *api.Error) {
	var rows []dbmodel.MappingSelectors

	// spec_selectors: { "<selectorKey>": { "fieldGroups": ["name", ...], ... }, ... }
	subq := `EXISTS (
		SELECT 1 FROM jsonb_each(spec_selectors) AS e(k, v)
		WHERE EXISTS (
			SELECT 1 FROM jsonb_array_elements_text(COALESCE(v->'fieldGroups', '[]'::jsonb)) AS elem
			WHERE elem = ?
		)
	)`

	res := tx.Model(&dbmodel.MappingSelectors{}).Where(subq, fieldGroupName)
	res = filterMappingSelectorsByTenant(res, ctx)
	res = res.Find(&rows)
	if res.Error != nil {
		dbErr := database.Error(database.LIST, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to list mappings referencing field group %s: %s: %v", fieldGroupName, dbErr.Message, res.Error)
		return nil, dbErr
	}

	return rows, nil
}

// GetMappingByGvk returns visible rows for the given GVK (system and/or tenant scoped)
func GetMappingByGvk(ctx context.Context, gvk api.Gvk) (MappingByScope, *api.Error) {
	var rows []dbmodel.MappingSelectors

	where := dbmodel.MappingSelectors{
		GvkGroup:   convutils.Deref(gvk.Group, ""),
		GvkKind:    gvk.Kind,
		GvkVersion: gvk.Version,
	}

	res := dao_utils.DB().WithContext(ctx).Model(&dbmodel.MappingSelectors{}).Where(where)

	res = filterMappingSelectorsByTenant(res, ctx)

	res = res.Find(&rows)
	if res.Error != nil {
		dbErr := database.Error(database.GET, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to get mapping: %s: %v", dbErr.Message, res.Error)
		return nil, dbErr
	}

	out := make(MappingByScope)

	for _, row := range rows {
		r := row
		out[r.ScopeType] = r
	}

	return out, nil
}

// GetMappingByGvkAndScope returns the tenant or system row for a GVK.
func GetMappingByGvkAndScope(ctx context.Context, gvk api.Gvk, scope api.MappingScope) (*dbmodel.MappingSelectors, *api.Error) {
	var row dbmodel.MappingSelectors

	where := dbmodel.MappingSelectors{
		GvkGroup:   convutils.Deref(gvk.Group, ""),
		GvkKind:    gvk.Kind,
		GvkVersion: gvk.Version,
		ScopeType:  scope,
	}
	if scope == api.MAPPINGSCOPE_TENANT {
		where.TenantId = auth.TenantIdOf(ctx)
	}

	res := dao_utils.DB().WithContext(ctx).Model(&dbmodel.MappingSelectors{}).Where(where).First(&row)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped mapping for %s/%s/%s not found", scope, convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version), nil)
		}
		dbErr := database.Error(database.GET, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to get mapping: %s: %v", dbErr.Message, res.Error)
		return nil, dbErr
	}

	return &row, nil
}

func SaveMapping(ctx context.Context, db *gorm.DB, row *dbmodel.MappingSelectors) *api.Error {
	if row.SpecSelectors == nil {
		row.SpecSelectors = map[string]api.SpecSelectorsWithScopeValue{}
	}

	query := dbmodel.MappingSelectors{
		ID:         row.ID,
		GvkGroup:   row.GvkGroup,
		GvkKind:    row.GvkKind,
		GvkVersion: row.GvkVersion,
		ScopeType:  row.ScopeType,
		TenantId:   row.TenantId,
	}
	res := db.WithContext(ctx).Model(&query).Where(query).Save(row)
	if res.Error != nil {
		dbErr := database.Error(database.UPDATE, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to update mapping: %s: %v", dbErr.Message, res.Error)
		return dbErr
	}

	if res.RowsAffected == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("mapping %s/%s/%s not found within scope %s/%d", row.GvkGroup, row.GvkKind, row.GvkVersion, row.ScopeType, row.TenantId), nil)
	}

	return nil
}

// DeleteMappingByGvkAndScope deletes the tenant (or system) row for a GVK.
func DeleteMappingByGvkAndScope(ctx context.Context, db *gorm.DB, gvk api.Gvk, scope api.MappingScope) *api.Error {
	where := dbmodel.MappingSelectors{
		GvkGroup:   convutils.Deref(gvk.Group, ""),
		GvkKind:    gvk.Kind,
		GvkVersion: gvk.Version,
		ScopeType:  scope,
	}
	if scope == api.MAPPINGSCOPE_TENANT {
		where.TenantId = auth.TenantIdOf(ctx)
	}

	res := db.WithContext(ctx).Model(&dbmodel.MappingSelectors{}).Where(where).Delete(&dbmodel.MappingSelectors{})
	if res.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingSelectors).WithDetailsFromDb(res)
		logging.Errorf("failed to delete mapping: %s: %v", dbErr.Message, res.Error)
		return dbErr
	}

	if res.RowsAffected == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("mapping for %s/%s/%s (scope %s) is not found", convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version, scope), nil)
	}

	return nil
}
