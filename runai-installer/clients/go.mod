module github.com/dsx-ai-factory/run-ai/runai-installer/clients

go 1.27.2

require (
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client v1.0.11
	github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	golang.org/x/oauth2 v0.32.0
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
)

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../../control-plane/authorization/pkg/client

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../control-plane/cluster-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client => ../../control-plane/org-unit-service/pkg/client

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.7 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.29 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
)
