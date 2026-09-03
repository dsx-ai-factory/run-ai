/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	idmclienttypes "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
)

type AuthnManager interface {
	CreateClient(ctx context.Context, tenantID int32, clusterUuid string, clientType idmclienttypes.ClientType) (string, error)
	DeleteClientByID(ctx context.Context, clientID string) error
	DeleteClients(ctx context.Context, clusterUuid string) error
}

type AuthzManager interface {
	CreateAccessRule(ctx context.Context, clientID, clusterUuid, clusterName string) error
	DeleteAccessRules(ctx context.Context, clusterUuid string) error
}

type ClusterAccessControlManager struct {
	authnManager AuthnManager
	authzManager AuthzManager
}

func NewClusterAccessControlManager(authnManager AuthnManager, authzManager AuthzManager) *ClusterAccessControlManager {
	return &ClusterAccessControlManager{
		authnManager: authnManager,
		authzManager: authzManager,
	}
}

func (c *ClusterAccessControlManager) CreateAccessControl(ctx context.Context, tennatID int32, clusterUuid, clusterName string) error {
	clusterAccessControlEntities := []idmclienttypes.ClientType{idmclienttypes.CLIENTTYPE_AGENT, idmclienttypes.CLIENTTYPE_WORKLOADS_AUTH}

	for _, entity := range clusterAccessControlEntities {
		clientID, err := c.authnManager.CreateClient(ctx, tennatID, clusterUuid, entity)
		if err != nil {
			return fmt.Errorf("failed to create client for cluster %s: %v", clusterUuid, err)
		}

		err = c.authzManager.CreateAccessRule(ctx, clientID, clusterUuid, clusterName)
		if err != nil {
			// Best effort rollback
			c.authnManager.DeleteClientByID(ctx, clientID)
			return fmt.Errorf("failed to create access rule for cluster %s: %v", clusterUuid, err)
		}
	}

	return nil
}

func (c *ClusterAccessControlManager) DeleteAccessControl(ctx context.Context, clusterUuid string) error {
	multiErr := &multierror.Error{}
	err := c.authzManager.DeleteAccessRules(ctx, clusterUuid)
	if err != nil {
		multiErr = multierror.Append(multiErr, fmt.Errorf("failed to delete access rules for cluster %s: %w", clusterUuid, err))
	}

	err = c.authnManager.DeleteClients(ctx, clusterUuid)
	if err != nil {
		multiErr = multierror.Append(multiErr, fmt.Errorf("failed to delete clients for cluster %s: %w", clusterUuid, err))
	}

	return multiErr.ErrorOrNil()
}
