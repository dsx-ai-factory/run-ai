/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

type DbOperation int

const (
	GET DbOperation = iota
	LIST
	COUNT
	CREATE
	UPDATE
	DELETE
	UPSERT
	COMMIT_CREATIONS
	COMMIT_UPDATES
	COMMIT_DELETIONS
)

func (op DbOperation) String() string {
	return [...]string{
		OperationGet,
		OperationList,
		OperationCount,
		OperationCreate,
		OperationUpdate,
		OperationDelete,
		OperationUpsert,
		OperationCommitCreations,
		OperationCommitUpdates,
		OperationCommitDeletions,
	}[op]
}

type DbEntity string

const (
	Policies           DbEntity = "policies"
	EffectivePolicy    DbEntity = "EffectivePolicy"
	UpdateCounter      DbEntity = "EffectivePolicySequence"
	ApprovedDeletions  DbEntity = "ApprovedDeletions"
	MappingFormatters  DbEntity = "MappingFormatters"
	MappingFieldGroups DbEntity = "MappingFieldGroups"
	MappingSelectors   DbEntity = "MappingSelectors"
	MappingLocking     DbEntity = "MappingLocking"
	EntitiesCache      DbEntity = "EntitiesCache"
	EntitiesDependency DbEntity = "EntitiesDependency"
)

func Error(operation DbOperation, entity DbEntity) *api.Error {
	msg := fmt.Sprintf("failed to %s requested %s", operation, entity)
	return api.NewError(http.StatusServiceUnavailable, msg, nil)
}

func CleanMap(aMap dbmodels.JsonBMap) dbmodels.JsonBMap {
	if aMap != nil {
		gojsonq.CleanMap(aMap)
	}
	if len(aMap) == 0 {
		return nil
	}
	return aMap
}

// tablesWithoutSystemRows are wiped fully between specs: nothing in them is
// shared by the suite-wide system seed. Listed together in a single TRUNCATE so
// the entities_dependency -> entities_cache foreign key is satisfied without
// CASCADE.
var tablesWithoutSystemRows = []string{
	dbtables.ApprovedDeletionsTable,
	dbtables.MappingLockingTable,
	dbtables.EntitiesDependencyTable,
	dbtables.EntitiesCacheTable,
}

// tablesWithSystemRows are partially wiped: only non-system rows are removed.
// System-scoped rows (one system policy + the system formatters / field groups /
// mappings) are seeded once at suite start by dao.HandleSystemPolicy() and
// services.HandleSystemMapping() and never change for the lifetime of the
// suite, so re-seeding them per spec is wasted work.
//
// Tests that exercise the seeding mechanism itself MUST use names that don't
// collide with the real system rows (test-* style) and clean up after
// themselves; otherwise their custom system-scoped rows will leak across specs.
var tablesWithSystemRows = []struct {
	table       string
	scopeColumn string
}{
	{dbtables.PoliciesTable, dbtables.PoliciesColumnScope},
	{dbtables.MappingFormattersTable, dbtables.MappingFormattersColumnScopeType},
	{dbtables.MappingFieldGroupsTable, dbtables.MappingFieldGroupsColumnScopeType},
	{dbtables.MappingSelectorsTable, dbtables.MappingSelectorsColumnScopeType},
}

// DeleteAllTablesForTesting wipes tenant/user data between specs while preserving
// the suite-wide system seed.
//
// Hard DELETE / TRUNCATE is used (not gorm's soft delete) so rows don't accumulate
// across the ~470 specs in the services suite. effective_policies is purged of any
// row whose linked policy is gone, so it implicitly preserves rows tied to the
// surviving system policy (currently none, but that path is preserved from the
// previous implementation in case future code starts persisting one).
func DeleteAllTablesForTesting() {
	db := runaidb.DB()

	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY",
		strings.Join(tablesWithoutSystemRows, ", ")))
	Expect(err).ToNot(HaveOccurred())

	for _, t := range tablesWithSystemRows {
		_, err = db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s != 'system'", t.table, t.scopeColumn))
		Expect(err).ToNot(HaveOccurred())
	}

	_, err = db.Exec(fmt.Sprintf(
		"DELETE FROM %s WHERE %s NOT IN (SELECT %s FROM %s)",
		dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId,
		dbtables.PoliciesColumnId, dbtables.PoliciesTable))
	Expect(err).ToNot(HaveOccurred())
}
