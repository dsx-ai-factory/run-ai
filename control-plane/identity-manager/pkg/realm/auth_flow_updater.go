/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realm_updater

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	iamServices "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	CookieAuthStep = "Cookie"
)

type AuthFlowUpdaterConfig struct{}

type AuthFlowUpdaterInterface interface {
	UpdateAuthFlows(ctx context.Context, realmName string) error
}

type AuthFlowUpdater struct {
	iamAuthFlow iamServices.AuthFlowService
}

func NewAuthFlowUpdater(ctx context.Context, iamServices iam.IamServiceContainer) *AuthFlowUpdater {
	return &AuthFlowUpdater{
		iamAuthFlow: iamServices.AuthFlowService,
	}
}

func (af *AuthFlowUpdater) UpdateAuthFlows(ctx context.Context, realmName string) error {
	err := af.alternativeAuthStepForAuthFlow(ctx, realmName, utils.BROWSER_FLOW_NAME, CookieAuthStep)
	if err != nil {
		return err
	}
	authFlow, err := af.iamAuthFlow.GetAuthFlow(ctx, realmName, utils.LOGIN_FLOW_NAME)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get auth flow for realm %s: %v", realmName, err)
		return err
	}
	if authFlow != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Auth flow %s already exists, skipping", utils.LOGIN_FLOW_NAME)
		return nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Auth flow %s does not exist, creating it", utils.LOGIN_FLOW_NAME)
	err = af.createAuthFlows(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create auth flow %s: %v", utils.LOGIN_FLOW_NAME, err)
		return err
	}
	return nil
}

func (af *AuthFlowUpdater) alternativeAuthStepForAuthFlow(ctx context.Context, realmName, flowAlias, authStep string) error {
	authExecutions, err := af.iamAuthFlow.GetAuthExecutions(ctx, realmName, flowAlias)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get auth execution for authflow %s. realm: %s err: %v", flowAlias, realmName, err)
		return err
	}
	cookieAuthRep := findItemByName(authExecutions, authStep)
	if cookieAuthRep == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("auth execution %s. is missing from authflow: %s, realm: %s err: %v", authStep, flowAlias, realmName, err)
		return err
	}
	cookieAuthRep.Requirement = ptr.To("ALTERNATIVE")
	err = af.iamAuthFlow.UpdateAuthExecution(ctx, realmName, flowAlias, *cookieAuthRep)
	if err != nil {
		return err
	}
	return nil
}

func (ru *AuthFlowUpdater) createAuthFlows(ctx context.Context, realmName string) error {
	authenticationExecution := []iamModels.AuthenticationExecutionRepresentation{
		{
			AuthenticatorFlow: ptr.To(true),
			Requirement:       ptr.To("REQUIRED"),
			UserSetupAllowed:  ptr.To(false),
			FlowAlias:         ptr.To("runai first login User creation or linking"),
			Priority:          ptr.To(10),
		},
	}
	authenticationFlows := iamModels.AuthenticationFlowRepresentation{
		Alias:                    ptr.To(utils.LOGIN_FLOW_NAME),
		Description:              ptr.To("Actions taken after first broker login with identity provider account, which is not yet linked to any Keycloak account"),
		ProviderID:               ptr.To("basic-flow"),
		TopLevel:                 ptr.To(true),
		BuiltIn:                  ptr.To(false),
		AuthenticationExecutions: &authenticationExecution,
	}
	err := ru.iamAuthFlow.CreateAuthFlow(ctx, realmName, authenticationFlows)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create auth flow %s: %v", utils.LOGIN_FLOW_NAME, err)
		return err
	}

	creationOrLinkingFlow := &iamModels.ModifyAuthenticationExecutionRepresentation{
		Requirement:        ptr.To("REQUIRED"),
		DisplayName:        ptr.To("runai first login User creation or linking"),
		Alias:              ptr.To("runai first login User creation or linking"),
		Description:        ptr.To("Flow for the existing/non-existing user alternatives"),
		Configurable:       ptr.To(false),
		AuthenticationFlow: ptr.To(true),
		ProviderID:         ptr.To("registration-page-form"),
		Level:              ptr.To(0),
		Index:              ptr.To(1),
		FlowID:             ptr.To(utils.LOGIN_FLOW_NAME),
	}
	err = ru.updateExecutionFlow(ctx, realmName, "basic-flow", *creationOrLinkingFlow)
	if err != nil {
		return err
	}
	createUserIfUnique := &iamModels.ModifyAuthenticationExecutionRepresentation{
		Requirement:  ptr.To("ALTERNATIVE"),
		DisplayName:  ptr.To("Create User If Unique"),
		Alias:        ptr.To("create unique user config"),
		Configurable: ptr.To(true),
		ProviderID:   ptr.To("idp-create-user-if-unique"),
		Level:        ptr.To(1),
		Index:        ptr.To(0),
		FlowID:       ptr.To("runai first login User creation or linking"),
	}
	err = ru.updateExecution(ctx, realmName, *createUserIfUnique)
	if err != nil {
		return err
	}
	automaticallySetUser := &iamModels.ModifyAuthenticationExecutionRepresentation{
		Requirement:  ptr.To("ALTERNATIVE"),
		DisplayName:  ptr.To("Automatically set existing user"),
		Alias:        ptr.To("automatically set existing user"),
		Configurable: ptr.To(false),
		ProviderID:   ptr.To("idp-auto-link"),
		Level:        ptr.To(1),
		Index:        ptr.To(1),
		FlowID:       ptr.To("runai first login User creation or linking"),
	}
	err = ru.updateExecution(ctx, realmName, *automaticallySetUser)
	if err != nil {
		return err
	}
	return nil
}

