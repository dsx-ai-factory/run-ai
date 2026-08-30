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

type samlIdentityProviderRepresentation struct {
	base            IdentityProviderRepresentation
	mapperMapper    mappers.UserAttributeMapper
	defaultEntityId string
}

func NewSamlIdentityProvider(defaultEntityId string) IdpRepresentationer {
	return &samlIdentityProviderRepresentation{
		IdentityProviderRepresentation{
			ProviderID:                models.SAML,
			FirstBrokerLoginFlowAlias: utils.LOGIN_FLOW_NAME,
			AddReadTokenRoleOnCreate:  false,
			hardCodedMappers:          models.ExternalUser,
			attributeMappers:          []models.IdpAttributeMapper{models.Email, models.Upn, models.Groups, models.SupGroup, models.Uid, models.Gid, models.FirstName, models.LastName},
		},
		mappers.NewSamlIdpMapperMapper(),
		defaultEntityId,
	}
}

func (sip *samlIdentityProviderRepresentation) GetRepresentation(ctx context.Context, data *models.IdpData, config *map[string]string) *iamModels.Idp {
	if data.Idp != sip.base.ProviderID {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot update Idp type from %s to %s", string(data.Idp), sip.base.ProviderID)
		return nil
	}
	entityId := (*data.Config)[utils.EntityId]
	if entityId == "" {
		entityId = sip.defaultEntityId
	}
	(*config)[utils.EntityId] = entityId
	(*config)["enabledFromMetadata"] = "true"
	(*config)["principalType"] = "SUBJECT"

	identityProvider := basicIdpRep(data.Alias, &sip.base, config)

	return &identityProvider
}

func (sip *samlIdentityProviderRepresentation) GetMappersRepresentation(ctx context.Context, alias string, values map[string]string) []*iamModels.IdpMapper {
	idpMappers := []*iamModels.IdpMapper{
		mappers.GenerateHardCodedMapper(alias, sip.base.hardCodedMappers),
	}
	return append(idpMappers, mappers.GenerateUserAttributeMappers(alias, sip.base.attributeMappers, values, sip.mapperMapper)...)
}
