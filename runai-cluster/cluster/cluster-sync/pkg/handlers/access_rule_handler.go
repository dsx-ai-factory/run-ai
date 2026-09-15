/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRuleHandler struct {
	runaiClient   client.Client
	backendClient *runaiBackendClient.BackendClient
}

func NewAccessRuleHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *AccessRuleHandler {
	return &AccessRuleHandler{
		runaiClient:   runaiClient,
		backendClient: runaiBackendClient,
	}
}

func (accessRuleHandler *AccessRuleHandler) HandleUpdateItem(key client.ObjectKey) error {
	accessRule := &v1alpha1.AccessRule{}
	err := accessRuleHandler.runaiClient.Get(context.Background(), key, accessRule)
	if err != nil {
		return err
	}

	return accessRuleHandler.HandleUpdate(accessRule)
}

func (accessRuleHandler *AccessRuleHandler) HandleUpdate(accessRule *v1alpha1.AccessRule) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.AccessRuleResourceName, constants.UpdateOperation)()
	return accessRuleHandler.HandleAddUpdate(accessRule)
}

func (accessRuleHandler *AccessRuleHandler) HandleAdd(accessRule *v1alpha1.AccessRule) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.AccessRuleResourceName, constants.AddOperation)()
	return accessRuleHandler.HandleAddUpdate(accessRule)
}

func (accessRuleHandler *AccessRuleHandler) HandleAddUpdate(accessRule *v1alpha1.AccessRule) error {
	return accessRuleHandler.updateAccessRuleStatus(accessRule)
}

func (accessRuleHandler *AccessRuleHandler) HandleDelete(accessRule *v1alpha1.AccessRule) error {
	accessRuleId, err := parsers.GetAccessRuleId(accessRule)
	if err != nil {
		zap.S().Errorf("Failed updating AccessRule <%s> status to deleted, err: %+v", accessRule.Name, err)
		return err
	}
	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.AccessRuleResourceName, constants.DeleteOperation)()
		toDelete := parsers.GetAccessRuleDeleteRequest()
		res, err := accessRuleHandler.backendClient.AccessRulesClient.UpdateStatusAccessRule(
			accessRuleHandler.backendClient.Context(),
			accessRuleId,
			config.Config().ClusterUuid,
			toDelete,
		)
		if res != nil {
			_ = res.Body.Close()

		}
		return accessRuleHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted access rule %s", accessRule.Name))
	}
}

func (accessRuleHandler *AccessRuleHandler) updateAccessRuleStatus(accessRule *v1alpha1.AccessRule) error {
	accessRuleId, err := parsers.GetAccessRuleId(accessRule)
	if err != nil {
		zap.S().Errorf("Failed updating AccessRule <%s> status, err: %+v", accessRule.Name, err)
		return err
	}

	toUpdate := parsers.ParseAccessRuleToUpdateStatusRequest(&accessRule.Status)
	res, err := accessRuleHandler.backendClient.AccessRulesClient.UpdateStatusAccessRule(
		accessRuleHandler.backendClient.Context(),
		accessRuleId,
		config.Config().ClusterUuid,
		toUpdate,
	)
	if res != nil {
		_ = res.Body.Close()

	}
	return accessRuleHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated access rule %s", accessRule.Name))
}

func (accessRuleHandler *AccessRuleHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start access rules Sync")

	// Don't convert to var since this sends an empty array when no data is in cluster vs null (which will result in bad request (400))
	accessRules := &v1alpha1.AccessRuleList{}
	kubeErr := accessRuleHandler.runaiClient.List(context.Background(), accessRules)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}
	toSync := parsers.ParseAccessRulesToSyncRequest(accessRules)
	if len(toSync) == 0 {
		zap.S().Infof("No AccessRules to sync")
		return
	}
	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.AccessRuleResourceName, constants.SyncOperation)()
		res, err := accessRuleHandler.backendClient.AccessRulesClient.AccessRulesSync(
			accessRuleHandler.backendClient.Context(),
			config.Config().ClusterUuid,
			toSync,
		)
		if res != nil {
			_ = res.Body.Close()

		}
		err = accessRuleHandler.backendClient.HandleHttpResponse(res, err, "Sync access rules")
		if err != nil {
			zap.S().Errorf("Had an error while syncing access rules: %+v", err)
		}
	}
}
