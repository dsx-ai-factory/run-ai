/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Scenario tests for up/03000001_migrate_dgxc_roles_to_custom.up.sql.
//
// Each case describes a starting database state - roles, access rules - runs the real
// migration file against a real Postgres, and asserts the roles, access rules and
// disabled_roles rows it produces. Assertions are on exact sets, so anything the
// migration creates or destroys unexpectedly shows up as a diff.
//
// To add a case, append a migrationCase to migrationCases. Nothing else changes.

package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	authtesting "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
)

const (
	runAiCreator  = "Run:ai"
	migrationUp   = "../migrations/up/03000001_migrate_dgxc_roles_to_custom.up.sql"
	migrationDown = "../migrations/down/03000001_migrate_dgxc_roles_to_custom.down.sql"
	// 03000002 drops role_type, which 03000001 reads. The suite builds its schema
	// from every migration except that one, so the tests run against exactly the
	// state 03000001 sees in production.
	skipMigration = "03000002"
	migrationsDir = "../migrations"
)

// ---------------------------------------------------------------- seed types

type seedRole struct {
	Name           string
	Type           string // DGXC, NON_DGXC or ALL
	Custom         bool
	Scope          string // "system" (default) or "tenant"
	ScopeID        string // tenant id, for tenant-scoped roles
	Creator        string // defaults to Run:ai
	PermissionSets []string
}

type seedRule struct {
	Subject string
	Role    string
	Tenant  string // "" writes tenant_id NULL
	Deleted bool   // soft-deleted
}

// ---------------------------------------------------------------- want types

type wantRole struct {
	Name     string
	Scope    string
	ScopeID  string
	Custom   bool
	Disabled bool // roles.enabled = false
}

type wantRule struct {
	Subject string
	Role    string
}

type wantDisabled struct {
	Role   string
	Tenant string
}

type migrationCase struct {
	Name  string
	Roles []seedRole
	Rules []seedRule

	WantRoles    []wantRole
	WantRules    []wantRule
	WantDisabled []wantDisabled
	// Optional: permission-set names expected on a role after the migration.
	WantPermissionSets map[string][]string
	// tenant id -> tenantIsCspManaged value, seeded into the settings table.
	CspSettings map[string]bool
	// Roles seeded after the main set, so they can point at one of them.
	ExtraTenantRoles []seedRole
	// role name -> name of the role it uses as its predefined cluster role
	PredefinedRoleOf map[string]string
	// role name -> expected predefined cluster role name after the migration
	WantPredefinedFor map[string]string
	// Set to the reason when up -> down cannot return to the seeded state, i.e. the
	// migration destroys something it has no way to rebuild. Empty means the case is
	// included in the round-trip test.
	NotReversible string
}

// ---------------------------------------------------------------- the cases

// Shorthands for the roles these tests lean on repeatedly.
func dgxc(name string, sets ...string) seedRole {
	return seedRole{Name: name, Type: "DGXC", PermissionSets: sets}
}

func twin(name string) seedRole {
	return seedRole{Name: name, Type: "NON_DGXC"}
}

func plain(name string) seedRole {
	return seedRole{Name: name, Type: "ALL"}
}

