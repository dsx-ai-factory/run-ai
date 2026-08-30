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

func NewLimitRange(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *LimitRange {
	return &LimitRange{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type LimitRange struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (lr *LimitRange) OperandName() string {
	return lr.operandName
}

func (lr *LimitRange) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = "default-limit-range"
		namespace = lr.runaiConfig.Namespace
		config    = lr.runaiConfig.Spec.LimitRange
	)

	configMap := &corev1.ConfigMap{}

	// Get the existing configMap if it exists to consume any cluster-set values
	err := lr.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = name
	configMap.ObjectMeta.Namespace = namespace

	data := map[string]string{}

	if config.CPUDefaultRequestGPUFactor != "" {
		data["cpuDefaultRequestGpuFactor"] = config.CPUDefaultRequestGPUFactor
	}

	if config.MemoryDefaultRequestGPUFactor != "" {
		data["memoryDefaultRequestGpuFactor"] = config.MemoryDefaultRequestGPUFactor
	}

	if config.CPUDefaultLimitGPUFactor != "" {
		data["cpuDefaultLimitGpuFactor"] = config.CPUDefaultLimitGPUFactor
	}

	if config.MemoryDefaultLimitGPUFactor != "" {
		data["memoryDefaultLimitGpuFactor"] = config.MemoryDefaultLimitGPUFactor
	}

	if config.CPUDefaultRequestCPULimitFactorNoGPU != nil {
		data["cpuDefaultRequestCpuLimitFactorNoGpu"] = fmt.Sprintf("%f", *config.CPUDefaultRequestCPULimitFactorNoGPU)
	}

	if config.MemoryDefaultRequestMemoryLimitFactorNoGPU != nil {
		data["memoryDefaultRequestMemoryLimitFactorNoGpu"] = fmt.Sprintf("%f", *config.MemoryDefaultRequestMemoryLimitFactorNoGPU)
	}

	if config.CPUDefaultRequest != "" {
		data["cpuDefaultRequest"] = config.CPUDefaultRequest
	}

	if config.MemoryDefaultRequest != "" {
		data["memoryDefaultRequest"] = config.MemoryDefaultRequest
	}

	if config.CPUDefaultLimit != "" {
		data["cpuDefaultLimit"] = config.CPUDefaultLimit
	}

	if config.MemoryDefaultLimit != "" {
		data["memoryDefaultLimit"] = config.MemoryDefaultLimit
	}

	if config.CPUMaxLimit != "" {
		data["cpuMaxLimit"] = config.CPUMaxLimit
	}

	if config.MemoryMaxLimit != "" {
		data["memoryMaxLimit"] = config.MemoryMaxLimit
	}

	configMap.Data = data

	return configMap, nil
}

func (lr *LimitRange) Status(ctx context.Context) error {
	configMap, err := lr.Object(ctx)
	if err != nil {
		return err
	}

	if configMap.GetUID() == "" {
		return errors.New(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName()))
	}

	return nil
}

func (lr *LimitRange) Enabled() bool {
	return true
}
