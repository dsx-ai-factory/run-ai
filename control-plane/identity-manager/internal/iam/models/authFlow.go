package models

type AuthenticationExecutionRepresentation struct {
	Authenticator       *string `json:"authenticator,omitempty"`
	AuthenticatorConfig *string `json:"authenticatorConfig,omitempty"`
	AuthenticatorFlow   *bool   `json:"authenticatorFlow,omitempty"`
	AutheticatorFlow    *bool   `json:"autheticatorFlow,omitempty"`
	FlowAlias           *string `json:"flowAlias,omitempty"`
	Priority            *int    `json:"priority,omitempty"`
	Requirement         *string `json:"requirement,omitempty"`
	UserSetupAllowed    *bool   `json:"userSetupAllowed,omitempty"`
}

type AuthenticationFlowRepresentation struct {
	Alias                    *string                                  `json:"alias,omitempty"`
	AuthenticationExecutions *[]AuthenticationExecutionRepresentation `json:"authenticationExecutions,omitempty"`
	BuiltIn                  *bool                                    `json:"builtIn,omitempty"`
	Description              *string                                  `json:"description,omitempty"`
	ID                       *string                                  `json:"id,omitempty"`
	ProviderID               *string                                  `json:"providerId,omitempty"`
	TopLevel                 *bool                                    `json:"topLevel,omitempty"`
}

type ModifyAuthenticationExecutionRepresentation struct {
	ID                   *string   `json:"id,omitempty"`
	ProviderID           *string   `json:"providerId,omitempty"`
	AuthenticationConfig *string   `json:"authenticationConfig,omitempty"`
	AuthenticationFlow   *bool     `json:"authenticationFlow,omitempty"`
	Requirement          *string   `json:"requirement,omitempty"`
	FlowID               *string   `json:"flowId"`
	DisplayName          *string   `json:"displayName,omitempty"`
	Alias                *string   `json:"alias,omitempty"`
	RequirementChoices   *[]string `json:"requirementChoices,omitempty"`
	Configurable         *bool     `json:"configurable,omitempty"`
	Level                *int      `json:"level,omitempty"`
	Index                *int      `json:"index,omitempty"`
	Description          *string   `json:"description"`
}

type CreateAuthenticationExecutionRepresentation struct {
	Provider *string `json:"provider,omitempty"`
}

type CreateAuthenticationExecutionFlowRepresentation struct {
	Alias       *string `json:"alias,omitempty"`
	Description *string `json:"description,omitempty"`
	Provider    *string `json:"provider,omitempty"`
	Type        *string `json:"type,omitempty"`
}
