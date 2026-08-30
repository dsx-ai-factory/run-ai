package workloads

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const (
	submitProjectID  = "4500002"
	submitClusterID  = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	submitWorkloadID = "4bfbe3a8-6659-44a3-914f-b82952b0d99b"
)

type fakeNativeResponse struct {
	workloadID string
	name       string
}

func (r *fakeNativeResponse) GetWorkloadId() string { return r.workloadID }
func (r *fakeNativeResponse) GetName() string       { return r.name }

type valueNativeResponse struct {
	workloadID string
	name       string
}

func (r valueNativeResponse) GetWorkloadId() string { return r.workloadID }
func (r valueNativeResponse) GetName() string       { return r.name }

func TestSubmitWorkloadInputValidate(t *testing.T) {
	validNative := &nativeWorkloadSubmission{WorkloadType: workloadTypeTraining, Request: map[string]any{
		"name": "train", "projectId": submitProjectID, "clusterId": submitClusterID,
	}}
	validYAML := &yamlWorkloadSubmission{
		Metadata:     map[string]any{"projectId": submitProjectID},
		ManifestYAML: "apiVersion: batch/v1\nkind: Job\n",
	}
	tests := []struct {
		name    string
		input   submitWorkloadInput
		wantErr string
	}{
		{name: "native", input: submitWorkloadInput{Source: submitSourceNative, Native: validNative}},
		{name: "native dry run", input: submitWorkloadInput{Source: submitSourceNative, Native: validNative, DryRun: true}, wantErr: "dryRun is supported only"},
		{name: "yaml", input: submitWorkloadInput{Source: submitSourceYAML, YAML: validYAML}},
		{name: "unknown source", input: submitWorkloadInput{Source: "file"}, wantErr: "source must be one of"},
		{name: "missing native", input: submitWorkloadInput{Source: submitSourceNative}, wantErr: "native is required"},
		{name: "both payloads", input: submitWorkloadInput{Source: submitSourceNative, Native: validNative, YAML: validYAML}, wantErr: "yaml must not be set"},
		{name: "missing native request", input: submitWorkloadInput{Source: submitSourceNative, Native: &nativeWorkloadSubmission{WorkloadType: workloadTypeTraining}}, wantErr: "native.request is required"},
		{name: "empty manifest", input: submitWorkloadInput{Source: submitSourceYAML, YAML: &yamlWorkloadSubmission{}}, wantErr: "manifestYAML is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSubmitWorkloadRegistrationKeepsNativeModeOnOlderVersions(t *testing.T) {
	toolNames := func(version string) []string {
		srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
		registry := runaimcp.New(srv, runaimcp.Options{APIVersion: version})
		registerSubmitWorkload(registry, nil)
		return registry.ToolNames()
	}

	assert.Contains(t, toolNames("2.22"), "submit_workload")
	assert.Contains(t, toolNames(submitWorkloadMinAPIVersion), "submit_workload")

	_, err := submitWorkloadRun(context.Background(), nil, submitWorkloadInput{Source: submitSourceYAML}, false)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.ErrorContains(t, err, "requires control-plane version")
}

func TestSubmitWorkloadDescriptionMatchesVersionCapabilities(t *testing.T) {
	nativeOnly := submitWorkloadDescription(false)
	assert.Contains(t, nativeOnly, "supports source=native only")
	assert.Contains(t, nativeOnly, submitWorkloadMinAPIVersion)

	withYAML := submitWorkloadDescription(true)
	assert.Contains(t, withYAML, "two submission paths")
	assert.Contains(t, withYAML, "native Run:ai workload API")
	assert.Contains(t, withYAML, "Kubernetes YAML manifest")
	assert.Contains(t, withYAML, "source=native")
	assert.Contains(t, withYAML, "source=yaml")
	assert.Contains(t, withYAML, "dryRun=true")
}

func TestSubmitNativeWorkloadPreservesCompleteRequest(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)

	wantSpec := map[string]any{
		"image":   "nvcr.io/nvidia/pytorch:latest",
		"command": "python train.py",
		"environmentVariables": []any{
			map[string]any{"name": "EPOCHS", "value": "10"},
		},
	}
	mux.HandleFunc("/api/v1/workloads/trainings", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "template-1", body["templateId"])
		assert.Equal(t, wantSpec, body["spec"])
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "train-final", "workloadId": submitWorkloadID,
			"projectId": submitProjectID, "clusterId": submitClusterID,
		})
	})

	out, err := submitNativeWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceNative,
		Native: &nativeWorkloadSubmission{
			WorkloadType: workloadTypeTraining,
			Request: map[string]any{
				"name": "train", "projectId": submitProjectID, "clusterId": submitClusterID,
				"templateId": "template-1", "spec": wantSpec,
			},
		},
		Confirm: true,
	})
	require.NoError(t, err)
	assert.Equal(t, submitWorkloadID, out.ID)
	assert.Equal(t, "train-final", out.Name)
	assert.Equal(t, workloadTypeTraining, out.WorkloadType)
	assert.False(t, out.DryRun)
}

