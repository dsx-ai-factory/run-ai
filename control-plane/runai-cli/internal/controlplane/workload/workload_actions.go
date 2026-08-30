package workload

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type ActionClient interface {
	Delete(ctx context.Context, uuid string) error
	Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error)
}

type WithSuspendableAction interface {
	ActionClient
	Suspend(ctx context.Context, uuid string) error
	Resume(ctx context.Context, uuid string) error
}

type WithUpdateAction interface {
	ActionClient
	Update(ctx context.Context, uuid string, clusterId string, spec *spec.GeneralSpec) (string, error)
}
