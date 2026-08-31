/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package advanced

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener"
	api2 "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/listeners"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/sink"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/service"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	runai_http "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/http"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/kevent"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/prom"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/smtp"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	router "github.com/run-ai/runai/control-plane/notifications-service/internal/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/util"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
)

var DefaultUpdateEventsChannelSize = 10

type ConfigProviderDao interface {
	http_listener.ListenerDao
	http_listener.TenantDao
	http_listener.UserDao
	http_listener.SinksDao
	http_listener.NotificationKindsDao
}

type ConfigurationProvider struct {
	service.Config
	dao ConfigProviderDao
}

func NewConfigurationProvider(serviceConfig *service.Config, dao ConfigProviderDao) *ConfigurationProvider {
	return &ConfigurationProvider{
		Config: *serviceConfig,
		dao:    dao,
	}
}

func (pcp *ConfigurationProvider) GetListeningProviders() map[string]listen.ListeningConfigProvider {
	listenerRows, err := pcp.dao.FetchListeners()
	if err != nil {
		log.Err(err).Msgf("failed to fetch listeners")
		return nil
	}

	listenersMap := map[string]listen.ListeningConfigProvider{}
	for _, listenerRow := range listenerRows {
		addListener(listenerRow, listenersMap)
	}
	return listenersMap
}

func (pcp *ConfigurationProvider) GetListeningProvider(id string) listen.ListeningConfigProvider {
	listenerRow, err := pcp.dao.FetchListener(id)
	if err != nil {
		log.Err(err).Msgf("failed to fetch listener %s", id)
		return nil
	}
	if listenerRow == nil {
		return nil
	}

	listenersMap := map[string]listen.ListeningConfigProvider{}
	addListener(listenerRow, listenersMap)
	c := listenersMap[id]
	return c
}

func (pcp *ConfigurationProvider) GetSinkProvider(event *events.Event, sinkIds ...string) (router.SinkConfigProvider, error) {
	sinkList, err := pcp.dao.FetchSinks(sinkIds...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sinks from DAO: %w", err)
	}

	eventConfig, err := pcp.getNotificationConfig(event)
	if err != nil {
		log.Err(err)
	}

	sinkConfigs := util.MigrateSliceToMap(sinkList, pcp.getDbSinkToSinkConfigMappingFunc(eventConfig))
	return &sink.ConfigProvider{SinkConfig: sinkConfigs}, nil
}

func (pcp *ConfigurationProvider) GetRouteProviders(listenerId string, event *events.Event) []router.RouteConfigProvider {
	routeProviders := []router.RouteConfigProvider{}
	tenantId, ok := event.Properties["tenant_id"].(string)
	if !ok {
		log.Warn().Msgf("failed to read route configurations for event: %v", event.UID)
		return routeProviders
	}

	tenantRoutes, err := pcp.dao.FetchRoutes(listenerId, tenantId)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to fetch tenant routes, event: %v", event.UID)
		return routeProviders
	}

	if len(tenantRoutes) == 0 {
		return routeProviders
	}

	userRoutes, err := pcp.getUserRoutes(event, listenerId)
	if err != nil {
		log.Warn().Msgf("failed to fetch user routes for user %s, error: %v", event.Properties["user_id"], err)
		return routeProviders
	}

	routes := PatchRoutes(userRoutes, tenantRoutes[0])

	for _, routeJson := range routes {
		routeProviders = pcp.addRoute(routeJson, routeProviders)
	}

	return routeProviders
}

func (pcp *ConfigurationProvider) GetName() string {
	return "Advanced"
}

func (pcp *ConfigurationProvider) isHistoryEnabled() bool {
	historyConfig := pcp.GetHistoryConfig()
	return historyConfig != nil && historyConfig.Enabled
}

