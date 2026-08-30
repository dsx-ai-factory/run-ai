package models

type Idp struct {
	Alias                     *string
	ProviderID                *string
	DisplayName               *string
	Enabled                   *bool
	Config                    *map[string]string
	AddReadTokenRoleOnCreate  *bool
	FirstBrokerLoginFlowAlias *string
	InternalID                *string
	LinkOnly                  *bool
	PostBrokerLoginFlowAlias  *string
	SupportTokenExchange      *bool
}

type IdpMapper struct {
	ID                     *string
	Name                   *string
	IdentityProviderMapper *string
	IdentityProviderAlias  *string
	Config                 *map[string]string
}
