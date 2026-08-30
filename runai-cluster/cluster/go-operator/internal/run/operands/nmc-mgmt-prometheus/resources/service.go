package resources

import (
	context "context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
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

	service := &corev1.Service{}
	err := r.k8sReader.Get(ctx, k8stypes.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace
	service.Spec.Selector = map[string]string{
		"app.kubernetes.io/instance": prometheusInstanceName,
	}
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "web",
			Port:       9090,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.IntOrString{IntVal: 9090},
		},
	}
	service.Spec.Type = corev1.ServiceTypeClusterIP

	return service, nil
}

func (r *ServiceResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}
	service := resource.(*corev1.Service)
	if service.UID == "" {
		return nil // Not created yet
	}
	return nil
}

func (r *ServiceResource) Enabled() bool {
	return true
}
