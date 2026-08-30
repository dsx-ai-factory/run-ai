package config_scale

import (
	"context"
	"fmt"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1global "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SmallResourceConfig struct {
	kubeClient client.Client
}

func NewSmallResourceConfig(kubeClient client.Client) ResourceScaleConfig {
	return &SmallResourceConfig{
		kubeClient: kubeClient,
	}
}

func (c *SmallResourceConfig) identifier() string {
	return "Small"
}

func (c *SmallResourceConfig) adjustFakeGPUOperator(ctx context.Context) error {
	return nil
}

func (c *SmallResourceConfig) adjustTrainingOperator(ctx context.Context) error {
	fmt.Println("Adjust training operator resources")
	deployment := &appsv1.Deployment{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "training-operator",
		Namespace: "kubeflow",
	}, deployment)

	if err != nil {
		return err
	}
	deployment.Spec.Template.Spec.Containers[0].Image = "kubeflow/training-operator:v1-3f15cb8"

	err = c.kubeClient.Update(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (c *SmallResourceConfig) adjustWorkloadServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
}

func (c *SmallResourceConfig) adjustSyncServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
}

func (c *SmallResourceConfig) adjustSchedulingServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust scheduler services - Small resources")

	if runaiConfig.Spec.Global.SchedulingServices == nil {
		runaiConfig.Spec.Global.SchedulingServices = &runaiv1global.SchedulingServices{}
	}
	runaiConfig.Spec.Global.SchedulingServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("700m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}

func (c *SmallResourceConfig) adjustThanos(ctx context.Context) error {
	return nil
}

func (c *SmallResourceConfig) adjustPostgres(ctx context.Context) error {
	return nil
}
