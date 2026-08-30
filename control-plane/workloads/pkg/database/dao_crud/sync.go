package dao_crud

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

const (
	SequenceLastValue     = "last_value"
	SequenceWorkloadsCrud = "crud_workloads_sequence"
)

func GetHighestUpdateCount(ctx context.Context) (int64, *api.Error) {
	var currentValue int64
	result := dao_utils.DB().WithContext(ctx).Table(SequenceWorkloadsCrud).Select(SequenceLastValue).Scan(&currentValue)
	if result.Error != nil {
		return 0, api.NewError(http.StatusServiceUnavailable, "failed to read highest update count", result.Error).WithDetailsFromDb(result)
	}
	return currentValue, nil
}
