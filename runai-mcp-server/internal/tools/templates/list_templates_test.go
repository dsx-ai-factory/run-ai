package templates

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestListTemplatesMapsQueryAndResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	projectID := int32(42)
	mux.HandleFunc("/api/v1/workload-templates", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "base", query.Get("search"))
		assert.Equal(t, "name", query.Get("sortBy"))
		assert.Equal(t, "desc", query.Get("sortOrder"))
		assert.Equal(t, "true", query.Get("includeAncestors"))
		assert.Equal(t, "10", query.Get("limit"))
		assert.Equal(t, "5", query.Get("offset"))
		assert.ElementsMatch(t, []string{
			"workloadType==Training", "scopeType==project", "projectId==42",
		}, strings.Split(query.Get("filterBy"), ","))

		entry := templateResponse(templateTypeTraining, "base")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"entries":    []any{entry},
			"pagination": map[string]any{"limit": 10, "offset": 5, "nextOffset": 15},
		})
	})

	out, err := listTemplatesRun(context.Background(), client, listTemplatesInput{
		WorkloadType: templateTypeTraining, ScopeType: "project", ProjectID: &projectID,
		Search: "base", IncludeAncestors: true, SortBy: "name", SortOrder: "desc", Limit: 10, Offset: 5,
	})
	require.NoError(t, err)
	require.Len(t, out.Templates, 1)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(15), out.NextOffset)
	assert.Equal(t, "base", out.Templates[0].Name)
}

func TestListTemplatesRejectsInvalidAPIPage(t *testing.T) {
	tests := []struct {
		name    string
		input   listTemplatesInput
		result  map[string]any
		wantErr string
	}{
		{
			name: "non-advancing cursor", input: listTemplatesInput{Limit: 10, Offset: 5},
			result: map[string]any{"entries": []any{}, "pagination": map[string]any{
				"limit": 10, "offset": 5, "nextOffset": 5,
			}},
			wantErr: "does not advance",
		},
		{
			name: "mismatched offset", input: listTemplatesInput{Limit: 10, Offset: 5},
			result: map[string]any{"entries": []any{}, "pagination": map[string]any{
				"limit": 10, "offset": 4,
			}},
			wantErr: "response offset 4 does not match requested offset 5",
		},
		{
			name: "oversized page", input: listTemplatesInput{Limit: 1},
			result: map[string]any{"entries": []any{
				templateResponse(templateTypeTraining, "one"),
				templateResponse(templateTypeTraining, "two"),
			}},
			wantErr: "exceeding requested limit 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workload-templates", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, tt.result)
			})

			_, err := listTemplatesRun(context.Background(), client, tt.input)
			assert.ErrorIs(t, err, runaimcp.ErrUpstream)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestListTemplatesInputValidation(t *testing.T) {
	projectID := int32(42)
	zero := int32(0)
	tests := []struct {
		name    string
		input   listTemplatesInput
		wantErr string
	}{
		{name: "defaults"},
		{name: "supported filters", input: listTemplatesInput{
			Name: "base", WorkloadType: templateTypeWorkspace, ScopeType: "project", ClusterID: testClusterID,
			DepartmentID: "department-1", ProjectID: &projectID, SortBy: "updatedAt", SortOrder: "desc", Limit: 200,
		}},
		{name: "unsupported workload", input: listTemplatesInput{WorkloadType: "Job"}, wantErr: "workloadType must be one of"},
		{name: "unsupported scope", input: listTemplatesInput{ScopeType: "namespace"}, wantErr: "scopeType must be one of"},
		{name: "invalid cluster", input: listTemplatesInput{ClusterID: "cluster"}, wantErr: "clusterId must be a UUID"},
		{name: "nonpositive project", input: listTemplatesInput{ProjectID: &zero}, wantErr: "projectId must be positive"},
		{name: "ancestors require scope", input: listTemplatesInput{IncludeAncestors: true}, wantErr: "exactly one"},
		{name: "ancestors reject ambiguous scope", input: listTemplatesInput{ClusterID: testClusterID, ProjectID: &projectID, IncludeAncestors: true}, wantErr: "exactly one"},
		{name: "blank search", input: listTemplatesInput{Search: "  "}, wantErr: "search must not be blank"},
		{name: "unsupported sort", input: listTemplatesInput{SortBy: "id"}, wantErr: "sortBy must be one of"},
		{name: "unsupported order", input: listTemplatesInput{SortOrder: "newest"}, wantErr: "sortOrder must be one of"},
		{name: "excessive limit", input: listTemplatesInput{Limit: 201}, wantErr: "limit must be between"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
