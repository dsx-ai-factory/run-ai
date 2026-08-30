package metrics

import (
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders/composite"
	metricStorebuilder "github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders/metrics/metric-store"
	metricStore "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/metric-store"
)

const (
	builderIdentifier = "metrics_builder"
)

type Builder struct {
	composite.CompositeBuilder
}

func NewBuilder(
	metricsStoreService metricStore.MetricStoreService,
) builders.Interface {
	var metricsBuilders []builders.Interface
	metricsBuilders = append(metricsBuilders,
		metricStorebuilder.NewBuilder(metricsStoreService))
	return &Builder{
		composite.NewCompositeBuilder(metricsBuilders, composite.Serial),
	}
}

func (h Builder) Name() string {
	return builderIdentifier
}
