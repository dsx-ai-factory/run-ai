package testing_test

import (
	"testing"

	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

func TestSetOrmMock(t *testing.T) {
	db := dbtesting.SetOrmMock(t)
	require.NotNil(t, db)
}

func TestSetSQLMock(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	require.NotNil(t, mocker)
}
