/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func newMetricsPusher(promRegistry *prometheus.Registry, serviceName string, pushGatewayUrl string) *metricsPusher {
	return &metricsPusher{
		pusher:          push.New(pushGatewayUrl, serviceName).Gatherer(promRegistry),
		metricsRegistry: newMetricsRegistry(promRegistry),
	}
}

type metricsPusher struct {
	pusher          *push.Pusher
	metricsRegistry MetricsRegistrar
}

func (mp *metricsPusher) GetRegistry() MetricsRegistrar {
	return mp.metricsRegistry
}

func (mp *metricsPusher) Push() error {
	if err := mp.pusher.Push(); err != nil {
		return fmt.Errorf("failed to push metrics, err: %w", err)
	}
	return nil
}
