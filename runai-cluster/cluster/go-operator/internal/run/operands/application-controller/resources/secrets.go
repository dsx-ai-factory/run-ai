package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/secrets"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretForRunaiConfig builds the whisper-application-controller-secret from the
// installation-configuration, mirroring the cluster-sync operand. It exposes the per-cluster agent
// client credentials (name/secret) and the cluster UUID so the application controller can obtain a
// control-plane bearer token for authenticated chart downloads. The secret is mounted into the pod
// by MountToDeployment (see deployments.go).
func SecretForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	ic, err := secrets.GetConfigurationSecret(ctx, k8sReader)
	if err != nil {
		return nil, err
	}
	deploymentSecret, err := secrets.GetDeploymentSecret(ctx, k8sReader, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment secret for application-controller: %w", err)
	}
	vault, err := secrets.NewVault(name, ic, deploymentSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault for application-controller: %w", err)
	}
	vault.AddFromSecret(secrets.ClientName, "CLIENT_NAME")
	vault.AddFromSecret(secrets.ClientSecret, "CLIENT_SECRET")
	vault.AddFromSecret(secrets.ClusterID, "CLUSTER_ID")
	return vault.Generate()
}

func SecretStatus(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := SecretForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}
	if resource.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("secret [%s] is missing", resource.GetName())), nil
	}
	return operands.ReadyStatus(), nil
}
