package runaictx

import (
	"context"
	"errors"
	"strings"
)

func IsManagementTenant(ctx context.Context) bool {
	value := GetString(ctx, CfIsManagementTenant)

	if value == nil || strings.ToLower(*value) != "true" {
		return false
	}

	return true
}

func ParseAuthData(ctx context.Context) (tenantId string, subjectId string, groups []string, err error) {
	tenantId, err = ParseTenantId(ctx)
	if err != nil {
		return
	}

	subjectId, err = ParseSubjectId(ctx)
	if err != nil {
		return
	}

	groups = ParseGroups(ctx)

	return
}

func ParseTenantId(ctx context.Context) (string, error) {
	tenantId := GetString(ctx, CfTenantId)

	if tenantId == nil {
		return "", errors.New("missing tenant id")
	}
	return *tenantId, nil
}

func ParseSubjectId(ctx context.Context) (string, error) {
	subjectId := GetString(ctx, CfSubjectID)

	if subjectId == nil {
		return "", errors.New("missing subject id")
	}
	return *subjectId, nil
}

func ParseGroups(ctx context.Context) []string {
	var ok bool
	var groups []string
	if groups, ok = ctx.Value(CfGroups).([]string); !ok {
		groups = []string{}
	}

	return groups
}
