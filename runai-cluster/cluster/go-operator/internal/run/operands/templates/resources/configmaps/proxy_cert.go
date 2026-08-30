package configmaps

import (
	"context"
	"errors"
	"fmt"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewProxyCert(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *ProxyCert {
	return &ProxyCert{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type ProxyCert struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (pc *ProxyCert) OperandName() string {
	return pc.operandName
}

func (pc *ProxyCert) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = pc.runaiConfig.Spec.Global.HTTPProxy.ConfigMapName
		namespace = pc.runaiConfig.Namespace
		certRef   = pc.runaiConfig.Spec.Global.HTTPProxy.TLSCertSecret
	)

	if certRef == nil {
		configMap := &corev1.ConfigMap{}
		configMap.ObjectMeta.Name = name
		configMap.ObjectMeta.Namespace = namespace
		return configMap, nil
	}

	secret := &corev1.Secret{}
	err := pc.k8sReader.Get(ctx, types.NamespacedName{
		Name:      certRef.Name,
		Namespace: namespace,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS cert secret %q: %w", certRef.Name, err)
	}

	certValue, ok := secret.Data[certRef.Key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in TLS cert secret %q", certRef.Key, certRef.Name)
	}

	configMap := &corev1.ConfigMap{}

	// Get the existing configMap if it exists to consume any cluster-set values
	err = pc.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = name
	configMap.ObjectMeta.Namespace = namespace

	configMap.Data = map[string]string{
		"tls.crt": string(certValue),
	}

	return configMap, nil
}

func (pc *ProxyCert) Status(ctx context.Context) error {
	configMap, err := pc.Object(ctx)
	if err != nil {
		return err
	}

	if configMap.GetUID() == "" {
		return errors.New(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName()))
	}

	return nil
}

func (pc *ProxyCert) Enabled() bool {
	return pc.runaiConfig.Spec.Global.HTTPProxy.TLSCertSecret != nil
}
