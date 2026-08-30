package resources

import (
	"context"
	"fmt"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleBindingResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewRoleBindingResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *RoleBindingResource {
	return &RoleBindingResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (rb *RoleBindingResource) OperandName() string {
	return rb.operandName
}

func (rb *RoleBindingResource) Object(ctx context.Context) (client.Object, error) {
	namespace := rb.runaiConfig.Namespace

	roleBinding := &rbacv1.RoleBinding{}

	err := rb.k8sReader.Get(ctx, types.NamespacedName{
		Name:      rb.operandName,
		Namespace: namespace,
	}, roleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	roleBinding.Name = rb.operandName
	roleBinding.Namespace = namespace

	roleBinding.RoleRef = rbacv1.RoleRef{
		Kind:     "Role",
		APIGroup: "rbac.authorization.k8s.io",
		Name:     rb.operandName,
	}
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      rb.operandName,
			Namespace: namespace,
		},
	}

	return roleBinding, nil
}

func (rb *RoleBindingResource) Status(ctx context.Context) error {
	resource, err := rb.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("role binding [%s] is missing", resource.GetName())
	}

	return nil
}

func (rb *RoleBindingResource) Enabled() bool {
	return true
}
