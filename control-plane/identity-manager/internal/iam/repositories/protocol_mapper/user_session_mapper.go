package protocol_mapper

import (
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

type UserSessionMapperConfig struct {
	MapperName        string
	ClaimName         string
	ClaimType         ClaimType
	UserSessionNote   string
	ShowInIdToken     bool
	ShowInAccessToken bool
}

func NewSubToClientIdProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserSessionMapper(UserSessionMapperConfig{
		MapperName:        "SubToClientId Mapper",
		ClaimName:         "sub",
		ClaimType:         StringClaim,
		UserSessionNote:   "clientId",
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewEmailToClientIdProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserSessionMapper(UserSessionMapperConfig{
		MapperName:        "EmailToClientId Mapper",
		ClaimName:         "email",
		ClaimType:         StringClaim,
		UserSessionNote:   "clientId",
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewIdentityProviderMapper() gocloak.ProtocolMapperRepresentation {
	return newUserSessionMapper(UserSessionMapperConfig{
		MapperName:        "IdentityProvider Mapper",
		ClaimName:         "identity_provider",
		ClaimType:         StringClaim,
		UserSessionNote:   "identity_provider",
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func newUserSessionMapper(mapperConfig UserSessionMapperConfig) gocloak.ProtocolMapperRepresentation {
	config := map[string]string{
		"claim.name":         mapperConfig.ClaimName,
		"user.session.note":  mapperConfig.UserSessionNote,
		"jsonType.label":     mapperConfig.ClaimType,
		"id.token.claim":     strconv.FormatBool(mapperConfig.ShowInIdToken),
		"access.token.claim": strconv.FormatBool(mapperConfig.ShowInAccessToken),
	}

	return gocloak.ProtocolMapperRepresentation{
		ProtocolMapper:  gocloak.StringP("oidc-usersessionmodel-note-mapper"),
		Name:            gocloak.StringP(mapperConfig.MapperName),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