func TestSubmitNativeWorkloadRequiresConfirmationBeforeCreate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	createCalled := false
	mux.HandleFunc("/api/v1/workloads/trainings", func(w http.ResponseWriter, _ *http.Request) {
		createCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	out, err := submitNativeWorkload(context.Background(), client, nativeSubmitInput(workloadTypeTraining))
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.False(t, createCalled)
	assert.Equal(t, "train", out.Name)
}

func TestSubmitNativeWorkloadDelegatesAuthorizationToCreateAPI(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	createCalled := false
	mux.HandleFunc("/api/v1/workloads/trainings", func(w http.ResponseWriter, _ *http.Request) {
		createCalled = true
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "denied"})
	})

	in := nativeSubmitInput(workloadTypeTraining)
	in.Confirm = true
	_, err := submitNativeWorkload(context.Background(), client, in)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	assert.True(t, createCalled)
}

func TestSubmitNativeWorkloadRejectsUnknownRequestFields(t *testing.T) {
	tests := []struct {
		name    string
		request map[string]any
		field   string
	}{
		{
			name: "top level",
			request: map[string]any{
				"name": "train", "projectId": submitProjectID, "clusterId": submitClusterID, "specTypo": map[string]any{},
			},
			field: "specTypo",
		},
		{
			name: "nested spec",
			request: map[string]any{
				"name": "train", "projectId": submitProjectID, "clusterId": submitClusterID, "spec": map[string]any{"imageTypo": "image"},
			},
			field: "imageTypo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (submitWorkloadInput{Source: submitSourceNative, Native: &nativeWorkloadSubmission{
				WorkloadType: workloadTypeTraining,
				Request:      tt.request,
			}}).Validate()
			assert.ErrorIs(t, err, runaimcp.ErrValidation)
			assert.ErrorContains(t, err, tt.field)
		})
	}
}

func TestPrepareNativeSubmissionDispatchesEveryWorkloadType(t *testing.T) {
	tests := []struct {
		workloadType string
		path         string
	}{
		{workloadTypeTraining, "/api/v1/workloads/trainings"},
		{workloadTypeWorkspace, "/api/v1/workloads/workspaces"},
		{workloadTypeDistributed, "/api/v1/workloads/distributed"},
		{workloadTypeInference, "/api/v1/workloads/inferences"},
		{workloadTypeDistribInf, "/api/v1/workloads/distributed-inferences"},
	}
	for _, tt := range tests {
		t.Run(tt.workloadType, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(tt.path, func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{
					"name": "submitted", "workloadId": submitWorkloadID,
					"projectId": submitProjectID, "clusterId": submitClusterID,
				})
			})

			prepared, err := prepareNativeSubmission(nativeSubmitInput(tt.workloadType).Native)
			require.NoError(t, err)
			id, name, err := prepared.create(context.Background(), client)
			require.NoError(t, err)
			assert.Equal(t, submitWorkloadID, id)
			assert.Equal(t, "submitted", name)
		})
	}
}

