/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	runaikaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ParseKaiDepartmentToUpdateStatusRequest", func() {
	buildKaiDepartment := func() *runaikaiv1alpha1.Department {
		return &runaikaiv1alpha1.Department{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "my-department",
				Labels: map[string]string{constants.SyncedDepartmentIdLabel: "department-id-123"},
			},
		}
	}

	It("sets the cluster id", func() {
		req := parsers.ParseKaiDepartmentToUpdateStatusRequest(buildKaiDepartment())
		Expect(req.ClusterId).To(HaveValue(Equal(config.Config().ClusterUuid)))
	})

	It("reads the last sync time from the annotation", func() {
		department := buildKaiDepartment()
		syncTime := time.Date(2026, 7, 12, 10, 30, 0, 0, time.UTC)
		department.Annotations = map[string]string{
			parsers.LastSyncTimestampAnnotation: syncTime.Format(time.RFC3339),
		}
		req := parsers.ParseKaiDepartmentToUpdateStatusRequest(department)
		Expect(req.Status.LastUpdatedTime.Get()).To(HaveValue(BeTemporally("==", syncTime)))
	})

	It("falls back to the zero time when the annotation is missing", func() {
		req := parsers.ParseKaiDepartmentToUpdateStatusRequest(buildKaiDepartment())
		Expect(req.Status.LastUpdatedTime.Get()).To(HaveValue(Equal(time.Time{})))
	})
})
