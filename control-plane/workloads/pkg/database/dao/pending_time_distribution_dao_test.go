/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
)

const (
	ptdP95Select            = "SELECT COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (NOW() - phase_updated_at))::BIGINT), 0) as p95_pending FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND cluster_id = ? AND phase IN (?)"
	ptdP95NoClusterSelect   = "SELECT COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (NOW() - phase_updated_at))::BIGINT), 0) as p95_pending FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND phase IN (?)"
	ptdP95WithNodepoolSelec = "SELECT COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (NOW() - phase_updated_at))::BIGINT), 0) as p95_pending FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND cluster_id = ? AND array_length(requested_nodepools, 1) = 1 AND ? = ANY(requested_nodepools) AND phase IN (?)"

	ptdBucketsSelect = "SELECT SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 0 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 900 THEN 1 ELSE 0 END) as bucket_1, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 900 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 1800 THEN 1 ELSE 0 END) as bucket_2, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 1800 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 2700 THEN 1 ELSE 0 END) as bucket_3, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 2700 THEN 1 ELSE 0 END) as bucket_4 FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND cluster_id = ? AND phase IN (?)"

	ptdBucketsGroupedSelect = "SELECT SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 0 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 1000 THEN 1 ELSE 0 END) as bucket_1, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 1000 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 2000 THEN 1 ELSE 0 END) as bucket_2, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 2000 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 3000 THEN 1 ELSE 0 END) as bucket_3, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 3000 THEN 1 ELSE 0 END) as bucket_4, cluster_id, cluster_name FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND ((false OR project_id IN ( 'p1'))) AND cluster_id = ? AND array_length(requested_nodepools, 1) = 1 AND ? = ANY(requested_nodepools) AND department_id = ? AND project_id = ? AND phase IN (?) GROUP BY cluster_id,cluster_name"

	ptdBucketsSingleGroupSelect = "SELECT SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 0 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 900 THEN 1 ELSE 0 END) as bucket_1, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 900 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 1800 THEN 1 ELSE 0 END) as bucket_2, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 1800 AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < 2700 THEN 1 ELSE 0 END) as bucket_3, SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= 2700 THEN 1 ELSE 0 END) as bucket_4, project_id FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND phase IN (?) GROUP BY project_id"
)

type ptdP95Row struct {
	P95Pending float64 `gorm:"column:p95_pending"`
}

type ptdBucketCountsRow struct {
	Bucket1 int64 `gorm:"column:bucket_1"`
	Bucket2 int64 `gorm:"column:bucket_2"`
	Bucket3 int64 `gorm:"column:bucket_3"`
	Bucket4 int64 `gorm:"column:bucket_4"`
}

type ptdBucketCountsWithClusterRow struct {
	Bucket1     int64  `gorm:"column:bucket_1"`
	Bucket2     int64  `gorm:"column:bucket_2"`
	Bucket3     int64  `gorm:"column:bucket_3"`
	Bucket4     int64  `gorm:"column:bucket_4"`
	ClusterId   string `gorm:"column:cluster_id"`
	ClusterName string `gorm:"column:cluster_name"`
}

type ptdBucketCountsWithProjectRow struct {
	Bucket1   int64  `gorm:"column:bucket_1"`
	Bucket2   int64  `gorm:"column:bucket_2"`
	Bucket3   int64  `gorm:"column:bucket_3"`
	Bucket4   int64  `gorm:"column:bucket_4"`
	ProjectId string `gorm:"column:project_id"`
}

func ptdDao() *WorkloadsDao {
	return &WorkloadsDao{view: WorkloadsViewName}
}

func ptdScopes() *openapi.PermittedScopes {
	system := false
	return &openapi.PermittedScopes{System: &system, Projects: []string{"p1"}}
}

var _ = Describe("getBucketColumnName", func() {
	It("maps zero based indexes to one based column names", func() {
		Expect(getBucketColumnName(0)).To(Equal("bucket_1"))
		Expect(getBucketColumnName(3)).To(Equal("bucket_4"))
	})
})

