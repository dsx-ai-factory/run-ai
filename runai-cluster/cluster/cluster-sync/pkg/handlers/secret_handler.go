/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	"go.uber.org/zap"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"

	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewSecretHandler(cachedReader client.Reader, runaiBackendClient *runaiBackendClient.BackendClient) *SecretHandler {
	return &SecretHandler{
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (secretHandler *SecretHandler) HandleUpdateItem(key client.ObjectKey) error {
	secret := &coreV1.Secret{}
	err := secretHandler.cachedReader.Get(context.Background(), key, secret)
	if err != nil {
		return err
	}

	return secretHandler.HandleSecretUpdate(secret)
}

func (secretHandler *SecretHandler) HandleSecretAdd(secret *coreV1.Secret) error {
	zap.S().Infof("Adding Secret <Name: %v, Namespace: %v>", secret.Name, secret.Namespace)
	return secretHandler.handleSecretAddUpdate(secret)
}

func (secretHandler *SecretHandler) HandleSecretUpdate(secret *coreV1.Secret) error {
	zap.S().Infof("Updating Secret <Name: %v, Namespace: %v>", secret.Name, secret.Namespace)
	return secretHandler.handleSecretAddUpdate(secret)
}

func (secretHandler *SecretHandler) handleSecretAddUpdate(secret *coreV1.Secret) error {
	parsedSecret, err := parsers.ParseSecretToServerScheme(secret)
	if err != nil {
		zap.S().Warnf("Error parsing secret %s:%s, error: %v", secret.Namespace, secret.Name, err)
		return err
	}

	var response *http.Response

	switch parsedSecret.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		_, response, err = secretHandler.backendClient.K8sObjectTrackerClient.CreateClusterWideSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, secret.Name,
			parsedSecret.CreationRequest.Type, parsedSecret.CreationRequest.Origin, parsedSecret.CreationRequest.Issues, parsedSecret.CreationRequest.Private)
		if response != nil {
			_ = response.Body.Close()

		}

	case auth_client.SCOPETYPE_DEPARTMENT:
		_, response, err = secretHandler.backendClient.K8sObjectTrackerClient.CreateDepartmentWideSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, parsedSecret.Scope.Id, secret.Name,
			parsedSecret.CreationRequest.Type, parsedSecret.CreationRequest.Origin, parsedSecret.CreationRequest.Issues)
		if response != nil {
			_ = response.Body.Close()

		}

	case auth_client.SCOPETYPE_PROJECT:
		_, response, err = secretHandler.backendClient.K8sObjectTrackerClient.CreateProjectSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, parsedSecret.Scope.Id, secret.Name,
			parsedSecret.CreationRequest.Type, parsedSecret.CreationRequest.Origin, parsedSecret.CreationRequest.Issues)
		if response != nil {
			_ = response.Body.Close()

		}
	}

	return secretHandler.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Added Secrets %s:%s",
		secret.Namespace, secret.Name))
}

func (secretHandler *SecretHandler) HandleSecretDelete(secret *coreV1.Secret) error {
	zap.S().Debugf("Deleting Secret <Name: %v, Namespace: %v>", secret.Name, secret.Namespace)

	parsedSecret, err := parsers.ParseSecretToServerScheme(secret)
	if err != nil {
		zap.S().Warnf("Error parsing secret %s:%s, error: %v", secret.Namespace, secret.Name, err)
		return err
	}

	var response *http.Response

	switch parsedSecret.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		response, err = secretHandler.backendClient.K8sObjectTrackerClient.DeleteClusterWideSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, secret.Name)
		if response != nil {
			_ = response.Body.Close()

		}

	case auth_client.SCOPETYPE_DEPARTMENT:
		response, err = secretHandler.backendClient.K8sObjectTrackerClient.DeleteDepartmentWideSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, parsedSecret.Scope.Id, secret.Name)
		if response != nil {
			_ = response.Body.Close()

		}

	case auth_client.SCOPETYPE_PROJECT:
		response, err = secretHandler.backendClient.K8sObjectTrackerClient.DeleteProjectSecret(
			secretHandler.backendClient.Context(), config.Config().ClusterUuid, parsedSecret.Scope.Id, secret.Name)
		if response != nil {
			_ = response.Body.Close()

		}
	}

	zap.S().Debugf("HandleSecretDelete res %v ", response)
	return secretHandler.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Deleted Secret %s:%s",
		secret.Namespace, secret.Name))
}

