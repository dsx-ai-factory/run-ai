package services_crud

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/stretchr/testify/require"
)

func TestWriteBackDistributedInferenceSnapshotPropagatesUpsertError(t *testing.T) {
	snapshot := &models_crud.WorkloadV2Snapshot{ID: uuid.New()}
	expected := api.NewError(http.StatusServiceUnavailable, "snapshot unavailable", nil)

	actual := writeBackDistributedInferenceSnapshot(
		context.Background(),
		snapshot,
		func(context.Context, *models_crud.WorkloadV2Snapshot) *api.Error {
			return expected
		},
	)

	require.NotNil(t, actual)
	require.Equal(t, expected.Code, actual.Code)
	require.Equal(t, expected.Message, actual.Message)
}
