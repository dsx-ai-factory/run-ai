/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

var _ = Describe("PostgresReportsDao GetReports sorting", func() {
	var (
		ctx        context.Context
		reportsDao PostgresReportsDao
		tenantId   int32
	)

	// Ascending id order, so a leaked "ORDER BY id" tiebreaker is distinguishable
	// from any requested sort.
	idFor := func(n int) uuid.UUID {
		return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	}

	insert := func(report models.Report) {
		report.TenantId = tenantId
		_, err := runaidb.DB().WithContext(ctx).Create(&report)
		Expect(err).To(BeNil())
	}

	namesOf := func(reports []models.Report) []string {
		names := make([]string, 0, len(reports))
		for _, report := range reports {
			names = append(names, report.Name)
		}
		return names
	}

	idsOf := func(reports []models.Report) []uuid.UUID {
		ids := make([]uuid.UUID, 0, len(reports))
		for _, report := range reports {
			ids = append(ids, report.ID)
		}
		return ids
	}

	BeforeEach(func() {
		ctx = context.WithValue(
			context.WithValue(context.Background(), runaictx.CfTenantId, "1"),
			runaictx.CfSubjectID,
			"user",
		)
		reportsDao = *NewPostgresReportsDao()
		tenantId = rand.New(rand.NewSource(time.Now().UnixNano())).Int31()
	})

	AfterEach(func() {
		_, err := runaidb.DB().WithContext(ctx).Where("tenant_id = ?", tenantId).Delete(&models.Report{})
		Expect(err).To(BeNil())
	})

	It("sorts by name ascending rather than by id", func() {
		insert(models.Report{ID: idFor(1), Name: "charlie"})
		insert(models.Report{ID: idFor(2), Name: "bravo"})
		insert(models.Report{ID: idFor(3), Name: "alpha"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "name", "asc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(namesOf(reports)).To(Equal([]string{"alpha", "bravo", "charlie"}))
	})

	It("sorts by name descending rather than by id", func() {
		insert(models.Report{ID: idFor(1), Name: "alpha"})
		insert(models.Report{ID: idFor(2), Name: "bravo"})
		insert(models.Report{ID: idFor(3), Name: "charlie"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "name", "desc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(namesOf(reports)).To(Equal([]string{"charlie", "bravo", "alpha"}))
	})

	It("sorts by status rather than by id", func() {
		insert(models.Report{ID: idFor(1), Name: "one", Status: "Ready"})
		insert(models.Report{ID: idFor(2), Name: "two", Status: "Failed"})
		insert(models.Report{ID: idFor(3), Name: "three", Status: "Deleting"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "status", "asc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(namesOf(reports)).To(Equal([]string{"three", "two", "one"}))
	})

	It("sorts by created_by rather than by id", func() {
		insert(models.Report{ID: idFor(1), Name: "one", CreatedBy: "zoe@run.ai"})
		insert(models.Report{ID: idFor(2), Name: "two", CreatedBy: "mia@run.ai"})
		insert(models.Report{ID: idFor(3), Name: "three", CreatedBy: "ada@run.ai"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "created_by", "asc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(namesOf(reports)).To(Equal([]string{"three", "two", "one"}))
	})

	It("sorts by created_at rather than by id", func() {
		base := time.Now().UTC().Truncate(time.Second)
		insert(models.Report{ID: idFor(1), Name: "one", CreatedAt: base.Add(2 * time.Hour)})
		insert(models.Report{ID: idFor(2), Name: "two", CreatedAt: base.Add(time.Hour)})
		insert(models.Report{ID: idFor(3), Name: "three", CreatedAt: base})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "created_at", "asc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(namesOf(reports)).To(Equal([]string{"three", "two", "one"}))
	})

	It("falls back to id as a tiebreaker when the sorted column ties", func() {
		insert(models.Report{ID: idFor(3), Name: "charlie", Status: "Ready"})
		insert(models.Report{ID: idFor(1), Name: "alpha", Status: "Ready"})
		insert(models.Report{ID: idFor(2), Name: "bravo", Status: "Ready"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "status", "asc", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(idsOf(reports)).To(Equal([]uuid.UUID{idFor(1), idFor(2), idFor(3)}))
	})

	It("keeps sorting stable across paged requests", func() {
		insert(models.Report{ID: idFor(1), Name: "charlie"})
		insert(models.Report{ID: idFor(2), Name: "bravo"})
		insert(models.Report{ID: idFor(3), Name: "alpha"})

		firstPage, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "name", "asc", 0, 2, "")
		Expect(apiErr).To(BeNil())
		Expect(namesOf(firstPage)).To(Equal([]string{"alpha", "bravo"}))

		secondPage, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "name", "asc", 2, 2, "")
		Expect(apiErr).To(BeNil())
		Expect(namesOf(secondPage)).To(Equal([]string{"charlie"}))
	})

	It("orders by id when no sort is requested", func() {
		insert(models.Report{ID: idFor(3), Name: "charlie"})
		insert(models.Report{ID: idFor(1), Name: "alpha"})
		insert(models.Report{ID: idFor(2), Name: "bravo"})

		reports, apiErr := reportsDao.GetReports(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(idsOf(reports)).To(Equal([]uuid.UUID{idFor(1), idFor(2), idFor(3)}))
	})
})
