/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster-service migration readiness", Ordered, func() {
	AfterEach(func() {
		// Ensure columns exist after each test
		Expect(addSchedulingColumns()).To(Succeed())
	})

	It("returns immediately when required columns exist", func() {
		Expect(addSchedulingColumns()).To(Succeed())

		err := waitForClusterServiceMigrations(context.Background(), 2*time.Second, 10*time.Millisecond)
		Expect(err).To(BeNil())
	})

	It("succeeds when columns appear before timeout", func() {
		Expect(dropSchedulingColumns()).To(Succeed())

		go func() {
			// Add columns shortly after, within timeout
			time.Sleep(50 * time.Millisecond)
			_ = addSchedulingColumns()
		}()

		err := waitForClusterServiceMigrations(context.Background(), 2*time.Second, 10*time.Millisecond)
		Expect(err).To(BeNil())
	})

	It("times out when columns do not exist", func() {
		Expect(dropSchedulingColumns()).To(Succeed())

		err := waitForClusterServiceMigrations(context.Background(), 150*time.Millisecond, 10*time.Millisecond)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("timeout waiting for cluster-service schema"))
	})

	It("times out when some columns do not exist", func() {
		Expect(dropPartialSchedulingColumns()).To(Succeed())

		err := waitForClusterServiceMigrations(context.Background(), 150*time.Millisecond, 10*time.Millisecond)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("timeout waiting for cluster-service schema"))
	})

	It("returns error when context is canceled", func() {
		Expect(dropSchedulingColumns()).To(Succeed())

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		err := waitForClusterServiceMigrations(ctx, 2*time.Second, 10*time.Millisecond)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("context canceled while waiting for cluster-service schema"))
	})
})

func dropSchedulingColumns() error {
	db := runaidb.DB()
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools DROP COLUMN IF EXISTS %v", models.AdjustableWeightsColumn)); err != nil {
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools DROP COLUMN IF EXISTS %v", models.CpuQuotaColumn)); err != nil {
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools DROP COLUMN IF EXISTS %v", models.QuotaColumn)); err != nil {
		return err
	}
	return nil
}

func dropPartialSchedulingColumns() error {
	db := runaidb.DB()
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools DROP COLUMN IF EXISTS %v", models.AdjustableWeightsColumn)); err != nil {
		return err
	}
	return nil
}

func addSchedulingColumns() error {
	db := runaidb.DB()
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS %v BOOLEAN NOT NULL DEFAULT false", models.AdjustableWeightsColumn)); err != nil {
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS %v BOOLEAN NOT NULL DEFAULT false", models.CpuQuotaColumn)); err != nil {
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS %v BOOLEAN NOT NULL DEFAULT true", models.QuotaColumn)); err != nil {
		return err
	}
	return nil
}
