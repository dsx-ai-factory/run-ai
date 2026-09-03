/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	utilsTableName = "workload_elements"
)

type utilsRow struct {
	ID string
}

func utilsPtr(s string) *uuid.UUID {
	parsed := uuid.MustParse(s)
	return &parsed
}

type utilsDeleteQueryCase struct {
	clusterId    string
	projectId    string
	workloadId   string
	from         *uuid.UUID
	to           *uuid.UUID
	excludingIds []string
	expectedSQL  string
	expectedArgs []any
}

const (
	utilsFromID = "11111111-1111-1111-1111-111111111111"
	utilsToID   = "22222222-2222-2222-2222-222222222222"
)

var _ = Describe("ConstructDeleteQueryForProjectCluster", func() {
	DescribeTable("builds the expected query for the given filters",
		func(testCase utilsDeleteQueryCase) {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(testCase.expectedSQL, testCase.expectedArgs...).Returns([]utilsRow{{ID: "row-1"}})

			query := ConstructDeleteQueryForProjectCluster(dao_utils.DB(), utilsTableName, 7,
				testCase.clusterId, testCase.projectId, testCase.from, testCase.to, testCase.excludingIds, testCase.workloadId)

			var rows []utilsRow
			result := query.Find(&rows)

			Expect(result.Error).ToNot(HaveOccurred())
			Expect(rows).To(HaveLen(1))
		},
		Entry("no optional filters", utilsDeleteQueryCase{
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ?",
			expectedArgs: []any{7},
		}),
		Entry("cluster only", utilsDeleteQueryCase{
			clusterId:    "cluster-1",
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND cluster_id = ?",
			expectedArgs: []any{7, "cluster-1"},
		}),
		Entry("project only", utilsDeleteQueryCase{
			projectId:    "project-1",
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND project_id = ?",
			expectedArgs: []any{7, "project-1"},
		}),
		Entry("workload only", utilsDeleteQueryCase{
			workloadId:   "workload-1",
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND workload_id = ?",
			expectedArgs: []any{7, "workload-1"},
		}),
		Entry("from only", utilsDeleteQueryCase{
			from:         utilsPtr(utilsFromID),
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND id >= ?",
			expectedArgs: []any{7, utilsFromID},
		}),
		Entry("to only", utilsDeleteQueryCase{
			to:           utilsPtr(utilsToID),
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND id <= ?",
			expectedArgs: []any{7, utilsToID},
		}),
		Entry("from and to", utilsDeleteQueryCase{
			from:         utilsPtr(utilsFromID),
			to:           utilsPtr(utilsToID),
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND id >= ? AND id <= ?",
			expectedArgs: []any{7, utilsFromID, utilsToID},
		}),
		Entry("excluding ids only", utilsDeleteQueryCase{
			excludingIds: []string{"e1", "e2"},
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND NOT id IN (?,?)",
			expectedArgs: []any{7, "e1", "e2"},
		}),
		Entry("cluster from to and excluding ids", utilsDeleteQueryCase{
			clusterId:    "cluster-1",
			from:         utilsPtr(utilsFromID),
			to:           utilsPtr(utilsToID),
			excludingIds: []string{"e1"},
			expectedSQL:  "SELECT * FROM workload_elements WHERE tenant_id = ? AND cluster_id = ? AND id >= ? AND id <= ? AND NOT id IN (?)",
			expectedArgs: []any{7, "cluster-1", utilsFromID, utilsToID, "e1"},
		}),
	)
})

var _ = Describe("createMapFromSlice", func() {
	It("returns an empty map for nil and empty input", func() {
		Expect(createMapFromSlice(nil)).To(Equal(map[string]bool{}))
		Expect(createMapFromSlice([]string{})).To(Equal(map[string]bool{}))
	})

	It("collects the items into a set and drops duplicates", func() {
		Expect(createMapFromSlice([]string{"a", "b", "a"})).To(Equal(map[string]bool{"a": true, "b": true}))
	})
})
