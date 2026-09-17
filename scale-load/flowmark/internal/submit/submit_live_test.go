/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	orgunitclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

func TestLiveSubmitAndCleanup(t *testing.T) {
	if os.Getenv("FLOWMARK_LIVE_TEST") == "" {
		t.Skip("set FLOWMARK_LIVE_TEST=1, FLOWMARK_CP_URL, KUBECONFIG and FLOWMARK_CP_* credentials to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	clients, err := client.New(ctx, core.Env{
		KubeconfigPath:  os.Getenv("KUBECONFIG"),
		ControlPlaneURL: os.Getenv("FLOWMARK_CP_URL"),
		Username:        os.Getenv("FLOWMARK_CP_USER"),
		Password:        os.Getenv("FLOWMARK_CP_PASSWORD"),
		Token:           os.Getenv("FLOWMARK_CP_TOKEN"),
	})
	if err != nil {
		t.Fatalf("client.New: %v", err)
	}

	clusterID, err := ClusterID(ctx, clients, os.Getenv("FLOWMARK_CLUSTER"))
	if err != nil {
		t.Fatalf("ClusterID: %v", err)
	}

	cfg := core.Config{Knobs: map[string]any{"project_ready_timeout": "3m"}}
	timeout, err := cfg.Duration("project_ready_timeout", 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	stamp := time.Now().Unix() % 1000000
	nameA := fmt.Sprintf("flowmark-a-%d", stamp)
	nameB := fmt.Sprintf("flowmark-b-%d", stamp)
	nameC := fmt.Sprintf("flowmark-c-%d", stamp)
	t.Cleanup(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 2*timeout+time.Minute)
		defer cancelCleanup()
		for _, name := range []string{nameA, nameB} {
			if err := CleanupProject(cleanupCtx, clients, clusterID, name, timeout); err != nil {
				t.Logf("cleanup %s: %v", name, err)
			}
		}
		deleteForeignProject(cleanupCtx, t, clients, clusterID, nameC)
	})

	started := time.Now()
	projectA, err := Project(ctx, clients, clusterID, nameA, timeout)
	if err != nil {
		t.Fatalf("Project A: %v", err)
	}
	t.Logf("project %s ready in %s, namespace %s", projectA.Name, time.Since(started).Round(time.Second), projectA.Namespace)

	again, err := Project(ctx, clients, clusterID, nameA, timeout)
	if err != nil {
		t.Fatalf("Project A again (409 reuse): %v", err)
	}
	if again.ID != projectA.ID {
		t.Fatalf("409 reuse returned id %s, want %s", again.ID, projectA.ID)
	}
	t.Log("existing project reused with the same id")

	training, err := Training(ctx, clients, WorkloadSpec{Name: fmt.Sprintf("fm-train-%d", stamp), Project: projectA, Image: "ubuntu", Command: "sleep infinity"})
	if err != nil {
		t.Fatalf("Training: %v", err)
	}
	workspace, err := Workspace(ctx, clients, WorkloadSpec{Name: fmt.Sprintf("fm-ws-%d", stamp), Project: projectA, Image: "ubuntu", Command: "sleep infinity"})
	if err != nil {
		t.Fatalf("Workspace: %v", err)
	}
	deployment, err := K8sDeployment(ctx, clients.Kube, DeploymentSpec{Name: "fm-dep", Namespace: projectA.Namespace, Image: "nginx", Replicas: 1})
	if err != nil {
		t.Fatalf("K8sDeployment: %v", err)
	}

	for _, ref := range []WorkloadRef{training, workspace} {
		detailed, _, err := clients.Workloads.WorkloadsAPI.GetWorkload(ctx, ref.ID).Execute()
		if err != nil {
			t.Fatalf("GetWorkload %s: %v", ref.Name, err)
		}
		if detailed.Name != ref.Name || detailed.ProjectId != projectA.ID {
			t.Fatalf("workload %s: name=%s project=%s, want %s in project %s", ref.ID, detailed.Name, detailed.ProjectId, ref.Name, projectA.ID)
		}
		t.Logf("workload %s exists (id %s, phase %s)", ref.Name, ref.ID, detailed.Phase)
	}
	var fetched appsv1.Deployment
	if err := clients.Kube.Get(ctx, k8sclient.ObjectKeyFromObject(deployment), &fetched); err != nil {
		t.Fatalf("deployment not found: %v", err)
	}
	t.Logf("deployment %s/%s exists", fetched.Namespace, fetched.Name)

	projectB, err := Project(ctx, clients, clusterID, nameB, timeout)
	if err != nil {
		t.Fatalf("Project B: %v", err)
	}

	cleanupStarted := time.Now()
	if err := CleanupProject(ctx, clients, clusterID, nameA, timeout); err != nil {
		t.Fatalf("CleanupProject A: %v", err)
	}
	t.Logf("project %s cleaned up in %s", nameA, time.Since(cleanupStarted).Round(time.Second))
	if _, err := managedProject(ctx, clients, clusterID, nameA); !errors.Is(err, errProjectNotFound) {
		t.Fatalf("project A must be gone after cleanup, got %v", err)
	}
	assertNamespaceGone(ctx, t, clients.Kube, projectA.Namespace)

	if _, err := managedProject(ctx, clients, clusterID, nameB); err != nil {
		t.Fatalf("project B must be untouched: %v", err)
	}
	var namespaceB corev1.Namespace
	if err := clients.Kube.Get(ctx, k8sclient.ObjectKey{Name: projectB.Namespace}, &namespaceB); err != nil || namespaceB.Status.Phase != corev1.NamespaceActive {
		t.Fatalf("namespace of project B must stay active: phase=%s err=%v", namespaceB.Status.Phase, err)
	}
	t.Logf("anti-nuke ok: %s deleted, %s untouched", nameA, nameB)

	pools, err := nodepools(ctx, clients, clusterID)
	if err != nil {
		t.Fatal(err)
	}
	foreign := projectRequest(nameC, clusterID, pools)
	foreign.Description = orgunitclient.PtrString("created by a human, not flowmark")
	if _, _, err := clients.OrgUnits.ProjectsAPI.CreateProject(ctx).ProjectCreationRequest(foreign).Execute(); err != nil {
		t.Fatalf("creating foreign project: %v", err)
	}
	err = CleanupProject(ctx, clients, clusterID, nameC, timeout)
	if err == nil || !strings.Contains(err.Error(), "not created by flowmark") {
		t.Fatalf("cleanup of a foreign project must be refused, got %v", err)
	}
	if _, err := Project(ctx, clients, clusterID, nameC, timeout); err == nil || !strings.Contains(err.Error(), "not created by flowmark") {
		t.Fatalf("409 reuse of a foreign project must be refused, got %v", err)
	}
	t.Logf("ownership guard ok: foreign project %s refused for both reuse and cleanup", nameC)
}

