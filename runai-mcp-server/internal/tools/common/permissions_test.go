package common_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const permittedScopesPath = "/api/v1/authorization/permitted-scopes"

func TestPermittedScopeForAction_PicksActionScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permittedScopesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"create": map[string]any{"clusters": []string{"c-create"}},
			"read":   map[string]any{"clusters": []string{"c-read"}},
			"update": map[string]any{"clusters": []string{"c-update"}},
			"delete": map[string]any{"tenants": []string{"1001"}},
		})
	})

	upd, err := common.PermittedScopeForAction(context.Background(), client, authz.RESOURCETYPE_NODEPOOLS, authz.ACTION_UPDATE)
	require.NoError(t, err)
	assert.Equal(t, []string{"c-update"}, upd.Clusters)

	del, err := common.PermittedScopeForAction(context.Background(), client, authz.RESOURCETYPE_NODEPOOLS, authz.ACTION_DELETE)
	require.NoError(t, err)
	assert.Equal(t, []string{"1001"}, del.Tenants)
}

func TestPermittedScopeForAction_Forbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permittedScopesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "nope"})
	})

	_, err := common.PermittedScopeForAction(context.Background(), client, authz.RESOURCETYPE_NODEPOOLS, authz.ACTION_CREATE)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func boolPtr(b bool) *bool { return &b }

func TestScopeAllowsCluster(t *testing.T) {
	cl := "c-1"
	cases := []struct {
		name  string
		scope authz.PermittedScopes
		want  bool
	}{
		{"system", authz.PermittedScopes{System: boolPtr(true)}, true},
		{"tenant", authz.PermittedScopes{Tenants: []string{"1001"}}, true},
		{"cluster match", authz.PermittedScopes{Clusters: []string{"c-0", "c-1"}}, true},
		{"cluster miss", authz.PermittedScopes{Clusters: []string{"c-0"}}, false},
		{"empty", authz.PermittedScopes{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, common.ScopeAllowsCluster(tc.scope, cl))
		})
	}
}

func TestScopeAllowsOrgUnit(t *testing.T) {
	cl, id, parent := "c-1", "proj-1", "dept-1"

	// cluster-level grants flow through ScopeAllowsCluster.
	assert.True(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{System: boolPtr(true)}, common.OrgTypeProject, cl, id, parent))
	assert.True(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Clusters: []string{cl}}, common.OrgTypeDepartment, cl, id, ""))

	// a grant on the project's PARENT department covers it...
	assert.True(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Departments: []string{"dept-1"}}, common.OrgTypeProject, cl, id, parent))
	// ...but a grant on an UNRELATED department does not (the A01 fix).
	assert.False(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Departments: []string{"dept-9"}}, common.OrgTypeProject, cl, id, parent))
	// ...and with no known parent, a department grant alone doesn't authorize the project.
	assert.False(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Departments: []string{"dept-1"}}, common.OrgTypeProject, cl, id, ""))

	// a department target matches only its own id.
	assert.False(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Departments: []string{"dept-9"}}, common.OrgTypeDepartment, cl, "dept-1", ""))
	assert.True(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Departments: []string{"dept-1"}}, common.OrgTypeDepartment, cl, "dept-1", ""))

	// a matching project id covers that specific project.
	assert.True(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Projects: []string{"proj-1"}}, common.OrgTypeProject, cl, id, parent))
	assert.False(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{Projects: []string{"proj-2"}}, common.OrgTypeProject, cl, id, parent))

	// nothing matches.
	assert.False(t, common.ScopeAllowsOrgUnit(authz.PermittedScopes{}, common.OrgTypeProject, cl, id, parent))
}
