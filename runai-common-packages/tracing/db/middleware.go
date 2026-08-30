package db

import (
	"github.com/run-ai/runai/runai-common-packages/tracing"
	"gorm.io/gorm"
	gormotel "gorm.io/plugin/opentelemetry/tracing"
)

func DBMiddleware(tp tracing.TracerProvider) (gorm.Plugin, bool) {
	switch t := tp.(type) {
	case *tracing.OtelTracerProvider:
		return gormotel.NewPlugin(
			gormotel.WithTracerProvider(t.Provider),
			gormotel.WithoutMetrics(),
		), true
	default:
		return nil, false
	}
}
