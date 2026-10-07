/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func fakeControlPlane(t *testing.T, clusters string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case versionPath:
			_, _ = w.Write([]byte(`{"version":"3.1.0-rc.4"}`))
		case "/api/v1/clusters":
			_, _ = w.Write([]byte(clusters))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestEnvironmentReadsVersionsPlatformAndGPUKind(t *testing.T) {
	server := fakeControlPlane(t, `[{"uuid":"c1","tenantId":1,"name":"kind-693","createdAt":"2026-10-05T08:00:00Z","version":"2.26.30","status":{"platform":{"type":"vanilla","kubeVersion":"v1.35.5"}}}]`)
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker-0", Labels: map[string]string{fakeGPULabel: "true"}}}
	clients := &Clients{
		Kube:           fake.NewClientBuilder().WithObjects(node).Build(),
		ClusterService: clusterclient.NewAPIClient(&clusterclient.Configuration{Servers: clusterclient.ServerConfigurations{{URL: server.URL}}}),
		http:           server.Client(),
		baseURL:        server.URL,
	}

	env, err := clients.Environment(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if env.ControlPlaneVersion != "3.1.0-rc.4" || env.ClusterVersion != "2.26.30" || env.Platform != "vanilla" || env.KubernetesVersion != "v1.35.5" || !env.FakeGPUs {
		t.Fatalf("unexpected environment: %+v", env)
	}
	if got := env.ClusterType(); got != "vanilla-fake-gpu" {
		t.Fatalf("cluster type = %q, want vanilla-fake-gpu", got)
	}
}

func TestEnvironmentRealGPUsAndNamedCluster(t *testing.T) {
	server := fakeControlPlane(t, `[{"uuid":"a","tenantId":1,"name":"other","createdAt":"2026-10-05T08:00:00Z"},{"uuid":"b","tenantId":1,"name":"ocp","createdAt":"2026-10-05T08:00:00Z","version":"2.24.95","status":{"platform":{"type":"openshift","kubeVersion":"v1.33.2"}}}]`)
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu-0", Labels: map[string]string{"nvidia.com/gpu.present": "true"}}}
	clients := &Clients{
		Kube:           fake.NewClientBuilder().WithObjects(node).Build(),
		ClusterService: clusterclient.NewAPIClient(&clusterclient.Configuration{Servers: clusterclient.ServerConfigurations{{URL: server.URL}}}),
		http:           server.Client(),
		baseURL:        server.URL,
	}

	if _, err := clients.Environment(context.Background(), ""); err == nil {
		t.Fatal("two clusters without a name must be an error")
	}
	env, err := clients.Environment(context.Background(), "ocp")
	if err != nil {
		t.Fatal(err)
	}
	if env.ClusterType() != "openshift-real-gpu" || env.ClusterVersion != "2.24.95" {
		t.Fatalf("unexpected environment: %+v", env)
	}
}
