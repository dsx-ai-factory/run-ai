/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const validProjectID = "4500002"

func TestGetWorkloadEffectivePolicyInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadEffectivePolicyInput
		wantErr bool
	}{
		{"empty projectId", getWorkloadEffectivePolicyInput{Kind: "training"}, true},
		{"whitespace projectId", getWorkloadEffectivePolicyInput{ProjectID: "   ", Kind: "training"}, true},
		{"non-numeric projectId", getWorkloadEffectivePolicyInput{ProjectID: "abc", Kind: "training"}, true},
		{"example named project rejected", getWorkloadEffectivePolicyInput{ProjectID: "p1", Kind: "training"}, true},
		{"alphanumeric projectId", getWorkloadEffectivePolicyInput{ProjectID: "12a", Kind: "training"}, true},
		{"empty kind", getWorkloadEffectivePolicyInput{ProjectID: validProjectID}, true},
		{"unknown kind", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "ci"}, true},
		{"example job kind rejected", getWorkloadEffectivePolicyInput{ProjectID: "4500001", Kind: "job"}, true},
		{"nim-service rejected", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "nim-service"}, true},
		{"distributed inference accepted", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "DistributedInference"}, false},
		{"nimService accepted", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "nimService"}, false},
		{"fallback accepted", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "fallback"}, false},
		{"valid training", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "training"}, false},
		{"valid workspace", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "workspace"}, false},
		{"valid distributed", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "distributed"}, false},
		{"valid inference", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "inference"}, false},
		{"valid uppercase Training accepted (case-insensitive)", getWorkloadEffectivePolicyInput{ProjectID: validProjectID, Kind: "Training"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

var kindURLMappings = []struct {
	kind  string
	route string
}{
	{policyKindTraining, "/api/v2/policy/trainings"},
	{policyKindWorkspace, "/api/v2/policy/workspaces"},
	{policyKindDistributed, "/api/v2/policy/distributed"},
	{policyKindInference, "/api/v2/policy/inferences"},
	{policyKindDistributedInference, "/api/v2/policy/distributed-inferences"},
	{policyKindNimService, "/api/v2/policy/nim-services"},
	{policyKindFallback, "/api/v2/policy/fallback"},
}

func TestGetWorkloadEffectivePolicyRun_KindToURLMapping(t *testing.T) {
	for _, m := range kindURLMappings {
		t.Run(m.kind, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)

			var gotScope, gotProjectID string
			mux.HandleFunc(m.route, func(w http.ResponseWriter, r *http.Request) {
				gotScope = r.URL.Query().Get("scope")
				gotProjectID = r.URL.Query().Get("projectId")
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"policy": map[string]any{"defaults": map[string]any{"image": "declared-image"}, "imposedAssets": []any{}},
					"effective": map[string]any{
						"defaults": map[string]any{"image": "nvcr.io/" + m.kind},
					},
				})
			})

			out, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
				ProjectID: validProjectID,
				Kind:      m.kind,
			})
			require.NoError(t, err)
			assert.Equal(t, "project", gotScope, "scope=project always")
			assert.Equal(t, validProjectID, gotProjectID, "projectId forwarded as the query param")
			assert.Equal(t, validProjectID, out.ProjectID, "projectId echoed in the output")
			assert.Equal(t, m.kind, out.Kind)
			assert.Equal(t, "declared-image", out.Policy["defaults"].(map[string]any)["image"])
			assert.Contains(t, out.Policy, "imposedAssets")
			require.NotNil(t, out.Effective)
			defaults, ok := out.Effective["defaults"].(map[string]any)
			require.True(t, ok, "modern shape preserves defaults map")
			assert.Equal(t, "nvcr.io/"+m.kind, defaults["image"])
		})
	}
}

func TestGetWorkloadEffectivePolicyRun_ModernShapeForwardedVerbatim(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"effective": map[string]any{
				"defaults": map[string]any{
					"image": "default-image",
					"compute": map[string]any{
						"gpuDevicesRequest": float64(1),
					},
				},
				"rules": map[string]any{
					"image": map[string]any{
						"required":     true,
						"sourceOfRule": "tenant",
						"message":      "image is required at tenant scope",
					},
					"compute.gpuDevicesRequest": map[string]any{
						"min":          float64(1),
						"max":          float64(8),
						"sourceOfRule": "department",
					},
				},
			},
		})
	})

	out, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindTraining,
	})
	require.NoError(t, err)

	defaults, ok := out.Effective["defaults"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "default-image", defaults["image"])
	rules, ok := out.Effective["rules"].(map[string]any)
	require.True(t, ok)
	imageRule, ok := rules["image"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, imageRule["required"], "leaf metadata forwarded as-is")
	assert.Equal(t, "tenant", imageRule["sourceOfRule"])
	assert.Equal(t, "image is required at tenant scope", imageRule["message"])
}

