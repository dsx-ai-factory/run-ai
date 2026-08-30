package validation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

type KubeClienter interface {
	CheckConnectivity(ctx context.Context) error
	GetServerUrl(ctx context.Context) string
}

type ClusterValidator struct {
	kubeClient KubeClienter
}

func NewClusterValidator(kubeClient KubeClienter) *ClusterValidator {
	return &ClusterValidator{
		kubeClient: kubeClient,
	}
}

func (v *ClusterValidator) ValidateConnectivity(ctx context.Context) error {
	return v.kubeClient.CheckConnectivity(ctx)
}

func (v *ClusterValidator) ValidateClusterIdentity(ctx context.Context) (bool, error) {
	parsedClusterDomain, err := url.Parse(config.GetInstance().GetConfigValue(config.ClusterDomain))
	if err != nil {
		return false, fmt.Errorf("failed to parse cluster domain: %w", err)
	}
	parsedServerUrl, err := url.Parse(v.kubeClient.GetServerUrl(ctx))
	if err != nil {
		return false, fmt.Errorf("failed to parse server url: %w", err)
	}
	return parsedClusterDomain.Hostname() == parsedServerUrl.Hostname(), nil
}
