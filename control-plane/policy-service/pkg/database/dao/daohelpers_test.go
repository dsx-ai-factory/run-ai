package dao

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t dbtesting.TestingT) dbtesting.Mocker {
	return dbtesting.SetSQLMock(t)
}

type countRow struct {
	Count int64
}

func ctxWithTenant(tenantId int32) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
}
