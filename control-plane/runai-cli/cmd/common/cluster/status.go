package cluster

import (
	"context"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	openapi "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"
)

const (
	WorkloadRunningStatus           = "Running"
	InferenceWorkloadRunningStatus  = "Ready"
	DefaultApiStatusDurationTimeout = "3s"
)

var clusterApiClient status.ClusterStatusClient

func SetClient(clusterClientSet status.ClusterStatusClient) {
	clusterApiClient = clusterClientSet
}

func GetStatus() (*status.Metadata, error) {
	timeout, err := time.ParseDuration(instance.GetConfigValueWithDefault(config.ApiStatusTimeoutDuration, DefaultApiStatusDurationTimeout))
	if err != nil {
		return nil, err
	}

	txn := instance.GetTxn()
	token := instance.GetToken()
	instance.GetLogger().Debug("checking cluster status", zap.String("cluster_domain", instance.GetConfigValue(config.ClusterDomain)))
	return GetClusterApiStatus(txn.GetCtx(), instance.GetConfigValue(config.ClusterDomain), timeout, token)
}

func GetClusterApiStatus(ctx context.Context, clusterUrl string, timeout time.Duration, token string) (*status.Metadata, error) {
	var (
		metadata *status.Metadata
	)
	clusterUrlStr, err := url.Parse(clusterUrl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	metadata, err = GetClusterStatus(ctx, clusterUrlStr.String(), token)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func GetClusterApiUrl() (*url.URL, error) {
	clusterUrlStr, err := url.Parse(instance.GetConfigValue(config.ClusterDomain))
	if err != nil {
		return nil, err
	}
	clusterUrlStr = clusterUrlStr.JoinPath(consts.ClusterApi)
	return clusterUrlStr, nil
}

func GetClusterStatus(ctx context.Context, clusterApiURL, token string) (*status.Metadata, error) {
	client, err := getClient(clusterApiURL, token)
	if err != nil {
		return nil, err
	}

	workloadStatusService := status.NewClusterService(client)
	return workloadStatusService.GetCluster(ctx)
}

func GetWorkloadStatus(ctx context.Context, workload *apis.Workload, clusterApiURL, token string) (*openapi.WorkloadStatus, error) {
	client, err := getClient(clusterApiURL, token)
	if err != nil {
		return nil, err
	}
	workloadStatusService := status.NewClusterService(client)
	return workloadStatusService.GetWorkloadStatus(ctx, workload)
}

func ValidateWorkloadStatus(ctx context.Context, workload *apis.Workload, clusterApiURL, token, expectedStatus string) error {
	client, err := getClient(clusterApiURL, token)
	if err != nil {
		return err
	}

	workloadStatusService := status.NewClusterService(client)
	return workloadStatusService.ValidateWorkloadStatus(ctx, workload, expectedStatus)
}

func getClient(clusterApiURL string, token string) (status.ClusterStatusClient, error) {
	if clusterApiClient == nil {
		return status.NewClusterApiClient(clusterApiURL, token)
	}
	return clusterApiClient, nil
}
