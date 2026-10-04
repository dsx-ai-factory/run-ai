/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestGetTemplateInputValidation(t *testing.T) {
	assert.Error(t, (getTemplateInput{}).Validate())
	assert.Error(t, (getTemplateInput{TemplateID: "not-a-uuid"}).Validate())
	assert.NoError(t, (getTemplateInput{TemplateID: testTemplateID}).Validate())
}

func TestGetTemplateRunReturnsCompleteTemplate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		response := templateResponse(templateTypeDistributed, "distributed-base")
		response["spec"] = map[string]any{
			"image":   "nvcr.io/nvidia/worker:latest",
			"command": "python train.py",
		}
		response["assets"] = map[string]any{
			"compute":     testClusterID,
			"datasources": []string{testTemplateID},
		}
		response["masterSpecSameAsWorker"] = false
		response["master"] = map[string]any{
			"spec": map[string]any{"image": "nvcr.io/nvidia/master:latest"},
		}
		response["runnable"] = map[string]any{"isRunnable": true}
		runaiclienttest.WriteJSON(w, http.StatusOK, response)
	})

	out, err := getTemplateRun(context.Background(), client, getTemplateInput{TemplateID: testTemplateID})
	require.NoError(t, err)
	assert.Equal(t, testTemplateID, out.TemplateID)
	assert.Equal(t, "distributed-base", out.Name)
	assert.Equal(t, templateTypeDistributed, out.WorkloadType)
	assert.Nil(t, out.UpdateRequest)

	meta, ok := out.Template["meta"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, testTemplateID, meta["id"])
	assert.Equal(t, "distributed-base", meta["name"])

	spec, ok := out.Template["spec"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "nvcr.io/nvidia/worker:latest", spec["image"])
	assert.Contains(t, out.Template, "assets")
	assert.Contains(t, out.Template, "masterSpecSameAsWorker")
	assert.Contains(t, out.Template, "master")
	assert.Contains(t, out.Template, "runnable")
}

func TestGetTemplateRunReturnsUpdateReadyRequest(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		response := templateResponse(templateTypeDistributed, "distributed-base")
		response["meta"].(map[string]any)["description"] = "editable description"
		response["spec"] = map[string]any{"image": "worker:1", "command": "python train.py"}
		response["assets"] = map[string]any{"compute": testClusterID}
		response["masterSpecSameAsWorker"] = false
		response["master"] = map[string]any{"spec": map[string]any{"image": "master:1"}}
		response["runnable"] = map[string]any{"isRunnable": true}
		response["complianceIssues"] = []map[string]any{}
		runaiclienttest.WriteJSON(w, http.StatusOK, response)
	})

	out, err := getTemplateRun(context.Background(), client, getTemplateInput{
		TemplateID: testTemplateID,
		ForUpdate:  true,
	})
	require.NoError(t, err)
	assert.Nil(t, out.Template)
	assert.Equal(t, testTemplateID, out.TemplateID)
	assert.Equal(t, templateTypeDistributed, out.WorkloadType)
	assert.NotEmpty(t, out.ExpectedFingerprint)
	assert.Equal(t, map[string]any{
		"meta": map[string]any{
			"name":        "distributed-base",
			"description": "editable description",
		},
		"spec":                   map[string]any{"image": "worker:1", "command": "python train.py"},
		"assets":                 map[string]any{"compute": testClusterID},
		"masterSpecSameAsWorker": false,
		"master":                 map[string]any{"spec": map[string]any{"image": "master:1"}},
	}, out.UpdateRequest)

	assert.NoError(t, (manageTemplateInput{
		Action:              "update",
		WorkloadType:        out.WorkloadType,
		TemplateID:          out.TemplateID,
		Request:             out.UpdateRequest,
		ExpectedFingerprint: out.ExpectedFingerprint,
	}).Validate(), "get_template updateRequest must be accepted by manage_template")

	nullable := templateResponse(templateTypeDistributedInference, "nullable-template")
	nullable["spec"] = map[string]any{"replicas": 1, "workers": 0}
	nullable["assets"] = nil
	nullable["master"] = nil
	nullable["masterSpecSameAsWorker"] = nil
	request, fingerprint, err := templateUpdateSnapshot(templateTypeDistributedInference, nullable)
	require.NoError(t, err)
	assert.NotEmpty(t, fingerprint)
	assert.NotContains(t, request, "assets")
	assert.NotContains(t, request, "master")
	assert.NotContains(t, request, "masterSpecSameAsWorker")
}

