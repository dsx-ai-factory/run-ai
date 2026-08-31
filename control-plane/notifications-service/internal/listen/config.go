/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package listen

import (
	runai_http "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/http"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/kevent"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/prom"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/smtp"
)

type ListeningConfigProvider interface {
	GetListeningConfig() *ListeningConfig
	GetListenerType() string
}

type ListeningConfig struct {
	KubeListener       *kevent.KubeEventListenerConfig `mapstructure:"kube_events" json:"kube_events"`
	HttpListenerConfig *runai_http.HttpListenerConfig  `mapstructure:"http_listener" json:"http_listener"`
	PrometheusListener *prom.PrometheusListenerConfig  `mapstructure:"prometheus_alerts" json:"prometheus_alerts"`
	SmtpListener       *smtp.SmtpListenerConfig        `mapstructure:"smtp_listener" json:"smtp_listener"`
}
