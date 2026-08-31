/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prom

import (
	"context"
	"encoding/json"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/spf13/viper"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/alertmanager/api/v2/restapi"
	"github.com/prometheus/alertmanager/api/v2/restapi/operations"
	"github.com/prometheus/alertmanager/api/v2/restapi/operations/alert"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"golang.org/x/exp/maps"
)

const (
	ListenerType = "prometheus_alerts"
)

type PrometheusListenerConfig struct {
	Port int `json:"port"`
}

func NewPrometheusListenerConfig(v *viper.Viper) *PrometheusListenerConfig {
	prometheusAlertsDefaults(v)
	cfg := PrometheusListenerConfig{}
	if err := v.Unmarshal(&cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed unmarshalling configuration of route %q", v.GetString("Id"))
	}

	return &cfg
}

func prometheusAlertsDefaults(v *viper.Viper) {
	config.SetDefault(v, "port", 9093)
}

func NewPrometheusListener(id string, eventCh chan *events.Event, config *PrometheusListenerConfig) *PrometheusListener {
	api := getAlertmanagerApi()
	server := restapi.NewServer(api)
	server.Port = 9093

	return &PrometheusListener{
		alertmanagerApi:    api,
		alertmanagerServer: server,
		eventCh:            eventCh,
	}
}

type PrometheusListener struct {
	id                 string
	alertmanagerServer *restapi.Server
	alertmanagerApi    *operations.AlertmanagerAPI
	eventCh            chan *events.Event
}

func (pl *PrometheusListener) Id() string {
	return pl.id
}

func (pl *PrometheusListener) Listen(ctx context.Context) {
	pl.alertmanagerApi.AlertPostAlertsHandler = alert.PostAlertsHandlerFunc(func(params alert.PostAlertsParams) middleware.Responder {
		for _, alert := range params.Alerts {
			b, _ := json.Marshal(params)
			log.Debug().Str("component", "prometheus_alerts_listener").Msgf("received postable alert %s", b)
			event := alertToRouteEvent(alert)
			pl.eventCh <- event
		}
		return alert.NewPostAlertsOK()
	})

	pl.alertmanagerServer.SetAPI(pl.alertmanagerApi)
	err := pl.alertmanagerServer.Serve()
	if err != nil {
		return
	}
	log.Info().Str("component", "prometheus_alerts_listener").Msgf("Serving alertmanager at http://[[::]:%v]", pl.alertmanagerServer.Port)
	<-ctx.Done()
	pl.alertmanagerServer.Shutdown()
}

func getAlertmanagerApi() *operations.AlertmanagerAPI {
	spec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to analyze alertmanager swagger file")
	}
	api := operations.NewAlertmanagerAPI(spec)
	api.Logger = func(s string, i ...interface{}) {
		log.Debug().Str("component", "prometheus_alerts_listener").Msgf(s, i...)
	}
	return api
}

func alertToRouteEvent(postableAlerts *models.PostableAlert) *events.Event {
	sprops := map[string]string{}
	maps.Copy(sprops, postableAlerts.Annotations)
	maps.Copy(sprops, postableAlerts.Labels)

	props := map[string]interface{}{}
	for key, value := range sprops {
		props[key] = value
	}

	return &events.Event{
		RawEvent:   nil,
		UID:        "",
		Properties: props,
	}
}
