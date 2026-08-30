package identity_provider

import (
	"context"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
)

//go:generate mockery --name IdpRepresentationer
type IdpRepresentationer interface {
	GetRepresentation(ctx context.Context, data *models.IdpData, config *map[string]string) *iamModels.Idp
	GetMappersRepresentation(ctx context.Context, alias string, mappers map[string]string) []*iamModels.IdpMapper
}
