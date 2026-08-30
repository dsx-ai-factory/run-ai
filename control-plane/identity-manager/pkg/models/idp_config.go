package models

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ProviderID string

const (
	Invalid ProviderID = ""
	OIDC    ProviderID = "oidc"
	SAML    ProviderID = "saml"
	OCP     ProviderID = "openshift-v4"
)

var (
	ExternalUser = IdpAttributeMapper{utils.IdpExternalUserMapperName, utils.IS_LOCAL_USER, "false"}
	OCPEmail     = IdpAttributeMapper{utils.IdpEmailMapperName, "email", "metadata.name"}
	OCPUid       = IdpAttributeMapper{utils.IdpUidMapperName, "UID", "metadata.uid"}
	OCPGid       = IdpAttributeMapper{utils.IdpGidMapperName, "GID", "metadata.gid"}
	OCPSupGroup  = IdpAttributeMapper{utils.IdpSupGroupsMapperName, "SUPPLEMENTARYGROUPS", "supplementarygroups"}
	OCPGroups    = IdpAttributeMapper{utils.IdpGroupsMapperName, "GROUPS", "groups"}
	Email        = IdpAttributeMapper{utils.IdpEmailMapperName, "email", "email"}
	Gid          = IdpAttributeMapper{utils.IdpGidMapperName, "GID", "GID"}
	Uid          = IdpAttributeMapper{utils.IdpUidMapperName, "UID", "UID"}
	FirstName    = IdpAttributeMapper{utils.IdpFirstNameMapperName, "firstName", "given_name"}
	LastName     = IdpAttributeMapper{utils.IdpLastNameMapperName, "lastName", "family_name"}
	SupGroup     = IdpAttributeMapper{utils.IdpSupGroupsMapperName, "SUPPLEMENTARYGROUPS", "SUPPLEMENTARYGROUPS"}
	Upn          = IdpAttributeMapper{utils.IdpUpnMapperName, "upn", "upn"}
	Groups       = IdpAttributeMapper{utils.IdpGroupsMapperName, "GROUPS", "GROUPS"}
)

type IdpAttributeMapper struct {
	Name      string
	Attribute string
	Value     string
}

type IdpData struct {
	Alias        string
	Idp          ProviderID
	Resource     string
	ResourceType string
	ResourceName string
	Mappers      map[string]string
	Config       *map[string]string
}

func GenerateIdpData(name string, request api.IdpCreationRequest, isUpdate bool) (*IdpData, error) {
	res := IdpData{}
	res.Idp = getRequestedProvider(request)
	if name == "" {
		if hash, ok := utils.ShortHash(); ok {
			name = string(request.Type) + "-" + hash
		} else {
			logging.Warnf("failed to generate random hash for IDP alias, using type-only alias")
			name = string(request.Type)
		}
	}
	res.Alias = name
	res.Mappers = MapperObjectToMap(request.Mappers)

	switch res.Idp {
	case Invalid:
		return nil, fmt.Errorf("unable to determin idp configuration")
	case SAML:
		setSamlData(&res, request)
	case OIDC:
		err := validateIdpScope(request.OidcData.Scopes)
		if err != nil {
			return nil, err
		}
		if isUpdate {
			updateOidcData(&res, request)
		} else {
			setOidcData(&res, request)
		}
	case OCP:
		err := validateIdpScope(request.OcpData.Scopes)
		if err != nil {
			return nil, err
		}
		if isUpdate {
			updateOcpData(&res, request)
		} else {
			setOcpData(&res, request)
		}
	}
	if res.Config == nil {
		return nil, fmt.Errorf("invalid %s configuration", string(res.Idp))
	}
	return &res, nil
}

func validateIdpScope(scopes []string) error {
	if len(scopes) > 0 && !validateScopes(scopes) {
		return fmt.Errorf("invalid scopes %s, scopes should be in this format %s", scopesToString(scopes), utils.ScopeRegex)
	}
	return nil
}

func setSamlData(res *IdpData, request api.IdpCreationRequest) {
	res.Config = getSamlConfig(request.SamlData)
	res.ResourceType = string(request.SamlData.MetadataXmlType)
	if res.ResourceType == utils.File {
		res.Resource = request.SamlData.MetadataXmlFile
		res.ResourceName = request.SamlData.FileName
	} else {
		res.Resource = request.SamlData.MetadataXmlUrl
		res.ResourceName = utils.Url
	}
}

func setOidcData(res *IdpData, request api.IdpCreationRequest) {
	res.Config = getOidcConfig(request.OidcData)
	res.Resource = request.OidcData.DiscoverDocumentUrl
	res.ResourceName = utils.Url
	res.ResourceType = utils.Url
}

func updateOidcData(res *IdpData, request api.IdpCreationRequest) {
	res.Config = getOidcConfigForUpdate(request.OidcData)
	if request.OidcData.DiscoverDocumentUrl != "" {
		res.Resource = request.OidcData.DiscoverDocumentUrl
		res.ResourceName = utils.Url
		res.ResourceType = utils.Url
	}
}

