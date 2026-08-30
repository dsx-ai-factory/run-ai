package remote_write_relabel_configs

import (
	"sync"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
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
