package protocol_mapper

import (
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

type UserAttributeMapperConfig struct {
	MapperName        string
	ClaimName         string
	ClaimType         ClaimType
	UserAttribute     string
	Multivalued       bool
	ShowInIdToken     bool
	ShowInUserInfo    bool
	ShowInAccessToken bool
}

func NewGIDProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserAttributeProtocolMapper(UserAttributeMapperConfig{
		MapperName:        "GID Mapper",
		ClaimName:         "gid",
		ClaimType:         StringClaim,
		UserAttribute:     utils.GID,
		Multivalued:       false,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewSupplementaryGroupsProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserAttributeProtocolMapper(UserAttributeMapperConfig{
		MapperName:        "SupplementaryGroups Mapper",
		ClaimName:         "supplementarygroups",
		ClaimType:         IntClaim,
		UserAttribute:     utils.SupplementaryGroups,
		Multivalued:       true,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewUIDProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserAttributeProtocolMapper(UserAttributeMapperConfig{
		MapperName:        "UID Mapper",
		ClaimName:         "uid",
		ClaimType:         StringClaim,
		UserAttribute:     utils.UID,
		Multivalued:       false,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func newUserAttributeProtocolMapper(mapperConfig UserAttributeMapperConfig) gocloak.ProtocolMapperRepresentation {
	config := map[string]string{
		"claim.name":           mapperConfig.ClaimName,
		"user.attribute":       mapperConfig.UserAttribute,
		"jsonType.label":       mapperConfig.ClaimType,
		"multivalued":          strconv.FormatBool(mapperConfig.Multivalued),
		"id.token.claim":       strconv.FormatBool(mapperConfig.ShowInIdToken),
		"access.token.claim":   strconv.FormatBool(mapperConfig.ShowInAccessToken),
		"userinfo.token.claim": strconv.FormatBool(mapperConfig.ShowInUserInfo),
	}
	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP(mapperConfig.MapperName),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("oidc-usermodel-attribute-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
