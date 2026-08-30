package training_operator

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/training-operator/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	name = "training-operator"
)

var (
	// resourceFunctions is a slice of functions that return resources.
	// Update this variable when adding resources to an operand.
	resourceFunctions = []operands.ResourceFunction{
		resources.ServiceMonitorForRunaiConfig,
	}

	// resourceStatusFunctions is a slice of functions that know how to calculate
	// the status of resources.
	// Update this slice when a resource has a status to report.
	resourceStatusFunctions = []operands.ResourceStatusFunction{
		resources.ServiceMonitorStatus,
	}
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=role,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=rolebinding,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=service,verbs=get;create;update;watch;patch;delete;list

type operand struct{}

func (o *operand) Name() string {
	return name
}

// ResourcesForRunaiConfig returns a slice of Kubernetes resources
// that should be Created/Updated/Deleted by the controller to satisfy
// the needs of this operand.
func (o *operand) ResourcesForRunaiConfig(ctx context.Context, hybridClient kc.HybridClient,
	runaiConfig *runv1.RunaiConfig) ([]operands.ResourceOld, error) {
	return utils.ResourcesForRunaiConfig(ctx, hybridClient, runaiConfig, resourceFunctions, o.Name())
}

func (o *operand) Status(ctx context.Context, hybridClient kc.HybridClient,
	runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	return utils.StatusOld(ctx, hybridClient, runaiConfig, resourceStatusFunctions)
}

func Operand() operands.OperandOld {
	return &operand{}
}
