package auth_flow

import (
	"context"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
)

type AuthFlowRepository interface {
	CreateAuthFlow(ctx context.Context, realm string, flow models.AuthenticationFlowRepresentation) error
	GetAuthFlow(ctx context.Context, realm string, flowAlias string) (*models.AuthenticationFlowRepresentation, error)
	GetAuthExecutions(ctx context.Context, realm string, flowAlias string) ([]*models.ModifyAuthenticationExecutionRepresentation, error)
	CreateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionRepresentation) error
	UpdateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.ModifyAuthenticationExecutionRepresentation) error
	CreateAuthExecutionFlow(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionFlowRepresentation) error
}
