package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RoleBindingForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		namespace = runaiConfig.Namespace
	)

	roleBinding := &rbacv1.RoleBinding{}

	// Get the existing roleBinding if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, roleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	roleBinding.Name = name
	roleBinding.Namespace = namespace

	roleBinding.RoleRef = rbacv1.RoleRef{
		Kind:     "Role",
		APIGroup: "rbac.authorization.k8s.io",
		Name:     name,
	}
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      name,
			Namespace: namespace,
		},
	}

	return roleBinding, nil
}

func RoleBindingStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	roleBinding, err := RoleBindingForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if roleBinding.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("role binding [%s] is missing",
			roleBinding.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