func TestParseNativeCreationResult(t *testing.T) {
	var typedNil *fakeNativeResponse
	tests := []struct {
		name      string
		created   nativeResponse
		response  *http.Response
		upstream  error
		wantID    string
		wantName  string
		wantError error
	}{
		{name: "success", created: &fakeNativeResponse{workloadID: submitWorkloadID, name: "train"}, wantID: submitWorkloadID, wantName: "train"},
		{name: "non-pointer success", created: valueNativeResponse{workloadID: submitWorkloadID, name: "train-value"}, wantID: submitWorkloadID, wantName: "train-value"},
		{name: "nil response", wantError: runaimcp.ErrUpstream},
		{name: "typed nil response", created: typedNil, wantError: runaimcp.ErrUpstream},
		{name: "sparse success", created: &fakeNativeResponse{name: "train"}, wantName: "train"},
		{name: "upstream forbidden", response: &http.Response{StatusCode: http.StatusForbidden}, upstream: errors.New("denied"), wantError: runaimcp.ErrForbidden},
		{name: "accepted but undecodable", response: &http.Response{StatusCode: http.StatusAccepted}, upstream: errors.New("decode failed"), wantError: runaimcp.ErrUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, name, err := parseNativeCreationResult("training", tt.created, tt.response, tt.upstream)
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
				if tt.name == "accepted but undecodable" {
					assert.ErrorContains(t, err, "outcome is unknown")
					assert.ErrorContains(t, err, "inspect workloads before retrying")
				}
				assert.Empty(t, id)
				assert.Empty(t, name)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

func TestSubmitWorkloadInputRejectsWhitespaceNativeIdentity(t *testing.T) {
	in := nativeSubmitInput(workloadTypeTraining)
	in.Native.Request["name"] = " \t "
	err := in.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "native.request.name is required")
}

func TestSubmitWorkloadInputRejectsMalformedNativeScopeIDs(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr string
	}{
		{name: "project id", field: "projectId", value: "project-a", wantErr: "must be a numeric string"},
		{name: "cluster id", field: "clusterId", value: "cluster-a", wantErr: "must be a UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := nativeSubmitInput(workloadTypeTraining)
			in.Native.Request[tt.field] = tt.value
			err := in.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSubmitYAMLWorkloadUsesServerDryRunAndPreservesMetadata(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitNamespaceProject(t, mux, "team-a")
	manifest := "apiVersion: batch/v1\nkind: Job\nmetadata:\n  name: yaml-job\n  namespace: team-a\nspec:\n  template:\n    spec:\n      restartPolicy: Never\n"

	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "true", r.URL.Query().Get("dryRun"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		metadata, ok := body["metadata"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "high", metadata["priority"])
		configuration, ok := metadata["configuration"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Required", configuration["mnnvl"])
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(manifest)), body["manifestBase64Encoded"])
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"metadata": map[string]any{
				"id": "must-not-leak-from-dry-run", "name": "yaml-job", "projectId": submitProjectID,
			},
		})
	})

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata: map[string]any{
				"clusterId": submitClusterID,
				"priority":  "high",
				"configuration": map[string]any{
					"mnnvl": "Required",
				},
			},
			ManifestYAML: manifest,
		},
		DryRun: true,
	})
	require.NoError(t, err)
	assert.True(t, out.DryRun)
	assert.Empty(t, out.ID, "Workloads V2 intentionally omits an ID during dry-run")
	assert.Equal(t, submitProjectID, out.ProjectID, "V2 resolves namespace to project during dry-run")
	assert.Equal(t, submitClusterID, out.ClusterID, "keep the cluster resolved before the sparse dry-run response")
	assert.Equal(t, "Job", out.WorkloadType)
}

func TestSubmitYAMLWorkloadDryRunAcceptsEmptyDesiredPhaseResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("dryRun"))
		runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{
			"metadata": map[string]any{
				"name": "yaml-job", "projectId": submitProjectID, "clusterId": submitClusterID,
				"gvk": map[string]any{"kind": "Job"},
			},
			"desiredPhase":     "",
			"workloadManifest": map[string]any{},
			"rawManifest":      map[string]any{},
		})
	})

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: batch/v1\nkind: Job\nmetadata:\n  name: yaml-job\n",
		},
		DryRun: true,
	})
	require.NoError(t, err)
	assert.True(t, out.DryRun)
	assert.Empty(t, out.ID)
	assert.Equal(t, "yaml-job", out.Name)
	assert.Equal(t, submitProjectID, out.ProjectID)
	assert.Equal(t, submitClusterID, out.ClusterID)
	assert.Equal(t, "Job", out.WorkloadType)
}

func TestSubmitYAMLWorkloadRequiresConfirmationBeforeCreate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	createCalled := false
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		createCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
	})
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.False(t, createCalled)
	assert.Equal(t, "pod-a", out.Name)
}

func TestSubmitYAMLWorkloadDelegatesAuthorizationToCreateAPI(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	createCalled := false
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		createCalled = true
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "denied"})
	})

	_, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
		Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	assert.True(t, createCalled)
}