var migrationCases = []migrationCase{
	// ---- the delete branch: never a CSP environment, nothing assigned -------------
	{
		Name:          "no CSP setting and no rules: DGXC roles are deleted",
		Roles:         []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator"), plain("Editor")},
		Rules:         []seedRule{{Subject: "alice", Role: "Editor", Tenant: "1"}},
		WantRoles:     []wantRole{{Name: "Cloud operator", Scope: "system"}, {Name: "Editor", Scope: "system"}},
		WantRules:     []wantRule{{Subject: "alice", Role: "Editor"}},
		NotReversible: "the roles are deleted, and nothing is left to rebuild them from",
	},
	{
		// The tenantIsCspManaged setting has no bearing on the decision - only what
		// actually references the roles does.
		Name:          "a CSP-managed tenant does not save an unreferenced role",
		Roles:         []seedRole{dgxc("Cloud support")},
		CspSettings:   map[string]bool{"1": true},
		WantRoles:     nil,
		NotReversible: "the role is deleted",
	},
	{
		Name:      "new environment: no DGXC roles at all, nothing changes",
		Roles:     []seedRole{plain("Editor"), twin("Department administrator")},
		Rules:     []seedRule{{Subject: "alice", Role: "Editor", Tenant: "1"}},
		WantRoles: []wantRole{{Name: "Department administrator", Scope: "system"}, {Name: "Editor", Scope: "system"}},
		WantRules: []wantRule{{Subject: "alice", Role: "Editor"}},
	},

	// ---- the rename branch: something references a DGXC role ---------------------
	{
		// No access rule anywhere - the pointer alone is enough to keep the set.
		Name:  "a custom role naming a DGXC role as its predefined cluster role keeps the set",
		Roles: []seedRole{dgxc("Cloud operator-DGXC", "Workspace edit"), twin("Cloud operator")},
		ExtraTenantRoles: []seedRole{
			{Name: "Tenant custom role", Type: "ALL", Custom: true, Scope: "tenant", ScopeID: "1", Creator: "user-1"},
		},
		PredefinedRoleOf: map[string]string{"Tenant custom role": "Cloud operator-DGXC"},
		// flattened past the renamed role to the twin it now points at
		WantPredefinedFor: map[string]string{"Tenant custom role": "Cloud operator"},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Tenant custom role", Scope: "tenant", ScopeID: "1", Custom: true},
		},
		WantPermissionSets: map[string][]string{"Cloud operator-Superseded": {"Workspace edit"}},
	},
	{
		// Same pointer, but from a non-custom role: not a reference, so the set goes.
		//
		// The pointer is left aimed at the deleted row rather than cleared. Only
		// creationRequestToRoleDao writes that column and it sets custom = true, so a
		// non-custom role carrying one is not expected to exist; the migration does not
		// pay for the case.
		Name:  "a non-custom role's predefined pointer does not keep the set",
		Roles: []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator")},
		ExtraTenantRoles: []seedRole{
			{Name: "Another system role", Type: "ALL"},
		},
		PredefinedRoleOf: map[string]string{"Another system role": "Cloud operator-DGXC"},
		WantRoles: []wantRole{
			{Name: "Another system role", Scope: "system"},
			{Name: "Cloud operator", Scope: "system"},
		},
		NotReversible: "the role is deleted",
	},
	{
		Name:  "an access rule on a DGXC role is enough",
		Roles: []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator")},
		Rules: []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},
	{
		// The agent prefers PredefinedRole.Name when deriving a ClusterRole name, so
		// each renamed role points at its twin and bindings keep resolving to the
		// ClusterRole they use today.
		Name: "each renamed role points at its non-DGXC twin",
		Roles: []seedRole{
			dgxc("Cloud operator-DGXC"), twin("Cloud operator"),
			dgxc("Department administrator-DGXC"), twin("Department administrator"),
			dgxc("Department administrator-DGXC (Legacy)"), twin("Department administrator (Legacy)"),
			dgxc("Application administrator"), dgxc("Cloud support"),
			// the cluster remaps these two to editor/viewer, so that is what the
			// predefined role has to name
			plain("Editor"), plain("Viewer"),
		},
		Rules: []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		WantPredefinedFor: map[string]string{
			"Cloud operator-Superseded":                    "Cloud operator",
			"Department administrator-Superseded":          "Department administrator",
			"Department administrator (Legacy)-Superseded": "Department administrator (Legacy)",
			// no twin: point at what the cluster's RoleNamesMapping resolves them to
			"Application administrator-Superseded": "Editor",
			"Cloud support-Superseded":             "Viewer",
		},
		WantRoles: []wantRole{
			{Name: "Application administrator-Superseded", Scope: "system", Custom: true},
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Cloud support-Superseded", Scope: "system", Custom: true},
			{Name: "Department administrator", Scope: "system"},
			{Name: "Department administrator (Legacy)", Scope: "system"},
			{Name: "Department administrator (Legacy)-Superseded", Scope: "system", Custom: true},
			{Name: "Department administrator-Superseded", Scope: "system", Custom: true},
			{Name: "Editor", Scope: "system"},
			{Name: "Viewer", Scope: "system"},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},
	{
		// A custom role already naming a DGXC role as its predefined cluster role would
		// otherwise be left pointing at the '-Superseded' name - getPredefinedRoleForSync
		// only resolves one hop, so the agent would derive a ClusterRole that does not
		// exist. The chain is flattened to the twin instead.
		Name:  "a custom role pointing at a migrated role is flattened to its target",
		Roles: []seedRole{dgxc("Department administrator-DGXC"), twin("Department administrator")},
		ExtraTenantRoles: []seedRole{
			{Name: "dep admin predefined", Type: "ALL", Custom: true,
				Scope: "tenant", ScopeID: "1001", Creator: "test@run.ai"},
		},
		PredefinedRoleOf: map[string]string{"dep admin predefined": "Department administrator-DGXC"},
		Rules:            []seedRule{{Subject: "alice", Role: "Department administrator-DGXC", Tenant: "1001"}},
		WantPredefinedFor: map[string]string{
			// flattened past the renamed role, straight to the twin
			"dep admin predefined": "Department administrator",
			// the renamed role keeps its own pointer
			"Department administrator-Superseded": "Department administrator",
		},
		WantRoles: []wantRole{
			{Name: "Department administrator", Scope: "system"},
			{Name: "Department administrator-Superseded", Scope: "system", Custom: true},
			{Name: "dep admin predefined", Scope: "tenant", ScopeID: "1001", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Department administrator-Superseded"}},
	},
	{
		// The flatten only applies to roles this migration renamed. A customer's own
		// system-scoped role carrying role_type DGXC and a predefined pointer is not
		// one of them, so the pointer at it is left alone.
		Name:  "a pointer at a DGXC-typed role this migration did not rename is left alone",
		Roles: []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator"), plain("Editor")},
		ExtraTenantRoles: []seedRole{
			{Name: "Customer DGXC role", Type: "DGXC", Custom: true, Creator: "user-7"},
			{Name: "Points at it", Type: "ALL", Custom: true, Scope: "tenant", ScopeID: "1", Creator: "user-7"},
		},
		PredefinedRoleOf: map[string]string{
			"Customer DGXC role": "Editor",
			"Points at it":       "Customer DGXC role",
		},
		Rules: []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		WantPredefinedFor: map[string]string{
			// not flattened to Editor - "Customer DGXC role" has no -Superseded suffix
			"Points at it": "Customer DGXC role",
		},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Customer DGXC role", Scope: "system", Custom: true},
			{Name: "Editor", Scope: "system"},
			{Name: "Points at it", Scope: "tenant", ScopeID: "1", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},
	{
		// A pointer already on the role wins over the twin - the migration only fills
		// in a NULL one.
		Name: "an existing predefined role is not overwritten",
		Roles: []seedRole{
			dgxc("Cloud operator-DGXC"), twin("Cloud operator"), plain("Editor"),
		},
		Rules:             []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		PredefinedRoleOf:  map[string]string{"Cloud operator-DGXC": "Editor"},
		WantPredefinedFor: map[string]string{"Cloud operator-Superseded": "Editor"},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Editor", Scope: "system"},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},
	{
		// The twin must be a real non-DGXC role. Here a system-scoped custom role
		// happens to carry both the twin's name and role_type DGXC; it is not renamed
		// (custom = false is required for that), so there is no name collision - but
		// pointing at it would derive a ClusterRole name that does not exist.
		Name: "a DGXC-typed role is not mistaken for the twin",
		Roles: []seedRole{
			dgxc("Department administrator-DGXC"),
			{Name: "Department administrator", Type: "DGXC", Custom: true, Creator: "user-3"},
		},
		Rules:             []seedRule{{Subject: "alice", Role: "Department administrator-DGXC", Tenant: "1"}},
		WantPredefinedFor: map[string]string{"Department administrator-Superseded": ""},
		WantRoles: []wantRole{
			{Name: "Department administrator", Scope: "system", Custom: true},
			{Name: "Department administrator-Superseded", Scope: "system", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Department administrator-Superseded"}},
	},
	{
		Name:  "a soft-deleted rule counts as a reference",
		Roles: []seedRole{dgxc("Cloud support")},
		Rules: []seedRule{{Subject: "erin", Role: "Cloud support", Tenant: "3", Deleted: true}},
		WantRoles: []wantRole{
			{Name: "Cloud support-Superseded", Scope: "system", Custom: true},
		},
		WantRules: []wantRule{{Subject: "erin", Role: "Cloud support-Superseded"}},
	},
	{
		Name:  "a rule with no tenant counts as a reference",
		Roles: []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator")},
		Rules: []seedRule{{Subject: "grace", Role: "Cloud operator-DGXC", Tenant: ""}},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
		},
		WantRules: []wantRule{{Subject: "grace", Role: "Cloud operator-Superseded"}},
	},
	{
		Name: "one role in use renames the whole DGXC set",
		Roles: []seedRole{
			dgxc("Cloud operator-DGXC"),
			dgxc("Application administrator"),
			dgxc("Cloud support"),
			dgxc("Department administrator-DGXC"),
			dgxc("Department administrator-DGXC (Legacy)"),
			twin("Cloud operator"),
			twin("Department administrator"),
			twin("Department administrator (Legacy)"),
		},
		Rules: []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		WantRoles: []wantRole{
			{Name: "Application administrator-Superseded", Scope: "system", Custom: true},
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Cloud support-Superseded", Scope: "system", Custom: true},
			{Name: "Department administrator", Scope: "system"},
			{Name: "Department administrator (Legacy)", Scope: "system"},
			{Name: "Department administrator (Legacy)-Superseded", Scope: "system", Custom: true},
			{Name: "Department administrator-Superseded", Scope: "system", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},
	{
		Name:  "a role pointing at a renamed role keeps its predefined cluster role",
		Roles: []seedRole{dgxc("Cloud operator-DGXC"), twin("Cloud operator")},
		Rules: []seedRule{{Subject: "alice", Role: "Cloud operator-DGXC", Tenant: "1"}},
		ExtraTenantRoles: []seedRole{
			{Name: "Tenant custom role", Type: "ALL", Custom: true, Scope: "tenant", ScopeID: "1", Creator: "user-1"},
		},
		PredefinedRoleOf: map[string]string{"Tenant custom role": "Cloud operator-DGXC"},
		// flattened past the renamed role to the twin it now points at
		WantPredefinedFor: map[string]string{"Tenant custom role": "Cloud operator"},
		WantRoles: []wantRole{
			{Name: "Cloud operator", Scope: "system"},
			{Name: "Cloud operator-Superseded", Scope: "system", Custom: true},
			{Name: "Tenant custom role", Scope: "tenant", ScopeID: "1", Custom: true},
		},
		WantRules: []wantRule{{Subject: "alice", Role: "Cloud operator-Superseded"}},
	},

	// ---- roles the migration must never touch ------------------------------------
	{
		Name: "a customer's custom role marked DGXC is left alone",
		Roles: []seedRole{
			{Name: "My custom role", Type: "DGXC", Custom: true, Scope: "tenant", ScopeID: "5", Creator: "user-9"},
		},
		Rules:     []seedRule{{Subject: "judy", Role: "My custom role", Tenant: "5"}},
		WantRoles: []wantRole{{Name: "My custom role", Scope: "tenant", ScopeID: "5", Custom: true}},
		WantRules: []wantRule{{Subject: "judy", Role: "My custom role"}},
	},
	{
		Name:        "NON_DGXC and ALL roles are never touched",
		Roles:       []seedRole{twin("Department administrator"), plain("Editor"), plain("Viewer")},
		CspSettings: map[string]bool{"1": true},
		Rules: []seedRule{
			{Subject: "alice", Role: "Department administrator", Tenant: "1"},
			{Subject: "bob", Role: "Editor", Tenant: "1"},
		},
		WantRoles: []wantRole{
			{Name: "Department administrator", Scope: "system"},
			{Name: "Editor", Scope: "system"},
			{Name: "Viewer", Scope: "system"},
		},
		WantRules: []wantRule{
			{Subject: "alice", Role: "Department administrator"},
			{Subject: "bob", Role: "Editor"},
		},
	},
}

// ---------------------------------------------------------------- the test

func TestDGXCMigration(t *testing.T) {
	authtesting.SetupPostgres(t, "file://"+migrationsWithoutDropColumn(t))
	require.NoError(t, authorization_config.Init())

	for _, c := range migrationCases {
		t.Run(c.Name, func(t *testing.T) {
			resetDB(t)
			seed(t, c)
			execFile(t, migrationUp)
			assertState(t, c)
		})
	}

	// The down migration must return the database to the state it started in, for
	// every case that does not destroy something unrecoverable.
	for _, c := range migrationCases {
		if c.NotReversible != "" {
			continue
		}
		t.Run("down restores: "+c.Name, func(t *testing.T) {
			resetDB(t)
			seed(t, c)
			before := currentState(t)
			execFile(t, migrationUp)
			execFile(t, migrationDown)
			require.Equal(t, before, currentState(t), c.NotReversible)
		})
	}

	t.Run("running the migration twice changes nothing", func(t *testing.T) {
		c := migrationCases[3] // many tenants - the richest case
		resetDB(t)
		seed(t, c)
		execFile(t, migrationUp)
		first := currentState(t)
		execFile(t, migrationUp)
		require.Equal(t, first, currentState(t), "second run must be a no-op")
	})
}

// ---------------------------------------------------------------- helpers

// migrationsWithoutDropColumn copies up/ and down/ into a temp dir, leaving out the
// migration that drops role_type. Returns the temp dir path.
func migrationsWithoutDropColumn(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"up", "down"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, sub), 0o755))
		entries, err := os.ReadDir(filepath.Join(migrationsDir, sub))
		require.NoError(t, err)
		for _, e := range entries {
			if strings.Contains(e.Name(), skipMigration) {
				continue
			}
			body, err := os.ReadFile(filepath.Join(migrationsDir, sub, e.Name()))
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(dir, sub, e.Name()), body, 0o644))
		}
	}
	return dir
}

func execFile(t *testing.T, path string) {
	t.Helper()
	sql, err := os.ReadFile(path)
	require.NoError(t, err, "read %s", path)
	_, err = runaidb.DB().Exec(string(sql))
	require.NoError(t, err, "exec %s", path)
}

func exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	_, err := runaidb.DB().Exec(sql, args...)
	require.NoError(t, err, sql)
}