func (ru *AuthFlowUpdater) updateExecutionFlow(ctx context.Context, realmName string, flowType string, execution iamModels.ModifyAuthenticationExecutionRepresentation) error {
	flowExecution := &iamModels.CreateAuthenticationExecutionFlowRepresentation{
		Alias:       execution.Alias,
		Description: execution.Description,
		Provider:    execution.ProviderID,
		Type:        ptr.To(flowType),
	}

	err := ru.iamAuthFlow.CreateAuthExecutionFlow(ctx, realmName, utils.LOGIN_FLOW_NAME, *flowExecution)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create auth execution flow for flow %s: %v", utils.LOGIN_FLOW_NAME, err)
		return err
	}
	existingExecutions, err := ru.iamAuthFlow.GetAuthExecutions(ctx, realmName, utils.LOGIN_FLOW_NAME)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get auth flow executions for flow %s: %v", utils.LOGIN_FLOW_NAME, err)
		return err
	}

	for _, existingExecution := range existingExecutions {
		if execution.DisplayName != nil && *existingExecution.DisplayName == *execution.DisplayName {
			execution.ID = existingExecution.ID
			execution.FlowID = existingExecution.FlowID
			return ru.iamAuthFlow.UpdateAuthExecution(ctx, realmName, utils.LOGIN_FLOW_NAME, execution)
		}
	}

	return fmt.Errorf("error: created flow %v but could not list to for modification", execution.ProviderID)
}

func (ru *AuthFlowUpdater) updateExecution(ctx context.Context, realmName string, execution iamModels.ModifyAuthenticationExecutionRepresentation) error {
	executionId := &iamModels.CreateAuthenticationExecutionRepresentation{
		Provider: execution.ProviderID,
	}
	err := ru.iamAuthFlow.CreateAuthExecution(ctx, realmName, *execution.FlowID, *executionId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create auth execution for flow %s: %v", *execution.FlowID, err)
		return err
	}
	existingExecutions, err := ru.iamAuthFlow.GetAuthExecutions(ctx, realmName, *execution.FlowID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get auth flow executions for flow %s: %v", *execution.FlowID, err)
		return err
	}

	for _, existingExecution := range existingExecutions {
		if *existingExecution.ProviderID == *execution.ProviderID {
			execution.ID = existingExecution.ID
			return ru.iamAuthFlow.UpdateAuthExecution(ctx, realmName, *execution.FlowID, execution)
		}
	}

	return fmt.Errorf("error: created flow %v but could not list to for modification", execution.ProviderID)
}

func findItemByName(items []*iamModels.ModifyAuthenticationExecutionRepresentation, name string) *iamModels.ModifyAuthenticationExecutionRepresentation {
	for _, item := range items {
		if item.DisplayName != nil && *item.DisplayName == name {
			return item
		}
	}
	return nil
}
