package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

func GetPermissions(ctx context.Context, tenantId uint, subjectId string, groups []string) ([]*models.Permission, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get subject %s permissions", subjectId), nil)

	permissions := make([]*models.Permission, 0)
	db := runaidb.DB().WithContext(ctx)
	err := db.
		Table("permissions").
		Joins("INNER JOIN access_rules ON permissions.role_id = access_rules.role_id").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.deleted_at IS NULL").
		Where(getSubjectClause(db, groups, subjectId)).
		Group("resource_type").
		Select("resource_type, bool_or(permissions.create) AS create, bool_or(permissions.read) AS read, bool_or(permissions.update) AS update, bool_or(permissions.delete) AS delete").
		Find(&permissions)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch user %s permissions. err: %+v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}

	return permissions, nil
}

func GetPermissionSets(ctx context.Context, tenantId uint, subjectId string, groups []string) ([]*models.RolePermission, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get subject %s permissions", subjectId), nil)

	rolePermissions := make([]*models.RolePermission, 0)
	db := runaidb.DB().WithContext(ctx)
	err := db.
		Table("role_permissions").
		Select("resource_type, "+
			"bool_or(role_permissions.create) AS create, "+
			"bool_or(role_permissions.read) AS read, "+
			"bool_or(role_permissions.update) AS update, "+
			"bool_or(role_permissions.delete) AS delete, "+
			"bool_or(role_permissions.sync) AS sync").
		Joins("INNER JOIN role_to_permission_set ON role_to_permission_set.permission_set_id = role_permissions.permission_set_id").
		Joins("INNER JOIN access_rules ON role_to_permission_set.role_id = access_rules.role_id").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.deleted_at IS NULL").
		Where(getSubjectClause(db, groups, subjectId)).
		Group("resource_type").
		Find(&rolePermissions)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch user %s permissions. err: %+v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}

	return rolePermissions, nil
}
