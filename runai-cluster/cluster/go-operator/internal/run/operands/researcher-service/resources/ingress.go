package resources

import (
	"context"
	"fmt"
	"maps"
	"reflect"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	researcherServicePath = "researcher"
	cliServicePath        = "cli"
	ingressName           = "researcher-service-ingress"
)

type IngressResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewIngressResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *IngressResource {
	return &IngressResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (i *IngressResource) OperandName() string {
	return i.operandName
}

func (i *IngressResource) Object(ctx context.Context) (client.Object, error) {
	if !i.Enabled() {
		return &v1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ingressName,
				Namespace: i.runaiConfig.Namespace,
			},
		}, nil
	}

	ingress := &v1.Ingress{}
	err := i.k8sReader.Get(ctx, types.NamespacedName{
		Name:      ingressName,
		Namespace: i.runaiConfig.Namespace,
	}, ingress)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if err := initIngress(i.runaiConfig, ingress); err != nil {
		return nil, err
	}
	return ingress, nil
}

func (i *IngressResource) Status(ctx context.Context) error {
	if !i.Enabled() {
		return nil
	}

	res, err := i.Object(ctx)
	if err != nil {
		return err
	}

	if res.GetUID() == "" {
		return fmt.Errorf("ingress [%s] is missing", res.GetName())
	}

	ingress := res.(*v1.Ingress)
	if reflect.DeepEqual(ingress.Status, v1.IngressStatus{}) {
		return fmt.Errorf("no ingress controller has reported status for [%s]", ingress.GetName())
	}

	return nil
}

func (i *IngressResource) Enabled() bool {
	spec := i.runaiConfig.Spec
	global := spec.Global
	internal := spec.Internal
	return spec.ResearcherService != nil &&
		spec.ResearcherService.Enabled != nil &&
		*spec.ResearcherService.Enabled &&
		global.Ingress != nil &&
		global.Ingress.Enabled != nil &&
		*global.Ingress.Enabled &&
		internal.Global.ClusterURL != "" &&
		internal.ResearcherService.Ingress.Enabled != nil &&
		*internal.ResearcherService.Ingress.Enabled &&
		!(internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift)
}

func IngressStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	r := &IngressResource{k8sReader: k8sReader, runaiConfig: runaiConfig}

	if !r.Enabled() {
		return operands.ReadyStatus(), nil
	}

	res, err := r.Object(ctx)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if res.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("ingress [%s] is missing",
			res.GetName())), nil
	}

	ingress := res.(*v1.Ingress)
	if reflect.DeepEqual(ingress.Status, v1.IngressStatus{}) {
		return operands.NotReadyStatus(fmt.Sprintf("no ingress controller has reported status for [%s]",
			ingress.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}

func initIngress(runaiConfig *runv1.RunaiConfig, ingress *v1.Ingress) error {
	var (
		global = runaiConfig.Spec.Global
		config = runaiConfig.Spec.ResearcherService
	)

	domain := utils.ParseClusterDomain(runaiConfig.Spec.Internal.Global.ClusterURL)
	if domain == nil {
		return fmt.Errorf("failed to parse cluster domain '%s'", runaiConfig.Spec.Internal.Global.ClusterURL)
	}

	ingress.ObjectMeta.Name = ingressName
	ingress.ObjectMeta.Namespace = runaiConfig.Namespace

	if ingress.ObjectMeta.Labels == nil {
		ingress.ObjectMeta.Labels = map[string]string{}
	}
	ingress.ObjectMeta.Labels["app"] = DeploymentName

	ingress.ObjectMeta.Annotations = make(map[string]string)
	if global.Ingress != nil {
		maps.Copy(ingress.ObjectMeta.Annotations, global.Ingress.Annotations)
	}
	if config.Ingress != nil {
		maps.Copy(ingress.ObjectMeta.Annotations, config.Ingress.Annotations)
	}

	var ingressClassName string
	if config.Ingress != nil && config.Ingress.Class != "" {
		ingressClassName = config.Ingress.Class
	} else if global.Ingress != nil && global.Ingress.Class != "" {
		ingressClassName = global.Ingress.Class
	} else {
		ingressClassName = "nginx"
	}
	ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"] = ingressClassName

	ingress.Spec = v1.IngressSpec{
		IngressClassName: &ingressClassName,
		Rules: []v1.IngressRule{{
			Host: domain.Host,
			IngressRuleValue: v1.IngressRuleValue{
				HTTP: &v1.HTTPIngressRuleValue{
					Paths: []v1.HTTPIngressPath{
						{
							Path: utils.GetIngressPath(runaiConfig.Spec.Internal.Global.ClusterURL, researcherServicePath),
							PathType: func() *v1.PathType {
								pathType := v1.PathTypePrefix
								return &pathType
							}(),
							Backend: v1.IngressBackend{
								Service: &v1.IngressServiceBackend{
									Name: DeploymentName,
									Port: v1.ServiceBackendPort{Number: 4180},
								},
							},
						},
						{
							Path: utils.GetIngressPath(runaiConfig.Spec.Internal.Global.ClusterURL, cliServicePath),
							PathType: func() *v1.PathType {
								pathType := v1.PathTypePrefix
								return &pathType
							}(),
							Backend: v1.IngressBackend{
								Service: &v1.IngressServiceBackend{
									Name: DeploymentName,
									Port: v1.ServiceBackendPort{Number: 4180},
								},
							},
						},
						{
							Path: "/oauth2",
							PathType: func() *v1.PathType {
								pathType := v1.PathTypePrefix
								return &pathType
							}(),
							Backend: v1.IngressBackend{
								Service: &v1.IngressServiceBackend{
									Name: DeploymentName,
									Port: v1.ServiceBackendPort{Number: 4180},
								},
							},
						},
					},
				},
			},
		},
		},
		TLS: func() []v1.IngressTLS {
			if runaiConfig.Spec.ResearcherService.Ingress.TLSEnabled != nil && !*runaiConfig.Spec.ResearcherService.Ingress.TLSEnabled {
				return nil
			}

			hosts := []string{domain.Host}
			tls := []v1.IngressTLS{
				{
					Hosts:      hosts,
					SecretName: runaiConfig.Spec.ResearcherService.Ingress.TlsSecret,
				},
			}
			return tls
		}(),
	}

	return nil
}
