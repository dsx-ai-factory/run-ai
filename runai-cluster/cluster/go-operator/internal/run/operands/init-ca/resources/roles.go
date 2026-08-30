package resources

import (
	"context"
	"fmt"

	certmgr "github.com/cert-manager/cert-manager/pkg/apis/certmanager"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RoleForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		namespace = runaiConfig.GetNamespace()
		name      = "init-ca"
	)

	role := &rbacv1.Role{}

	// Get the existing role if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, role)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	role.ObjectMeta.Namespace = namespace
	role.ObjectMeta.Name = name
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "patch", "delete", "create"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list", "delete", "deletecollection"},
		},
		{
			APIGroups: []string{v1.GroupName},
			Resources: []string{"deployments"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{certmgr.GroupName},
			Resources: []string{"certificates"},
			Verbs:     []string{"get", "list", "create", "delete"},
		},
	}
	return role, nil
}

func RoleStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	role, err := RoleForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if role.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("role [%s] is missing",
			role.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
