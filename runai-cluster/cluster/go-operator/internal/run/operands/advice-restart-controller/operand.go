package advice_restart_controller

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/advice-restart-controller/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	Name = operands.RunaiAdviceRestartControllerName
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingadmissionpolicies,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingadmissionpolicybindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	k8sReader   kc.HybridClient
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return Name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	status, err := utils.Status(ctx, o.resources)
	if !status.Ready {
		return status, err
	}

	discoveryClient, dcErr := o.k8sReader.NewDiscoveryClient()
	if dcErr != nil {
		return operands.NotReadyStatus(fmt.Sprintf("failed to create discovery client: %s", dcErr)), dcErr
	}

	apiList, listErr := discoveryClient.ServerResourcesForGroupVersion("heimdall.nvidia.com/v1alpha1")
	if listErr != nil {
		if kerrors.IsNotFound(listErr) {
			return operands.NotReadyStatus("Advice CRD (heimdall.nvidia.com/v1alpha1) is not installed"), nil
		}
		return operands.NotReadyStatus(fmt.Sprintf("failed to check Advice CRD: %s", listErr)), listErr
	}

	for _, api := range apiList.APIResources {
		if api.Kind == "Advice" {
			return status, err
		}
	}
	return operands.NotReadyStatus("Advice CRD (heimdall.nvidia.com/v1alpha1) is not installed"), nil
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.AdviceRestartController.Enabled
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	operandResources := []operands.Resource{
		resources.NewDeploymentResource(Name, k8sReader, runaiConfig),
		resources.NewServiceAccountResource(Name, k8sReader, runaiConfig),
		resources.NewServiceResource(Name, k8sReader, runaiConfig),
	}
	operandResources = append(operandResources, admissionpolicy.NewResources(
		Name,
		k8sReader,
		resources.ValidatingAdmissionPolicyDefinition(),
	)...)
	operandResources = append(operandResources, resources.NewServiceMonitorResource(Name, k8sReader, runaiConfig))

	return &operand{
		runaiConfig: runaiConfig,
		k8sReader:   k8sReader,
		resources:   operandResources,
	}
}
