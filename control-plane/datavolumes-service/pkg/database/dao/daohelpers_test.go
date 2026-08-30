package dao

import (
	"context"
	"testing"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t *testing.T) dbtesting.Mocker {
	return dbtesting.SetSQLMock(t)
}

type countRow struct {
	Count int64
}

func testCtx(tenantId string, permittedScopes *authorization_api.PermittedScopes) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test-user")
	if permittedScopes != nil {
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)
	}
	return ctx
}