var _ = Describe("getClusterP95PendingTime", func() {
	It("returns the truncated p95 value", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95WithNodepoolSelec).Returns([]ptdP95Row{{P95Pending: 1234.9}})

		p95, err := ptdDao().getClusterP95PendingTime(testCtx(1), 1, "c1", "np1")

		Expect(err).ToNot(HaveOccurred())
		Expect(p95).To(Equal(int64(1234)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95NoClusterSelect).ReturnsError(errors.New("boom"))

		p95, err := ptdDao().getClusterP95PendingTime(testCtx(1), 1, "", "")

		Expect(err).To(HaveOccurred())
		Expect(p95).To(Equal(int64(0)))
	})
})

var _ = Describe("DbGetPendingTimeDistribution", func() {
	It("returns an error when the p95 query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95Select).ReturnsError(errors.New("boom"))

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "c1", "", "", "", nil, nil)

		Expect(values).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the bucket query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95Select).Returns([]ptdP95Row{{P95Pending: 0}})
		m.ExpectQuery(ptdBucketsSelect).ReturnsError(errors.New("boom"))

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "c1", "", "", "", nil, nil)

		Expect(values).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns empty buckets when there are no results", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95Select).Returns([]ptdP95Row{{P95Pending: 0}})
		m.ExpectQuery(ptdBucketsSelect).Returns([]ptdBucketCountsRow{})

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "c1", "", "", "", nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(4))
		for _, value := range *values {
			Expect(value.Data).To(Equal("0"))
			Expect(value.Group1MainName).To(Equal(bucketGroupKey))
			Expect(value.Group1SecondaryName).To(Equal(bucketGroupKey))
		}
		Expect((*values)[0].Group1MainValue).To(Equal("(0)-(15m)"))
		Expect((*values)[3].Group1MainValue).To(Equal("(45m)+"))
	})

	It("returns the bucket counts", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95Select).Returns([]ptdP95Row{{P95Pending: 0}})
		m.ExpectQuery(ptdBucketsSelect).Returns([]ptdBucketCountsRow{{Bucket1: 5, Bucket2: 3, Bucket3: 0, Bucket4: 2}})

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "c1", "", "", "", nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(4))
		Expect((*values)[0].Data).To(Equal("5"))
		Expect((*values)[1].Data).To(Equal("3"))
		Expect((*values)[2].Data).To(Equal("0"))
		Expect((*values)[3].Data).To(Equal("2"))
		Expect((*values)[0].Group2MainValue).To(BeEmpty())
	})

	It("groups by a value and a name field", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95WithNodepoolSelec).Returns([]ptdP95Row{{P95Pending: 4000}})
		m.ExpectQuery(ptdBucketsGroupedSelect).Returns([]ptdBucketCountsWithClusterRow{{
			Bucket1: 1, Bucket2: 2, Bucket3: 3, Bucket4: 4, ClusterId: "c1", ClusterName: "cluster-one",
		}})

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "c1", "np1", "d1", "p1", ptdScopes(),
			[][]string{{"cluster_id", "cluster_name"}})

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(4))
		for _, value := range *values {
			Expect(value.Group2MainValue).To(Equal("c1"))
			Expect(value.Group2SecondaryValue).To(Equal("cluster-one"))
		}
		Expect((*values)[0].Data).To(Equal("1"))
		Expect((*values)[3].Data).To(Equal("4"))
	})

	It("groups by a single field", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(ptdP95NoClusterSelect).Returns([]ptdP95Row{{P95Pending: 0}})
		m.ExpectQuery(ptdBucketsSingleGroupSelect).Returns([]ptdBucketCountsWithProjectRow{{
			Bucket1: 1, Bucket2: 1, Bucket3: 1, Bucket4: 1, ProjectId: "p1",
		}})

		values, apiErr := ptdDao().DbGetPendingTimeDistribution(testCtx(1), 1, "", "", "", "", nil, [][]string{{"project_id"}})

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(4))
		for _, value := range *values {
			Expect(value.Group2MainValue).To(Equal("p1"))
			Expect(value.Group2SecondaryValue).To(Equal("p1"))
		}
	})
})

