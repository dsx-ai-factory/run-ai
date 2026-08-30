package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	prom_common "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

type ServiceResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceResource {
	return &ServiceResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *ServiceResource) OperandName() string {
	return r.operandName
}

func (r *ServiceResource) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = prom_common.MgmtPrometheusServiceName
		namespace = r.runaiConfig.Namespace
	)

	service := &v1.Service{}

	// Get the existing service if it exists to consume any cluster-set values
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}

	service.ObjectMeta.Labels["app"] = name

	if service.ObjectMeta.Annotations == nil {
		service.ObjectMeta.Annotations = map[string]string{}
	}
	service.ObjectMeta.Annotations["dgxc.nvidia.com/enable-endpoints-update"] = "true"

	service.Spec.ClusterIP = v1.ClusterIPNone
	service.Spec.Ports = []v1.ServicePort{
		{
			Protocol:   v1.ProtocolTCP,
			Port:       443,
			TargetPort: intstr.FromInt(443),
		},
	}

	return service, nil
}

func (r *ServiceResource) Status(ctx context.Context) error {
	service, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if service.GetUID() == "" {
		return fmt.Errorf("service [%s] is missing", service.GetName())
	}

	return nil
}

func (r *ServiceResource) Enabled() bool {
	return r.runaiConfig.Spec.Internal.Global.NMC.IsEnabled()
}