func createSinks(provider *ConfigurationProvider) {
	allSinks, _ := provider.SinkProvider().GetSinks()
	for id, config := range allSinks {
		if config.Email != nil {
			insertSinkConfig(id, sinks.EmailSinkType, *config.Email, provider)
		} else if config.Slack != nil {
			insertSinkConfig(id, sinks.SlackSinkType, *config.Slack, provider)
		} else if config.Postgres != nil {
			insertSinkConfig(id, sinks.PostgresSinkType, *config.Postgres, provider)
		} else if config.History != nil {
			insertSinkConfig(id, sinks.HistorySinkType, *config.History, provider)
		}
	}
}

func insertSinkConfig(sId, sType string, sConfig interface{}, provider *ConfigurationProvider) {
	parsed, err := toJson(sConfig)
	if err != nil {
		log.Err(err).Msgf("failed to parse sink config: %s", sId)
		return
	}

	err = provider.dao.UpsertSink(sId, sType, datatypes.JSON(parsed))
	if err != nil {
		log.Err(err).Msgf("failed to upsert sink: %s", sId)
		return
	}
}

func toJson(conf interface{}) ([]byte, error) {
	result := map[string]interface{}{}
	err := mapstructure.Decode(conf, &result)
	if err != nil {
		return nil, err
	}

	s, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func createListeners(provider *ConfigurationProvider) {
	for n, l := range provider.Config.GetListeningProviders() {
		lj, err := json.Marshal(getListeningConfig(l))
		if err != nil {
			log.Err(err).Msgf("failed to parse listener config")
		}
		err = provider.dao.CreateListener(n, l.GetListenerType(), lj)
		if err == nil {
			continue
		}

		if daoError, ok := err.(*configprovider.Error); !ok || daoError.Code != configprovider.ErrCodeDuplicate {
			log.Err(err).Msgf("failed to create listener")
			continue
		}

		if err = provider.dao.UpdateListener(n, l.GetListenerType(), lj); err != nil {
			log.Err(err).Msgf("failed to update listener")
		}
	}
}

func getListeningConfig(l listen.ListeningConfigProvider) interface{} {
	switch l.GetListenerType() {
	case prom.ListenerType:
		return l.GetListeningConfig().PrometheusListener
	case runai_http.ListenerType:
		return l.GetListeningConfig().HttpListenerConfig
	case kevent.ListenerType:
		return l.GetListeningConfig().KubeListener
	case smtp.ListenerType:
		return l.GetListeningConfig().SmtpListener
	}
	return nil
}

func (pcp *ConfigurationProvider) getNotificationConfig(event *events.Event) (*models.Notification, error) {
	if event == nil {
		return nil, nil
	}
	if event.Properties == nil {
		return nil, fmt.Errorf("no event properties found, cannot fetch notification config")
	}
	category, ok := event.Properties["category"].(string)
	if !ok || category == "" {
		return nil, fmt.Errorf("failed to read category from event properties, cannot fetch notification config")
	}
	typ, ok := event.Properties["type"].(string)
	if !ok || typ == "" {
		return nil, fmt.Errorf("failed to read type from event properties, cannot fetch notification config")
	}
	notificationConfig, err := pcp.dao.FetchNotifications("", category, typ)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch notification config for category %s and type %s", category, typ)
	}
	if len(notificationConfig) == 0 {
		return nil, errors.Wrapf(err, "no notification config found for category %s and type %s", category, typ)
	}
	return notificationConfig[0], nil
}

func (pcp *ConfigurationProvider) getDbSinkToSinkConfigMappingFunc(notification *models.Notification) func(int, *models.SinkConfig) (string, *sinks.SinkConfig, error) {
	templateFile := ""
	if notification != nil {
		templateFile = notification.Template
	}
	return func(_ int, dbSink *models.SinkConfig) (string, *sinks.SinkConfig, error) {
		v := viper.New()
		if templateFile != "" {
			v.Set("template_file", fmt.Sprintf("%v.tpl", templateFile))
		}
		if err := util.ReadJsonConfig(dbSink.Config.String(), v); err != nil {
			return "", nil, err
		}

		v.Set("history_enabled", pcp.isHistoryEnabled())

		conf, err := sink.NewSinkConfig(v, dbSink.SinkId, dbSink.Type)
		if err != nil {
			return "", nil, err
		}
		return dbSink.SinkId, conf, nil
	}
}

