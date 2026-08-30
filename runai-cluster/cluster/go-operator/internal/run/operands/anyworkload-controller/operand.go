package anyworkload_controller

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/anyworkload-controller/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
)

const (
	// Name specifies the name of this operand for logging purposes.
	name = operands.RunaiAnyworkloadControllerName
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=secrets,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingadmissionpolicies,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingadmissionpolicybindings,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.AnyworkloadController.Enabled
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	operandResources := []operands.Resource{
		resources.NewDeploymentResource(name, k8sReader, runaiConfig),
		resources.NewServiceAccountResource(name, k8sReader, runaiConfig),
		resources.NewServiceResource(name, k8sReader, runaiConfig),
		resources.NewServiceMonitorResource(name, k8sReader, runaiConfig),
		resources.NewSecretResource(name, k8sReader, runaiConfig),
	}
	operandResources = append(operandResources, admissionpolicy.NewResources(
		name,
		k8sReader,
		resources.ValidatingAdmissionPolicyDefinition(runaiConfig),
	)...)

	return &operand{
		runaiConfig: runaiConfig,
		resources:   operandResources,
	}
}
