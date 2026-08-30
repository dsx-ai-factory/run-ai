package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestGetUserAttributeName_SAML(t *testing.T) {
	attributeName := GetUserAttributeName(context.Background(), ptr.To(SAML))
	assert.Equal(t, MapperAttributeName, attributeName)
}

func TestGetUserAttributeName_OIDC(t *testing.T) {
	attributeName := GetUserAttributeName(context.Background(), ptr.To(OIDC))
	assert.Equal(t, Claim, attributeName)
}

func TestGetUserAttributeName_OCP(t *testing.T) {
	attributeName := GetUserAttributeName(context.Background(), ptr.To(OCP))
	assert.Equal(t, MapperValueOcp, attributeName)
}

func TestGetUserAttributeName_nil_idp(t *testing.T) {
	attributeName := GetUserAttributeName(context.Background(), nil)
	assert.Equal(t, "", attributeName)
}

func TestGetUserAttributeKey_SAML(t *testing.T) {
	attributeKey := GetUserAttributeKey(context.Background(), ptr.To(SAML))
	assert.Equal(t, MapperUserAttribute, attributeKey)
}

func TestGetUserAttributeKey_OIDC(t *testing.T) {
	attributeKey := GetUserAttributeKey(context.Background(), ptr.To(OIDC))
	assert.Equal(t, MapperAttributeName, attributeKey)
}

func TestGetUserAttributeKey_OCP(t *testing.T) {
	attributeKey := GetUserAttributeKey(context.Background(), ptr.To(OCP))
	assert.Equal(t, MapperUserAttributeOcp, attributeKey)
}

func TestGetUserAttributeKey_nil_idp(t *testing.T) {
	attributeKey := GetUserAttributeKey(context.Background(), nil)
	assert.Equal(t, "", attributeKey)
}