func TestSubmitNativeWorkloadRejectsProjectClusterMismatch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects/"+submitProjectID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": submitProjectID, "name": "team-a", "clusterId": "00000000-0000-0000-0000-000000000001",
		})
	})

	_, err := submitNativeWorkload(context.Background(), client, submitWorkloadInput{
		Source:  submitSourceNative,
		Native:  nativeSubmitInput(workloadTypeTraining).Native,
		Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.ErrorContains(t, err, "belongs to cluster")
}

func TestSubmitWorkloadInputValidatesYAMLSubmission(t *testing.T) {
	tests := []struct {
		name    string
		input   *yamlWorkloadSubmission
		wantErr string
	}{
		{name: "malformed", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID}, ManifestYAML: "kind: ["}, wantErr: "decode yaml.manifestYAML"},
		{name: "missing kind", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID}, ManifestYAML: "apiVersion: v1"}, wantErr: "apiVersion and kind"},
		{name: "missing target", input: &yamlWorkloadSubmission{ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "projectId or clusterId"},
		{name: "both targets", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID, "clusterId": submitClusterID}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "mutually exclusive"},
		{name: "cluster without namespace", input: &yamlWorkloadSubmission{Metadata: map[string]any{"clusterId": submitClusterID}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "metadata.namespace is required"},
		{name: "multiple documents", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID}, ManifestYAML: "apiVersion: v1\nkind: Pod\n---\napiVersion: v1\nkind: Service"}, wantErr: "exactly one YAML document"},
		{name: "whitespace target", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": "  "}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "projectId is required"},
		{name: "whitespace project with cluster", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": "  ", "clusterId": submitClusterID}, ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  namespace: team-a"}, wantErr: "mutually exclusive"},
		{name: "project with whitespace cluster", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID, "clusterId": "  "}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "mutually exclusive"},
		{name: "invalid project id", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": "project-a"}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "must be a numeric string"},
		{name: "invalid cluster id", input: &yamlWorkloadSubmission{Metadata: map[string]any{"clusterId": "cluster-a"}, ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  namespace: team-a"}, wantErr: "must be a UUID"},
		{name: "cluster filter injection", input: &yamlWorkloadSubmission{Metadata: map[string]any{"clusterId": submitClusterID + ",name==other"}, ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  namespace: team-a"}, wantErr: "must be a UUID"},
		{name: "unknown metadata field", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID, "future": true}, ManifestYAML: "apiVersion: v1\nkind: Pod"}, wantErr: "unknown field"},
		{name: "manifest exceeds API limit", input: &yamlWorkloadSubmission{Metadata: map[string]any{"projectId": submitProjectID}, ManifestYAML: strings.Repeat("x", maxEncodedYAMLManifestLength)}, wantErr: "maximum encoded size"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (submitWorkloadInput{Source: submitSourceYAML, YAML: tt.input}).Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestYAMLGenerateNameIsIncludedInConfirmation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: batch/v1\nkind: Job\nmetadata:\n  generateName: nightly-\n",
		},
	})

	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Equal(t, "nightly-", out.Name)
	assert.True(t, out.NameIsPrefix)
	assert.ErrorContains(t, err, "nameIsPrefix:true")
}

func TestResolveProjectForNamespaceContinuesPastListResultCap(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	requests := 0
	mux.HandleFunc("/api/v1/org-unit/projects/metadata", func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch r.URL.Query().Get("offset") {
		case "0":
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"projects": []map[string]any{{
					"id": "4500001", "name": "other", "clusterId": submitClusterID, "namespace": "other-team",
				}}, "next": common.MaxPaginatedResults,
			})
		case "2000":
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"projects": []map[string]any{{
					"id": submitProjectID, "name": "team-a", "clusterId": submitClusterID, "namespace": "late-team",
				}},
			})
		default:
			http.Error(w, "unexpected offset", http.StatusBadRequest)
		}
	})
	registerSubmitProject(mux)

	project, err := resolveProjectForNamespace(context.Background(), client, submitClusterID, "late-team")
	require.NoError(t, err)
	assert.Equal(t, submitProjectID, project.Id)
	assert.Equal(t, 2, requests)
}

func TestResolveProjectForNamespaceRejectsEmptyPageWithNextOffset(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects/metadata", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []any{}, "next": int32(200),
		})
	})

	_, err := resolveProjectForNamespace(context.Background(), client, submitClusterID, "team-a")
	require.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "empty page reported a next offset")
}

func TestGetSubmissionProjectTrustsAPIResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects/"+submitProjectID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "9999999", "name": "wrong", "clusterId": submitClusterID,
		})
	})

	project, err := getSubmissionProject(context.Background(), client, submitProjectID)
	require.NoError(t, err)
	assert.Equal(t, "9999999", project.Id)
}

