package testing

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
)

func AddMnnvlToContextWrapper(ctx context.Context, status interface{}) context.Context {
	mnnvlStatus, ok := status.(mnnvlutils.Status)
	if !ok {
		panic("status is not mnnvl.Status")
	}
	return mnnvlutils.ToContext(ctx, mnnvlStatus)
}
