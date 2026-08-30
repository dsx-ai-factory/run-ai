package identity_provider

import (
	"context"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/identity_provider/mappers"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type oidcIdentityProviderRepresentation struct {
	base         IdentityProviderRepresentation
	mapperMapper mappers.UserAttributeMapper
}

func NewOidcIdentityProvider() IdpRepresentationer {
	return &oidcIdentityProviderRepresentation{
		base: IdentityProviderRepresentation{
			ProviderID:                models.OIDC,
			FirstBrokerLoginFlowAlias: utils.LOGIN_FLOW_NAME,
			AddReadTokenRoleOnCreate:  false,
			hardCodedMappers:          models.ExternalUser,
			attributeMappers:          []models.IdpAttributeMapper{models.Email, models.Upn, models.Groups, models.SupGroup, models.Uid, models.Gid, models.FirstName, models.LastName},
		},
		mapperMapper: mappers.NewOidcIdpMapperMapper(),
	}
}

func (oip *oidcIdentityProviderRepresentation) GetRepresentation(ctx context.Context, data *models.IdpData, config *map[string]string) *iamModels.Idp {
	if data.Idp != oip.base.ProviderID {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot update Idp type from %s to %s", string(data.Idp), oip.base.ProviderID)
		return nil
	}

	(*config)["clientId"] = (*data.Config)["clientId"]
	(*config)["clientSecret"] = (*data.Config)["clientSecret"]
	(*config)["syncMode"] = "IMPORT"
	(*config)["clientAuthMethod"] = "client_secret_post"

	identityProvider := basicIdpRep(data.Alias, &oip.base, config)

	return &identityProvider
}

func (oip *oidcIdentityProviderRepresentation) GetMappersRepresentation(ctx context.Context, alias string, values map[string]string) []*iamModels.IdpMapper {
	idpMappers := []*iamModels.IdpMapper{
		mappers.GenerateHardCodedMapper(alias, oip.base.hardCodedMappers),
	}
	return append(idpMappers, mappers.GenerateUserAttributeMappers(alias, oip.base.attributeMappers, values, oip.mapperMapper)...)
}
