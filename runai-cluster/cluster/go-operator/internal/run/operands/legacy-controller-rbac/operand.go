// Package legacy_controller_rbac removes the Helm-managed RBAC left behind
// when one of the legacy workload controllers is disabled. The allowlist is
// intentionally explicit so this cleanup cannot expand to similarly named
// resources or credentials.
package legacy_controller_rbac

import (
	"context"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;delete

const name = "legacy-controller-rbac-cleanup"

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

// Enabled is always false so every selected resource is routed to deletion.
func (o *operand) Enabled() bool { return false }

func Operand(_ kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	namespace := runaiConfig.GetNamespace()
	resources := []operands.Resource{}

	if explicitlyDisabled(runaiConfig.Spec.WorkloadController.Enabled) {
		resources = append(resources,
			clusterRole("workload-controller-rw"),
			clusterRole("workload-controller-ro"),
			clusterRoleBinding("workload-controller-ro"),
			role(namespace, "workload-controller-migration"),
			roleBinding(namespace, "workload-controller-migration"),
		)
	}

	if explicitlyDisabled(runaiConfig.Spec.InferenceWorkloadController.Enabled) {
		resources = append(resources,
			clusterRole("runai-inference-workload-controller-project"),
			clusterRole("runai-inference-workload-controller"),
			clusterRoleBinding("runai-inference-workload-controller"),
		)
	}

	if explicitlyDisabled(runaiConfig.Spec.JobController.Enabled) {
		resources = append(resources,
			clusterRole("runai-job-controller"),
			clusterRoleBinding("runai-job-controller"),
			clusterRole("runai-job-controller-project"),
		)
	}

	return &operand{resources: resources}
}

func explicitlyDisabled(enabled *bool) bool {
	return enabled != nil && !*enabled
}

func clusterRole(resourceName string) operands.Resource {
	return deleteOnly(func() client.Object {
		return &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
	})
}

func clusterRoleBinding(resourceName string) operands.Resource {
	return deleteOnly(func() client.Object {
		return &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
	})
}

func role(namespace, resourceName string) operands.Resource {
	return deleteOnly(func() client.Object {
		return &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: namespace}}
	})
}

func roleBinding(namespace, resourceName string) operands.Resource {
	return deleteOnly(func() client.Object {
		return &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: namespace}}
	})
}

func deleteOnly(build func() client.Object) operands.Resource {
	return operands.NewDeleteOnlyResource(name, build)
}
