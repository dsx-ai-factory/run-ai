/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package listeners

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen"
	http_listener "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/http"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/kevent"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/prom"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/smtp"
	"github.com/spf13/viper"
)

type ListeningConfigProvider struct {
	Config listen.ListeningConfig
}

func NewListeningConfigProvider(typ string, v *viper.Viper) listen.ListeningConfigProvider {
	config := listen.ListeningConfig{}
	switch typ {
	case kevent.ListenerType:
		config.KubeListener = kevent.NewKubeEventListenerConfig(v)
	case prom.ListenerType:
		config.PrometheusListener = prom.NewPrometheusListenerConfig(v)
	case http_listener.ListenerType:
		config.HttpListenerConfig = http_listener.NewHttpListenerConfig(v)
	case smtp.ListenerType:
		config.SmtpListener = smtp.NewSmtpListenerConfig(v)
	}

	return &ListeningConfigProvider{Config: config}
}

func (scp *ListeningConfigProvider) GetListeningConfig() *listen.ListeningConfig {
	return &scp.Config
}

func (scp *ListeningConfigProvider) GetListenerType() string {
	if !config.IsConfigFieldEmpty(scp.Config.KubeListener) {
		return kevent.ListenerType
	}
	if !config.IsConfigFieldEmpty(scp.Config.PrometheusListener) {
		return prom.ListenerType
	}
	if !config.IsConfigFieldEmpty(scp.Config.HttpListenerConfig) {
		return http_listener.ListenerType
	}
	if !config.IsConfigFieldEmpty(scp.Config.SmtpListener) {
		return smtp.ListenerType
	}
	return ""
}
