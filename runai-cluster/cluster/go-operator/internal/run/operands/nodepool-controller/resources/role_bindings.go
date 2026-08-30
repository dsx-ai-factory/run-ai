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

func (r *RoleBindingResource) OperandName() string {
	return r.operandName
}

func (r *RoleBindingResource) Object(ctx context.Context) (client.Object, error) {
	namespace := r.runaiConfig.Namespace

	roleBinding := &rbacv1.RoleBinding{}

	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: namespace,
	}, roleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	roleBinding.ObjectMeta.Name = DeploymentName
	roleBinding.ObjectMeta.Namespace = namespace

	roleBinding.RoleRef = rbacv1.RoleRef{
		Kind:     "Role",
		APIGroup: rbacv1.SchemeGroupVersion.Group,
		Name:     DeploymentName,
	}

	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      DeploymentName,
			Namespace: "runai",
		},
	}

	return roleBinding, nil
}

func (r *RoleBindingResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("role binding [%s] is missing", resource.GetName())
	}

	return nil
}

func (r *RoleBindingResource) Enabled() bool {
	return true
}
