package models

import (
	"reflect"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	m.Run()
}

const (
	URL       = "https://www.example-runai123.com/"
	idp_alias = "alias"
)

func TestOCP_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OCP)
	clientId := "test"
	idpBaseUrl := URL
	clientSecret := "clientSecret"
	defaultScope := "defaultScope1 defaultScope2"
	mapperValues := getMapWithValues()

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			ClientId:     clientId,
			IdpBaseUrl:   idpBaseUrl,
			ClientSecret: clientSecret,
			Scopes:       []string{"defaultScope1", "defaultScope2"},
		},
		Mappers: api.Mappers{
			Groups:    "groups_2",
			Email:     "email_1",
			FirstName: "firstName_1",
			LastName:  "lastName_1",
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := &map[string]string{
		"baseUrl":      idpBaseUrl,
		"clientId":     clientId,
		"clientSecret": clientSecret,
		"defaultScope": defaultScope,
	}

	assert.Equal(t, data.Alias, name)
	assert.Equal(t, "", data.Resource)
	assert.True(t, reflect.DeepEqual(data.Config, config))
	assert.True(t, reflect.DeepEqual(data.Mappers, mapperValues))
}

func TestOCP_HappyMissingInputFlow(t *testing.T) {
	name := ""
	rType := string(OCP)
	idpBaseUrl := URL
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope1"}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			IdpBaseUrl:   idpBaseUrl,
			ClientSecret: clientSecret,
			Scopes:       defaultScope,
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := &map[string]string{
		"baseUrl":      idpBaseUrl,
		"clientId":     utils.RUNAI_CLIENT_ID,
		"clientSecret": clientSecret,
		"defaultScope": "defaultScope1",
	}

	assert.True(t, strings.HasPrefix(data.Alias, rType+"-"))
	assert.Equal(t, "", data.Resource)
	assert.True(t, reflect.DeepEqual(data.Config, config))
	mapsValues := getEmptyMap()
	assert.True(t, reflect.DeepEqual(data.Mappers, mapsValues))
}

func TestOCP_BadType(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			IdpBaseUrl:   "idpBaseUrl",
			ClientSecret: "clientSecret",
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestOCP_BadData(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestOIDC_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)
	clientId := "test"
	discoverDocumentUrl := URL
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope1", "defaultScope2"}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverDocumentUrl,
			Scopes:              defaultScope,
		},
		Mappers: mappers,
	}
	data, err := GenerateIdpData(name, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":            clientId,
		"clientSecret":        clientSecret,
		"discoverDocumentUrl": discoverDocumentUrl,
		"defaultScope":        "defaultScope1 defaultScope2",
	}

	assert.Equal(t, name, data.Alias)
	assert.Equal(t, discoverDocumentUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, MapperObjectToMap(mappers), data.Mappers)
}

func TestOIDC_Update_WithoutUrl_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)
	clientId := "test"
	discoverDocumentUrl := ""
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope1"}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverDocumentUrl,
			Scopes:              defaultScope,
		},
		Mappers: mappers,
	}
	data, err := GenerateIdpData(name, request, true)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":     clientId,
		"clientSecret": clientSecret,
		"defaultScope": "defaultScope1",
	}

	assert.Equal(t, name, data.Alias)
	assert.Equal(t, discoverDocumentUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, MapperObjectToMap(mappers), data.Mappers)
}