func resetDB(t *testing.T) {
	t.Helper()
	for _, table := range []string{"access_rules", "disabled_roles", "role_to_permission_set", "permissions", "roles", "permission_sets", "settings"} {
		exec(t, "DELETE FROM "+table)
	}
}

func seed(t *testing.T, c migrationCase) {
	t.Helper()
	for _, r := range c.Roles {
		seedOneRole(t, r)
	}

	for tenant, enabled := range c.CspSettings {
		exec(t, `INSERT INTO settings (id, tenant_id, key, value, updated_at, updated_by_customer)
		         VALUES (gen_random_uuid(), ?, ?, ?::jsonb, NOW(), false)`,
			tenant, "tenantIsCspManaged", fmt.Sprintf(`{"enabled": %t}`, enabled))
	}

	for _, r := range c.ExtraTenantRoles {
		seedOneRole(t, r)
	}
	for from, to := range c.PredefinedRoleOf {
		exec(t, `UPDATE roles SET cluster_role_predefined_id = (SELECT id FROM roles WHERE name = ?) WHERE name = ?`, to, from)
	}

	for _, rule := range c.Rules {
		var tenant any
		if rule.Tenant != "" {
			tenant = rule.Tenant
		}
		deletedAt := "NULL"
		if rule.Deleted {
			deletedAt = "NOW()"
		}
		exec(t, fmt.Sprintf(`INSERT INTO access_rules (created_at, updated_at, created_by, subject_id,
		                             subject_type, role_id, scope_id, scope_type, scope_name,
		                             tenant_id, cluster_uuid, deleted_at)
		         SELECT NOW(), NOW(), 'seed', ?, 'user', r.id, '1', 'tenant', 't', ?, '', %s
		         FROM roles r WHERE r.name = ?`, deletedAt),
			rule.Subject, tenant, rule.Role)
	}
}

