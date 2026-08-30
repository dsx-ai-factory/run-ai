package metrics

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func measurement(metricType string, vals ...[2]string) map[string]any {
	values := make([]map[string]any, 0, len(vals))
	for _, v := range vals {
		values = append(values, map[string]any{"value": v[0], "timestamp": v[1]})
	}
	return map[string]any{"type": metricType, "values": values}
}

func TestGetOrgUnitMetricsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getOrgUnitMetricsInput
		wantErr string
	}{
		{name: "happy_project_id", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p-1"}},
		{name: "happy_department_name", in: getOrgUnitMetricsInput{OrgType: "department", OrgName: "eng"}},
		{name: "bad_org_type", in: getOrgUnitMetricsInput{OrgType: "team", OrgID: "x"}, wantErr: "orgType must be"},
		{name: "both_id_and_name", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p", OrgName: "n"}, wantErr: "exactly one"},
		{name: "neither_id_nor_name", in: getOrgUnitMetricsInput{OrgType: "project"}, wantErr: "exactly one"},
		{name: "unknown_metric", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p", MetricTypes: []string{"diskIO"}}, wantErr: "unknown metric"},
		{name: "bad_window", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p", Window: "last_year"}, wantErr: "window must be"},
		{name: "start_without_end", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p", Start: "2026-06-01T00:00:00Z"}, wantErr: "together"},
		{name: "samples_over_max", in: getOrgUnitMetricsInput{OrgType: "project", OrgID: "p", NumberOfSamples: 1001}, wantErr: "numberOfSamples"},
		{name: "comma_in_name", in: getOrgUnitMetricsInput{OrgType: "project", OrgName: "a,b"}, wantErr: "commas"},
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

func TestGetOrgUnitMetricsRun_ProjectHappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotSamples string
	mux.HandleFunc("/api/v1/org-unit/projects/p-1/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotSamples = r.URL.Query().Get("numberOfSamples")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				// two per-GPU variants at the same timestamp → summed for allocation
				measurement("GPU_ALLOCATION", [2]string{"1", "2026-06-01T00:00:00Z"}),
				measurement("GPU_ALLOCATION", [2]string{"1", "2026-06-01T00:00:00Z"}),
				measurement("CPU_ALLOCATION_MILLICORES", [2]string{"2000", "2026-06-01T00:00:00Z"}),
				// two utilization variants at the same timestamp → averaged
				measurement("GPU_UTILIZATION", [2]string{"40", "2026-06-01T00:00:00Z"}),
				measurement("GPU_UTILIZATION", [2]string{"60", "2026-06-01T00:00:00Z"}),
			},
		})
	})

	out, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType:         "project",
		OrgID:           "p-1",
		MetricTypes:     []string{"gpuAllocation", "cpuAllocation", "gpuUtilization"},
		NumberOfSamples: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, "5", gotSamples)
	assert.Nil(t, out.NodePool)
	assert.False(t, out.SeriesTruncated)
	assert.False(t, out.PartialData)
	require.NotNil(t, out.Series)

	require.Len(t, out.Series.GPUAllocation, 1)
	assert.Equal(t, float64(2), out.Series.GPUAllocation[0].V, "two variants summed")
	require.Len(t, out.Series.CPUAllocation, 1)
	assert.Equal(t, float64(2), out.Series.CPUAllocation[0].V, "2000 millicores → 2 cores")
	require.Len(t, out.Series.GPUUtilization, 1)
	assert.Equal(t, float64(50), out.Series.GPUUtilization[0].V, "40 and 60 averaged")
	// unrequested series stay empty
	assert.Empty(t, out.Series.CPUMemoryAllocation)
}

func TestGetOrgUnitMetricsRun_ResolvesDepartmentByName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/departments", func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("filterBy"), "name==eng")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"departments": []map[string]any{{"id": "d-1", "name": "eng", "clusterId": "c-1"}}})
	})
	mux.HandleFunc("/api/v1/org-unit/departments/d-1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{measurement("GPU_ALLOCATION", [2]string{"3", "2026-06-01T00:00:00Z"})},
		})
	})

	out, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "department", OrgName: "eng", MetricTypes: []string{"gpuAllocation"},
	})
	require.NoError(t, err)
	assert.Equal(t, "d-1", out.OrgID)
	assert.Equal(t, "eng", out.OrgName)
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.GPUAllocation, 1)
	assert.Equal(t, float64(3), out.Series.GPUAllocation[0].V)
}

