package autorization

import (
	"context"
	"fmt"
	"strconv"

	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx"
)

func ResolveTenantID(ctx context.Context, tenantIdParam string) (int, error) {
	if IsManagementTenant(ctx) {
		if tenantIdParam != "" {
			tenantID, err := strconv.Atoi(tenantIdParam)
			if err != nil {
				return 0, fmt.Errorf("invalid tenant Id query parameter: %w", err)
			}
			return tenantID, nil
		}
	}

	tenantID, err := extractTenantID(ctx)
	if err != nil {
		return 0, fmt.Errorf("invalid tenant id")
	}
	return tenantID, nil
}

func IsManagementTenant(ctx context.Context) bool {
	val := runaictx.GetString(ctx, runaictx.CfIsManagementTenant)
	return val != nil && *val == "true"
}

func extractTenantID(ctx context.Context) (int, error) {
	tenantIDStr := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantIDStr == nil || *tenantIDStr == "" {
		return 0, fmt.Errorf("tenant id not found in context")
	}
	return strconv.Atoi(*tenantIDStr)
}