type dbState struct {
	Roles    []wantRole
	Rules    []wantRule
	Disabled []wantDisabled
}

func currentState(t *testing.T) dbState {
	t.Helper()
	s := dbState{Roles: []wantRole{}, Rules: []wantRule{}, Disabled: []wantDisabled{}}

	type roleRow struct {
		Name      string
		ScopeType string
		ScopeID   *string
		Custom    bool
		Enabled   bool
	}
	var roles []roleRow
	require.NoError(t, runaidb.DB().Actual().
		Raw(`SELECT name, scope_type, scope_id, custom, enabled FROM roles ORDER BY name, scope_id`).
		Scan(&roles).Error)
	for _, r := range roles {
		id := ""
		if r.ScopeID != nil {
			id = *r.ScopeID
		}
		s.Roles = append(s.Roles, wantRole{Name: r.Name, Scope: r.ScopeType, ScopeID: id, Custom: r.Custom, Disabled: !r.Enabled})
	}

	require.NoError(t, runaidb.DB().Actual().
		Raw(`SELECT a.subject_id AS subject, r.name AS role
		     FROM access_rules a JOIN roles r ON r.id = a.role_id
		     ORDER BY a.subject_id`).
		Scan(&s.Rules).Error)

	require.NoError(t, runaidb.DB().Actual().
		Raw(`SELECT r.name AS role, d.scope_id AS tenant
		     FROM disabled_roles d JOIN roles r ON r.id = d.role_id
		     ORDER BY r.name, d.scope_id`).
		Scan(&s.Disabled).Error)

	return s
}

