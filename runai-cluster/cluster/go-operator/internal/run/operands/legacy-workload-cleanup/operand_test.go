/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacy_workload_cleanup

import (
	"testing"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResourcesAreGatedByDerivedCleanupCondition(t *testing.T) {
	client := clientfake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()
	hybrid := kc.New(client, client, nil)

	for _, test := range []struct {
		name      string
		condition *metav1.Condition
		resources int
	}{
		{name: "no condition", resources: 0},
		{name: "waiting for readiness", condition: cleanupCondition(metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupWaitingForReady), resources: 0},
		{name: "creating", condition: cleanupCondition(metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupCreating), resources: 1},
		{name: "running", condition: cleanupCondition(metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupRunning), resources: 1},
		{name: "failed", condition: cleanupCondition(metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupFailed), resources: 1},
		{name: "blocked", condition: cleanupCondition(metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupBlocked), resources: 0},
		{name: "completed", condition: cleanupCondition(metav1.ConditionTrue, runaiv1.LegacyWorkloadCleanupCompleted), resources: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := &runaiv1.RunaiConfig{}
			if test.condition != nil {
				config.Status.Conditions = []metav1.Condition{*test.condition}
			}
			require.Len(t, Operand(hybrid, config).Resources(), test.resources)
		})
	}
}

func cleanupCondition(status metav1.ConditionStatus, reason runaiv1.ConditionReason) *metav1.Condition {
	return &metav1.Condition{
		Type: string(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete), Status: status, Reason: string(reason),
	}
}
