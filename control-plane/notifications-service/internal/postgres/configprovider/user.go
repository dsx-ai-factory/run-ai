/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type userDao struct {
	db     *gorm.DB
	schema string
}

func newUserDao(db *gorm.DB, schema string) *userDao {
	return &userDao{db: db, schema: schema}
}

func (pud *userDao) FetchUserRoutes(listenerId string, tenantId string, userId string) ([]*models.RouteGlobalConfig, error) {
	routes := make([]*models.RouteGlobalConfig, 0)
	query := pud.db.
		Table(pud.routeUserConfigTable()).
		Where(tenantIdCol, tenantId).
		Where(userIdCol, userId)

	result := query.Find(&routes)

	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch routes")
		return routes, result.Error
	}

	filterRoutesByListenerId(listenerId, routes)
	return routes, nil
}

func (pud *userDao) FetchUserRoute(listenerId string, tenantId string, userId string, routeId string) (*models.RouteGlobalConfig, error) {
	routes := make([]*models.RouteGlobalConfig, 0)
	query := pud.db.
		Table(pud.routeUserConfigTable()).
		Where(models.RouteUserConfig{UserId: userId, TenantId: tenantId, RouteId: routeId, ListenerId: listenerId})

	result := query.Find(&routes)

	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch routes")
		return nil, result.Error
	}

	if len(routes) == 0 {
		return nil, nil
	}

	return routes[0], nil
}

func (pud *userDao) UpdateUserRoute(tenantId string, userId string, routeId string, listenerId string, route datatypes.JSON) error {
	return pud.db.Begin().Table(pud.routeUserConfigTable()).
		Where(models.RouteUserConfig{UserId: userId, TenantId: tenantId, ListenerId: listenerId, RouteId: routeId}).
		Assign(models.RouteUserConfig{RouteConfig: route}).
		FirstOrCreate(&models.RouteUserConfig{}).
		Commit().Error
}

func (pud *userDao) CreateUserRoute(tenantId string, userId string, routeId string, listenerId string, route datatypes.JSON) error {
	return pud.db.Table(pud.routeUserConfigTable()).
		Create(&models.RouteUserConfig{TenantId: tenantId, UserId: userId, RouteId: routeId, ListenerId: listenerId, RouteConfig: route}).
		Error
}

func (pud *userDao) DeleteUserRoute(tenantId string, userId string, routeId string) error {
	result := pud.db.Begin().Table(pud.routeUserConfigTable()).
		Where(models.RouteUserConfig{UserId: userId, TenantId: tenantId, RouteId: routeId}).
		Unscoped().
		Delete(&models.RouteUserConfig{}).
		Commit()

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return NewError(fmt.Sprintf("route %s not found for tenant %s and user %s", routeId, tenantId, userId), ErrCodeNotFound)
	}

	return nil
}

func (pud *userDao) DeleteAllTenantsUserRoutes(tenantId string) error {
	return pud.db.Begin().Table(pud.routeUserConfigTable()).
		Where(models.RouteUserConfig{TenantId: tenantId}).
		Unscoped().
		Delete(&models.RouteUserConfig{}).
		Commit().Error
}

func (pud *userDao) DeleteNotifyIdFromAllUserRoutes(tenantId string, notifyId string) error {
	return pud.db.
		Table(pud.routeUserConfigTable()).
		Where("tenant_id = ? AND (route_config->'notify_ids') @> jsonb_build_array(?)", tenantId, notifyId).
		UpdateColumn("route_config", gorm.Expr(`
		jsonb_set(
			route_config,
			'{notify_ids}',
			(
				SELECT COALESCE(jsonb_agg(elem), '[]'::jsonb)
				FROM jsonb_array_elements_text(route_config->'notify_ids') AS elem
				WHERE elem <> ?
			)
		)
	`, notifyId)).Error
}

func (pud *userDao) IsHealthy() (bool, error) {
	res := pud.db.Exec(fmt.Sprintf("select count (*) from %s", pud.routeUserConfigTable()))
	return res.Error == nil, res.Error
}

func (pud *userDao) routeUserConfigTable() string {
	return fmt.Sprintf("%s.%s", pud.schema, models.RouteUserConfigTable)
}
