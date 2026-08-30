package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func ValidateAuthenticationHeaders(request *http.Request) bool {
	err := loadAuthenticationHeaders(request)
	if err != nil {
		logging.Warnf("%v", err)
		return false
	}
	return true
}

func loadAuthenticationHeaders(request *http.Request) error {
	requester, err := runaictx.ParseSubjectId(request.Context())
	if err != nil {
		return err
	}

	if requester == "" {
		return fmt.Errorf("%s header should not be empty", runaictx.CfSubjectID)
	}

	_, err = runaictx.ParseTenantId(request.Context())
	if err != nil {
		return err
	}

	return nil
}

func TenantIdOf(ctx context.Context) int32 {
	value, ok := ctx.Value(runaictx.CfTenantId).(string)
	if !ok {
		return 0
	}
	valueAsInt, _ := strconv.Atoi(value)
	return int32(valueAsInt)
}

func SubjectIdOf(ctx context.Context) string {
	value, ok := ctx.Value(runaictx.CfSubjectID).(string)
	if !ok {
		return "unknown"
	}
	return value
}

func MockAuthInfo() context.Context {
	return MockAuthContext("test", 1)
}

func MockAuthContext(requester string, tenantId int32) context.Context {
	r := context.WithValue(context.Background(), runaictx.CfSubjectID, requester)
	r = context.WithValue(r, runaictx.CfTenantId, tenantId)
	return r
}
