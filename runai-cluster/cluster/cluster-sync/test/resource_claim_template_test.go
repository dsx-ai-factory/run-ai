/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaijobv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Exercises the ResourceClaimTemplate controller/handler. Unlike DeviceClass,
// RCTs intentionally go through a single collapsed /sync endpoint (no per-item
// CRUD surface - see the handler for the why), so every informer event funnels
// through the same coalesced resync key. The test asserts that:
//
//   - the full-set sync carries exactly the RCTs the handler is expected to
//     keep (project-scope with the control-plane project id resolved from the
//     kai Project, cluster/department-scope in the runai namespace, and
//     asset-owned templates), and
//   - RCTs the handler must drop as replicas never leak out: cluster/
//     department-scope templates outside the runai namespace, and project-
//     scope templates in namespaces that are not surfaced by GetAllProjects
//     or whose Project does not expose its id label.
var _ = Describe("A resource_claim_template controller test", func() {
	var (
		server           *ghttp.Server
		fakeCachedClient cachedclient.CachedClient
		backendClient    *backendclient.BackendClient

		stopper chan struct{}
		ctx     context.Context
		cancel  context.CancelFunc

		// Captures every sync body the handler POSTs. The RCT handler
		// coalesces informer events through a single workqueue key, so
		// we see one call per rate-limited tick rather than one per
		// event - tests look at the latest captured body under the
		// mutex rather than racing on a specific call ordinal.
		syncMu     sync.Mutex
		syncBodies []trackerclient.ResourceClaimTemplates
	)

	BeforeEach(func() {
		server = CreateBasicRunaiBackend()
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		DeferCleanup(server.Close)
		server.AllowUnhandledRequests = true
	})

	// Resets the shared sync accumulator for every spec. The handler's
	// coalesced sync key means older spec bodies would otherwise bleed
	// across Serial It-blocks that share this Context's BeforeEach.
	resetSyncBodies := func() {
		syncMu.Lock()
		defer syncMu.Unlock()
		syncBodies = nil
	}

	latestSync := func() trackerclient.ResourceClaimTemplates {
		syncMu.Lock()
		defer syncMu.Unlock()
		if len(syncBodies) == 0 {
			return trackerclient.ResourceClaimTemplates{}
		}
		return syncBodies[len(syncBodies)-1]
	}

	latestItemNames := func() []string {
		latest := latestSync()
		names := make([]string, 0, len(latest.Items))
		for i := range latest.Items {
			names = append(names, latest.Items[i].Name)
		}
		return names
	}

	findItem := func(name string) *trackerclient.ResourceClaimTemplate {
		latest := latestSync()
		for i := range latest.Items {
			if latest.Items[i].Name == name {
				return &latest.Items[i]
			}
		}
		return nil
	}

	Context("runai backend is alive, mix of scoped and replica RCTs live in the cluster", Serial, func() {
		// --- Projects wiring the handler's namespace -> project-id map. ---
		// The handler reads kai.resources Projects (not legacy v2.Projects)
		// via util.GetAllProjects; the project id label is what the sync
		// payload's scopeId must carry for project-scoped RCTs.
		var (
			projectWithIdNamespace = "runai-ofer"
			projectWithIdValue     = "4500002"
			projectNoIdNamespace   = "runai-no-id"

			departmentIdLabel = "4500003"

			projectWithId    = BuildKaiProject("ofer", projectWithIdValue, projectWithIdNamespace)
			projectWithoutId = BuildKaiProject("no-id", "", projectNoIdNamespace)
		)

		// --- RCTs the handler must ship. ---
		var (
			projectRct = BuildResourceClaimTemplate(
				"rct-project", projectWithIdNamespace, nil)

			clusterRct = BuildResourceClaimTemplate(
				"rct-cluster", runaijobv1.DefaultNamespace, map[string]string{
					commonconstants.ClusterWideResourceLabel: "true",
				})

			departmentRct = BuildResourceClaimTemplate(
				"rct-department", runaijobv1.DefaultNamespace, map[string]string{
					commonconstants.DepartmentScopeLabel: departmentIdLabel,
				})

			assetRct = BuildResourceClaimTemplate(
				"rct-asset", projectWithIdNamespace, map[string]string{
					commonconstants.RunaiAssetLabel:   "ResourceClaimTemplate",
					commonconstants.RunaiAssetIdLabel: "asset-xyz",
				})
		)

		// --- RCTs the handler must drop silently. ---
		var (
			// Project namespace exists but the Project has no id
			// label; the handler can't resolve a scopeId so this is
			// treated as a replica.
			projectNoIdRct = BuildResourceClaimTemplate(
				"rct-project-no-id", projectNoIdNamespace, nil)

			// Namespace is not Run:ai-managed at all (no Project
			// backs it); same code path as the no-id case.
			projectUnmanagedRct = BuildResourceClaimTemplate(
				"rct-project-unmanaged", "random-ns", nil)

			// Cluster-scoped template outside the runai namespace -
			// a per-project copy left behind by replication - must
			// never be reported back to the control plane.
			clusterReplicaRct = BuildResourceClaimTemplate(
				"rct-cluster-replica", projectWithIdNamespace, map[string]string{
					commonconstants.ClusterWideResourceLabel: "true",
				})

			// Same rule applies to department-scope templates.
			departmentReplicaRct = BuildResourceClaimTemplate(
				"rct-department-replica", projectWithIdNamespace, map[string]string{
					commonconstants.DepartmentScopeLabel: departmentIdLabel,
				})
		)

		shippedNames := []string{
			projectRct.Name,
			clusterRct.Name,
			departmentRct.Name,
			assetRct.Name,
		}
		droppedNames := []string{
			projectNoIdRct.Name,
			projectUnmanagedRct.Name,
			clusterReplicaRct.Name,
			departmentReplicaRct.Name,
		}

		BeforeEach(func() {
			resetSyncBodies()

			server.RouteToHandler("POST", "/api/v1/resource-claim-templates/sync", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.ResourceClaimTemplates
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				syncMu.Lock()
				syncBodies = append(syncBodies, payload)
				syncMu.Unlock()
			})

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().
				WithScheme(GetTestScheme()).
				WithObjects(
					projectWithId, projectWithoutId,
					projectRct, clusterRct, departmentRct, assetRct,
					projectNoIdRct, projectUnmanagedRct,
					clusterReplicaRct, departmentReplicaRct,
				)

			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			controllers.AddResourceClaimTemplateHandlerToCache(backendClient, fakeCachedClient)

			stopper = make(chan struct{})
			ctx, cancel = context.WithCancel(context.TODO())
			go backendClient.StartSendingLivenessPeriodically(stopper)
			fakeCachedClient.Start(ctx, stopper)
		})

		AfterEach(func() {
			close(stopper)
			cancel()
			server.Reset()
		})

		It("ships every scoped RCT and drops every replica/unresolvable RCT", func() {
			// Wait for the sync body to stabilize at the expected
			// set. Successive coalesced syncs may arrive before the
			// informer finishes replaying every initial Add, so poll
			// the latest body rather than the first one.
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			// The dropped set must never appear - even on an
			// intermediate sync. Scan every captured body.
			syncMu.Lock()
			bodies := append([]trackerclient.ResourceClaimTemplates{}, syncBodies...)
			syncMu.Unlock()
			for _, body := range bodies {
				for _, item := range body.Items {
					Expect(droppedNames).ToNot(ContainElement(item.Name),
						"replica/unresolvable RCT %q leaked into a sync body", item.Name)
				}
			}
		})

		It("rewrites the project RCT scopeId to the control-plane project id", func() {
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			item := findItem(projectRct.Name)
			Expect(item).ToNot(BeNil())
			Expect(item.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_PROJECT)))
			Expect(item.ScopeId).To(Equal(projectWithIdValue))
			Expect(item.Namespace).To(Equal(projectWithIdNamespace))
			Expect(item.Origin.IsSet()).To(BeTrue())
			Expect(*item.Origin.Get()).To(Equal(trackerclient.ORIGIN_EXISTING))
		})

		It("stamps the cluster uuid on cluster-scope RCTs and the department id on department-scope RCTs", func() {
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			cluster := findItem(clusterRct.Name)
			Expect(cluster).ToNot(BeNil())
			Expect(cluster.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_CLUSTER)))
			Expect(cluster.ScopeId).To(Equal(ClusterIdForTests))

			department := findItem(departmentRct.Name)
			Expect(department).ToNot(BeNil())
			Expect(department.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_DEPARTMENT)))
			Expect(department.ScopeId).To(Equal(departmentIdLabel))
		})

		It("marks asset-owned RCTs as origin=assets and propagates the asset id", func() {
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			item := findItem(assetRct.Name)
			Expect(item).ToNot(BeNil())
			Expect(item.Origin.IsSet()).To(BeTrue())
			Expect(*item.Origin.Get()).To(Equal(trackerclient.ORIGIN_ASSETS))
			Expect(item.AssetId.IsSet()).To(BeTrue())
			Expect(*item.AssetId.Get()).To(Equal("asset-xyz"))
		})

		It("re-syncs the full inventory after an informer Add (any event triggers the coalesced resync)", func() {
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			added := BuildResourceClaimTemplate("rct-project-added", projectWithIdNamespace, nil)
			Expect(fakeCachedClient.Create(context.TODO(), added)).To(Succeed())

			expected := append([]string{added.Name}, shippedNames...)
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(expected))
		})

		It("re-syncs after a Delete, surfacing the shrunk inventory", func() {
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedNames))

			Expect(fakeCachedClient.Delete(context.TODO(), departmentRct)).To(Succeed())

			remaining := []string{projectRct.Name, clusterRct.Name, assetRct.Name}
			Eventually(latestItemNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(remaining))
		})
	})
})