func TestSubmitYAMLWorkloadUsesReturnedScope(t *testing.T) {
	tests := []struct {
		name            string
		responseProject string
		responseCluster string
	}{
		{name: "project", responseProject: "9999999", responseCluster: submitClusterID},
		{name: "cluster", responseProject: submitProjectID, responseCluster: "00000000-0000-0000-0000-000000000001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerSubmitProject(mux)
			mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusCreated, map[string]any{
					"metadata": map[string]any{
						"id": submitWorkloadID, "name": "pod-a", "projectId": tt.responseProject,
						"clusterId": tt.responseCluster, "gvk": map[string]any{"kind": "Pod"},
					},
				})
			})

			out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
				Source: submitSourceYAML,
				YAML: &yamlWorkloadSubmission{
					Metadata:     map[string]any{"projectId": submitProjectID},
					ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
				},
				Confirm: true,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.responseProject, out.ProjectID)
			assert.Equal(t, tt.responseCluster, out.ClusterID)
		})
	}
}

func TestSubmitYAMLWorkloadAcceptsSparseSuccessfulResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{
			"metadata": map[string]any{"name": "pod-a", "projectId": submitProjectID, "clusterId": submitClusterID},
		})
	})

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
		Confirm: true,
	})
	require.NoError(t, err)
	assert.Empty(t, out.ID)
	assert.Equal(t, "pod-a", out.Name)
}

func TestSubmitYAMLWorkloadRealCreateReturnsUpstreamIdentity(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "false", r.URL.Query().Get("dryRun"))
		runaiclienttest.WriteJSON(w, http.StatusCreated, map[string]any{
			"metadata": map[string]any{
				"id": submitWorkloadID, "name": "server-name", "projectId": submitProjectID,
				"clusterId": submitClusterID, "gvk": map[string]any{"kind": "Job"},
			},
		})
	})

	out, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: batch/v1\nkind: Job\nmetadata:\n  name: requested-name\n",
		},
		Confirm: true,
	})
	require.NoError(t, err)
	assert.Equal(t, submitWorkloadID, out.ID)
	assert.Equal(t, "server-name", out.Name)
	assert.Equal(t, submitProjectID, out.ProjectID)
	assert.Equal(t, submitClusterID, out.ClusterID)
	assert.Equal(t, "Job", out.WorkloadType)
	assert.False(t, out.DryRun)
}

func TestSubmitYAMLWorkloadRejectsNilUpstreamResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	_, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
		Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "nil response")
}

func TestSubmitYAMLWorkloadAcceptedButInvalidResponseHasUnknownOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"invalid"`))
	})

	_, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
		Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect workloads before retrying")
}

func TestSubmitYAMLWorkloadDryRunFailureNeverSuggestsCreationMayHaveOccurred(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerSubmitProject(mux)
	mux.HandleFunc("/api/v2/workloads", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"invalid"`))
	})

	_, err := submitYAMLWorkload(context.Background(), client, submitWorkloadInput{
		Source: submitSourceYAML,
		YAML: &yamlWorkloadSubmission{
			Metadata:     map[string]any{"projectId": submitProjectID},
			ManifestYAML: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-a\n",
		},
		DryRun: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "dry-run YAML workload")
	assert.NotContains(t, err.Error(), "outcome is unknown")
	assert.NotContains(t, err.Error(), "inspect workloads")
}

func nativeSubmitInput(workloadType string) submitWorkloadInput {
	return submitWorkloadInput{
		Source: submitSourceNative,
		Native: &nativeWorkloadSubmission{
			WorkloadType: workloadType,
			Request: map[string]any{
				"name": "train", "projectId": submitProjectID, "clusterId": submitClusterID,
			},
		},
	}
}

func registerSubmitProject(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/org-unit/projects/"+submitProjectID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": submitProjectID, "name": "team-a", "clusterId": submitClusterID, "parentId": "department-1",
		})
	})
}

func registerSubmitNamespaceProject(t *testing.T, mux *http.ServeMux, namespace string) {
	t.Helper()
	mux.HandleFunc("/api/v1/org-unit/projects/metadata", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "clusterId=="+submitClusterID, r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{{
				"id": submitProjectID, "name": "team-a", "clusterId": submitClusterID, "namespace": namespace,
			}},
		})
	})
	registerSubmitProject(mux)
}
