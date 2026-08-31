/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
)

func integrationCreds(t *testing.T) (baseURL, clientID, clientSecret string) {
	baseURL = os.Getenv("RUNAI_BASE_URL")
	clientID = os.Getenv("RUNAI_CLIENT_ID")
	clientSecret = os.Getenv("RUNAI_CLIENT_SECRET")
	if baseURL == "" || clientID == "" || clientSecret == "" {
		t.Fatal("RUNAI_BASE_URL, RUNAI_CLIENT_ID and RUNAI_CLIENT_SECRET must be set for the live test")
	}
	return baseURL, clientID, clientSecret
}

func TestIntegrationManageOrgUnit(t *testing.T) {
	if os.Getenv("RUNAI_INTEGRATION") != "1" {
		t.Skip("set RUNAI_INTEGRATION=1 to run the integration test")
	}
	baseURL, clientID, clientSecret := integrationCreds(t)
	ctx := context.Background()
	cc, err := auth.NewClientCredentials(ctx, baseURL, clientID, clientSecret)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := runaiclient.New(baseURL, cc)

	clusterID, nodePool := discoverTarget(ctx, t, client)
	t.Logf("target cluster=%s nodePool=%s", clusterID, nodePool)

	dump := func(label string, v any) {
		b, _ := json.MarshalIndent(v, "", "  ")
		t.Logf("\n===== %s =====\n%s", label, b)
	}
	run := func(label string, in manageOrgUnitInput) manageOrgUnitOutput {
		out, rErr := manageOrgUnitRun(ctx, client, in)
		if rErr != nil {
			t.Fatalf("%s failed: %v", label, rErr)
		}
		dump(label, out)
		return out
	}

	name := "claude-itest-" + time.Now().Format("150405")
	// CPU/memory quota is omitted here: the test clusters run with the CPU-memory quota
	// feature OFF, so any cpu/memory deserved value is rejected upstream. Their cores/GiB
	// conversion is covered by the unit test instead.
	createSpec := &orgUnitSpec{
		Description: ptr("claude integration test project"),
		Quota:       []orgUnitQuota{{NodePoolName: nodePool, GPUDeserved: f(1), GPULimit: f(2)}},
		SchedulingRules: &orgUnitSchedulingRules{
			TrainingJobTimeLimitSeconds:          i32(3600),
			InteractiveJobMaxIdleDurationSeconds: i32(600),
		},
	}

	run("CREATE dry-run", manageOrgUnitInput{Action: "create", OrgType: "project", Name: name, ClusterID: clusterID, DryRun: true, Spec: createSpec})

	created := run("CREATE", manageOrgUnitInput{Action: "create", OrgType: "project", Name: name, ClusterID: clusterID, Spec: createSpec})
	id := created.OrgUnitID
	if id == "" {
		t.Fatal("create returned empty id")
	}
	// Fail-safe: a later t.Fatalf must not leak the project into the live environment.
	defer func() {
		if _, dErr := manageOrgUnitRun(ctx, client, manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: id, Confirm: true, IgnoreMissing: true}); dErr != nil {
			t.Logf("cleanup delete of %s failed: %v", id, dErr)
		}
	}()

	sr, ok := created.After["schedulingRules"].(map[string]any)
	require.True(t, ok, "schedulingRules should be present after create")
	assert.EqualValues(t, 3600, sr["trainingJobTimeLimitSeconds"])
	assert.EqualValues(t, 600, sr["interactiveJobMaxIdleDurationSeconds"])

	// Partial schedulingRules update must merge: change one field, preserve the other.
	merged := run("UPDATE (partial schedulingRules)", manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: id,
		Spec: &orgUnitSpec{SchedulingRules: &orgUnitSchedulingRules{InteractiveJobMaxIdleDurationSeconds: i32(900)}}})
	msr := merged.After["schedulingRules"].(map[string]any)
	assert.EqualValues(t, 900, msr["interactiveJobMaxIdleDurationSeconds"], "changed field")
	assert.EqualValues(t, 3600, msr["trainingJobTimeLimitSeconds"], "untouched field must be preserved")

	run("UPDATE dry-run (desc only)", manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: id, DryRun: true, Spec: &orgUnitSpec{Description: ptr("updated by claude")}})
	upd := run("UPDATE (desc only)", manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: id, Spec: &orgUnitSpec{Description: ptr("updated by claude")}})
	if q, ok := upd.After["quota"].([]map[string]any); ok && len(q) == 0 {
		t.Error("quota was wiped on a description-only update")
	}

	run("UPDATE (bump gpu quota)", manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: id, Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: nodePool, GPUDeserved: f(2), GPULimit: f(4)}}}})

	if _, dErr := manageOrgUnitRun(ctx, client, manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: id}); dErr != nil {
		t.Logf("DELETE without confirm correctly refused: %v", dErr)
	} else {
		t.Error("delete without confirm should have failed")
	}

	run("DELETE (confirm)", manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: id, Confirm: true})
}