var _ = Describe("createEmptyPendingTimeBuckets", func() {
	It("returns one zeroed telemetry value per bucket", func() {
		buckets := services_utils.CalculatePendingTimeBuckets(0)

		values := createEmptyPendingTimeBuckets(buckets)

		Expect(*values).To(HaveLen(4))
		for i, value := range *values {
			Expect(value.Data).To(Equal("0"))
			Expect(value.Group1MainValue).To(Equal(buckets[i].Name))
		}
	})
})

var _ = Describe("createPendingTimeBucketTelemetry", func() {
	It("populates the first grouping from the bucket", func() {
		bucket := services_utils.Bucket{MinSeconds: 0, MaxSeconds: 900, Name: "(0)-(15m)"}

		value := createPendingTimeBucketTelemetry("12", bucket)

		Expect(value.Data).To(Equal("12"))
		Expect(value.Group1MainName).To(Equal(bucketGroupKey))
		Expect(value.Group1MainValue).To(Equal("(0)-(15m)"))
		Expect(value.Group1SecondaryName).To(Equal(bucketGroupKey))
		Expect(value.Group1SecondaryValue).To(Equal("(0)-(15m)"))
	})
})

var _ = Describe("transformPendingTimeResultsToTelemetry", func() {
	It("treats missing and nil counts as zero", func() {
		buckets := services_utils.CalculatePendingTimeBuckets(0)
		var nilPtr *int64
		results := []map[string]interface{}{{
			"bucket_1": nil,
			"bucket_2": "",
			"bucket_3": nilPtr,
			"bucket_4": int64(7),
		}}

		values := transformPendingTimeResultsToTelemetry(results, buckets, nil)

		Expect(values).To(HaveLen(4))
		Expect(values[0].Data).To(Equal("0"))
		Expect(values[1].Data).To(Equal("0"))
		Expect(values[2].Data).To(Equal("0"))
		Expect(values[3].Data).To(Equal("7"))
	})

	It("emits a bucket set per result row", func() {
		buckets := services_utils.CalculatePendingTimeBuckets(0)
		results := []map[string]interface{}{
			{"bucket_1": 1, "bucket_2": 2, "bucket_3": 3, "bucket_4": 4, "project_id": "p1"},
			{"bucket_1": 5, "bucket_2": 6, "bucket_3": 7, "bucket_4": 8, "project_id": "p2"},
		}

		values := transformPendingTimeResultsToTelemetry(results, buckets, [][]string{{"project_id"}})

		Expect(values).To(HaveLen(8))
		Expect(values[0].Group2MainValue).To(Equal("p1"))
		Expect(values[4].Group2MainValue).To(Equal("p2"))
	})
})

var _ = Describe("addPendingTimeEntityGrouping", func() {
	It("leaves the grouping empty when both fields are nil", func() {
		value := createPendingTimeBucketTelemetry("0", services_utils.Bucket{Name: "b"})

		addPendingTimeEntityGrouping(&value, map[string]interface{}{"cluster_id": nil, "cluster_name": nil},
			[]string{"cluster_id", "cluster_name"})

		Expect(value.Group2MainValue).To(BeEmpty())
		Expect(value.Group2SecondaryValue).To(BeEmpty())
	})

	It("leaves the grouping empty when the single field is nil", func() {
		value := createPendingTimeBucketTelemetry("0", services_utils.Bucket{Name: "b"})

		addPendingTimeEntityGrouping(&value, map[string]interface{}{"project_id": nil}, []string{"project_id"})

		Expect(value.Group2MainValue).To(BeEmpty())
		Expect(value.Group2SecondaryValue).To(BeEmpty())
	})

	It("leaves the grouping empty when there are no fields", func() {
		value := createPendingTimeBucketTelemetry("0", services_utils.Bucket{Name: "b"})

		addPendingTimeEntityGrouping(&value, map[string]interface{}{}, nil)

		Expect(value.Group2MainValue).To(BeEmpty())
		Expect(value.Group2SecondaryValue).To(BeEmpty())
	})
})
