/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	idmclienttypes "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	iamutils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/utils"

	errtypes "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/common/errors"

	"k8s.io/utils/ptr"
)

const (
	internalClientsPrefix = "runai-internal"
)

var clientIdTemplate = template.Must(template.New("").Parse(fmt.Sprintf("%s.{{.scope}}.{{.clientType}}", internalClientsPrefix)))

func (a *AuthnManager) CreateClient(ctx context.Context, tenantId int32, clusterUuid string, clientType idmclienttypes.ClientType) (string, error) {
	clientId, err := generateClientID(clusterUuid, clientType)
	if err != nil {
		return "", fmt.Errorf("failed to generate client ID: %w", err)
	}

	req := idmclienttypes.ClientCreationRequest{
		Type:     clientType,
		ClientId: clientId,
		TenantId: ptr.To(strconv.Itoa(int(tenantId))),
		Attributes: &idmclienttypes.ClientAttributeCreationRequest{
			RunaiClusterId: ptr.To(clusterUuid),
		},
	}

	_, resp, err := a.idmClient.CreateInternalClient(ctx, req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusConflict {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cluster %s already has a client", clusterUuid)
			return clientId, nil
		}
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	return clientId, nil
}

func (a *AuthnManager) GetClient(ctx context.Context, clusterUuid string, clientType idmclienttypes.ClientType) (idmclienttypes.Client, error) {
	filter := generateClientFilter(map[string]string{
		iamutils.GetClusterIdAttr():  clusterUuid,
		iamutils.GetClientTypeAttr(): strcase.ToSnake(string(clientType)), // Keycloak client attributes use snake_case; clientType uses camelCase
	})

	internalClients, err := a.idmClient.GetInternalClientByFilter(ctx, filter)
	if err != nil {
		return idmclienttypes.Client{}, fmt.Errorf("failed to get internal clients: %w", err)
	}

	if len(internalClients) == 0 {
		return idmclienttypes.Client{}, errors.Join(errtypes.ErrNotFound, fmt.Errorf("%s client not found for cluster %s", clientType, clusterUuid))
	}

	if len(internalClients) > 1 {
		return idmclienttypes.Client{}, fmt.Errorf("found more than one %s client for cluster %s", clientType, clusterUuid)
	}

	return internalClients[0], nil
}

func (a *AuthnManager) DeleteClientByID(ctx context.Context, clientID string) error {
	_, err := a.idmClient.DeleteClient(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}

func (a *AuthnManager) DeleteClients(ctx context.Context, clusterUuid string) error {
	filter := fmt.Sprintf("%s:%s", iamutils.GetClusterIdAttr(), clusterUuid)
	internalClients, err := a.idmClient.GetInternalClientByFilter(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get internal clients: %w", err)
	}

	multiErr := &multierror.Error{}
	for _, client := range internalClients {
		_, err := a.idmClient.DeleteClient(ctx, client.Id)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr.ErrorOrNil()
}

func generateClientID(scope string, clientType idmclienttypes.ClientType) (string, error) {
	buf := &bytes.Buffer{}

	err := clientIdTemplate.Execute(buf, map[string]interface{}{
		"scope":      scope,
		"clientType": strcase.ToSnake(string(clientType)),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func generateClientFilter(keyValues map[string]string) string {
	filter := ""
	for key, value := range keyValues {
		filter += fmt.Sprintf("%s:%s ", key, value)
	}
	return filter
}