func TestGetOrgUnitMetricsRun_AmbiguousName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/departments", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"departments": []map[string]any{
			{"id": "d-1", "name": "eng", "clusterId": "c-1"},
			{"id": "d-2", "name": "eng", "clusterId": "c-2"},
		}})
	})

	_, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "department", OrgName: "eng",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_org_unit")
}

func TestGetOrgUnitMetricsRun_TruncationGuard(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t) // no metrics handler: must not be called
	out, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType:         "project",
		OrgID:           "p-1",
		NumberOfSamples: 1000,
		MetricTypes:     []string{"gpuAllocation", "cpuAllocation", "cpuMemoryAllocation", "gpuUtilization", "gpuMemoryUtilization", "cpuUtilization"}, // 6 × 1000 = 6000 > 5000
	})
	require.NoError(t, err)
	assert.True(t, out.SeriesTruncated)
	assert.Nil(t, out.Series, "series omitted when truncated")
	assert.Equal(t, int32(1000), out.TimeRange.NumberOfSamples)
}

func TestGetOrgUnitMetricsRun_PartialData207(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects/p-1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{measurement("GPU_ALLOCATION", [2]string{"1", "2026-06-01T00:00:00Z"})},
		})
	})

	out, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "project", OrgID: "p-1", MetricTypes: []string{"gpuAllocation", "cpuUtilization"},
	})
	require.NoError(t, err)
	assert.True(t, out.PartialData, "207 maps to partialData")
	require.NotNil(t, out.Series)
	assert.Len(t, out.Series.GPUAllocation, 1)
	assert.Empty(t, out.Series.CPUUtilization, "failed metric absent from series")
}

func TestGetOrgUnitMetricsRun_MemoryNormalization(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects/p-1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				measurement("CPU_MEMORY_ALLOCATION_MB", [2]string{"2147.483648", "2026-06-01T00:00:00Z"}),
			},
		})
	})

	out, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "project", OrgID: "p-1", MetricTypes: []string{"cpuMemoryAllocation"},
	})
	require.NoError(t, err)
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.CPUMemoryAllocation, 1)
	assert.Equal(t, float64(2), out.Series.CPUMemoryAllocation[0].V, "2147.483648 MB → 2 GiB")
}

func TestGetOrgUnitMetricsRun_TimeRange(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotStart, gotEnd string
	mux.HandleFunc("/api/v1/org-unit/projects/p-1/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []map[string]any{}})
	})

	// Explicit start/end override the window.
	_, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "project", OrgID: "p-1", MetricTypes: []string{"gpuAllocation"},
		Start: "2026-06-01T00:00:00Z", End: "2026-06-02T00:00:00Z", Window: "last_7d",
	})
	require.NoError(t, err)
	st, perr := time.Parse(time.RFC3339, gotStart)
	require.NoError(t, perr, "start forwarded as RFC3339; got %q", gotStart)
	en, perr := time.Parse(time.RFC3339, gotEnd)
	require.NoError(t, perr)
	assert.True(t, st.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)), "explicit start forwarded; got %q", gotStart)
	assert.Equal(t, 24*time.Hour, en.Sub(st), "explicit end forwarded, window ignored")

	// Default window (last_24h) when no start/end given.
	_, err = getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "project", OrgID: "p-1", MetricTypes: []string{"gpuAllocation"},
	})
	require.NoError(t, err)
	st2, _ := time.Parse(time.RFC3339, gotStart)
	en2, _ := time.Parse(time.RFC3339, gotEnd)
	assert.InDelta(t, float64(24*time.Hour), float64(en2.Sub(st2)), float64(time.Minute), "default window ≈ last_24h")
}

func TestGetOrgUnitMetricsRun_NodePoolScopeAndForbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotNodePool string
	mux.HandleFunc("/api/v1/org-unit/projects/p-1/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotNodePool = r.URL.Query().Get("nodepoolName")
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getOrgUnitMetricsRun(context.Background(), client, getOrgUnitMetricsInput{
		OrgType: "project", OrgID: "p-1", NodePoolName: "h100",
	})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	assert.Equal(t, "h100", gotNodePool, "nodePoolName forwarded to upstream")
}
