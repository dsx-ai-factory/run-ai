package status_updater

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/status-updater/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	Name = "status-updater"
)

//+kubebuilder:rbac:namespace="",groups="",resources=pods,verbs=get;watch;patch;list
//+kubebuilder:rbac:namespace=runai,groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return Name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.StatusUpdater.Enabled
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	return &operand{
		runaiConfig: runaiConfig,
		resources: []operands.Resource{
			resources.NewDeploymentResource(Name, k8sReader, runaiConfig),
			resources.NewServiceAccountResource(Name, k8sReader, runaiConfig),
			resources.NewVPAResource(Name, k8sReader, runaiConfig),
		},
	}
}
