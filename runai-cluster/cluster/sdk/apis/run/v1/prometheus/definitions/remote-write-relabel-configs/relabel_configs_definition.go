/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package remote_write_relabel_configs

import (
	"sync"

	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var once sync.Once

type RemoteWriteRelabelConfigs interface {
	Get() []monitoringv1.RelabelConfig
	GetRemoteWriteMatchingLabel() string
}

type RemoteWriteRelabelConfigsHandler interface {
	GetRemoteWriteRelabelConfigs(label string) []monitoringv1.RelabelConfig
}

type RemoteWriteRelabelConfigsHandlerImpl struct {
	remoteWriteRelabelConfigMap map[string]RemoteWriteRelabelConfigs
}

var handler RemoteWriteRelabelConfigsHandler

func GetHandler(internalSpec *runv1_internal.Internal) RemoteWriteRelabelConfigsHandler {
	once.Do(func() {
		handler = newRemoteWriteRelabelConfigsHandler(internalSpec)
	})
	return handler
}

func newRemoteWriteRelabelConfigsHandler(internalSpec *runv1_internal.Internal) RemoteWriteRelabelConfigsHandler {
	handler := &RemoteWriteRelabelConfigsHandlerImpl{
		remoteWriteRelabelConfigMap: make(map[string]RemoteWriteRelabelConfigs),
	}
	clusterRemoteWriteRelabelConfigs := newDefaultClusterRemoteWriteRelabelConfigs(internalSpec)
	applicationRemoteWriteRelabelConfigs := newDefaultApplicationRemoteWriteRelabelConfigs(internalSpec)
	handler.remoteWriteRelabelConfigMap[clusterRemoteWriteRelabelConfigs.GetRemoteWriteMatchingLabel()] = clusterRemoteWriteRelabelConfigs
	handler.remoteWriteRelabelConfigMap[applicationRemoteWriteRelabelConfigs.GetRemoteWriteMatchingLabel()] = applicationRemoteWriteRelabelConfigs
	return handler
}

func (h *RemoteWriteRelabelConfigsHandlerImpl) GetRemoteWriteRelabelConfigs(label string) []monitoringv1.RelabelConfig {
	conf, ok := h.remoteWriteRelabelConfigMap[label]
	if !ok {
		return nil
	}
	return conf.Get()
}