func (secretHandler *SecretHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	zap.S().Infof("Start syncing Secrets")

	secrets, err := secretHandler.listSecretsWithLabels(config.Config().ExistingSecretToSyncLabel, commonconstants.RunaiAssetLabel)
	if err != nil {
		return
	}

	// get a map of all the project namespaces
	namespacesMap := util.GetAllRunAiManagedNamespacesMap(secretHandler.cachedReader)

	// initialize a map that arranges the secrets by their scope
	secretsByScope := make(map[parsers.ScopeTypeAndId]*k8sobjectstracker.Secrets)

	for _, secret := range secrets {
		if !util.IsUserExposedCredentialsForRunai(&secret) && !util.IsAssetRelatedResource(&secret.ObjectMeta) {
			continue
		}

		parsedSecret, err := parsers.ParseSecretToServerScheme(&secret)
		if err != nil {
			zap.S().Warnf("Error parsing secret %s:%s, error: %v", secret.Namespace, secret.Name, err)
			continue
		}

		if parsedSecret.Scope.Type == auth_client.SCOPETYPE_PROJECT {
			_, namespaceFound := namespacesMap[secret.Namespace]
			if !namespaceFound {
				continue // project secrets only in project related namespace (all others are replicates)
			}
		} else {
			if secret.Namespace != v1.DefaultNamespace {
				continue // non-project secrets only in runai namespace
			}
		}

		secretToSync := k8sobjectstracker.Secret{
			Name:    secret.Name,
			Type:    parsedSecret.CreationRequest.Type,
			Origin:  parsedSecret.CreationRequest.Origin,
			Issues:  parsedSecret.CreationRequest.Issues,
			Private: parsedSecret.CreationRequest.Private,
		}

		if secretsOfThisType, ok := secretsByScope[parsedSecret.Scope]; ok {
			secretsOfThisType.Secrets = append(secretsOfThisType.Secrets, secretToSync)
		} else {
			secretsByScope[parsedSecret.Scope] = &k8sobjectstracker.Secrets{
				Secrets: []k8sobjectstracker.Secret{secretToSync},
			}
		}
	}

	for scope, secretList := range secretsByScope {
		zap.S().Debugf("apply %d secrets of %s %s", len(secretList.Secrets), scope.Type, scope.Id)

		var execError error
		var response *http.Response

		switch scope.Type {
		case auth_client.SCOPETYPE_CLUSTER:
			_, response, execError = secretHandler.backendClient.K8sObjectTrackerClient.SyncClusterWideSecrets(
				secretHandler.backendClient.Context(), config.Config().ClusterUuid, secretList.Secrets)
			if response != nil {
				_ = response.Body.Close()

			}

		case auth_client.SCOPETYPE_DEPARTMENT:
			_, response, execError = secretHandler.backendClient.K8sObjectTrackerClient.SyncDepartmentWideSecrets(
				secretHandler.backendClient.Context(), config.Config().ClusterUuid, scope.Id, secretList.Secrets)
			if response != nil {
				_ = response.Body.Close()

			}

		case auth_client.SCOPETYPE_PROJECT:
			_, response, execError = secretHandler.backendClient.K8sObjectTrackerClient.SyncProjectSecrets(
				secretHandler.backendClient.Context(), config.Config().ClusterUuid, scope.Id, secretList.Secrets)
			if response != nil {
				_ = response.Body.Close()

			}
		}

		if execError != nil {
			zap.S().Warnf("Error while syncing secrets of %s %s, error: %v", scope.Type, scope.Id, execError)
		} else {
			zap.S().Debugf("HandleSecretSync of %s %s res %v ", scope.Type, scope.Id, response)
		}
	}
}

func (secretHandler *SecretHandler) listSecretsWithLabels(labels ...string) ([]coreV1.Secret, error) {
	var results []coreV1.Secret

	// When filtering by labels, the k8s list API doesn't support or operators, so we need to query each label separately
	for _, l := range labels {
		// get a list of all the secrets in the cluster. Don't convert to var since this sends an empty array when
		// no data is in cluster vs null (which will result in bad request (400))
		secrets := &coreV1.SecretList{}
		err := secretHandler.cachedReader.List(context.TODO(), secrets,
			client.HasLabels{l})
		if err != nil {
			zap.S().Errorf("Failed listing secrets by label %v, error: %v", l, err)
			return nil, err
		}
		results = append(results, secrets.Items...)
	}
	return results, nil
}
