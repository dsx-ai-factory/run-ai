package protocol_mapper

import (
	"testing"
)

func TestNewRunaiGroupsCountProtocolMapper_Config(t *testing.T) {
	mapper := NewRunaiGroupsCountProtocolMapper(true)
	if mapper.Config == nil {
		t.Fatalf("Config should not be nil")
	}
	if (*mapper.Config)["claim.name"] != "groups_count" {
		t.Errorf("Expected claim.name to be 'groups_count', got %v", (*mapper.Config)["claim.name"])
	}
	if (*mapper.Config)["list.user.attribute"] != "GROUPS" {
		t.Errorf("Expected list.user.attribute to be 'GROUPS', got %v", (*mapper.Config)["list.user.attribute"])
	}
	if _, ok := (*mapper.Config)["tenant.id"]; ok {
		t.Errorf("Config should not contain tenant.id")
	}
	if (*mapper.Config)["id.token.claim"] != "true" {
		t.Errorf("Expected id.token.claim to be 'true', got %v", (*mapper.Config)["id.token.claim"])
	}
	if (*mapper.Config)["access.token.claim"] != "true" {
		t.Errorf("Expected access.token.claim to be 'true', got %v", (*mapper.Config)["access.token.claim"])
	}
	if (*mapper.Config)["userinfo.token.claim"] != "true" {
		t.Errorf("Expected userinfo.token.claim to be 'true', got %v", (*mapper.Config)["userinfo.token.claim"])
	}
	if (*mapper.Config)["lightweight.claim"] != "true" {
		t.Errorf("Expected lightweight.claim to be 'true', got %v", (*mapper.Config)["lightweight.claim"])
	}
	if (*mapper.Config)["introspection.token.claim"] != "false" {
		t.Errorf("Expected introspection.token.claim to be 'false', got %v", (*mapper.Config)["introspection.token.claim"])
	}
}

func TestNewRunaiGroupsCountProtocolMapper_IdTokenFalse(t *testing.T) {
	mapper := NewRunaiGroupsCountProtocolMapper(false)
	if (*mapper.Config)["id.token.claim"] != "false" {
		t.Errorf("Expected id.token.claim to be 'false', got %v", (*mapper.Config)["id.token.claim"])
	}
}
