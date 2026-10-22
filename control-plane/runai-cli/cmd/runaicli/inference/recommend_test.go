// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"

	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
)

func TestRecommendationHTTPContract(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/api/v1/inference/right-sizing/recommendation" {
					t.Errorf("wrong route: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer <test-token>" {
					t.Error("missing authentication")
				}
				var request map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				if request["image"] != "nvcr.io/nim/nvidia/example" || request["tag"] != "1.2.3" || request["clusterId"] != "cluster-id" || request["gpuType"] != "A100" {
					t.Errorf("wrong request: %+v", request)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status != http.StatusOK {
					_, _ = w.Write([]byte(`{"code":503,"message":"NGC sizing metadata unavailable"}`))
					return
				}
				_, _ = w.Write([]byte(`{"image":"nvcr.io/nim/nvidia/example","tag":"1.2.3","results":[{"gpuType":"A100","status":"recommended","minimumGpuMemoryPerDeviceGb":36,"recommendation":{"allocationType":"gpu_memory","gpuCount":1,"gpuMemoryPerDeviceBytes":36000000000,"estimated":false}},{"gpuType":"T4","status":"no_recommendation","reason":"Missing capacity"}]}`))
			}))
			defer server.Close()
			request := catalog.NewRightSizingRequest("cluster-id", "nvcr.io/nim/nvidia/example", "1.2.3")
			request.SetGpuType("A100")
			var output bytes.Buffer
			err := runRecommendation(context.Background(), server.URL, "<test-token>", *request, &output)
			if status != http.StatusOK {
				if err == nil || !strings.Contains(err.Error(), "NGC sizing metadata unavailable") || output.Len() != 0 {
					t.Fatalf("upstream failure hidden: %v %s", err, output.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				Results []map[string]interface{} `json:"results"`
			}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.Results) != 2 || result.Results[0]["minimumGpuMemoryPerDeviceGb"] != float64(36) {
				t.Fatalf("bad output: %s", output.String())
			}
			recommendation := result.Results[0]["recommendation"].(map[string]interface{})
			if recommendation["estimated"] != false {
				t.Fatalf("lost false estimated: %s", output.String())
			}
			if result.Results[1]["recommendation"] != nil {
				t.Fatalf("invented recommendation: %s", output.String())
			}
		})
	}
}

func TestRecommendLifecycleKeepsJSONClean(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := instance.Initializing(config.Options{ConfigPath: directory, ConfigFile: "config.json"}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	root := &cobra.Command{Use: "runai", PersistentPostRunE: func(cmd *cobra.Command, _ []string) error { cmd.Println("A CLI update is available"); return nil }}
	command := newRecommendCommand()
	// Replace only the network operation; execute Cobra's real post-run chain.
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		_, err := cmd.OutOrStdout().Write([]byte(`{"results":[]}`))
		return err
	}
	root.AddCommand(command)
	root.AddGroup(&cobra.Group{ID: command.GroupID, Title: "Inference"})
	root.SetOut(&output)
	root.SetArgs([]string{"recommend", "--image", "nvcr.io/nim/nvidia/example", "--tag", "1.2.3"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(output.Bytes()) {
		t.Fatalf("post-run output corrupted JSON: %s", output.String())
	}
	if instance.GetTxn().GetInfo().Cmd != "runai recommend" {
		t.Fatal("transaction reporting was skipped")
	}
}

func TestRecommendCommandPreservesCallerCancellation(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := instance.Initializing(config.Options{ConfigPath: directory, ConfigFile: "config.json"}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"image":"nvcr.io/nim/nvidia/example","tag":"1.2.3","results":[]}`))
	}))
	defer server.Close()
	if err := instance.SetCPURL(server.URL); err != nil {
		t.Fatal(err)
	}
	if err := instance.SetCluster("cluster-id", "test", ""); err != nil {
		t.Fatal(err)
	}
	instance.SetToken("<test-token>")
	var output bytes.Buffer
	root := &cobra.Command{Use: "runai", SilenceErrors: true, SilenceUsage: true}
	command := newRecommendCommand()
	root.AddCommand(command)
	root.AddGroup(&cobra.Group{ID: command.GroupID, Title: "Inference"})
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs([]string{"recommend", "--image", "nvcr.io/nim/nvidia/example", "--tag", "1.2.3"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := root.ExecuteContext(ctx); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("command discarded caller cancellation: %v output=%s", err, output.String())
	}
}
