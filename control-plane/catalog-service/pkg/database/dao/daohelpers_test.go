package dao

import (
	"testing"

	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t *testing.T) dbtesting.Mocker {
	mocker := dbtesting.SetSQLMock(t)
	return mocker
}