func TestIntegrationManageDepartment(t *testing.T) {
	if os.Getenv("RUNAI_INTEGRATION") != "1" {
		t.Skip("set RUNAI_INTEGRATION=1 to run the integration test")
	}
	baseURL, clientID, clientSecret := integrationCreds(t)
	ctx := context.Background()
	cc, err := auth.NewClientCredentials(ctx, baseURL, clientID, clientSecret)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := runaiclient.New(baseURL, cc)
	clusterID, nodePool := discoverTarget(ctx, t, client)

	run := func(label string, in manageOrgUnitInput) manageOrgUnitOutput {
		out, rErr := manageOrgUnitRun(ctx, client, in)
		if rErr != nil {
			t.Fatalf("%s failed: %v", label, rErr)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		t.Logf("\n===== %s =====\n%s", label, b)
		return out
	}

	stamp := time.Now().Format("150405")
	deptQuota := []orgUnitQuota{{NodePoolName: nodePool, GPUDeserved: f(2)}}

	dept := run("CREATE department", manageOrgUnitInput{Action: "create", OrgType: "department", Name: "claude-dept-" + stamp, ClusterID: clusterID, Spec: &orgUnitSpec{Description: ptr("claude dept"), Quota: deptQuota}})
	deptID := dept.OrgUnitID
	if deptID == "" {
		t.Fatal("department create returned empty id")
	}
	cleanup := func(in manageOrgUnitInput) {
		if _, err := manageOrgUnitRun(ctx, client, in); err != nil {
			t.Errorf("cleanup delete of %s %q failed (may leak state): %v", in.OrgType, in.OrgUnitID, err)
		}
	}
	defer cleanup(manageOrgUnitInput{Action: "delete", OrgType: "department", OrgUnitID: deptID, Confirm: true, IgnoreMissing: true})

	run("UPDATE department (desc)", manageOrgUnitInput{Action: "update", OrgType: "department", OrgUnitID: deptID, Spec: &orgUnitSpec{Description: ptr("claude dept updated")}})

	child := run("CREATE child project", manageOrgUnitInput{Action: "create", OrgType: "project", Name: "claude-child-" + stamp, ClusterID: clusterID, Spec: &orgUnitSpec{ParentID: ptr(deptID), Quota: []orgUnitQuota{{NodePoolName: nodePool, GPUDeserved: f(1)}}}})
	childID := child.OrgUnitID
	defer cleanup(manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: childID, Confirm: true, IgnoreMissing: true})

	if _, dErr := manageOrgUnitRun(ctx, client, manageOrgUnitInput{Action: "delete", OrgType: "department", OrgUnitID: deptID, Confirm: true}); dErr != nil {
		t.Logf("DELETE department correctly blocked by child: %v", dErr)
	} else {
		t.Error("expected department delete to be blocked by its child project")
	}

	run("DELETE child project", manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: childID, Confirm: true})
	run("DELETE department", manageOrgUnitInput{Action: "delete", OrgType: "department", OrgUnitID: deptID, Confirm: true})
}

func discoverTarget(ctx context.Context, t *testing.T, client *runaiclient.Client) (clusterID, nodePool string) {
	if c := os.Getenv("RUNAI_CLUSTER_ID"); c != "" {
		np := os.Getenv("RUNAI_NODE_POOL")
		if np == "" {
			np = "default"
		}
		return c, np
	}
	resp, httpResp, err := client.Cluster.NodePoolsAPI.GetNodepools(ctx).Limit(50).Execute()
	if err != nil || resp == nil {
		t.Fatalf("discover node pools: %v (http=%v)", err, httpResp)
	}
	if len(resp.Nodepools) == 0 {
		t.Fatal("no node pools found to target")
	}
	for _, np := range resp.Nodepools {
		if np.IsDefault {
			return np.ClusterId, np.Name
		}
	}
	return resp.Nodepools[0].ClusterId, resp.Nodepools[0].Name
}

func ptr(s string) *string { return &s }
func f(v float64) *float64 { return &v }
