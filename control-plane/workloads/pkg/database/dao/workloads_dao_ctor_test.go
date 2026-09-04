/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
)

var _ = Describe("NewWorkloadsDao", func() {
	It("uses the plain view when the refresh rate is zero", func() {
		dao := NewWorkloadsDao(&config.ServerConfig{ViewRefreshRateMilliseconds: 0})

		Expect(dao.view).To(Equal(WorkloadsViewName))
	})

	It("uses the partial materialized view when it is enabled", func() {
		dao := NewWorkloadsDao(&config.ServerConfig{ViewRefreshRateMilliseconds: 100, UsePartialMaterializedView: true})

		Expect(dao.view).To(Equal(WorkloadsApiViewName))
	})

	It("uses the materialized view when the partial view is disabled", func() {
		dao := NewWorkloadsDao(&config.ServerConfig{ViewRefreshRateMilliseconds: 100, UsePartialMaterializedView: false})

		Expect(dao.view).To(Equal(WorkloadsMaterializedViewName))
	})
})
