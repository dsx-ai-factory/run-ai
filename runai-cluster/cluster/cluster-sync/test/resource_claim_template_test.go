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
	cs_constants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaijobv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Exercises the ResourceClaimTemplate controller/handler after it was aligned
// with the DeviceClass pattern:
//
//   - informer Add -> per-item POST (CreateResourceClaimTemplate)
//   - informer Update -> rate-limited PUT (UpdateResourceClaimTemplate)
//   - informer Delete -> per-item DELETE (DeleteResourceClaimTemplate)
//   - liveness sync -> full-set /sync (SyncResourceClaimTemplates) which also
//     prunes anything the control plane still has but the cluster does not.
//
// Pre-existing objects reach the control plane through the startup /sync -
// their Add events typically race the liveness ping and are rejected by the
// `IsRunaiBackendAvailable` filter, so per-item CRUD assertions only cover
// events that happen AFTER backend-available flips true.
//
// Replicas (cluster/department-scope outside the runai namespace, project-
// scope in unmanaged namespaces, project-scope whose Project has no id label)
// must never reach any endpoint.
var _ = Describe("A resource_claim_template controller test", func() {
	var (
		server           *ghttp.Server
		fakeCachedClient cachedclient.CachedClient
		backendClient    *backendclient.BackendClient

		stopper chan struct{}
		ctx     context.Context
		cancel  context.CancelFunc

		callsMu      sync.Mutex
		createdCalls []trackerclient.ResourceClaimTemplate
		updatedCalls []trackerclient.ResourceClaimTemplate
		deletedCalls []deletedRctCall
		syncBodies   []trackerclient.ResourceClaimTemplates
	)

	resetCalls := func() {
		callsMu.Lock()
		defer callsMu.Unlock()
		createdCalls = nil
		updatedCalls = nil
		deletedCalls = nil
		syncBodies = nil
	}

	namesOf := func(items []trackerclient.ResourceClaimTemplate) []string {
		names := make([]string, 0, len(items))
		for i := range items {
			names = append(names, items[i].Name)
		}
		return names
	}

	createdNames := func() []string {
		callsMu.Lock()
		defer callsMu.Unlock()
		return namesOf(createdCalls)
	}

	updatedNames := func() []string {
		callsMu.Lock()
		defer callsMu.Unlock()
		return namesOf(updatedCalls)
	}

	currentDeletedCalls := func() []deletedRctCall {
		callsMu.Lock()
		defer callsMu.Unlock()
		return append([]deletedRctCall{}, deletedCalls...)
	}

	latestSyncNames := func() []string {
		callsMu.Lock()
		defer callsMu.Unlock()
		if len(syncBodies) == 0 {
			return nil
		}
		return namesOf(syncBodies[len(syncBodies)-1].Items)
	}

	createdByName := func(name string) (trackerclient.ResourceClaimTemplate, bool) {
		callsMu.Lock()
		defer callsMu.Unlock()
		for i := range createdCalls {
			if createdCalls[i].Name == name {
				return createdCalls[i], true
			}
		}
		return trackerclient.ResourceClaimTemplate{}, false
	}

	BeforeEach(func() {
		server = CreateBasicRunaiBackend()
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		DeferCleanup(server.Close)
		server.AllowUnhandledRequests = true
	})

	Context("backend is alive, mix of scoped and replica RCTs live in the cluster", Serial, func() {
		// Projects wiring the handler's namespace -> project-id resolver.
		// The per-item path goes through util.GetProjectIdFromNamespaceName
		// (which reads the Namespace RunaiQueueLabel then fetches the kai
		// Project), so the fixture must include the Namespace.
		var (
			projectWithIdNamespace = "runai-ofer"
			projectName            = "ofer"
			projectIdValue         = "4500002"
			departmentIdLabel      = "4500003"

			runaiManagedNs = BuildNamespace(projectWithIdNamespace,
				map[string]string{cs_constants.RunaiQueueLabel: projectName})
			kaiProject    = BuildKaiProject(projectName, projectIdValue, projectWithIdNamespace)
			legacyProject = BuildProject(projectName, projectIdValue, projectWithIdNamespace)
		)

		// Pre-existing RCTs - these reach the server through startup /sync.
		var (
			existingProjectRct = BuildResourceClaimTemplate(
				"rct-project", projectWithIdNamespace, nil)
			existingClusterRct = BuildResourceClaimTemplate(
				"rct-cluster", runaijobv1.DefaultNamespace, map[string]string{
					commonconstants.ClusterWideResourceLabel: "true",
				})
			existingDepartmentRct = BuildResourceClaimTemplate(
				"rct-department", runaijobv1.DefaultNamespace, map[string]string{
					commonconstants.DepartmentScopeLabel: departmentIdLabel,
				})
		)

		// Replicas the handler must drop (and so must never show up in
		// any per-item or /sync body).
		var (
			clusterReplicaRct = BuildResourceClaimTemplate(
				"rct-cluster-replica", projectWithIdNamespace, map[string]string{
					commonconstants.ClusterWideResourceLabel: "true",
				})
			projectUnmanagedRct = BuildResourceClaimTemplate(
				"rct-project-unmanaged", "random-ns", nil)
		)

		shippedExistingNames := []string{
			existingProjectRct.Name,
			existingClusterRct.Name,
			existingDepartmentRct.Name,
		}

		BeforeEach(func() {
			resetCalls()

			server.RouteToHandler("POST", "/api/v1/resource-claim-templates", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.ResourceClaimTemplate
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				Expect(request.URL.Query().Get("clusterId")).To(Equal(ClusterIdForTests))
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				callsMu.Lock()
				createdCalls = append(createdCalls, payload)
				callsMu.Unlock()
			})

			server.RouteToHandler("PUT", "/api/v1/resource-claim-templates", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.ResourceClaimTemplate
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				Expect(request.URL.Query().Get("clusterId")).To(Equal(ClusterIdForTests))
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				callsMu.Lock()
				updatedCalls = append(updatedCalls, payload)
				callsMu.Unlock()
			})

			server.RouteToHandler("DELETE", "/api/v1/resource-claim-templates", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusNoContent, "")(writer, request)
				q := request.URL.Query()

				callsMu.Lock()
				deletedCalls = append(deletedCalls, deletedRctCall{
					clusterId: q.Get("clusterId"),
					scopeType: q.Get("scopeType"),
					scopeId:   q.Get("scopeId"),
					name:      q.Get("name"),
				})
				callsMu.Unlock()
			})

			server.RouteToHandler("POST", "/api/v1/resource-claim-templates/sync", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.ResourceClaimTemplates
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				Expect(request.URL.Query().Get("clusterId")).To(Equal(ClusterIdForTests))
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				callsMu.Lock()
				syncBodies = append(syncBodies, payload)
				callsMu.Unlock()
			})

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().
				WithScheme(GetTestScheme()).
				WithObjects(
					runaiManagedNs, kaiProject, legacyProject,
					existingProjectRct, existingClusterRct, existingDepartmentRct,
					clusterReplicaRct, projectUnmanagedRct,
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

		It("ships the full inventory (and only the full inventory) through startup /sync", func() {
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			// No replica may appear in any sync body - even intermediate ones.
			callsMu.Lock()
			bodies := append([]trackerclient.ResourceClaimTemplates{}, syncBodies...)
			callsMu.Unlock()
			for _, body := range bodies {
				for _, item := range body.Items {
					Expect(item.Name).ToNot(BeElementOf(
						clusterReplicaRct.Name,
						projectUnmanagedRct.Name,
					), "replica leaked into a /sync body")
				}
			}
		})

		It("POSTs a newly-added project RCT per-item, with the resolved project id", func() {
			// Wait for startup /sync so the per-item filter guard
			// (IsRunaiBackendAvailable) has flipped true.
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			added := BuildResourceClaimTemplate("rct-project-new", projectWithIdNamespace, nil)
			Expect(fakeCachedClient.Create(context.TODO(), added)).To(Succeed())

			Eventually(createdNames).WithTimeout(threeSecondTimeout).Should(ContainElement(added.Name))

			captured, ok := createdByName(added.Name)
			Expect(ok).To(BeTrue())
			Expect(captured.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_PROJECT)))
			Expect(captured.ScopeId).To(Equal(projectIdValue))
			Expect(captured.Namespace).To(Equal(projectWithIdNamespace))
		})

		It("POSTs a newly-added cluster-scope RCT per-item with the cluster uuid as scopeId", func() {
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			added := BuildResourceClaimTemplate("rct-cluster-new", runaijobv1.DefaultNamespace, map[string]string{
				commonconstants.ClusterWideResourceLabel: "true",
			})
			Expect(fakeCachedClient.Create(context.TODO(), added)).To(Succeed())

			Eventually(createdNames).WithTimeout(threeSecondTimeout).Should(ContainElement(added.Name))

			captured, ok := createdByName(added.Name)
			Expect(ok).To(BeTrue())
			Expect(captured.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_CLUSTER)))
			Expect(captured.ScopeId).To(Equal(ClusterIdForTests))
		})

		It("does NOT POST a newly-added replica RCT (cluster-scope label in a project namespace)", func() {
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			// Capture the names seen so far, then add a replica and
			// make sure the per-item POST stream never grows past it.
			replica := BuildResourceClaimTemplate("rct-cluster-replica-new", projectWithIdNamespace, map[string]string{
				commonconstants.ClusterWideResourceLabel: "true",
			})
			Expect(fakeCachedClient.Create(context.TODO(), replica)).To(Succeed())

			// Also add a non-replica so we have something to wait on;
			// if the replica had leaked, it would show up alongside.
			bystander := BuildResourceClaimTemplate("rct-bystander", runaijobv1.DefaultNamespace, map[string]string{
				commonconstants.ClusterWideResourceLabel: "true",
			})
			Expect(fakeCachedClient.Create(context.TODO(), bystander)).To(Succeed())

			Eventually(createdNames).WithTimeout(threeSecondTimeout).Should(ContainElement(bystander.Name))
			Expect(createdNames()).ToNot(ContainElement(replica.Name))
		})

		It("PUTs a mutated RCT per-item via the rate-limited queue", func() {
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			current := &resourcev1.ResourceClaimTemplate{}
			Expect(fakeCachedClient.Get(context.TODO(),
				types.NamespacedName{Name: existingProjectRct.Name, Namespace: projectWithIdNamespace},
				current)).To(Succeed())
			if current.Annotations == nil {
				current.Annotations = map[string]string{}
			}
			current.Annotations["edited"] = "true"
			Expect(fakeCachedClient.Update(context.TODO(), current)).To(Succeed())

			Eventually(updatedNames).WithTimeout(threeSecondTimeout).Should(ContainElement(existingProjectRct.Name))
		})

		It("DELETEs an RCT per-item with its full scope identity", func() {
			Eventually(latestSyncNames).WithTimeout(threeSecondTimeout).Should(ConsistOf(shippedExistingNames))

			toDelete := &resourcev1.ResourceClaimTemplate{}
			Expect(fakeCachedClient.Get(context.TODO(),
				types.NamespacedName{Name: existingClusterRct.Name, Namespace: runaijobv1.DefaultNamespace},
				toDelete)).To(Succeed())
			Expect(fakeCachedClient.Delete(context.TODO(), toDelete)).To(Succeed())

			// Cluster-scope delete carries scopeType=cluster and
			// scopeId=<cluster uuid>, matching the identity the
			// parser stamps on cluster-scoped create requests.
			Eventually(currentDeletedCalls).WithTimeout(threeSecondTimeout).Should(ContainElement(deletedRctCall{
				clusterId: ClusterIdForTests,
				scopeType: string(auth_client.SCOPETYPE_CLUSTER),
				scopeId:   ClusterIdForTests,
				name:      existingClusterRct.Name,
			}))
		})
	})
})

type deletedRctCall struct {
	clusterId string
	scopeType string
	scopeId   string
	name      string
}
