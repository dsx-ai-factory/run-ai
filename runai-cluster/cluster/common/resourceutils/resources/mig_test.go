package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var nonExistingResourceName = "nvidia.com/mig-1g.5gb111"

func TestExtractGpuAndMemoryFromMigDeviceName(t *testing.T) {
	gpu, memory, err := ExtractGpuAndMemoryFromMigResourceName(fmt.Sprintf("%s%s", constants.AnnotationMigProfilePrefix, "1g.5gb"))
	assert.Equal(t, 1, gpu)
	assert.Equal(t, 5, memory)
	assert.Nil(t, err)
}

func TestExtractGpuAndMemoryFromMigDeviceNameError(t *testing.T) {
	gpu, memory, err := ExtractGpuAndMemoryFromMigResourceName(nonExistingResourceName)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("failed to extract gpu/memory from %s", nonExistingResourceName), err.Error())
	assert.Equal(t, -1, gpu)
	assert.Equal(t, -1, memory)
}

func Test_isMigResource(t *testing.T) {
	tests := []struct {
		name  string
		rName string
		want  bool
	}{
		{
			"is mig",
			"nvidia.com/mig-1g.5gb",
			true,
		},
		{
			"not mig",
			"nvidia.com/gpu",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMigResource(tt.rName); got != tt.want {
				t.Errorf("isMigResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAllocatedMigPortion(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		node *v1.Node
	}
	tests := []struct {
		name           string
		args           args
		deviceTypeName string
		migGPUs        float64
		wantErr        bool
	}{
		{
			"Valid mig",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							migDeviceAnnotationName: "{ \"name\": \"nvidia.com/mig-2g.10gb\" }",
						},
					},
					Spec: v1.PodSpec{NodeName: "n1"},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "NVIDIA-A100-SXM4-40GB"},
					},
				},
			},
			"nvidia.com/mig-2g.10gb",
			float64(2) / float64(7),
			false,
		},
		{
			"No mig device",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
					Spec: v1.PodSpec{NodeName: "n1"},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "NVIDIA-A100-SXM4-40GB"},
					},
				},
			},
			"",
			0,
			true,
		},
		{
			"Invalid mig annotation",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							migDeviceAnnotationName: "{ \"name\": \"mig-2g.10gb\" }",
						},
					},
					Spec: v1.PodSpec{NodeName: "n1"},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "NVIDIA-A100-SXM4-40GB"},
					},
				},
			},
			"",
			0,
			true,
		},
		{
			"No gpu product label",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							migDeviceAnnotationName: "{ \"name\": \"nvidia.com/mig-2g.10gb\" }",
						},
					},
					Spec: v1.PodSpec{NodeName: "n1"},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{},
					},
				},
			},
			"",
			0,
			true,
		},
		{
			"Invalid gpu product type for mig",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							migDeviceAnnotationName: "{ \"name\": \"nvidia.com/mig-2g.10gb\" }",
						},
					},
					Spec: v1.PodSpec{NodeName: "n1"},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "NVIDIA-A20-SXM4-40GB"},
					},
				},
			},
			"",
			0,
			true,
		},
		{
			"Unallocated pod",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							migDeviceAnnotationName: "{ \"name\": \"nvidia.com/mig-2g.10gb\" }",
						},
					},
				},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "NVIDIA-A100-SXM4-40GB"},
					},
				},
			},
			"",
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			err := v1.AddToScheme(scheme)
			if err != nil {
				t.Fatal(err)
			}
			kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.args.node).Build()

			got, got1, err := calculateAllocatedMigPortion(context.TODO(), tt.args.pod, kubeClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateAllocatedMigPortion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.deviceTypeName {
				t.Errorf("calculateAllocatedMigPortion() got = %v, want %v", got, tt.deviceTypeName)
			}
			if got1 != tt.migGPUs {
				t.Errorf("calculateAllocatedMigPortion() got1 = %v, want %v", got1, tt.migGPUs)
			}
		})
	}
}

func Test_extractGpuAndMemoryFromMigResourceName(t *testing.T) {
	tests := []struct {
		name            string
		migResourceName string
		compute         int
		memory          int
		wantErr         bool
	}{
		{
			"Valid mig 1",
			"nvidia.com/mig-2g.10gb",
			2,
			10,
			false,
		},
		{
			"Valid mig 2",
			"nvidia.com/mig-1g.5gb",
			1,
			5,
			false,
		},
		{
			"invalid mig resource",
			"nvidia.com/gpu3",
			-1,
			-1,
			true,
		},
		{
			"Invalid mig 2",
			"nvidia.com/mig-100000000000000000000000000g.5gb",
			-1,
			-1,
			true,
		},
		{
			"Invalid mig 2",
			"nvidia.com/mig-1g.100000000000000000000000000gb",
			-1,
			-1,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := extractGpuAndMemoryFromMigResourceName(tt.migResourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractGpuAndMemoryFromMigResourceName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.compute {
				t.Errorf("extractGpuAndMemoryFromMigResourceName() got = %v, want %v", got, tt.compute)
			}
			if got1 != tt.memory {
				t.Errorf("extractGpuAndMemoryFromMigResourceName() got1 = %v, want %v", got1, tt.memory)
			}
		})
	}
}

func Test_getNodeGpuType(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		node *v1.Node
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"pod with gpu",
			args{
				&v1.Pod{Spec: v1.PodSpec{NodeName: "n1"}},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "gpu-type-1"},
					},
				},
			},
			"gpu-type-1",
			false,
		},
		{
			"pod without gpu",
			args{
				&v1.Pod{Spec: v1.PodSpec{NodeName: "n1"}},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{},
					},
				},
			},
			"",
			true,
		},
		{
			"pod with gpu - but no node in cluster",
			args{
				&v1.Pod{Spec: v1.PodSpec{NodeName: "n2"}},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1", Labels: map[string]string{constants.NvidiaGPUProduct: "gpu-type-1"},
					},
				},
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			err := v1.AddToScheme(scheme)
			if err != nil {
				t.Fatal(err)
			}
			kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.args.node).Build()

			got, err := getNodeGpuType(context.Background(), tt.args.pod, kubeClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeGpuType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNodeGpuType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDeviceComputePortions(t *testing.T) {
	tests := []struct {
		name           string
		gpuProduct     string
		computePortion float64
		wantErr        bool
	}{
		{
			"A100",
			"NVIDIA-A100-SXM4-40GB",
			defaultComputePortions,
			false,
		},
		{
			"H100",
			H100,
			defaultComputePortions,
			false,
		},
		{
			"nonDevice",
			"nonDevice",
			-1,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDeviceComputePortions(tt.gpuProduct)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDeviceComputePortions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.computePortion {
				t.Errorf("getDeviceComputePortions() got = %v, want %v", got, tt.computePortion)
			}
		})
	}
}

func TestAnnotationDontCountAsMigDeviceRequest(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"nvidia.com/mig-3g.20gb": "1",
			},
		},
	}

	assert.False(t, HasMigDeviceRequest(&pod))
}

func TestHasMigDeviceRequest(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"nvidia.com/mig-3g.20gb": resource.MustParse("1"),
					},
				}},
			},
		},
	}

	assert.True(t, HasMigDeviceRequest(&pod))
}

func TestDoesntHaveMigDeviceRequest(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{},
				}},
			},
		},
	}

	assert.False(t, HasMigDeviceRequest(&pod))
}