func TestGetTemplatePreservesNewReadFieldsButFailsClosedForUpdate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		response := templateResponse(templateTypeTraining, "future-template")
		response["spec"] = map[string]any{
			"image":           "worker:1",
			"futureSpecField": map[string]any{"enabled": true},
		}
		response["futureTopLevelField"] = map[string]any{"value": "preserved"}
		runaiclienttest.WriteJSON(w, http.StatusOK, response)
	})

	out, err := getTemplateRun(context.Background(), client, getTemplateInput{TemplateID: testTemplateID})
	require.NoError(t, err)
	assert.Contains(t, out.Template, "futureTopLevelField")
	spec, ok := out.Template["spec"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, spec, "futureSpecField")

	_, err = getTemplateRun(context.Background(), client, getTemplateInput{TemplateID: testTemplateID, ForUpdate: true})
	assert.ErrorIs(t, err, runaimcp.ErrConflict)
	assert.ErrorContains(t, err, "cannot safely preserve")
}

func TestGetTemplateSessionTokenHandling(t *testing.T) {
	tests := []struct {
		name               string
		configure          func(map[string]any)
		wantUpdateConflict bool
		wantBucketInUpdate bool
	}{
		{
			name: "mutable token blocks update",
			configure: func(response map[string]any) {
				response["spec"] = map[string]any{
					"image": "worker:1",
					"storage": map[string]any{"s3": []map[string]any{{
						"name": "models", "bucket": "model-bucket", "path": "/models",
						"awsSessionToken": "session-token",
					}}},
				}
			},
			wantUpdateConflict: true,
		},
		{
			name: "response-only token does not block update",
			configure: func(response map[string]any) {
				response["spec"] = map[string]any{"image": "worker:1"}
				response["combinedSpec"] = map[string]any{
					"image": "worker:1",
					"storage": map[string]any{"s3": []map[string]any{{
						"name": "models", "bucket": "model-bucket", "path": "/models",
						"awsSessionToken": "session-token",
					}}},
				}
			},
		},
		{
			name: "null mutable token is omitted",
			configure: func(response map[string]any) {
				response["spec"] = map[string]any{
					"image": "worker:1",
					"storage": map[string]any{"s3": []map[string]any{{
						"name": "models", "bucket": "model-bucket", "path": "/models",
						"awsSessionToken": nil,
					}}},
				}
			},
			wantBucketInUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
				response := templateResponse(templateTypeTraining, "secret-template")
				tt.configure(response)
				runaiclienttest.WriteJSON(w, http.StatusOK, response)
			})

			readOutput, err := getTemplateRun(context.Background(), client, getTemplateInput{TemplateID: testTemplateID})
			require.NoError(t, err)
			encoded, err := json.Marshal(readOutput)
			require.NoError(t, err)
			assert.NotContains(t, string(encoded), "awsSessionToken")
			assert.NotContains(t, string(encoded), "session-token")
			assert.Contains(t, string(encoded), "model-bucket")

			updateOutput, err := getTemplateRun(context.Background(), client, getTemplateInput{
				TemplateID: testTemplateID,
				ForUpdate:  true,
			})
			if tt.wantUpdateConflict {
				assert.ErrorIs(t, err, runaimcp.ErrConflict)
				assert.ErrorContains(t, err, "awsSessionToken")
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, updateOutput.ExpectedFingerprint)
			encoded, err = json.Marshal(updateOutput.UpdateRequest)
			require.NoError(t, err)
			assert.NotContains(t, string(encoded), "awsSessionToken")
			assert.Equal(t, tt.wantBucketInUpdate, strings.Contains(string(encoded), "model-bucket"))
			assert.NotContains(t, updateOutput.UpdateRequest, "combinedSpec")
		})
	}
}

func TestGetTemplateRunRejectsNilSuccessResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	_, err := getTemplateRun(context.Background(), client, getTemplateInput{TemplateID: testTemplateID})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "nil response")
}

func TestGetTemplateIsRegistered(t *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	registry := runaimcp.New(srv, runaimcp.Options{Logger: zerolog.New(io.Discard)})
	Register(registry, nil)
	assert.Contains(t, registry.ToolNames(), "get_template")
}
