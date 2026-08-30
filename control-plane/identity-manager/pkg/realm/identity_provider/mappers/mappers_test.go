package mappers

import (
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/providers"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestOcpUserAttributeMapper(t *testing.T) {
	mapper := NewOcpIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, ptr.To(value))
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, value, (*configMap)["jsonField"])
	assert.Equal(t, attribute, (*configMap)["userAttribute"])
	assert.Equal(t, 3, len(*configMap))
}

func TestOcpUserAttributeMapper_nil_value(t *testing.T) {
	mapper := NewOcpIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, nil)
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)["jsonField"])
	assert.Equal(t, attribute, (*configMap)["userAttribute"])
	assert.Equal(t, 3, len(*configMap))
}

func TestOidcUserAttributeMapper(t *testing.T) {
	mapper := NewOidcIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, ptr.To(value))
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, value, (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestOidcUserAttributeMapper_nil_value(t *testing.T) {
	mapper := NewOidcIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, nil)
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestOidcUserAttributeMapper_empty_value(t *testing.T) {
	mapper := NewOidcIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, ptr.To(""))
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestSamlUserAttributeMapper(t *testing.T) {
	mapper := NewSamlIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, ptr.To(value))
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, value, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}

func TestSamlUserAttributeMapper_nil_value(t *testing.T) {
	mapper := NewSamlIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, nil)
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}

func TestSamlUserAttributeMapper_empty_value(t *testing.T) {
	mapper := NewSamlIdpMapperMapper()
	attributeMapper := models.IdpAttributeMapper{
		Name:      "a",
		Attribute: attribute,
		Value:     attribute,
	}
	m := mapper.GenerateUserAttributeMapper("alias", attributeMapper, ptr.To(""))
	configMap := m.Config
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}
