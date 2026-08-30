package tracing

const (
	OtelTracerName  = "otel"
	DummyTracerName = "dummy"
)

type TracerProviderFactory struct {
}

func (tpf *TracerProviderFactory) CreateTracerProvider(config *Config) (TracerProvider, error) {
	if !config.Enabled {
		initOtelPropagation()
		return &DummyTracerProvider{}, nil
	}

	switch config.UnderlineTracer {
	case OtelTracerName:
		return NewOtelTracerProvider(config)
	}

	return &DummyTracerProvider{}, nil
}
