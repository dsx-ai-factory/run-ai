package api

import openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

type NodeTelemetriesByType map[openapi.NodeTelemetryType]*openapi.TelemetryResponse
