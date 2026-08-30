package orgunits

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func projectJSON(id, name, parentID, phase string) map[string]any {
	p := map[string]any{
		"id": id, "name": name, "clusterId": mouClusterUUID, "parentId": parentID,
		"description": "", "totalResources": map[string]any{}, "resources": []map[string]any{},
	}
	if phase != "" {
		p["status"] = map[string]any{"phase": phase}
	}
	return p
}

func TestListProjectsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      listProjectsInput
		wantErr string
	}{
		{name: "happy_no_filter", in: listProjectsInput{}},
		{name: "happy_department", in: listProjectsInput{DepartmentID: "dept-1"}},
		{name: "reject_dept_id_and_name", in: listProjectsInput{DepartmentID: "dept-1", DepartmentName: "eng"}, wantErr: "mutually exclusive"},
		{name: "reject_bad_limit", in: listProjectsInput{Limit: 201}, wantErr: "limit must be between"},
		{name: "reject_comma", in: listProjectsInput{DepartmentName: "a,b"}, wantErr: "must not contain commas"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// A project with no quota cell on any node pool must still be listed — this is
// the entity view, not the quota-cell view.
func TestListProjects(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				projectJSON("p-2", "beta", "dept-1", "Ready"),
				projectJSON("p-1", "alpha", "", ""),
				projectJSON("p-3", "gamma", "dept-2", "Creating"),
			},
		})
	})

	out, err := listProjectsRun(context.Background(), client, listProjectsInput{})
	require.NoError(t, err)
	assert.Equal(t, int32(3), out.TotalCount)
	assert.False(t, out.HasMore)
	require.Len(t, out.Projects, 3)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, []string{out.Projects[0].Name, out.Projects[1].Name, out.Projects[2].Name})
	assert.Nil(t, out.Projects[0].DepartmentID) // empty parent → null
	assert.Nil(t, out.Projects[0].Phase)
	require.NotNil(t, out.Projects[1].DepartmentID)
	assert.Equal(t, "dept-1", *out.Projects[1].DepartmentID)
	require.NotNil(t, out.Projects[1].Phase)
	assert.Equal(t, "Ready", *out.Projects[1].Phase)
}

func TestListProjects_DepartmentFilter(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				projectJSON("p-1", "alpha", "dept-1", "Ready"),
				projectJSON("p-2", "beta", "dept-2", "Ready"),
			},
		})
	})

	out, err := listProjectsRun(context.Background(), client, listProjectsInput{DepartmentID: "dept-2"})
	require.NoError(t, err)
	assert.Equal(t, int32(1), out.TotalCount)
	require.Len(t, out.Projects, 1)
	assert.Equal(t, "beta", out.Projects[0].Name)
}

func TestListProjects_DepartmentNameResolves(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{{
				"id": "dept-9", "name": "research", "clusterId": mouClusterUUID,
				"totalResources": map[string]any{}, "resources": []map[string]any{},
			}},
		})
	})
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				projectJSON("p-1", "alpha", "dept-9", "Ready"),
				projectJSON("p-2", "beta", "dept-other", "Ready"),
			},
		})
	})

	out, err := listProjectsRun(context.Background(), client, listProjectsInput{DepartmentName: "research"})
	require.NoError(t, err)
	assert.Equal(t, int32(1), out.TotalCount)
	require.Len(t, out.Projects, 1)
	assert.Equal(t, "p-1", out.Projects[0].ID)
}

func TestListProjects_Pagination(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				projectJSON("p-1", "alpha", "", "Ready"),
				projectJSON("p-2", "beta", "", "Ready"),
				projectJSON("p-3", "gamma", "", "Ready"),
			},
		})
	})

	out, err := listProjectsRun(context.Background(), client, listProjectsInput{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, int32(3), out.TotalCount) // count spans all pages
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(2), out.NextOffset)
	require.Len(t, out.Projects, 2)

	out, err = listProjectsRun(context.Background(), client, listProjectsInput{Limit: 2, Offset: 2})
	require.NoError(t, err)
	assert.False(t, out.HasMore)
	require.Len(t, out.Projects, 1)
	assert.Equal(t, "gamma", out.Projects[0].Name)
}

func TestListProjects_Forbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	_, err := listProjectsRun(context.Background(), client, listProjectsInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}
