package queue

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/run-ai/runai/runai-common-packages/tracing"
)

func QueueMiddleware(redisClient *redis.Client, tp tracing.TracerProvider) error {
	switch t := tp.(type) {
	case *tracing.OtelTracerProvider:
		opts := redisotel.WithTracerProvider(t.Provider)
		return redisotel.InstrumentTracing(redisClient, opts)
	default:
		return nil
	}
}
