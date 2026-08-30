package authorizers

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/auth"
)

type Authorizer interface {
	Authorize(ctx context.Context, request *Request) (allowed bool, reason Reason, err error)
}

type Reason int

const (
	ReasonAuthorized Reason = iota
	ReasonUnauthorized
	ReasonError
)

type Request struct {
	AuthInfo    *auth.Info
	Project     string
	Namespace   string
	Workload    string
	Kind        api.WorkloadType
	Action      api.WorkloadAction
	WorkloadUID string
}
