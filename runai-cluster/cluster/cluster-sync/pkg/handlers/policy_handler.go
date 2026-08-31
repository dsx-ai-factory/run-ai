/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"

	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PolicyHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedClient  client.Client
}

type TrainingPolicyHandler struct {
	*PolicyHandler
}

func (h *TrainingPolicyHandler) HandleUpdateItem(key client.ObjectKey) error {
	tp := &tpv1.TrainingPolicy{}
	err := h.cachedClient.Get(context.Background(), key, tp)
	if err != nil {
		return err
	}

	return h.HandleUpdate(tp)
}

type InteractivePolicyHandler struct {
	*PolicyHandler
}

func (h *InteractivePolicyHandler) HandleUpdateItem(key client.ObjectKey) error {
	ip := &ipv1.InteractivePolicy{}
	err := h.cachedClient.Get(context.Background(), key, ip)
	if err != nil {
		return err
	}

	return h.HandleUpdate(ip)
}

type InferencePolicyHandler struct {
	*PolicyHandler
}

func (h *InferencePolicyHandler) HandleUpdateItem(key client.ObjectKey) error {
	ip := &infpv1.InferencePolicy{}
	err := h.cachedClient.Get(context.Background(), key, ip)
	if err != nil {
		return err
	}

	return h.HandleUpdate(ip)
}

type DistributedPolicyHandler struct {
	*PolicyHandler
}

func (h *DistributedPolicyHandler) HandleUpdateItem(key client.ObjectKey) error {
	dp := &dpv1.DistributedPolicy{}
	err := h.cachedClient.Get(context.Background(), key, dp)
	if err != nil {
		return err
	}

	return h.HandleUpdate(dp)
}

func NewPolicyHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedClient client.Client) *PolicyHandler {
	return &PolicyHandler{
		backendClient: runaiBackendClient,
		cachedClient:  cachedClient,
	}
}

func (h *PolicyHandler) handlePolicyGenericSyncStatus(policy client.Object, operation string) error {
	updateSyncStatusPayload, err := parsers.ParsePolicySyncStatusToServerScheme(policy, operation)
	if err != nil {
		return err
	}

	switch policy.(type) {
	case *ipv1.InteractivePolicy:
		_, err = h.backendClient.PolicyClient.UpdateSyncStatusWorkspace(h.backendClient.Context(), updateSyncStatusPayload)
		if err != nil {
			return err
		}
	case *tpv1.TrainingPolicy:
		_, err = h.backendClient.PolicyClient.UpdateSyncStatusTraining(h.backendClient.Context(), updateSyncStatusPayload)
		if err != nil {
			return err
		}
	case *dpv1.DistributedPolicy:
		_, err = h.backendClient.PolicyClient.UpdateSyncStatusDistributed(h.backendClient.Context(), updateSyncStatusPayload)
		if err != nil {
			return err
		}
	case *infpv1.InferencePolicy:
		_, err = h.backendClient.PolicyClient.UpdateSyncStatusInference(h.backendClient.Context(), updateSyncStatusPayload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *PolicyHandler) HandleAdd(policy client.Object) error {
	return h.HandleAddUpdate(policy, constants.AddOperationString)
}

func (h *PolicyHandler) HandleUpdate(policy client.Object) error {
	return h.HandleAddUpdate(policy, constants.UpdateOperationString)
}

func (h *PolicyHandler) HandleDelete(policy client.Object) error {
	err := h.handlePolicyGenericSyncStatus(policy, constants.DeleteOperationString)
	if err != nil {
		zap.S().Warnf("The sync status update for the %s operation couldn't be completed: %s", constants.DeleteOperationString, err.Error())
	}

	return nil
}

func (h *PolicyHandler) HandleAddUpdate(policy client.Object, operation string) error {
	err := h.handlePolicyGenericSyncStatus(policy, operation)
	if err != nil {
		zap.S().Warnf("The sync status update for the %s operation couldn't be completed: %s", operation, err.Error())
	}

	return err
}
