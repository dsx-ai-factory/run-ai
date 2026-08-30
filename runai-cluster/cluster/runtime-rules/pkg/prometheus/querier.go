package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

const (
	gpuIdleTimeMetric = "runai_gpu_idle_seconds_per_workload"
)

type PromQuerier struct {
	client *PromClient
}

func NewPromQuerier(promServiceUrl string) (*PromQuerier, error) {
	client, err := NewPromClient(promServiceUrl)
	if err != nil {
		return nil, err
	}
	return &PromQuerier{client: client}, nil
}

func (pq *PromQuerier) GetWorkloadIdleDuration(ctx context.Context, workloadName string, workloadID string) (*time.Duration, error) {
	query := fmt.Sprintf("%v{workload_name=\"%v\",workload_id=\"%v\"}", gpuIdleTimeMetric,
		workloadName, workloadID)
	result, err := pq.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	idleDuration, err := extractDurationFromResult(result)
	if err != nil {
		return nil, err
	}
	return idleDuration, nil
}

func extractDurationFromResult(result model.Value) (*time.Duration, error) {
	if result.Type() != model.ValVector {
		return nil, fmt.Errorf("failed to decode query result - not a vector type")
	}

	vectorResult := result.(model.Vector)
	if vectorResult.Len() == 0 {
		return nil, nil
	}

	duration := time.Duration(vectorResult[0].Value) * time.Second
	return &duration, nil
}
