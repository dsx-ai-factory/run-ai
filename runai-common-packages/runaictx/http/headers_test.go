package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

func TestContextHeaders_WithOriginal_SetHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdValue)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectValue)
	ctx = context.WithValue(ctx, runaictx.CfGroups, []string{groupsValue1, groupsValue2})
	ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, supplementalGroupsValue)
	ctx = context.WithValue(ctx, runaictx.CfClientId, clientIdValue)

	req := &http.Request{}
	SetHeaders(ctx, req)
	assert.Equal(t, tenantIdValue, req.Header.Get(tenantIdHeader))
	assert.Equal(t, subjectValue, req.Header.Get(subjectIdHeader))
	assert.Equal(t, []string{groupsValue1, groupsValue2}, req.Header.Values(groupsHeader))
	assert.Equal(t, supplementalGroupsValue, req.Header.Get(supplementalGroupsHeader))
	assert.Equal(t, clientIdValue, req.Header.Get(clientIdHeader))
}

func TestContextHeaders_WithContextFields_SetHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdValue)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectValue)
	ctx = context.WithValue(ctx, runaictx.CfGroups, []string{groupsValue1, groupsValue2})
	ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, supplementalGroupsValue)
	ctx = context.WithValue(ctx, runaictx.CfClientId, clientIdValue)

	ctx = context.WithValue(ctx, runaictx.CfTraceId, fmt.Sprintf("%s-value", string(runaictx.CfTraceId)))
	ctx = context.WithValue(ctx, runaictx.CfSpanId, fmt.Sprintf("%s-value", string(runaictx.CfSpanId)))

	req := &http.Request{}
	SetHeaders(ctx, req)
	assert.Equal(t, tenantIdValue, req.Header.Get(tenantIdHeader))
	assert.Equal(t, subjectValue, req.Header.Get(subjectIdHeader))
	assert.Equal(t, []string{groupsValue1, groupsValue2}, req.Header.Values(groupsHeader))
	assert.Equal(t, supplementalGroupsValue, req.Header.Get(supplementalGroupsHeader))
	assert.Equal(t, clientIdValue, req.Header.Get(clientIdHeader))
}

func TestContextHeaders_WithUidGid_SetHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfGId, "some-gid")
	ctx = context.WithValue(ctx, runaictx.CfUId, "some-uid")

	req := &http.Request{}
	SetHeaders(ctx, req)
	assert.Equal(t, "some-gid", req.Header.Get(gIdHeader))
	assert.Equal(t, "some-uid", req.Header.Get(uIdHeader))
}

func TestContextHeaders_WithSid_SetHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSId, "some-sid")

	req := &http.Request{}
	SetHeaders(ctx, req)
	assert.Equal(t, "some-sid", req.Header.Get(sIdHeader))
}
