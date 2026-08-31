/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authflow

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
)

func FromAuthFlowRepModel(authFlowRep *models.AuthenticationFlowRepresentation) *gocloak.AuthenticationFlowRepresentation {
	authFlowKcRep := &gocloak.AuthenticationFlowRepresentation{
		ID:          authFlowRep.ID,
		Alias:       authFlowRep.Alias,
		BuiltIn:     authFlowRep.BuiltIn,
		Description: authFlowRep.Description,
		ProviderID:  authFlowRep.ProviderID,
		TopLevel:    authFlowRep.TopLevel,
	}
	if authFlowRep.AuthenticationExecutions != nil {
		modelArray := *authFlowRep.AuthenticationExecutions
		dst := make([]gocloak.AuthenticationExecutionRepresentation, len(modelArray))
		for i, item := range modelArray {
			dst[i] = gocloak.AuthenticationExecutionRepresentation{
				Authenticator:       item.Authenticator,
				AuthenticatorConfig: item.AuthenticatorConfig,
				AuthenticatorFlow:   item.AuthenticatorFlow,
				AutheticatorFlow:    item.AutheticatorFlow,
				FlowAlias:           item.FlowAlias,
				Priority:            item.Priority,
				Requirement:         item.Requirement,
				UserSetupAllowed:    item.UserSetupAllowed,
			}
			authFlowKcRep.AuthenticationExecutions = &dst
		}
	}
	return authFlowKcRep
}

func ToAuthFlowRepModel(kcAuthFlowRep *gocloak.AuthenticationFlowRepresentation) *models.AuthenticationFlowRepresentation {
	authFlowKcRep := &models.AuthenticationFlowRepresentation{
		ID:          kcAuthFlowRep.ID,
		Alias:       kcAuthFlowRep.Alias,
		BuiltIn:     kcAuthFlowRep.BuiltIn,
		Description: kcAuthFlowRep.Description,
		ProviderID:  kcAuthFlowRep.ProviderID,
		TopLevel:    kcAuthFlowRep.TopLevel,
	}
	return authFlowKcRep
}

func FromModifyAutheExecutionRepModel(modifyAuthRep *models.ModifyAuthenticationExecutionRepresentation) *gocloak.ModifyAuthenticationExecutionRepresentation {
	kcModifyAuthRep := &gocloak.ModifyAuthenticationExecutionRepresentation{
		ID:                   modifyAuthRep.ID,
		ProviderID:           modifyAuthRep.ProviderID,
		AuthenticationConfig: modifyAuthRep.AuthenticationConfig,
		AuthenticationFlow:   modifyAuthRep.AuthenticationFlow,
		Requirement:          modifyAuthRep.Requirement,
		FlowID:               modifyAuthRep.FlowID,
		DisplayName:          modifyAuthRep.DisplayName,
		Alias:                modifyAuthRep.Alias,
		RequirementChoices:   modifyAuthRep.RequirementChoices,
		Configurable:         modifyAuthRep.Configurable,
		Level:                modifyAuthRep.Level,
		Index:                modifyAuthRep.Index,
		Description:          modifyAuthRep.Description,
	}
	return kcModifyAuthRep
}

func ToModifyAutheExecutionRepModel(kcModifyAuthRep *gocloak.ModifyAuthenticationExecutionRepresentation) *models.ModifyAuthenticationExecutionRepresentation {
	modifyAuthRep := &models.ModifyAuthenticationExecutionRepresentation{
		ID:                   kcModifyAuthRep.ID,
		ProviderID:           kcModifyAuthRep.ProviderID,
		AuthenticationConfig: kcModifyAuthRep.AuthenticationConfig,
		AuthenticationFlow:   kcModifyAuthRep.AuthenticationFlow,
		Requirement:          kcModifyAuthRep.Requirement,
		FlowID:               kcModifyAuthRep.FlowID,
		DisplayName:          kcModifyAuthRep.DisplayName,
		Alias:                kcModifyAuthRep.Alias,
		RequirementChoices:   kcModifyAuthRep.RequirementChoices,
		Configurable:         kcModifyAuthRep.Configurable,
		Level:                kcModifyAuthRep.Level,
		Index:                kcModifyAuthRep.Index,
		Description:          kcModifyAuthRep.Description,
	}
	return modifyAuthRep
}

func FromCreateAutheExecutionRepModel(kcModifyAuthRep *models.CreateAuthenticationExecutionFlowRepresentation) *gocloak.CreateAuthenticationExecutionFlowRepresentation {
	createAuthRep := &gocloak.CreateAuthenticationExecutionFlowRepresentation{
		Alias:       kcModifyAuthRep.Alias,
		Description: kcModifyAuthRep.Description,
		Provider:    kcModifyAuthRep.Provider,
		Type:        kcModifyAuthRep.Type,
	}
	return createAuthRep
}
