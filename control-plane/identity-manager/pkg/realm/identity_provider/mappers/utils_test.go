package mappers

import (
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/providers"
	"github.com/stretchr/testify/assert"
)

const (
	attribute = "attribute"
	value     = "value"
)

func TestOcpUserAttributeConfig(t *testing.T) {
	configMap := ocpUserAttributeConfig(attribute, value)
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, value, (*configMap)["jsonField"])
	assert.Equal(t, attribute, (*configMap)["userAttribute"])
	assert.Equal(t, 3, len(*configMap))
}

func TestOcpUserAttributeConfig_nil_value(t *testing.T) {
	configMap := ocpUserAttributeConfig(attribute, "")
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, "", (*configMap)["jsonField"])
	assert.Equal(t, attribute, (*configMap)["userAttribute"])
	assert.Equal(t, 3, len(*configMap))
}

func TestOidcUserAttributeConfig(t *testing.T) {
	configMap := oidcUserAttributeConfig(attribute, value)
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, value, (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestOidcUserAttributeConfig_nil_value(t *testing.T) {
	configMap := oidcUserAttributeConfig(attribute, "")
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, "", (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestOidcUserAttributeConfig_empty_value(t *testing.T) {
	configMap := oidcUserAttributeConfig(attribute, "")
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, attribute, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, "", (*configMap)[providers.Claim])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 5, len(*configMap))
}

func TestSamlUserAttributeConfig(t *testing.T) {
	configMap := samlUserAttributeConfig(attribute, value)
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, value, (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}

func TestSamlUserAttributeConfig_nil_value(t *testing.T) {
	configMap := samlUserAttributeConfig(attribute, "")
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, "", (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}

func TestSamlUserAttributeConfig_empty_value(t *testing.T) {
	configMap := samlUserAttributeConfig(attribute, "")
	assert.Equal(t, "FORCE", (*configMap)["syncMode"])
	assert.Equal(t, attribute, (*configMap)[providers.MapperUserAttribute])
	assert.Equal(t, "", (*configMap)[providers.MapperAttributeName])
	assert.Equal(t, attribute, (*configMap)[providers.AttributeFriendlyName])
	assert.Equal(t, 4, len(*configMap))
}