func (pcp *ConfigurationProvider) getUserRoutes(event *events.Event, listenerId string) ([]*models.RouteGlobalConfig, error) {
	userId, ok := event.Properties["user_id"].(string)
	if !ok {
		return []*models.RouteGlobalConfig{}, nil
	}

	tenantId, ok := event.Properties["tenant_id"].(string)
	if !ok {
		return []*models.RouteGlobalConfig{}, nil
	}

	userRoutes, err := pcp.dao.FetchUserRoutes(listenerId, tenantId, userId)
	if len(userRoutes) == 0 {
		log.Debug().Msgf("no user routes found for user %s in tenant %s", userId, tenantId)
	}
	return userRoutes, err
}

func addListener(listenerRow *models.ListenersConfig, listenersMap map[string]listen.ListeningConfigProvider) {
	v := viper.New()
	err := util.ReadJsonConfig(listenerRow.ListenerConfig.String(), v)
	if err != nil {
		log.Err(err).Msgf("failed to read listener configuration %s", listenerRow.ListenerId)
	} else {
		listenersMap[listenerRow.ListenerId] = listeners.NewListeningConfigProvider(listenerRow.ListenerType, v)
	}
}

func (pcp *ConfigurationProvider) addRoute(routeJson *models.RouteGlobalConfig, routeProviders []router.RouteConfigProvider) []router.RouteConfigProvider {
	v := viper.New()
	err := util.ReadJsonConfig(routeJson.RouteConfig.String(), v)
	routeName := "tenantId:" + routeJson.TenantId + ":routeId:" + routeJson.RouteId
	if err != nil {
		log.Err(err).Msg("failed to initialize viper to read route config: " + routeName)
		return routeProviders
	}

	return append(routeProviders, route.NewBasicRouteConfigurationProvider(v, routeName))
}

func addRoutes(router chi.Router, routers ...api2.Router) chi.Router {
	for _, aRouter := range routers {
		for _, route := range aRouter.Routes() {
			router.Method(route.Method, route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// PatchRoutes completes configurations, which are missing in userRoutes - from tenant route
func PatchRoutes(userRoutes []*models.RouteGlobalConfig, tenantRoute *models.RouteGlobalConfig) []*models.RouteGlobalConfig {
	if len(userRoutes) == 0 {
		return []*models.RouteGlobalConfig{tenantRoute}
	}

	for _, userRoute := range userRoutes {
		if err := patchRoute(userRoute, tenantRoute); err != nil {
			log.Err(err).Msgf("failed patching user route: %s", userRoute.RouteId)
		}
	}
	return userRoutes
}

func patchRoute(userRoute *models.RouteGlobalConfig, tenantRoute *models.RouteGlobalConfig) error {
	tenantRouteMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(tenantRoute.RouteConfig.String()), &tenantRouteMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tenant route: %w", err)
	}

	userRouteMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(userRoute.RouteConfig.String()), &userRouteMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user route: %w", err)
	}

	_, err = patchJSON(userRouteMap, tenantRouteMap)
	if err != nil {
		return fmt.Errorf("failed merging routes: %w", err)
	}

	mergedConfig, err := json.Marshal(userRouteMap)
	if err != nil {
		return fmt.Errorf("failed marshaling merged route: %w", err)
	}

	userRoute.RouteConfig = mergedConfig
	return nil
}

func patchJSON(target, patch any) (any, error) {
	if target == nil {
		return patch, nil
	}
	if patch == nil {
		return target, nil
	}

	if reflect.TypeOf(target).Kind() != reflect.TypeOf(patch).Kind() {
		return target, fmt.Errorf("patch and target type missmatch")
	}

	if reflect.TypeOf(target).Kind() != reflect.Map {
		return target, nil
	}

	t := target.(map[string]any)
	p := patch.(map[string]any)

	for k, pVal := range p {
		res, err := patchJSON(t[k], pVal)
		if err != nil {
			return nil, err
		}
		t[k] = res
	}
	return target, nil
}
