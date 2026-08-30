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

func departmentJSON(id, name string) map[string]any {
	return map[string]any{
		"id": id, "name": name, "clusterId": mouClusterUUID,
		"totalResources": map[string]any{}, "resources": []map[string]any{},
	}
}

func emptyProjectsHandler(w http.ResponseWriter, _ *http.Request) {
	runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"projects": []map[string]any{}})
}

func TestListDepartmentsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      listDepartmentsInput
		wantErr string
	}{
		{name: "happy_no_filter", in: listDepartmentsInput{}},
		{name: "happy_cluster", in: listDepartmentsInput{ClusterID: "c-1"}},
		{name: "reject_bad_limit", in: listDepartmentsInput{Limit: 201}, wantErr: "limit must be between"},
		{name: "reject_negative_offset", in: listDepartmentsInput{Offset: -1}, wantErr: "offset must be non-negative"},
		{name: "reject_comma", in: listDepartmentsInput{ClusterID: "a,b"}, wantErr: "must not contain commas"},
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

// A department with no quota configured must still be listed — the quota-cell
// tool returns zero rows for it, which is exactly the blind spot this covers.
// Child counts come from the projects listing, not the department payload.
func TestListDepartments(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{
				departmentJSON("d-2", "rnd"),
				departmentJSON("d-1", "default"),
			},
		})
	})
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				projectJSON("p-1", "alpha", "d-2", "Ready"),
				projectJSON("p-2", "beta", "d-2", "Ready"),
			},
		})
	})

	out, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{})
	require.NoError(t, err)
	assert.Equal(t, int32(2), out.TotalCount)
	assert.False(t, out.HasMore)
	require.Len(t, out.Departments, 2)
	assert.Equal(t, "default", out.Departments[0].Name)
	require.NotNil(t, out.Departments[0].ProjectCount)
	assert.Equal(t, int32(0), *out.Departments[0].ProjectCount)
	assert.Equal(t, "rnd", out.Departments[1].Name)
	require.NotNil(t, out.Departments[1].ProjectCount)
	assert.Equal(t, int32(2), *out.Departments[1].ProjectCount)
}

// allocated (point-in-time, from quotaStatus) and gpuQuota (from totalResources) must be
// surfaced so over-quota departments can be found — the value the UI shows as "Allocated
// GPUs", distinct from the get_org_unit_metrics over-time average. CPU/memory are normalized
// to cores/GiB exactly as list_project_resources does.
func TestListDepartments_AllocatedAndQuota(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		// Verbose is required for the response to carry status.quotaStatus.
		assert.Equal(t, "verbose", r.URL.Query().Get("verbosity"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{
				{
					"id": "d-2", "name": "rnd", "clusterId": mouClusterUUID,
					"totalResources": map[string]any{"gpuQuota": 10},
					"resources":      []map[string]any{},
					"status": map[string]any{
						"quotaStatus": map[string]any{"allocated": map[string]any{"gpu": 9, "cpu": 2000, "memory": 2147.483648}},
					},
				},
				// No status block → allocated stays null (not a misleading zero).
				departmentJSON("d-1", "default"),
			},
		})
	})
	mux.HandleFunc(projectsPath, emptyProjectsHandler)

	out, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{})
	require.NoError(t, err)
	require.Len(t, out.Departments, 2)

	def, rnd := out.Departments[0], out.Departments[1]
	require.Equal(t, "default", def.Name)
	assert.Nil(t, def.Allocated, "no status → null, not zero")

	require.Equal(t, "rnd", rnd.Name)
	require.NotNil(t, rnd.Allocated)
	require.NotNil(t, rnd.Allocated.GPU)
	assert.Equal(t, float64(9), *rnd.Allocated.GPU)
	require.NotNil(t, rnd.Allocated.CPUCores)
	assert.Equal(t, float64(2), *rnd.Allocated.CPUCores) // 2000 millicores → 2 cores
	require.NotNil(t, rnd.Allocated.MemoryGiB)
	assert.InDelta(t, 2.0, *rnd.Allocated.MemoryGiB, 0.001) // 2147.483648 MB → 2 GiB
	require.NotNil(t, rnd.GPUQuota)
	assert.Equal(t, float64(10), *rnd.GPUQuota)
}

// gpuQuota -1 means unlimited: it must survive as -1 (not dropped) so the model can treat
// it as never-over-quota, even when the department has live allocation.
func TestListDepartments_UnlimitedQuotaSurfacedAsNegativeOne(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{{
				"id": "d-1", "name": "default", "clusterId": mouClusterUUID,
				"totalResources": map[string]any{"gpuQuota": -1},
				"resources":      []map[string]any{},
				"status": map[string]any{
					"quotaStatus": map[string]any{"allocated": map[string]any{"gpu": 19}},
				},
			}},
		})
	})
	mux.HandleFunc(projectsPath, emptyProjectsHandler)

	out, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{})
	require.NoError(t, err)
	require.Len(t, out.Departments, 1)
	d := out.Departments[0]
	require.NotNil(t, d.GPUQuota)
	assert.Equal(t, float64(-1), *d.GPUQuota)
	require.NotNil(t, d.Allocated)
	require.NotNil(t, d.Allocated.GPU)
	assert.Equal(t, float64(19), *d.Allocated.GPU)
}

// The projects fetch is best-effort for child counts: if it fails, the
// department listing must still succeed, with null counts rather than zeros.
func TestListDepartments_ProjectsFetchFailureDegrades(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{departmentJSON("d-1", "default")},
		})
	})
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	out, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), out.TotalCount)
	require.Len(t, out.Departments, 1)
	assert.Nil(t, out.Departments[0].ProjectCount)
}

func TestListDepartments_Pagination(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{
				departmentJSON("d-1", "a"),
				departmentJSON("d-2", "b"),
				departmentJSON("d-3", "c"),
			},
		})
	})
	mux.HandleFunc(projectsPath, emptyProjectsHandler)

	out, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, int32(3), out.TotalCount)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(2), out.NextOffset)
	require.Len(t, out.Departments, 2)
}

func TestListDepartments_Forbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	_, err := listDepartmentsRun(context.Background(), client, listDepartmentsInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}
