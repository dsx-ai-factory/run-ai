/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	tenantIdCol    = "tenant_id"
	userIdCol      = "user_id"
	routeIdCol     = "route_id"
	routeConfigCol = "route_config"
)

type tenantDao struct {
	db     *gorm.DB
	schema string
}

func newTenantDao(db *gorm.DB, schema string) *tenantDao {
	return &tenantDao{db: db, schema: schema}
}

func (ptd *tenantDao) FetchRoutes(listenerId string, tenantId string) ([]*models.RouteGlobalConfig, error) {
	routes := make([]*models.RouteGlobalConfig, 0)
	query := ptd.db.
		Table(ptd.routeGlobalConfigTable()).
		Where(tenantIdCol, tenantId)

	result := query.Find(&routes)

	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch routes")
		return routes, result.Error
	}

	filterRoutesByListenerId(listenerId, routes)
	return routes, nil
}

func (ptd *tenantDao) FetchAllTenantsWithRoutes() ([]string, error) {
	ids := make([]string, 0)
	query := ptd.db.
		Table(ptd.routeGlobalConfigTable()).
		Select(fmt.Sprintf("distinct(%s)", tenantIdCol))

	result := query.Find(&ids)
	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch distinct routes")
		return nil, result.Error
	}

	return ids, nil
}

func (ptd *tenantDao) FetchRoute(listenerId string, tenantId string, routeId string, defaultRoute ...*models.RouteGlobalConfig) (*models.RouteGlobalConfig, error) {
	routes := make([]*models.RouteGlobalConfig, 0)
	query := ptd.db.
		Table(ptd.routeGlobalConfigTable()).
		Where(models.RouteGlobalConfig{TenantId: tenantId, RouteId: routeId, ListenerId: listenerId})
	result := query.Find(&routes)

	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch routes")
		return nil, result.Error
	}

	return getFirstOrDefault(routes, defaultRoute...), nil
}

func (ptd *tenantDao) CreateRoute(tenantId string, routeId string, listenerId string, route datatypes.JSON) error {
	return ptd.db.Table(ptd.routeGlobalConfigTable()).
		Create(&models.RouteGlobalConfig{
			TenantId:    tenantId,
			RouteId:     routeId,
			ListenerId:  listenerId,
			RouteConfig: route,
		}).Error
}

func (ptd *tenantDao) UpdateRoute(tenantId string, routeId string, listenerId string, route datatypes.JSON) error {
	return ptd.db.Begin().Table(ptd.routeGlobalConfigTable()).
		Where(models.RouteGlobalConfig{TenantId: tenantId, ListenerId: listenerId, RouteId: routeId}).
		Assign(models.RouteGlobalConfig{RouteConfig: route}).
		FirstOrCreate(&models.RouteGlobalConfig{}).
		Commit().Error
}

func (ptd *tenantDao) DeleteRoute(tenantId string, routeId string) error {
	result := ptd.db.Begin().Table(ptd.routeGlobalConfigTable()).
		Where(models.RouteGlobalConfig{TenantId: tenantId, RouteId: routeId}).
		Unscoped().
		Delete(&models.RouteGlobalConfig{}).
		Commit()

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return NewError(fmt.Sprintf("route not found for tenant: %s and route id: %s", tenantId, routeId), ErrCodeNotFound)
	}

	return nil
}

func (ptd *tenantDao) IsHealthy() (bool, error) {
	res := ptd.db.Exec(fmt.Sprintf("select count (*) from %s", ptd.routeGlobalConfigTable()))
	return res.Error == nil, res.Error
}

func (ptd *tenantDao) routeGlobalConfigTable() string {
	return fmt.Sprintf("%s.%s", ptd.schema, models.RouteGlobalConfigTable)
}

func getFirstOrDefault[T any](items []*T, defaultValue ...*T) *T {
	if len(items) != 0 {
		return items[0]
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}