func assertState(t *testing.T, c migrationCase) {
	t.Helper()
	got := currentState(t)

	require.ElementsMatch(t, c.WantRoles, got.Roles, "roles")
	require.ElementsMatch(t, c.WantRules, got.Rules, "access rules")
	require.ElementsMatch(t, c.WantDisabled, got.Disabled, "disabled roles")

	for roleName, wantTarget := range c.WantPredefinedFor {
		var got []string
		require.NoError(t, runaidb.DB().Actual().
			Raw(`SELECT p.name FROM roles r JOIN roles p ON p.id = r.cluster_role_predefined_id WHERE r.name = ?`, roleName).
			Scan(&got).Error)
		if wantTarget == "" {
			// "" means the pointer must be cleared, not left dangling.
			require.Empty(t, got, "predefined cluster role of %s should be cleared", roleName)
			var dangling []int64
			require.NoError(t, runaidb.DB().Actual().
				Raw(`SELECT cluster_role_predefined_id FROM roles WHERE name = ? AND cluster_role_predefined_id IS NOT NULL`, roleName).
				Scan(&dangling).Error)
			require.Empty(t, dangling, "%s still points at a deleted role", roleName)
			continue
		}
		require.Equal(t, []string{wantTarget}, got, "predefined cluster role of %s", roleName)
	}

	for roleName, want := range c.WantPermissionSets {
		var got []string
		require.NoError(t, runaidb.DB().Actual().
			Raw(`SELECT p.name FROM permission_sets p
			     JOIN role_to_permission_set rps ON rps.permission_set_id = p.id
			     JOIN roles r ON r.id = rps.role_id
			     WHERE r.name = ? ORDER BY p.name`, roleName).
			Scan(&got).Error)
		sort.Strings(want)
		require.Equal(t, want, got, "permission sets of %s", roleName)
	}
}

