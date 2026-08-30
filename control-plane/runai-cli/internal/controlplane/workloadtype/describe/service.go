package describe

import (
	"context"
	"fmt"
	"strings"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type Lister interface {
	List(ctx context.Context, filterBy []string, externalTypesOnly *bool) (*openapi.ListWorkloadTypes200Response, error)
}

type Getter interface {
	Get(ctx context.Context, workloadTypeId string) (*openapi.WorkloadTypeConfig, error)
}

type Service struct {
	lister Lister
	getter Getter
}

func NewService(lister Lister, getter Getter) *Service {
	return &Service{
		lister: lister,
		getter: getter,
	}
}

func (s *Service) Describe(ctx context.Context, name string) (*openapi.WorkloadTypeConfig, error) {
	match, err := s.findByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if match.Id == nil {
		return nil, fmt.Errorf("workload type '%s' has no ID", name)
	}

	result, err := s.getter.Get(ctx, *match.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload type '%s': %w", name, err)
	}

	return result, nil
}

func (s *Service) findByName(ctx context.Context, name string) (*openapi.WorkloadTypeConfig, error) {
	response, err := s.lister.List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workload types: %w", err)
	}

	for i := range response.Types {
		if strings.EqualFold(response.Types[i].Name, name) {
			return &response.Types[i], nil
		}
	}

	return nil, fmt.Errorf("workload type '%s' not found", name)
}
