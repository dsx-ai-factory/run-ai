package testing

import (
	"context"
	"fmt"
	"strings"
	"testing"

	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"k8s.io/apimachinery/pkg/runtime"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	commontesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJobBuilder(inputYaml []byte, t *testing.T, _ error, _ string,
	podLabels map[string]string, podAnnotations map[string]string, _ ...*runtime.Object) (factory.Builder, context.Context, error) {
	var jobWorkload jobs.JobWorkload
	if err := yaml.Unmarshal(inputYaml, &jobWorkload); err != nil {
		return nil, nil, err
	}

	fakePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing",
			Namespace: jobWorkload.ObjectMeta.Namespace,
		},
	}

	projectName := strings.TrimPrefix(jobWorkload.ObjectMeta.Namespace, "runai-")

	fakeProject := runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
		Spec: runaiv2.ProjectSpec{
			InteractiveJobTimeLimitSecs:                  120,
			InteractiveJobMaxIdleDurationSecs:            1200,
			InteractivePreemptibleJobMaxIdleDurationSecs: 1800,
			TrainingJobMaxIdleDurationSecs:               2400,
			TrainingJobTimeLimitSecs:                     120,
		},
	}

	kaiNodePool := kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: "nodepoola"},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}

	fakeClient := commontesting.NewFakeClient(t, &fakePvc, &fakeProject, &kaiNodePool)
	builder := jobWorkload.Builder(factory.BuilderOptions{DryRun: true}, fakeClient,
		podLabels, podAnnotations)
	ctx := workloadproject.SetProjectInContext(context.TODO(), projectName)
	err := builder.Build(ctx)
	return builder, ctx, err
}

// MockUuidGenerator set up a mock UUID generator for the tests, with real UUID generator we cannot have a static result
// to compare to
type MockUuidGenerator struct {
	utils.UuidGenerator
}

func (self *MockUuidGenerator) Get(idx int) string {
	return fmt.Sprintf("uuid-%d", idx)
}
