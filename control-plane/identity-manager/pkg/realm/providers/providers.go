package providers

import (
	"context"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	OIDC string = "oidc"
	SAML string = "saml"
	OCP  string = "openshift-v4"
)

const (
	MapperGroups              string = "Groups Mapper"
	MapperUid                 string = "UID Mapper"
	MapperGid                 string = "GID Mapper"
	MapperSupplementaryGroups string = "SupplementaryGroups Mapper"
	MapperAttributeName       string = "attribute.name"
	MapperUserAttribute       string = "user.attribute"
	MapperValueOcp            string = "jsonField"
	Claim                     string = "claim"
	AttributeFriendlyName     string = "attribute.friendly.name"
	MapperUserAttributeOcp    string = "userAttribute"
)

func IsMapperConfigurable(mapper *iamModels.IdpMapper) bool {
	if mapper.Name == nil {
		return false
	}

	switch *mapper.Name {
	case models.Email.Name, models.SupGroup.Name, models.Upn.Name, models.Groups.Name,
		models.Uid.Name, models.Gid.Name, models.FirstName.Name, models.LastName.Name,
		MapperGroups, MapperUid, MapperGid, MapperSupplementaryGroups:
		return true
	}
	attribute := (*mapper.Config)[MapperAttributeName]
	switch attribute {
	case models.Email.Attribute, models.SupGroup.Attribute, models.Upn.Attribute, models.Groups.Attribute, models.Uid.Attribute, models.Gid.Attribute, models.FirstName.Attribute, models.LastName.Attribute:
		return true
	}

	return false
}

func GetUserAttributeName(ctx context.Context, identityProviderId *string) string {
	if identityProviderId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing identity provider to map to user attribute.")
		return ""
	}
	if *identityProviderId == OIDC {
		return Claim
	}
	if *identityProviderId == SAML {
		return MapperAttributeName
	}
	if *identityProviderId == OCP {
		return MapperValueOcp
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("missing mapping to user attribute for identity provider: %v.", identityProviderId)
	return ""
}

func GetUserAttributeKey(ctx context.Context, identityProviderId *string) string {
	if identityProviderId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing identity provider to map to user attribute key.")
		return ""
	}
	if *identityProviderId == OIDC {
		return MapperAttributeName
	}
	if *identityProviderId == SAML {
		return MapperUserAttribute
	}
	if *identityProviderId == OCP {
		return MapperUserAttributeOcp
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("missing mapping to user attribute key for identity provider: %v.", identityProviderId)
	return ""
}
