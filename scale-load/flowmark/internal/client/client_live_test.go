/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

func TestLiveEnvironment(t *testing.T) {
	if os.Getenv("FLOWMARK_LIVE_TEST") == "" {
		t.Skip("set FLOWMARK_LIVE_TEST=1, FLOWMARK_CP_URL, KUBECONFIG and FLOWMARK_CP_* credentials to run")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	env := core.Env{
		KubeconfigPath:  os.Getenv("KUBECONFIG"),
		ControlPlaneURL: os.Getenv("FLOWMARK_CP_URL"),
		Username:        os.Getenv("FLOWMARK_CP_USER"),
		Password:        os.Getenv("FLOWMARK_CP_PASSWORD"),
		Token:           os.Getenv("FLOWMARK_CP_TOKEN"),
	}

	clients, err := New(ctx, env)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	projects, _, err := clients.OrgUnits.ProjectsAPI.GetProjects(ctx).Execute()
	if err != nil {
		t.Fatalf("GET /projects: %v", err)
	}
	t.Logf("login ok, projects visible: %d", len(projects.Projects))

	var nodes corev1.NodeList
	if err := clients.Kube.List(ctx, &nodes); err != nil {
		t.Fatalf("listing nodes: %v", err)
	}
	t.Logf("kube ok, nodes: %d", len(nodes.Items))
}
