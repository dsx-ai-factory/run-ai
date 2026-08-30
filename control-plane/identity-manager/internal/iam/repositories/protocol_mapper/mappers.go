package protocol_mapper

import (
	"strconv"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

type ClaimType = string

const (
	IntClaim    ClaimType = "int"
	StringClaim ClaimType = "String"
)

func buildBaseTokenClaimsConfig(claimName string, idTokenClaim bool) map[string]string {
	return map[string]string{
		"claim.name":                claimName,
		"id.token.claim":            strconv.FormatBool(idTokenClaim),
		"access.token.claim":        strconv.FormatBool(true),
		"userinfo.token.claim":      strconv.FormatBool(true),
		"lightweight.claim":         strconv.FormatBool(true),
		"introspection.token.claim": strconv.FormatBool(false),
	}
}

func buildAccessKeyAttributeMapperConfig(claimName, userAttribute, tenantId string, multivalued, intersectionEnabled bool) map[string]string {
	config := buildBaseTokenClaimsConfig(claimName, true)
	config["groups.user.attribute"] = userAttribute
	config["multivalued"] = strconv.FormatBool(multivalued)
	config["groups.intersection.enabled"] = strconv.FormatBool(intersectionEnabled)
	config["tenant.id"] = tenantId
	return config
}

func buildGroupsMapperConfig(claimName, tenantId string) map[string]string {
	return buildAccessKeyAttributeMapperConfig(claimName, utils.Groups, tenantId, true, true)
}