func setOcpData(res *IdpData, request api.IdpCreationRequest) {
	res.Config = getOcpConfig(request.OcpData)
}

func updateOcpData(res *IdpData, request api.IdpCreationRequest) {
	res.Config = getOcpConfigForUpdate(request.OcpData)
}

func getRequestedProvider(request api.IdpCreationRequest) ProviderID {
	switch request.Type {
	case api.IdpCreationRequestTypeEnum(OIDC):
		if request.OidcData != nil {
			return OIDC
		}
	case api.IdpCreationRequestTypeEnum(SAML):
		if request.SamlData != nil {
			return SAML
		}
	case api.IdpCreationRequestTypeEnum(OCP):
		if request.OcpData != nil {
			return OCP
		}
	}
	return Invalid
}

func getOcpConfig(data *api.OcpCreationData) *map[string]string {
	if !utils.IsValidURL(data.IdpBaseUrl) || data.ClientSecret == "" {
		return nil
	}
	if data.ClientId == "" {
		data.ClientId = utils.RUNAI_CLIENT_ID
	}
	return &map[string]string{
		utils.IdpBaseUrl:       data.IdpBaseUrl,
		utils.Ocp_ClientId:     data.ClientId,
		utils.Ocp_ClientSecret: data.ClientSecret,
		utils.Scopes:           scopesToString(data.Scopes),
	}
}

func getOcpConfigForUpdate(data *api.OcpCreationData) *map[string]string {
	if data.ClientId == "" || data.ClientSecret == "" || data.IdpBaseUrl != "" && !utils.IsValidURL(data.IdpBaseUrl) {
		return nil
	}
	if data.IdpBaseUrl != "" {
		return &map[string]string{
			utils.Ocp_ClientId:     data.ClientId,
			utils.Ocp_ClientSecret: data.ClientSecret,
			utils.IdpBaseUrl:       data.IdpBaseUrl,
			utils.Scopes:           scopesToString(data.Scopes),
		}
	} else {
		return &map[string]string{
			utils.Ocp_ClientId:     data.ClientId,
			utils.Ocp_ClientSecret: data.ClientSecret,
			utils.Scopes:           scopesToString(data.Scopes),
		}
	}
}

func getSamlConfig(data *api.SamlCreationData) *map[string]string {
	if !utils.IsValidURL(data.MetadataXmlUrl) && data.MetadataXmlFile == "" {
		return nil
	}
	resourceName := utils.Url
	if data.MetadataXmlType == utils.File {
		resourceName = data.FileName
	}
	return &map[string]string{
		utils.MetadataXmlUrl:  data.MetadataXmlUrl,
		utils.MetadataXmlFile: data.MetadataXmlFile,
		utils.EntityId:        data.EntityId,
		utils.FileName:        resourceName,
		utils.MetadataXmlType: string(data.MetadataXmlType),
	}
}

func getOidcConfig(data *api.OidcCreationData) *map[string]string {
	if data.ClientId == "" || data.ClientSecret == "" || !utils.IsValidURL(data.DiscoverDocumentUrl) {
		return nil
	}
	return &map[string]string{
		utils.Oidc_ClientId:       data.ClientId,
		utils.Oidc_ClientSecret:   data.ClientSecret,
		utils.DiscoverDocumentUrl: data.DiscoverDocumentUrl,
		utils.Scopes:              scopesToString(data.Scopes),
	}
}

func getOidcConfigForUpdate(data *api.OidcCreationData) *map[string]string {
	if data.ClientId == "" || data.ClientSecret == "" || data.DiscoverDocumentUrl != "" && !utils.IsValidURL(data.DiscoverDocumentUrl) {
		return nil
	}
	if data.DiscoverDocumentUrl != "" {
		return &map[string]string{
			utils.Oidc_ClientId:       data.ClientId,
			utils.Oidc_ClientSecret:   data.ClientSecret,
			utils.DiscoverDocumentUrl: data.DiscoverDocumentUrl,
			utils.Scopes:              scopesToString(data.Scopes),
		}
	} else {
		return &map[string]string{
			utils.Oidc_ClientId:     data.ClientId,
			utils.Oidc_ClientSecret: data.ClientSecret,
			utils.Scopes:            scopesToString(data.Scopes),
		}
	}
}

func MapperObjectToMap(mapper api.Mappers) map[string]string {
	return map[string]string{
		Email.Attribute:     mapper.Email,
		Uid.Attribute:       mapper.Uid,
		Gid.Attribute:       mapper.Gid,
		Groups.Attribute:    mapper.Groups,
		SupGroup.Attribute:  mapper.SupplementaryGroups,
		FirstName.Attribute: mapper.FirstName,
		LastName.Attribute:  mapper.LastName,
	}
}

func scopesToString(scope []string) string {
	return strings.Join(scope, " ")
}

func validateScope(s string) bool {
	if s == "" {
		return true
	}
	regex := regexp.MustCompile(utils.ScopeRegex)
	return regex.MatchString(s)
}

func validateScopes(scope []string) bool {
	for _, s := range scope {
		if !validateScope(s) {
			return false
		}
	}
	return true
}