func TestOIDC_Update_With_Bad_Scopes(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)
	clientId := "test"
	discoverDocumentUrl := ""
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope1```", "defaultScope2"}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverDocumentUrl,
			Scopes:              defaultScope,
		},
		Mappers: mappers,
	}
	_, err := GenerateIdpData(name, request, true)
	assert.NotNil(t, err)
}

func TestOIDC_Update_WithUrl_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)
	clientId := "test"
	discoverDocumentUrl := URL
	clientSecret := "clientSecret"
	defaultScope := []string{"scopes"}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverDocumentUrl,
			Scopes:              defaultScope,
		},
		Mappers: mappers,
	}
	data, err := GenerateIdpData(name, request, true)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":            clientId,
		"clientSecret":        clientSecret,
		"discoverDocumentUrl": discoverDocumentUrl,
		"defaultScope":        "scopes",
	}

	assert.Equal(t, name, data.Alias)
	assert.Equal(t, discoverDocumentUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, MapperObjectToMap(mappers), data.Mappers)
}

func TestOCP_Update_WithUrl_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OCP)
	clientId := "test"
	idpBaseUrl := URL
	clientSecret := "clientSecret"
	defaultScope := []string{""}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			IdpBaseUrl:   idpBaseUrl,
			Scopes:       defaultScope,
		},
		Mappers: mappers,
	}
	data, err := GenerateIdpData(name, request, true)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":     clientId,
		"clientSecret": clientSecret,
		"baseUrl":      idpBaseUrl,
		"defaultScope": "",
	}

	assert.Equal(t, name, data.Alias)
	assert.Equal(t, "", data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, MapperObjectToMap(mappers), data.Mappers)
}

func TestOCP_Update_WithoutUrl_HappyFlow(t *testing.T) {
	name := idp_alias
	rType := string(OCP)
	clientId := "test"
	idpBaseUrl := ""
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope1", "defaultScope2"}
	mappers := api.Mappers{
		Groups: "groups_2",
		Email:  "email_1",
	}

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			IdpBaseUrl:   idpBaseUrl,
			Scopes:       defaultScope,
		},
		Mappers: mappers,
	}
	data, err := GenerateIdpData(name, request, true)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":     clientId,
		"clientSecret": clientSecret,
		"defaultScope": "defaultScope1 defaultScope2",
	}

	assert.Equal(t, name, data.Alias)
	assert.Equal(t, idpBaseUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, MapperObjectToMap(mappers), data.Mappers)
}

func TestOIDC_HappyMissingInputFlow(t *testing.T) {
	rType := string(OIDC)
	name := ""
	clientId := "test"
	discoverDocumentUrl := URL
	clientSecret := "clientSecret"
	defaultScope := []string{"defaultScope2"}

	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverDocumentUrl,
			Scopes:              defaultScope,
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := map[string]string{
		"clientId":            clientId,
		"clientSecret":        clientSecret,
		"discoverDocumentUrl": discoverDocumentUrl,
		"defaultScope":        "defaultScope2",
	}

	assert.True(t, strings.HasPrefix(data.Alias, rType+"-"))
	assert.Equal(t, discoverDocumentUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	mappers := getEmptyMap()
	assert.Equal(t, mappers, data.Mappers)
}

func TestOIDC_MissingclientSecretInputFlow(t *testing.T) {
	rType := string(OIDC)
	name := ""
	clientId := "test"
	discoverDocumentUrl := URL

	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(rType),
		OidcData: &api.OidcCreationData{
			ClientId:            clientId,
			DiscoverDocumentUrl: discoverDocumentUrl,
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestOIDC_BadData(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)

	request := api.IdpCreationRequest{
		Name:     name,
		Type:     api.IdpCreationRequestTypeEnum(rType),
		SamlData: nil,
		OidcData: nil,
		OcpData: &api.OcpCreationData{
			IdpBaseUrl:   URL,
			ClientSecret: "clientSecret",
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestOIDC_BadType(t *testing.T) {
	name := idp_alias
	rType := string(SAML)

	request := api.IdpCreationRequest{
		Name:     name,
		Type:     api.IdpCreationRequestTypeEnum(rType),
		SamlData: nil,
		OidcData: &api.OidcCreationData{},
		OcpData:  nil,
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestSAML_HappyFlow_URL(t *testing.T) {
	entityId := "entityId"
	mappers := getEmptyMap()

	request := api.IdpCreationRequest{
		Name: idp_alias,
		Type: api.IdpCreationRequestTypeEnum(SAML),
		SamlData: &api.SamlCreationData{
			MetadataXmlUrl:  URL,
			EntityId:        entityId,
			MetadataXmlFile: "",
			FileName:        "",
			MetadataXmlType: utils.Url,
		},
	}
	data, err := GenerateIdpData(idp_alias, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := getBasicSamlConfig(URL, "entityId", "", utils.Url, utils.Url)

	assert.Equal(t, idp_alias, data.Alias)
	assert.Equal(t, URL, data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, mappers, data.Mappers)
}

func TestSAML_HappyFlow_File(t *testing.T) {
	entityId := "entityId"
	mappers := getEmptyMap()

	request := api.IdpCreationRequest{
		Name: idp_alias,
		Type: api.IdpCreationRequestTypeEnum(SAML),
		SamlData: &api.SamlCreationData{
			MetadataXmlUrl:  "",
			EntityId:        entityId,
			MetadataXmlFile: "file",
			FileName:        "file-name",
			MetadataXmlType: utils.File,
		},
	}
	data, err := GenerateIdpData(idp_alias, request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := getBasicSamlConfig("", "entityId", "file", "file-name", utils.File)

	assert.Equal(t, idp_alias, data.Alias)
	assert.Equal(t, "file", data.Resource)
	assert.Equal(t, config, *data.Config)
	assert.Equal(t, mappers, data.Mappers)
}

func TestSAML_HappyMissingInputFlow(t *testing.T) {
	metadataXmlUrl := URL
	entityId := "entityId"

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(SAML),
		SamlData: &api.SamlCreationData{MetadataXmlUrl: metadataXmlUrl, EntityId: entityId, MetadataXmlType: utils.Url},
	}
	data, err := GenerateIdpData("", request, false)
	assert.NotNil(t, data)
	assert.Nil(t, err)

	config := getBasicSamlConfig(metadataXmlUrl, "entityId", "", utils.Url, utils.Url)

	assert.True(t, strings.HasPrefix(data.Alias, string(SAML)+"-"))
	assert.Equal(t, metadataXmlUrl, data.Resource)
	assert.Equal(t, config, *data.Config)
	mapsValues := getEmptyMap()
	assert.Equal(t, mapsValues, data.Mappers)
}

func TestSAML_MissingInputFlow(t *testing.T) {
	rType := string(OIDC)
	name := ""

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(rType),
		SamlData: &api.SamlCreationData{},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestSAML_BadType(t *testing.T) {
	name := idp_alias
	rType := string(OIDC)

	request := api.IdpCreationRequest{
		Name:     name,
		Type:     api.IdpCreationRequestTypeEnum(rType),
		SamlData: &api.SamlCreationData{},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func TestSAML_BadData(t *testing.T) {
	name := idp_alias
	rType := string(SAML)

	request := api.IdpCreationRequest{
		Name: name,
		Type: api.IdpCreationRequestTypeEnum(rType),
		OcpData: &api.OcpCreationData{
			IdpBaseUrl:   "idpBaseUrl",
			ClientSecret: "clientSecret",
		},
	}
	data, err := GenerateIdpData(name, request, false)
	assert.Nil(t, data)
	assert.NotNil(t, err)
}

func Test_MappersEmpty(t *testing.T) {
	api := api.Mappers{
		Groups:    "groups_2",
		Email:     "email_1",
		FirstName: "firstName_1",
		LastName:  "lastName_1",
	}
	mapsValues := getMapWithValues()
	maps := MapperObjectToMap(api)
	assert.Equal(t, maps, mapsValues)
}

func getMapWithValues() map[string]string {
	return map[string]string{
		"email":               "email_1",
		"GROUPS":              "groups_2",
		"GID":                 "",
		"SUPPLEMENTARYGROUPS": "",
		"UID":                 "",
		"firstName":           "firstName_1",
		"lastName":            "lastName_1",
	}
}

func getEmptyMap() map[string]string {
	return map[string]string{
		"email":               "",
		"GROUPS":              "",
		"GID":                 "",
		"SUPPLEMENTARYGROUPS": "",
		"UID":                 "",
		"firstName":           "",
		"lastName":            "",
	}
}

func getBasicSamlConfig(metadataXmlUrl, entityId, metadataXmlFile, fileName, metadataXmlType string) map[string]string {
	return map[string]string{
		utils.MetadataXmlUrl:  metadataXmlUrl,
		utils.EntityId:        entityId,
		utils.MetadataXmlFile: metadataXmlFile,
		utils.FileName:        fileName,
		utils.MetadataXmlType: metadataXmlType,
	}
}
