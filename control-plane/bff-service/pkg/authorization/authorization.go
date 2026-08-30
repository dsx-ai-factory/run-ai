package business

import (
	"context"
	"errors"

	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type AuthManagerInterface interface {
	ValidatePrivilegeEscalation(ctx context.Context, authzClient auth_client.IAuthorizationClient, subjectId string) error
	GetTenantID(ctx context.Context, tenantId string) (*string, error)
	HandleSystemPermissions(ctx context.Context) error
}

type AuthManager struct{}

func NewAuthManager() AuthManagerInterface {
	return &AuthManager{}
}

func (a *AuthManager) GetTenantID(ctx context.Context, tenantId string) (*string, error) {
	if tenantId == "" {
		tenant := runaictx.GetString(ctx, runaictx.CfTenantId)
		return tenant, nil
	}

	if a.HandleSystemPermissions(ctx) != nil {
		return nil, errors.New("unauthorized to manage UI settings")
	}
	return &tenantId, nil
}

func (a *AuthManager) HandleSystemPermissions(ctx context.Context) error {
	if !runaictx.IsManagementTenant(ctx) {
		return errors.New("unauthorized to manage UI settings")
	}

	return nil
}

func (authManager *AuthManager) ValidatePrivilegeEscalation(ctx context.Context, authzClient auth_client.IAuthorizationClient, subjectId string) error {
	_, err := authzClient.PrivilegeEscalationEvaluate(ctx).SubjectId(subjectId).GroupIds([]string{}).Execute()
	return err
}
