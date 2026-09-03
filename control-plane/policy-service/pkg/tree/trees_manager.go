/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/concurrency"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type WorkloadType string

const (
	WorkloadTypeWorkspace            WorkloadType = "Workspace"
	WorkloadTypeTraining                          = "Training"
	WorkloadTypeDistributed                       = "Distributed"
	WorkloadTypeInference                         = "Inference"
	WorkloadTypeDistributedInference              = "DistributedInference"
	WorkloadTypeNimService                        = "NimService"
	WorkloadTypeFallback                          = dao.FallbackPolicyAsType
)

var AllWorkloadTypes = []WorkloadType{
	WorkloadTypeWorkspace,
	WorkloadTypeTraining,
	WorkloadTypeDistributed,
	WorkloadTypeInference,
	WorkloadTypeDistributedInference,
	WorkloadTypeNimService,
	WorkloadTypeFallback,
}

// TreesManager holds the memory trees. we have one tree for
// each tenant x workloadType
type TreesManager struct {
	treesMap concurrency.ConcurrentMap[*HierarchyTree]
}

var (
	HierarchyTrees = newTreesManager()
)

func newTreesManager() *TreesManager {
	return &TreesManager{
		treesMap: concurrency.NewConcurrentMap[*HierarchyTree](nil),
	}
}

func (tm *TreesManager) Get(ctx context.Context, tenantId int32, workloadType WorkloadType) (*HierarchyTree, bool) {
	// xsync.Map requires the key to be a string, so we concatenate the tree identifiers to a string
	treeId := MakeTreeId(tenantId, workloadType)

	tree, alreadyIn := tm.treesMap.Load(treeId.ToString())
	if alreadyIn {
		return tree, true
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("[%s] attempt to create tree", treeId.ToString())

	newTree := NewHierarchyTree(treeId.TenantId, treeId.WorkloadType)
	tm.treesMap.Store(treeId.ToString(), newTree)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] tree created", treeId.ToString())

	return newTree, false
}

// Restart resets the in-memory tree manager. It does NOT touch the database -
// callers in the test suite are expected to wipe DB state via
// `database.DeleteAllTablesForTesting()` (a single TRUNCATE), and combining the
// two operations here would just duplicate that work for every spec.
func Restart() {
	HierarchyTrees = newTreesManager()
}
