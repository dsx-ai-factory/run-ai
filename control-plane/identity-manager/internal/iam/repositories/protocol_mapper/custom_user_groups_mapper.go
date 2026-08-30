package protocol_mapper

import (
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

func NewRunaiGroupsProtocolMapper(idTokenGroupsClaim bool, tenantId string) gocloak.ProtocolMapperRepresentation {
	config := buildGroupsMapperConfig("groups", tenantId)
	config["id.token.claim"] = strconv.FormatBool(idTokenGroupsClaim)

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP("Runai Groups Mapper"),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("runai-oidc-custom-groups-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
