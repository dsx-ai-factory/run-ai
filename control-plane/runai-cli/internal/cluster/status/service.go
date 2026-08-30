package status

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	openapi "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"
	"go.uber.org/zap"
)

const (
	FailedToGetStatusTemplate = "failed to validate workload status: %s"
	workloadIsNilMsg          = "workload is nil"
	statusMismatchTemplate    = "workload '%s' status is not '%s'"

	ResponseStatusOk = "\"OK\""
)

type ClusterStatusClient interface {
	GetClusterStatus(ctx context.Context) ([]byte, error)
	GetWorkloadStatus(ctx context.Context, workload *apis.Workload) (*openapi.WorkloadStatus, error)
}

type ClusterService struct {
	statusClient ClusterStatusClient
}

func NewClusterService(statusClient ClusterStatusClient) *ClusterService {
	return &ClusterService{statusClient: statusClient}
}

func (s *ClusterService) GetWorkloadStatus(ctx context.Context, workload *apis.Workload) (*openapi.WorkloadStatus, error) {
	if workload == nil {
		return nil, errors.New(workloadIsNilMsg)
	}

	return s.statusClient.GetWorkloadStatus(ctx, workload)
}

func (s *ClusterService) ValidateWorkloadStatus(ctx context.Context, workload *apis.Workload, expectedStatus string) error {
	workloadStatus, err := s.GetWorkloadStatus(ctx, workload)
	if err != nil {
		return err
	}
	if workloadStatus.GetStatus() != expectedStatus {
		return fmt.Errorf(statusMismatchTemplate, workload.Name, expectedStatus)
	}
	return nil
}

func (s *ClusterService) GetCluster(ctx context.Context) (*Metadata, error) {
	payload, err := s.statusClient.GetClusterStatus(ctx)
	if err != nil {
		return nil, err
	}
	metadata, err := getClusterMetadata(payload)
	if err != nil {
		config.GetInstance().GetLogger().Error("couldn't get cluster metadata, some feature might be disabled", zap.Error(err))
	}
	return metadata, nil
}

func getClusterMetadata(payload []byte) (*Metadata, error) {
	metadata := &Metadata{}
	if strings.Contains(string(payload), ResponseStatusOk) {
		config.GetInstance().GetLogger().Debug("older terminal does not support cluster features")
		return metadata, nil
	}
	err := json.Unmarshal(payload, metadata)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}
