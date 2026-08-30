package resolve

import (
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

func TestShouldResolve(t *testing.T) {
	tests := []struct {
		name      string
		framework api.Framework
		want      bool
	}{
		{name: "runai (workspace/training) is resolved", framework: api.Runai, want: true},
		{name: "knative inference is resolved", framework: api.Knative, want: true},
		{name: "lws distributed inference is resolved", framework: api.LWS, want: true},
		{name: "pytorch distributed is resolved", framework: api.PyTorch, want: true},
		{name: "mpi distributed is resolved", framework: api.MPI, want: true},
		{name: "tf distributed is resolved", framework: api.TF, want: true},
		{name: "xgboost distributed is resolved", framework: api.XGBoost, want: true},
		{name: "jax distributed is resolved", framework: api.Jax, want: true},
		{name: "nimservice is skipped (own external routing)", framework: api.NimService, want: false},
		{name: "undefined is skipped (already resolved)", framework: api.UndefinedFramework, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldResolve(tt.framework); got != tt.want {
				t.Errorf("shouldResolve(%v) = %v, want %v", tt.framework, got, tt.want)
			}
		})
	}
}

// TestClusterWorkloadNimServiceIsShortCircuited verifies that ClusterWorkload skips the
// control-plane lookup for NimService (whose external cluster API routing does not depend on
// workloads-manager) and returns its inputs unchanged.
func TestClusterWorkloadNimServiceIsShortCircuited(t *testing.T) {
	const name = "my-nim"
	gotType, gotFramework, gotName, err := ClusterWorkload(nil, nil, api.Inference, api.NimService, "proj", name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotType != api.Inference || gotFramework != api.NimService || gotName != name {
		t.Errorf("ClusterWorkload changed NimService inputs: got (%v, %v, %q)", gotType, gotFramework, gotName)
	}
}

// TestClusterWorkloadUndefinedFrameworkIsIdempotent verifies that a workload already
// resolved to UndefinedFramework (e.g. by the generic `runai workload` command's initializer
// or a prior ClusterWorkload call) is not resolved again.
func TestClusterWorkloadUndefinedFrameworkIsIdempotent(t *testing.T) {
	const name = "already-resolved-external-name"
	gotType, gotFramework, gotName, err := ClusterWorkload(nil, nil, api.Training, api.UndefinedFramework, "proj", name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotType != api.Training || gotFramework != api.UndefinedFramework || gotName != name {
		t.Errorf("ClusterWorkload changed pre-resolved inputs: got (%v, %v, %q)", gotType, gotFramework, gotName)
	}
}