func seedOneRole(t *testing.T, r seedRole) {
	t.Helper()
	scope := r.Scope
	if scope == "" {
		scope = "system"
	}
	creator := r.Creator
	if creator == "" {
		creator = runAiCreator
	}
	var scopeID, tenantID any
	if scope == "tenant" {
		scopeID, tenantID = r.ScopeID, r.ScopeID
	}
	exec(t, `INSERT INTO roles (created_at, updated_at, created_by, name, description,
	                            internal_only, role_type, scope_type, scope_id, tenant_id,
	                            custom, enabled, no_permission_set, deprecated)
	         VALUES (NOW(), NOW(), ?, ?, '', false, ?, ?, ?, ?, ?, true, false, false)`,
		creator, r.Name, r.Type, scope, scopeID, tenantID, r.Custom)

	for _, ps := range r.PermissionSets {
		exec(t, `INSERT INTO permission_sets (id, name, description, internal, created_at, updated_at)
		         VALUES (gen_random_uuid(), ?, '', false, NOW(), NOW())
		         ON CONFLICT (name) DO NOTHING`, ps)
		exec(t, `INSERT INTO role_to_permission_set (role_id, permission_set_id)
		         SELECT r.id, p.id FROM roles r, permission_sets p
		         WHERE r.name = ? AND p.name = ?`, r.Name, ps)
	}
}
