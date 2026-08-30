package mappers

import "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/providers"

func ocpUserAttributeConfig(attribute string, value string) *map[string]string {
	return &map[string]string{
		"syncMode":      "FORCE",
		"jsonField":     value,
		"userAttribute": attribute,
	}
}

func oidcUserAttributeConfig(attribute string, value string) *map[string]string {
	return &map[string]string{
		providers.MapperUserAttribute:   attribute,
		"syncMode":                      "FORCE",
		providers.MapperAttributeName:   attribute,
		providers.Claim:                 value,
		providers.AttributeFriendlyName: attribute,
	}
}

func samlUserAttributeConfig(attribute string, value string) *map[string]string {
	return &map[string]string{
		providers.MapperAttributeName:   value,
		"syncMode":                      "FORCE",
		providers.MapperUserAttribute:   attribute,
		providers.AttributeFriendlyName: attribute,
	}
}
