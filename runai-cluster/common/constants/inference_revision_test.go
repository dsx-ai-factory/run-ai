package constants

import "testing"

func TestIsManagedInferenceRevision(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{
			name: "native inference workload",
			labels: map[string]string{
				ControlPlaneWorkloadKindLabelForWorkload: NativeInferenceWorkloadKind,
			},
			want: true,
		},
		{
			name: "managed V2 inference workload",
			labels: map[string]string{
				ControlPlaneWorkloadKindLabelForWorkload: V2InferenceWorkloadKind,
				ControlPlaneWorkloadNameLabelForWorkload: "inference-workload",
			},
			want: true,
		},
		{
			name: "unmanaged V2-like Knative revision",
			labels: map[string]string{
				ControlPlaneWorkloadKindLabelForWorkload: V2InferenceWorkloadKind,
			},
		},
		{
			name: "managed non-inference revision",
			labels: map[string]string{
				ControlPlaneWorkloadKindLabelForWorkload: "Training",
				ControlPlaneWorkloadNameLabelForWorkload: "training-workload",
			},
		},
		{name: "unlabeled revision"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsManagedInferenceRevision(tt.labels); got != tt.want {
				t.Fatalf("IsManagedInferenceRevision() = %v, want %v", got, tt.want)
			}
		})
	}
}
