package testing

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	merger "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	common2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	commontesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// encryptedData holds a byte array with encrypted "private-value-1"
var encryptedData = []byte{
	168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
	224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
	201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
}

func GetMergedWorkload(workload *v2alpha1.DistributedWorkload, client client.Client) (*v2alpha1.DistributedWorkload, error) {
	projectName := strings.TrimPrefix(workload.ObjectMeta.Namespace, "runai-")
	ctx := workloadproject.SetProjectInContext(context.TODO(), projectName)

	DistributedWorkloadMerger := merger.NewDistributedWorkloadMergeCreator(client, true)

	mergeStack, masterMergeStack, err := DistributedWorkloadMerger.CreateDistributedMergeStack(ctx, workload, client)
	if err != nil {
		return nil, err
	}

	err = mergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
	if err != nil {
		return nil, err
	}
	if masterMergeStack != nil {
		err = masterMergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
		if err != nil {
			return nil, err
		}
	}

	values := mergeStack.MergedValues()
	var masterValues *gojsonq.JSONQ
	if masterMergeStack != nil {
		masterValues = masterMergeStack.MergedValues()
	}
	mergedWorkload, err := v2alpha1.UnmarshalWorkload(values, masterValues, &workload.TypeMeta, &workload.ObjectMeta)
	if err != nil {
		return nil, err
	}

	return mergedWorkload, nil
}

func GetDistributedBuilder(inputYaml []byte, t *testing.T, _ error, _ string,
	podLabels map[string]string, podAnnotations map[string]string, initObjs ...*runtime.Object) (factory.Builder, context.Context, error) {
	var workload v2alpha1.DistributedWorkload
	if err := yaml.Unmarshal(inputYaml, &workload); err != nil {
		return nil, nil, err
	}

	fakeDataVolumePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-volume-pvc",
			Namespace: "runai-team-a",
			Labels: map[string]string{
				v1alpha1.DataVolumeIdLabel: "5e0445ea-6d67-4641-a791-b4a372ef722c",
			},
		},
	}

	fakePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing",
			Namespace: workload.ObjectMeta.Namespace,
		},
	}

	fakePvc2 := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing2",
			Namespace: workload.ObjectMeta.Namespace,
		},
	}

	fakePvc3 := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing3",
			Namespace: workload.ObjectMeta.Namespace,
		},
	}

	volumeMode := corev1.PersistentVolumeBlock
	fakePvcBlock := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-block",
			Namespace: workload.ObjectMeta.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeMode: &volumeMode,
		},
	}

	projectName := strings.TrimPrefix(workload.ObjectMeta.Namespace, "runai-")

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
	fakeCredential := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-user-credential",
			Namespace: "runai",
		},
		Data: map[string][]byte{
			"my-user-key": encryptedData,
		},
	}

	fakeCredential2 := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ips-2",
			Namespace: "runai",
		},
		Data: map[string][]byte{
			"my-user-key": encryptedData,
		},
	}
	ctx := workloadproject.SetProjectInContext(context.TODO(), projectName)

	// in order to apply the fake objects above plus the objects given in initObjs,
	// we need to concatenate them to one list to be passed to NewFakeClient
	var allObjs []runtime.Object
	allObjs = append(allObjs, &fakePvc, &fakePvc2, &fakePvc3, &fakePvcBlock, &fakeProject, &fakeDataVolumePvc,
		&fakeCredential, &fakeCredential2)

	allObjs = append(allObjs,
		kaiNodePoolFixture(v1alpha1.DefaultNodePoolName, "test-key", "test-value", ""),
		kaiNodePoolFixture("nodepoola", "test-key", "test-value", ""),
		kaiNodePoolFixture("nodepoolb", "test-key-b", "test-value-b", ""),
		kaiNodePoolFixture("mnnvl-node-1", "nodepool-key-1", "nodepool-value-1", "nvidia.com/gpu.imex-domain"),
		kaiNodePoolFixture("mnnvl-node-2", "nodepool-key-2", "nodepool-value-2", "custom.run.ai/mnnvl-key"),
		kaiNodePoolFixture("mnnvl-node-3", "nodepool-key-3", "nodepool-value-3", "nvidia.com/nvlink"),
	)

	for _, obj := range initObjs {
		if obj != nil {
			allObjs = append(allObjs, *obj)
		}
	}
	fakeClient := commontesting.NewFakeClient(t, allObjs...)

	mergedWorkload, err := GetMergedWorkload(&workload, fakeClient)
	if err != nil {
		return nil, nil, err
	}

	builder, err := common2.GetBuilder(factory.BuilderOptions{DryRun: true}, fakeClient, mergedWorkload,
		podLabels, podAnnotations)

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

// kaiNodePoolFixture builds a kai.resources node pool fixture. An empty
// accelerationLabelKey leaves the annotations unset, which the readers treat as
// "acceleration not detected" and the schema default for the label key.
func kaiNodePoolFixture(name, labelKey, labelValue, accelerationLabelKey string) *kaiv1alpha1.NodePool {
	pool := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
		},
		Status: kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolReady},
	}
	if accelerationLabelKey != "" {
		pool.Annotations = map[string]string{
			kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
			kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: accelerationLabelKey,
		}
	}
	return pool
}
