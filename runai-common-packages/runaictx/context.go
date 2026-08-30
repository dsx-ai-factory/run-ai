package runaictx

import (
	"context"
)

type ContextField string

const (
	CfIsManagementTenant ContextField = "is-management-tenant"
	CfTenantId           ContextField = "tenant-id"
	CfClientId           ContextField = "client-id"
	CfClientType         ContextField = "client-type"
	CfSubjectID          ContextField = "subject-id"
	CfSubjectType        ContextField = "subject-type"
	CfGroups             ContextField = "groups"
	CfSupplementalGroups ContextField = "supplemental-groups"
	CfRequestOrigin      ContextField = "request-origin"
	CfTraceId            ContextField = "trace-id"
	CfSpanId             ContextField = "span-id"
	CfApplicationId      ContextField = "application-id"
	CfAccept             ContextField = "accept"
	CfGId                ContextField = "gid"
	CfUId                ContextField = "uid"
	CfSId                ContextField = "sid"
)

func GetField(ctx context.Context, field ContextField) any {
	return ctx.Value(field)
}

func GetString(ctx context.Context, field ContextField) *string {
	if value, ok := ctx.Value(field).(string); ok {
		return &value
	}
	return nil
}

func ToMap(ctx context.Context) map[string]interface{} {
	ctxMap := make(map[string]interface{})
	for ctxField, fieldIdentifier := range getRunaiCtxMapping() {
		if value := ctx.Value(ctxField); value != nil {
			ctxMap[fieldIdentifier] = value
		}
	}
	return ctxMap
}

func FromMap(ctx context.Context, ctxMap map[string]interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	for ctxField, fieldIdentifier := range getRunaiCtxMapping() {
		fieldVal := extractCtxFieldValue(ctxMap, fieldIdentifier)
		if fieldVal == nil {
			continue
		}
		ctx = context.WithValue(ctx, ctxField, fieldVal)
	}
	return ctx
}