func deleteForeignProject(ctx context.Context, t *testing.T, clients *client.Clients, clusterID, name string) {
	t.Helper()
	projects, _, err := clients.OrgUnits.ProjectsAPI.GetProjects(ctx).FilterBy(projectFilter(clusterID, name)).Execute()
	if err != nil {
		t.Logf("looking up foreign project %s: %v", name, err)
		return
	}
	project, err := projectByName(projects.Projects, name)
	if err != nil {
		return
	}
	if _, err := clients.OrgUnits.ProjectsAPI.DeleteProject(ctx, project.Id).Execute(); err != nil {
		t.Logf("deleting foreign project %s: %v", name, err)
		return
	}
	if err := waitClusterProjectGone(ctx, clients.Kube, name, time.Minute); err != nil {
		t.Logf("%v", err)
		return
	}
	if namespace, ok := project.Status.GetNamespaceOk(); ok && namespace != nil && *namespace != "" {
		if err := deleteProjectNamespace(ctx, clients.Kube, *namespace, name); err != nil {
			t.Logf("%v", err)
		}
	}
}

func assertNamespaceGone(ctx context.Context, t *testing.T, kube k8sclient.Client, name string) {
	t.Helper()
	var namespace corev1.Namespace
	err := kube.Get(ctx, k8sclient.ObjectKey{Name: name}, &namespace)
	switch {
	case apierrors.IsNotFound(err):
	case err != nil:
		t.Fatalf("checking namespace %s: %v", name, err)
	case namespace.DeletionTimestamp == nil:
		t.Fatalf("namespace %s still active after cleanup", name)
	}
}
