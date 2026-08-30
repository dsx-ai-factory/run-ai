package pod_mutators

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
)

func TestResourceMutation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							types.GpuResource: resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			CpuDefaultRequestGpuFactor:    "2",
			CpuDefaultLimitGpuFactor:      "3",
			MemoryDefaultRequestGpuFactor: "4",
			MemoryDefaultLimitGpuFactor:   "5",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m

	err := mutator.mutatePodResources(&pod, nil)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources

	request, found := resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("2")))

	request, found = resources.Limits[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("3")))

	request, found = resources.Requests[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("4")))

	request, found = resources.Limits[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("5")))
}

func TestResourceMutationRequestSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU: resource.MustParse("1"),
						},

						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			CpuDefaultRequestCpuLimitFactorNoGpu: "2",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m

	err := mutator.mutatePodResources(&pod, nil)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources
	request, found := resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("1")))
}

func TestResourceMutationNoGpu(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			MemoryDefaultRequestMemoryLimitFactorNoGpu: "3",
			CpuDefaultRequestCpuLimitFactorNoGpu:       "2",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m

	err := mutator.mutatePodResources(&pod, nil)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources
	request, found := resources.Requests[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("12")))

	request, found = resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("2")))
}

func TestResourceMutationWithNilConfigMap(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{},
				},
			},
		},
	}

	mutator := createMutator(testNamespaceName, false)
	err := mutator.mutatePodResources(&pod, nil)
	assert.Nil(t, err)
}

func draTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = v1.AddToScheme(s)
	_ = resourcev1.AddToScheme(s)
	_ = resourcev1beta1.AddToScheme(s)
	_ = resourcev1beta2.AddToScheme(s)
	return s
}

func createDRAParser(namespace string) *dra.DRAParser {
	claim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gpu-claim",
			Namespace: namespace,
		},
		Spec: resourcev1.ResourceClaimSpec{
			Devices: resourcev1.DeviceClaim{
				Requests: []resourcev1.DeviceRequest{
					{
						Name: "gpu",
						Exactly: &resourcev1.ExactDeviceRequest{
							DeviceClassName: "gpu.nvidia.com",
							AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
							Count:           4,
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(draTestScheme()).WithObjects(claim).Build()
	fakeKubeClient := kubefake.NewClientset(claim)
	return dra.NewDRAParser(fakeClient, fakeKubeClient)
}

func TestResourceMutationWithDRAClaims(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{},
				},
			},
			ResourceClaims: []v1.PodResourceClaim{
				{
					Name:              "gpu-claim",
					ResourceClaimName: ptr.To("my-gpu-claim"),
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			CpuDefaultRequestGpuFactor:    "2",
			CpuDefaultLimitGpuFactor:      "3",
			MemoryDefaultRequestGpuFactor: "4",
			MemoryDefaultLimitGpuFactor:   "5",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	mutator.draParser = createDRAParser(testNamespaceName)

	req := &v1beta1.AdmissionRequest{
		Namespace: testNamespaceName,
	}

	err := mutator.mutatePodResources(&pod, req)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources

	// DRA claim has 4 GPUs, so CPU/memory should be proportional to 4 GPUs
	cpuRequest, found := resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, cpuRequest.Equal(resource.MustParse("8"))) // 2 * 4

	cpuLimit, found := resources.Limits[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, cpuLimit.Equal(resource.MustParse("12"))) // 3 * 4

	memRequest, found := resources.Requests[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, memRequest.Equal(resource.MustParse("16"))) // 4 * 4

	memLimit, found := resources.Limits[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, memLimit.Equal(resource.MustParse("20"))) // 5 * 4
}

func TestResourceMutationTraditionalGpuTakesPriorityOverDRA(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							types.GpuResource: resource.MustParse("1"),
						},
					},
				},
			},
			ResourceClaims: []v1.PodResourceClaim{
				{
					Name:              "gpu-claim",
					ResourceClaimName: ptr.To("my-gpu-claim"),
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			CpuDefaultRequestGpuFactor:    "2",
			CpuDefaultLimitGpuFactor:      "3",
			MemoryDefaultRequestGpuFactor: "4",
			MemoryDefaultLimitGpuFactor:   "5",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	mutator.draParser = createDRAParser(testNamespaceName)

	req := &v1beta1.AdmissionRequest{
		Namespace: testNamespaceName,
	}

	err := mutator.mutatePodResources(&pod, req)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources

	// Traditional GPU (1) takes priority, not DRA GPU (4)
	cpuRequest, found := resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, cpuRequest.Equal(resource.MustParse("2"))) // 2 * 1

	cpuLimit, found := resources.Limits[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, cpuLimit.Equal(resource.MustParse("3"))) // 3 * 1

	memRequest, found := resources.Requests[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, memRequest.Equal(resource.MustParse("4"))) // 4 * 1

	memLimit, found := resources.Limits[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, memLimit.Equal(resource.MustParse("5"))) // 5 * 1
}

func TestResourceMutationNilDRAParserFallsThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceMemory: resource.MustParse("4"),
							v1.ResourceCPU:    resource.MustParse("1"),
						},
					},
				},
			},
			ResourceClaims: []v1.PodResourceClaim{
				{
					Name:              "gpu-claim",
					ResourceClaimName: ptr.To("my-gpu-claim"),
				},
			},
		},
	}

	configMap := v1.ConfigMap{
		Data: map[string]string{
			MemoryDefaultRequestMemoryLimitFactorNoGpu: "3",
			CpuDefaultRequestCpuLimitFactorNoGpu:       "2",
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetRunaiConfigMap().Return(&configMap, nil)

	// nil DRA parser — should fall through to no-GPU path
	mutator := createMutator(testNamespaceName, false)
	mutator.client = m

	req := &v1beta1.AdmissionRequest{
		Namespace: testNamespaceName,
	}

	err := mutator.mutatePodResources(&pod, req)
	assert.Nil(t, err)
	resources := pod.Spec.Containers[0].Resources

	// Should behave as no-GPU path (same as TestResourceMutationNoGpu)
	request, found := resources.Requests[v1.ResourceMemory]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("12")))

	request, found = resources.Requests[v1.ResourceCPU]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("2")))
}