func TestGetWorkloadEffectivePolicyRun_LegacyShapeForwardedVerbatim(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/workspaces", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"effective": map[string]any{
				"image": map[string]any{
					"default":  "legacy-image",
					"required": true,
				},
				"compute": map[string]any{
					"gpuDevicesRequest": map[string]any{
						"default": float64(1),
						"min":     float64(1),
						"max":     float64(4),
					},
				},
			},
		})
	})

	out, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindWorkspace,
	})
	require.NoError(t, err)

	_, hasModernDefaults := out.Effective["defaults"]
	assert.False(t, hasModernDefaults, "legacy responses must not be normalised into a modern {defaults, rules} shape")
	image, ok := out.Effective["image"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "legacy-image", image["default"])
	assert.Equal(t, true, image["required"])
}

func TestGetWorkloadEffectivePolicyRun_EmptyEffectiveBecomesEmptyMap(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{}) // upstream returned no "effective" field at all.
	})

	out, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindTraining,
	})
	require.NoError(t, err)
	require.NotNil(t, out.Effective, "absent effective key surfaces as a non-nil empty map (never null)")
	assert.Empty(t, out.Effective)
}

func TestGetWorkloadEffectivePolicyRun_EmptyBodyBecomesEmptyMap(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 2xx with zero-length body — must not surface as a decode error.
	})

	out, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindTraining,
	})
	require.NoError(t, err, "an empty body on a 2xx must not error out")
	require.NotNil(t, out.Effective, "empty body still surfaces a non-nil empty map")
	assert.Empty(t, out.Effective)
}

// Guards the dispatcher's defensive default branch: a kind that slipped past Validate() must
// keep its ErrValidation sentinel and not be silently downgraded to ErrUpstream by MapHTTPError.
func TestGetWorkloadEffectivePolicyRun_UnknownKindPreservesValidationSentinel(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t)

	_, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      "bogus-kind",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "validation sentinel must propagate from the dispatcher default branch")
	assert.NotErrorIs(t, err, runaimcp.ErrUpstream, "must not be wrapped as an upstream error")
}

func TestGetWorkloadEffectivePolicyRun_NotFoundSurfacedDirectly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "no training policy at project scope"})
	})

	_, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindTraining,
	})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadEffectivePolicyRun_AuthSurfacedDirectly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/inferences", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusUnauthorized, map[string]any{"message": "token expired"})
	})

	_, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindInference,
	})
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func TestGetWorkloadEffectivePolicyRun_ForbiddenSurfacedDirectly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v2/policy/distributed", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no read on policy.distributed"})
	})

	_, err := getWorkloadEffectivePolicyRun(context.Background(), client, getWorkloadEffectivePolicyInput{
		ProjectID: validProjectID,
		Kind:      policyKindDistributed,
	})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestEffectiveLegacyPolicyScopes(t *testing.T) {
	for _, tc := range []struct{ scope, cluster, department, project string }{
		{scope: "tenant"},
		{scope: "cluster", cluster: "6fced230-c208-4b73-98a8-964fc1a32787"},
		{scope: "department", department: "4500001"},
		{scope: "project", project: validProjectID},
	} {
		t.Run(tc.scope, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.scope, r.URL.Query().Get("scope"))
				assert.Equal(t, tc.cluster, r.URL.Query().Get("clusterId"))
				assert.Equal(t, tc.department, r.URL.Query().Get("departmentId"))
				assert.Equal(t, tc.project, r.URL.Query().Get("projectId"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"policy": map[string]any{"defaults": map[string]any{"image": "declared"}}, "effective": map[string]any{"defaults": map[string]any{"image": "effective"}}})
			})
			in := getWorkloadEffectivePolicyInput{Scope: tc.scope, ClusterID: tc.cluster, DepartmentID: tc.department, ProjectID: tc.project, Kind: "training"}
			require.NoError(t, in.Validate())
			out, err := getWorkloadEffectivePolicyRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.Equal(t, tc.scope, out.Scope)
			assert.Equal(t, "declared", out.Policy["defaults"].(map[string]any)["image"])
			assert.Equal(t, "effective", out.Effective["defaults"].(map[string]any)["image"])
		})
	}
}
