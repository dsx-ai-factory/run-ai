package resources

import (
	"context"

	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VPAResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewVPAResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *VPAResource {
	return &VPAResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (v *VPAResource) OperandName() string {
	return v.operandName
}

func (v *VPAResource) Object(ctx context.Context) (client.Object, error) {
	vpaSpec := v.runaiConfig.Spec.ExternalWorkloadIntegrator.VPA

	if v.Enabled() {
		vpa := outils.BuildVPA(
			DeploymentName,
			v.runaiConfig.GetNamespace(),
			autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       DeploymentName,
				APIVersion: "apps/v1",
			},
			vpaSpec,
		)
		return vpa, nil
	}

	return &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName,
			Namespace: v.runaiConfig.GetNamespace(),
		},
	}, nil
}

func (v *VPAResource) Status(ctx context.Context) error {
	return nil
}

func (v *VPAResource) Enabled() bool {
	vpaSpec := v.runaiConfig.Spec.ExternalWorkloadIntegrator.VPA
	return vpaSpec != nil && vpaSpec.Enabled != nil && *vpaSpec.Enabled
}
