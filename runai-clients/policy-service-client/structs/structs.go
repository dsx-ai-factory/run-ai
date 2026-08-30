package structs

import (
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

type PolicyWorkloadTypeConfig struct {
	CategoryName   *string
	PriorityName   *string
	Preemptibility *openapi.Preemptibility
}
