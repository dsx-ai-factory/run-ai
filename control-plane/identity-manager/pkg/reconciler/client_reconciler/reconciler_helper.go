/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_reconciler

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

//go:generate mockery --name ClientReconcilerHelperInterface
type ClientReconcilerHelperInterface interface {
	ReconcileClient(ctx context.Context, realm string, client realm_client.Client) error
}

type ClientReconcilerHelper struct {
	iamServices    iam.IamServiceContainer
	tenantResolver reconciler.TenantResolver
}

func NewClientReconcilerHelper(ctx context.Context, iamServices iam.IamServiceContainer, tenantResolver reconciler.TenantResolver) *ClientReconcilerHelper {
	cr := ClientReconcilerHelper{
		iamServices:    iamServices,
		tenantResolver: tenantResolver,
	}
	return &cr
}

func (rh *ClientReconcilerHelper) ReconcileClient(ctx context.Context, realm string, client realm_client.Client) error {
	clientID := client.GetClientID()
	existsClient, err := rh.iamServices.ClientService.GetClientByClientId(ctx, realm, clientID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get client %s in realm %s: %v", clientID, realm, err)
		return err
	}
	newClient := client.GetClientData(ctx, existsClient)
	if newClient == nil {
		err := fmt.Errorf("failed to prepare client data for client %s in realm %s", clientID, realm)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%v", err)
		return err
	}
	if newClient.ID != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Client %s already exists in realm %s, performing update", *newClient.ClientID, realm)
		if err := rh.iamServices.ClientService.UpdateClient(ctx, realm, newClient); err != nil {
			return err
		}
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Client %s does not exist in realm %s, performing create", *newClient.ClientID, realm)
		id, err := rh.iamServices.ClientService.CreateClient(ctx, realm, newClient)
		if err != nil {
			return err
		}
		newClient.ID = &id
	}
	if client.NeedsEmailScopeDefault() {
		if err := rh.iamServices.ClientService.SetEmailScopeDefault(ctx, realm, *newClient.ID); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set email scope default for client %s in realm %s: %v", clientID, realm, err)
			return err
		}
	}
	if client.NeedsGeneratedSecret() {
		if err := rh.ensureClientSecret(ctx, realm, clientID, *newClient.ID); err != nil {
			return err
		}
	}
	return nil
}

// ensureClientSecret regenerates the client's secret if it doesn't have one yet.
// A client that transitions from public to confidential via an update does not
// get a secret automatically from Keycloak; it has to be generated explicitly.
func (rh *ClientReconcilerHelper) ensureClientSecret(ctx context.Context, realm string, clientID string, internalID string) error {
	secret, err := rh.iamServices.ClientService.GetClientSecret(ctx, realm, internalID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get secret for client %s in realm %s: %v", clientID, realm, err)
		return err
	}
	if secret != "" {
		return nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Client %s in realm %s has no secret, regenerating", clientID, realm)
	if _, err := rh.iamServices.ClientService.RegenerateClientSecret(ctx, realm, internalID); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to regenerate secret for client %s in realm %s: %v", clientID, realm, err)
		return err
	}
	return nil
}
