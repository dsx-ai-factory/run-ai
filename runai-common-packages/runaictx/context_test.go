package runaictx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetField(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfTenantId, "12345")
	value := GetField(ctx, CfTenantId)
	if value != "12345" {
		t.Errorf("expected '12345', got %v", value)
	}
}

func TestGetString(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfTenantId, "12345")
	expectedValue := "12345"
	value := GetString(ctx, CfTenantId)
	assert.Equal(t, &expectedValue, value)

	// Test for non-string value
	ctx = context.WithValue(context.Background(), CfTenantId, 12345)
	value = GetString(ctx, CfTenantId)
	var expectedNotStringValue *string = nil
	assert.Equal(t, expectedNotStringValue, value)

}

func TestToMap(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfTenantId, "12345")
	ctx = context.WithValue(ctx, CfGroups, []string{"group-1", "group-2"})
	ctxMap := ToMap(ctx)

	mapping := getRunaiCtxMapping()
	assert.Equal(t, "12345", ctxMap[mapping[CfTenantId]])
	assert.Equal(t, []string{"group-1", "group-2"}, ctxMap[mapping[CfGroups]])
}

func TestFromMapNullCtx(t *testing.T) {
	ctxMap := map[string]interface{}{
		"tenant_id": "12345",
		"client_id": "client-1",
		"groups":    []string{"group-1", "group-2"},
	}
	ctx := FromMap(nil, ctxMap)
	assert.Equal(t, "12345", ctx.Value(CfTenantId))
	assert.Equal(t, "client-1", ctx.Value(CfClientId))
	assert.Equal(t, []string{"group-1", "group-2"}, ctx.Value(CfGroups))
}

func TestFromMapCtxProvided(t *testing.T) {
	ctxMap := map[string]interface{}{
		"tenant_id": "12345",
		"client_id": "client-1",
		"groups":    []string{"group-1", "group-2"},
	}
	ctx := context.Background()
	ctx = FromMap(ctx, ctxMap)
	assert.Equal(t, "12345", ctx.Value(CfTenantId))
	assert.Equal(t, "client-1", ctx.Value(CfClientId))
	assert.Equal(t, []string{"group-1", "group-2"}, ctx.Value(CfGroups))
}

func TestFromMapInterfaceType(t *testing.T) {
	var group1 interface{} = "group-1"
	var group2 interface{} = "group-2"

	ctxMap := map[string]interface{}{
		"tenant_id": "12345",
		"client_id": "client-1",
		"groups":    []interface{}{group1, group2},
	}
	ctx := context.Background()
	ctx = FromMap(ctx, ctxMap)
	assert.Equal(t, "12345", ctx.Value(CfTenantId))
	assert.Equal(t, "client-1", ctx.Value(CfClientId))
	assert.Equal(t, []string{"group-1", "group-2"}, ctx.Value(CfGroups))
}
