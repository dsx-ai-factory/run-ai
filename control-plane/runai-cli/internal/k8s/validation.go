package k8s

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/k8s/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/k8s/validation"
)

var (
	k8sClientHolder validation.KubeClienter
)

func SetClient(k8sClientSet validation.KubeClienter) {
	k8sClientHolder = k8sClientSet
}

func ValidateClusterIdentity(ctx context.Context) (bool, error) {
	k8sClient, err := getClient()
	if err != nil {
		return false, err
	}

	cv := validation.NewClusterValidator(k8sClient)
	return cv.ValidateClusterIdentity(ctx)
}

// GetServerUrl uses the K8S client to fetch the k8s server URL.
func GetServerUrl(ctx context.Context) (string, error) {
	k8sClient, err := getClient()
	if err != nil {
		return "", err
	}
	return k8sClient.GetServerUrl(ctx), nil
}

func getClient() (validation.KubeClienter, error) {
	if k8sClientHolder == nil {
		var err error
		k8sClientHolder, err = client.NewClient()
		if err != nil {
			return nil, err
		}
	}
	return k8sClientHolder, nil
}
