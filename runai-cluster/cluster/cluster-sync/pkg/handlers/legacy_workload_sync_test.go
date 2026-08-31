/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("legacy workload read-only sync filtering", func() {
	It("recognizes an unchanged workload without mutating its legacy sync status", func() {
		generation := int64(7)
		lastStatusChangeTime := metav1.Now()
		parsed := &openapi.WorkloadSyncFields{
			Id:        "legacy-uid",
			Name:      "legacy-training",
			Namespace: "project-namespace",
			Type:      "Training",
		}
		externalStatus := &v1alpha1.WorkloadSyncStatus{}
		Expect(setExternalWorkloadSuccessfulSyncStatus(
			externalStatus, generation, &lastStatusChangeTime, parsed,
		)).To(Succeed())
		legacyStatus := &common.WorkloadSyncStatus{
			LastSyncGeneration:       externalStatus.LastSyncGeneration,
			LastSyncStatusChangeTime: externalStatus.LastSyncStatusChangeTime,
			LastSyncStatusChecksum:   externalStatus.LastSyncStatusChecksum,
		}
		originalStatus := legacyStatus.DeepCopy()

		Expect(isLegacyWorkloadUpToDate(
			legacyStatus, generation, &lastStatusChangeTime, parsed,
		)).To(BeTrue())
		Expect(legacyStatus).To(Equal(originalStatus))
	})

	It("requires synchronization when the workload changed", func() {
		status := &common.WorkloadSyncStatus{LastSyncGeneration: 3}

		Expect(isLegacyWorkloadUpToDate(
			status, 4, nil, &openapi.WorkloadSyncFields{Id: "legacy-uid"},
		)).To(BeFalse())
	})

	It("requires synchronization when no prior status exists", func() {
		Expect(isLegacyWorkloadUpToDate(
			nil, 1, nil, &openapi.WorkloadSyncFields{Id: "legacy-uid"},
		)).To(BeFalse())
	})
})
