/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node labels", func() {
	const tenantID = int32(19451)
	var ctx context.Context
	var clusterID string
	var nodeService *services.NodeService

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "19451")
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{"19451"}})
		cluster, err := dao.CreateCluster(ctx, &models.Cluster{
			Name: clusterNameGenerator.Generate("node-labels"), TenantId: tenantID,
			Domain: "node-labels.example.com", Version: "3.1",
		})
		Expect(err).To(BeNil())
		clusterID = cluster.ID.String()
		DeferCleanup(func() {
			nodesErr := dao.DeleteClusterNodes(ctx, tenantID, clusterID)
			clusterErr := dao.DeleteCluster(ctx, clusterID, tenantID)
			Expect(nodesErr).To(BeNil())
			Expect(clusterErr).To(BeNil())
		})
		nodeService = services.NewNodeService(nil, nil)
	})

	nodeInfoWithLabels := func(labels map[string]string) api.NodeInfo {
		return api.NodeInfo{Status: "Ready", NodePool: "default", Labels: labels}
	}
	saveNodeLabels := func(labels map[string]string) {
		GinkgoHelper()
		response, err := nodeService.UpsertNode(ctx, clusterID, "node-1", nodeInfoWithLabels(labels))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	}
	syncNodes := func(nodes ...api.NodeForSync) {
		GinkgoHelper()
		response, err := nodeService.SyncNodes(ctx, clusterID, api.NodesSyncRequest{NodesForSync: nodes})
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	}
	expectStoredLabels := func(expectedLabelsByNode map[string]map[string]string) {
		GinkgoHelper()
		permittedScopes := authorization_api.PermittedScopes{Clusters: []string{clusterID}}
		nodes, dbErr := dao.GetNodesByClusterUuid(ctx, tenantID, clusterID, permittedScopes, nil, "", "", 0, 0, "")
		Expect(dbErr).To(BeNil())
		actualLabelsByNode := make(map[string]map[string]string, len(*nodes))
		for _, node := range *nodes {
			actualLabelsByNode[node.Name] = node.Labels
		}
		Expect(actualLabelsByNode).To(Equal(expectedLabelsByNode))
	}

	It("stores every label received from node sync", func() {
		labels := map[string]string{
			"gpu-type":                       "a100",
			"example.com/gpu_model":          "A100-SXM4.80GB",
			"node-role.kubernetes.io/worker": "",
		}
		saveNodeLabels(labels)
		expectStoredLabels(map[string]map[string]string{"node-1": labels})
	})

	DescribeTable("updating labels for one node",
		func(updatedLabels map[string]string) {
			By("Saving the node's original labels")
			saveNodeLabels(map[string]string{"gpu-type": "a100", "removed-label": "old"})

			By("Sending a node update containing its complete new label set")
			saveNodeLabels(updatedLabels)

			By("Checking that only the new labels are stored, with no old labels left behind")
			expectStoredLabels(map[string]map[string]string{"node-1": updatedLabels})
		},
		Entry("replaces old labels with the new label set", map[string]string{"gpu-type": "h100", "new-label": "value"}),
		Entry("clears labels when the new label set is empty", map[string]string{}),
	)

	It("replaces labels for each node when multiple nodes are sent in one sync request", func() {
		node1Labels := map[string]string{"gpu-type": "a100", "removed-label": "old"}
		node2Labels := map[string]string{"gpu-type": "h100"}

		By("Sending two nodes with different labels in one sync request")
		syncNodes(
			api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(node1Labels)},
			api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(node2Labels)},
		)

		By("Checking that both nodes were saved with their own labels")
		expectStoredLabels(map[string]map[string]string{"node-1": node1Labels, "node-2": node2Labels})

		By("Sending another sync request that changes node-1's labels and clears node-2's labels")
		updatedNode1Labels := map[string]string{"gpu-type": "l40", "new-label": "value"}
		syncNodes(
			api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(updatedNode1Labels)},
			api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(map[string]string{})},
		)

		By("Checking that node-1 has only its new labels and node-2 has no labels")
		expectStoredLabels(map[string]map[string]string{"node-1": updatedNode1Labels, "node-2": {}})
	})

	It("accepts a new node from an older sender that omits labels", func() {
		By("Creating a node without the optional labels field")
		saveNodeLabels(nil)
		expectStoredLabels(map[string]map[string]string{"node-1": nil})
	})

	It("clears stored labels when a node update omits the labels field", func() {
		By("Saving labels before receiving an update without labels")
		saveNodeLabels(map[string]string{"gpu-type": "a100"})
		saveNodeLabels(nil)

		By("Checking that the old labels are no longer stored")
		expectStoredLabels(map[string]map[string]string{"node-1": nil})
	})

	It("clears stored labels when a full sync omits the labels field", func() {
		By("Saving labels before receiving a full sync without labels")
		saveNodeLabels(map[string]string{"gpu-type": "a100"})
		syncNodes(api.NodeForSync{
			Name:     "node-1",
			NodeInfo: api.NodeInfo{Status: "Ready", NodePool: "default"},
		})
		expectStoredLabels(map[string]map[string]string{"node-1": nil})
	})

	Context("suggesting labels for node pool creation", func() {
		BeforeEach(func() {
			By("Storing the same key on several nodes with different and duplicate values")
			syncNodes(
				api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(map[string]string{"a": "b", "b": "d", "worker": ""})},
				api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(map[string]string{"a": "c"})},
				api.NodeForSync{Name: "node-3", NodeInfo: nodeInfoWithLabels(map[string]string{"a": "ba"})},
				api.NodeForSync{Name: "node-4", NodeInfo: nodeInfoWithLabels(map[string]string{"a": "b"})},
				api.NodeForSync{Name: "node-5", NodeInfo: nodeInfoWithLabels(nil)},
			)
		})

		requestLabelSuggestions := func(path string) *httptest.ResponseRecorder {
			controller := api.NewNodesAPIController(nodeService)
			routes := controller.Routes()
			router := chi.NewRouter()
			router.Get(routes["GetNodeLabelKeys"].Pattern, controller.GetNodeLabelKeys)
			router.Get(routes["GetNodeLabelValues"].Pattern, controller.GetNodeLabelValues)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/clusters/"+clusterID+"/nodes/labels/"+path, nil).WithContext(ctx)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			return response
		}

		DescribeTable("checks node read permissions through the generated route",
			func(path string, authorizationStatus int, grantsClusterAccess bool, expectedStatus int) {
				authorizationRequests := make(chan authorization_api.BackendEvaluationRequest, 1)
				authorizationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()
					Expect(r.Method).To(Equal(http.MethodPost))
					Expect(r.URL.Path).To(Equal("/api/v1/authorization/evaluate"))
					var evaluation authorization_api.BackendEvaluationRequest
					Expect(json.NewDecoder(r.Body).Decode(&evaluation)).To(Succeed())
					authorizationRequests <- evaluation
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(authorizationStatus)
					if authorizationStatus == http.StatusOK {
						grantedCluster := "another-cluster"
						if grantsClusterAccess {
							grantedCluster = clusterID
						}
						Expect(json.NewEncoder(w).Encode(authorization_api.PermittedScopes{
							Clusters: []string{grantedCluster},
						})).To(Succeed())
					}
				}))
				DeferCleanup(authorizationServer.Close)

				originalConfig := *runai_config.Get[*authorization_client.Config]()
				runai_config.Set(&authorization_client.Config{Url: authorizationServer.URL})
				DeferCleanup(func() { runai_config.Set(&originalConfig) })

				By("Calling the generated route including its authorization middleware")
				router := api.NewRouter(api.NewNodesAPIController(nodeService))
				request := httptest.NewRequest(http.MethodGet,
					"/api/v1/clusters/"+clusterID+"/nodes/labels/"+path, nil).WithContext(ctx)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)

				Expect(response.Code).To(Equal(expectedStatus))
				var evaluation authorization_api.BackendEvaluationRequest
				Eventually(authorizationRequests).Should(Receive(&evaluation))
				Expect(string(evaluation.GetResourceType())).To(Equal("nodes"))
				Expect(string(evaluation.GetRequestAction())).To(Equal("read"))
				if expectedStatus == http.StatusOK {
					expectedOptions := `["a", "b", "worker"]`
					if path == "values?key=a" {
						expectedOptions = `["b", "ba", "c"]`
					}
					Expect(response.Body.String()).To(MatchJSON(expectedOptions))
				}
			},
			Entry("returns keys when authorization grants cluster access", "keys", http.StatusOK, true, http.StatusOK),
			Entry("returns values when authorization grants cluster access", "values?key=a", http.StatusOK, true, http.StatusOK),
			Entry("rejects keys when authorization denies access", "keys", http.StatusForbidden, false, http.StatusForbidden),
			Entry("rejects values when authorization denies access", "values?key=a", http.StatusForbidden, false, http.StatusForbidden),
			Entry("rejects keys when authorization grants a different cluster", "keys", http.StatusOK, false, http.StatusForbidden),
			Entry("rejects values when authorization grants a different cluster", "values?key=a", http.StatusOK, false, http.StatusForbidden),
			Entry("returns an error for keys when authorization is unavailable", "keys", http.StatusServiceUnavailable, false, http.StatusServiceUnavailable),
			Entry("returns an error for values when authorization is unavailable", "values?key=a", http.StatusServiceUnavailable, false, http.StatusServiceUnavailable),
		)

		DescribeTable("returns unique, sorted label keys",
			func(contains string, expected []string) {
				By("Requesting label keys using the typed text")
				response, err := nodeService.GetNodeLabelKeys(ctx, clusterID, contains, 0, 50)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body.([]string)).To(Equal(expected))
			},
			Entry("lists each key once", "", []string{"a", "b", "worker"}),
			Entry("filters keys by the typed text", "wo", []string{"worker"}),
			Entry("matches text in the middle of a key", "ork", []string{"worker"}),
			Entry("matches text at the end of a key", "ker", []string{"worker"}),
			Entry("matches keys case-sensitively", "WORK", []string{}),
		)

		DescribeTable("returns unique, sorted values for an exact label key",
			func(key, contains string, expected []string) {
				By("Requesting values for the selected key using the typed text")
				response, err := nodeService.GetNodeLabelValues(ctx, clusterID, key, contains, 0, 50)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body.([]string)).To(Equal(expected))
			},
			Entry("returns only values belonging to key a", "a", "", []string{"b", "ba", "c"}),
			Entry("offers b and ba when typing b for key a", "a", "b", []string{"b", "ba"}),
			Entry("matches text at the end of a value", "a", "a", []string{"ba"}),
			Entry("returns d only for key b", "b", "", []string{"d"}),
			Entry("omits empty values that cannot be used for a node pool", "worker", "", []string{}),
			Entry("returns no values for an unknown key", "missing", "", []string{}),
			Entry("matches substrings case-sensitively", "a", "B", []string{}),
			Entry("treats percent as literal text", "a", "%", []string{}),
			Entry("treats SQL text in a key as literal text", "a' OR true --", "", []string{}),
		)

		It("rejects a malformed cluster UUID in both dropdown APIs", func() {
			originalClusterID := clusterID
			clusterID = "not-a-uuid"
			defer func() { clusterID = originalClusterID }()

			for _, path := range []string{"keys", "values?key=a"} {
				By("Requesting " + path + " with a malformed cluster UUID")
				response := requestLabelSuggestions(path)
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			}
		})

		It("accepts uppercase cluster UUIDs with cluster permissions", func() {
			originalClusterID := clusterID
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Clusters: []string{originalClusterID}})
			clusterID = strings.ToUpper(originalClusterID)
			defer func() { clusterID = originalClusterID }()

			for path, expected := range map[string]string{
				"keys":         `["a", "b", "worker"]`,
				"values?key=a": `["b", "ba", "c"]`,
			} {
				By("Requesting " + path + " with an uppercase UUID and permission for the same cluster")
				response := requestLabelSuggestions(path)
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body.String()).To(MatchJSON(expected))
			}
		})

		DescribeTable("validates label suggestion requests and pagination",
			func(path string, expectedStatus int, expected []string) {
				By("Sending an HTTP request to the generated controller with the documented route")
				response := requestLabelSuggestions(path)
				Expect(response.Code).To(Equal(expectedStatus))
				if expectedStatus == http.StatusOK {
					var suggestions []string
					Expect(json.Unmarshal(response.Body.Bytes(), &suggestions)).To(Succeed())
					Expect(suggestions).To(Equal(expected))
				}
			},
			Entry("keys endpoint matches text inside a key", "keys?contains=ork", http.StatusOK, []string{"worker"}),
			Entry("values endpoint filters values for the selected key", "values?key=a&contains=a", http.StatusOK, []string{"ba"}),
			Entry("keys endpoint skips already returned keys", "keys?offset=1&limit=1", http.StatusOK, []string{"b"}),
			Entry("values endpoint paginates after substring filtering and duplicate removal", "values?key=a&contains=b&offset=1&limit=1", http.StatusOK, []string{"ba"}),
			Entry("returns an empty page past the final value", "values?key=a&offset=10&limit=1", http.StatusOK, []string{}),
			Entry("keys endpoint rejects more than 500 items", "keys?limit=501", http.StatusBadRequest, []string(nil)),
			Entry("values endpoint rejects more than 500 items", "values?key=a&limit=501", http.StatusBadRequest, []string(nil)),
			Entry("keys endpoint rejects a negative offset", "keys?offset=-1", http.StatusBadRequest, []string(nil)),
			Entry("values endpoint rejects a negative offset", "values?key=a&offset=-1", http.StatusBadRequest, []string(nil)),
			Entry("rejects a zero page size", "keys?limit=0", http.StatusBadRequest, []string(nil)),
			Entry("rejects a negative page size", "values?key=a&limit=-1", http.StatusBadRequest, []string(nil)),
			Entry("values endpoint requires a key", "values", http.StatusBadRequest, []string(nil)),
			Entry("values endpoint rejects an empty key", "values?key=", http.StatusBadRequest, []string(nil)),
		)

		DescribeTable("applies the shared default and maximum page sizes",
			func(path string) {
				By("Storing 501 different label keys and values")
				nodes := make([]api.NodeForSync, 501)
				expected := make([]string, len(nodes))
				for i := range nodes {
					label := fmt.Sprintf("item-%03d", i)
					expected[i] = label
					nodes[i] = api.NodeForSync{
						Name: label, NodeInfo: nodeInfoWithLabels(map[string]string{label: "unused", "chosen-key": label}),
					}
				}
				syncNodes(nodes...)

				for _, page := range []struct {
					name, query string
					start, end  int
				}{
					{"Returns the first 50 items when limit is omitted", "", 0, 50},
					{"Allows a page of 500 items", "&limit=500", 0, 500},
					{"Uses offset to return the remaining item", "&offset=500&limit=50", 500, 501},
				} {
					By(page.name)
					response := requestLabelSuggestions(path + page.query)
					Expect(response.Code).To(Equal(http.StatusOK))
					var suggestions []string
					Expect(json.Unmarshal(response.Body.Bytes(), &suggestions)).To(Succeed())
					Expect(suggestions).To(Equal(expected[page.start:page.end]))
				}
			},
			Entry("label keys", "keys?contains=item-"),
			Entry("label values", "values?key=chosen-key&contains=item-"),
		)

		It("paginates after excluding empty label values", func() {
			By("Saving one empty value and two valid values for the same key")
			syncNodes(
				api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(map[string]string{"role": ""})},
				api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(map[string]string{"role": "compute"})},
				api.NodeForSync{Name: "node-3", NodeInfo: nodeInfoWithLabels(map[string]string{"role": "worker"})},
			)

			By("Skipping the first valid value and checking that the second is returned")
			response := requestLabelSuggestions("values?key=role&offset=1&limit=1")
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(MatchJSON(`["worker"]`))
		})

		It("omits keys longer than node pool creation allows before pagination", func() {
			By("Saving valid Kubernetes keys at 253 and 254 characters")
			domain := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." +
				strings.Repeat("c", 63) + "." + strings.Repeat("d", 56)
			allowedKey := domain + "/name"
			tooLongKey := domain + "/namex"
			labels := map[string]string{allowedKey: "allowed", tooLongKey: "too-long", "z": "last"}
			syncNodes(api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(labels)})

			By("Keeping the original labels in storage")
			expectStoredLabels(map[string]map[string]string{"node-1": labels})

			By("Offering the longest allowed key but omitting the longer key")
			response := requestLabelSuggestions("keys")
			Expect(response.Code).To(Equal(http.StatusOK))
			var keys []string
			Expect(json.Unmarshal(response.Body.Bytes(), &keys)).To(Succeed())
			Expect(keys).To(Equal([]string{allowedKey, "z"}))

			By("Skipping the first eligible key when loading the next page")
			response = requestLabelSuggestions("keys?offset=1&limit=1")
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(MatchJSON(`["z"]`))

			By("Returning no value suggestions for a key that cannot be used to create a node pool")
			response = requestLabelSuggestions("values?key=" + tooLongKey)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(MatchJSON(`[]`))
		})

		It("matches the selected label key exactly", func() {
			By("Storing two label keys that share the same starting text but have different values")
			syncNodes(api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(map[string]string{"gpu": "a100", "gpu-type": "h100"})})
			response, err := nodeService.GetNodeLabelValues(ctx, clusterID, "gpu", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.([]string)).To(Equal([]string{"a100"}))
		})

		DescribeTable("returns empty arrays when no node labels are available",
			func(nodes []api.NodeForSync) {
				By("Replacing the cluster's nodes with the supplied snapshot")
				syncNodes(nodes...)
				for _, path := range []string{"keys", "values?key=a"} {
					response := requestLabelSuggestions(path)
					Expect(response.Code).To(Equal(http.StatusOK))
					Expect(response.Body.String()).To(MatchJSON(`[]`))
				}
			},
			Entry("the cluster has no nodes", []api.NodeForSync{}),
			Entry("the nodes have no labels", []api.NodeForSync{{Name: "node-1", NodeInfo: api.NodeInfo{Status: "Ready", NodePool: "default"}}}),
		)

		It("treats underscores in search text as literal characters", func() {
			By("Storing labels that would both match if underscore were treated as a SQL wildcard")
			syncNodes(
				api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(map[string]string{"gpu_type": "b_a"})},
				api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(map[string]string{"gpu-type": "bxa", "gpu_type": "bxa"})},
			)
			response, err := nodeService.GetNodeLabelKeys(ctx, clusterID, "gpu_", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{"gpu_type"}))

			By("Searching inside values without treating underscore as a wildcard")
			response, err = nodeService.GetNodeLabelValues(ctx, clusterID, "gpu_type", "_", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{"b_a"}))
		})

		It("updates suggestions when node labels change or are removed", func() {
			By("Replacing the cluster's nodes with a node carrying only the new label value")
			syncNodes(api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(map[string]string{"a": "new"})})

			By("Checking that old values are absent from the dropdown")
			response, err := nodeService.GetNodeLabelValues(ctx, clusterID, "a", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{"new"}))

			By("Removing the remaining node's labels and checking that both dropdowns are empty")
			saveNodeLabels(map[string]string{})
			for _, path := range []string{"keys", "values?key=a"} {
				response := requestLabelSuggestions(path)
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body.String()).To(MatchJSON(`[]`))
			}
		})

		It("does not suggest labels from another cluster", func() {
			By("Storing a label on a node in another cluster belonging to the same tenant")
			otherCluster, clusterErr := dao.CreateCluster(ctx, &models.Cluster{
				Name: clusterNameGenerator.Generate("label-suggestions"), TenantId: tenantID,
				Domain: "label-suggestions.example.com", Version: "3.1",
			})
			Expect(clusterErr).NotTo(HaveOccurred())
			otherClusterID := otherCluster.ID.String()
			DeferCleanup(func() {
				Expect(dao.DeleteClusterNodes(ctx, tenantID, otherClusterID)).To(Succeed())
				Expect(dao.DeleteCluster(ctx, otherClusterID, tenantID)).To(Succeed())
			})
			update, err := nodeService.UpsertNode(ctx, otherClusterID, "other-node", nodeInfoWithLabels(map[string]string{"other-cluster": "hidden"}))
			Expect(err).NotTo(HaveOccurred())
			Expect(update.Code).To(Equal(http.StatusNoContent))

			By("Checking that the selected cluster returns only its own keys")
			response, err := nodeService.GetNodeLabelKeys(ctx, clusterID, "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{"a", "b", "worker"}))
			response, err = nodeService.GetNodeLabelValues(ctx, clusterID, "other-cluster", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{}))
		})

		It("does not expose another tenant's labels", func() {
			By("Requesting the original cluster as a different tenant")
			otherTenantContext := context.WithValue(ctx, runaictx.CfTenantId, "19452")
			otherTenantContext = context.WithValue(otherTenantContext, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{"19452"}})
			response, err := nodeService.GetNodeLabelKeys(otherTenantContext, clusterID, "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.([]string)).To(Equal([]string{}))
			response, err = nodeService.GetNodeLabelValues(otherTenantContext, clusterID, "a", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.([]string)).To(Equal([]string{}))
		})

		DescribeTable("allows label suggestions with cluster or system permissions",
			func(grantedScopes func() *authorization_api.PermittedScopes) {
				By("Granting access at the selected cluster or system scope")
				scopedContext := context.WithValue(ctx, authorization_utils.PermittedScopesKey, grantedScopes())
				keys, err := nodeService.GetNodeLabelKeys(scopedContext, clusterID, "", 0, 50)
				Expect(err).NotTo(HaveOccurred())
				Expect(keys.Code).To(Equal(http.StatusOK))
				Expect(keys.Body.([]string)).To(Equal([]string{"a", "b", "worker"}))

				values, err := nodeService.GetNodeLabelValues(scopedContext, clusterID, "a", "", 0, 50)
				Expect(err).NotTo(HaveOccurred())
				Expect(values.Code).To(Equal(http.StatusOK))
				Expect(values.Body.([]string)).To(Equal([]string{"b", "ba", "c"}))
			},
			Entry("cluster permissions", func() *authorization_api.PermittedScopes {
				return &authorization_api.PermittedScopes{Clusters: []string{clusterID}}
			}),
			Entry("system permissions", func() *authorization_api.PermittedScopes {
				allowed := true
				return &authorization_api.PermittedScopes{System: &allowed}
			}),
		)

		It("rejects requests without granted permissions", func() {
			By("Making the request with a tenant identity but no authorization scopes")
			unauthorizedContext := context.WithValue(context.Background(), runaictx.CfTenantId, "19451")
			keys, err := nodeService.GetNodeLabelKeys(unauthorizedContext, clusterID, "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(keys.Code).To(Equal(http.StatusForbidden))
			values, err := nodeService.GetNodeLabelValues(unauthorizedContext, clusterID, "a", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(values.Code).To(Equal(http.StatusForbidden))
		})

		It("rejects requests outside the permitted cluster", func() {
			By("Granting access only to a different cluster")
			restrictedContext := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{"another-cluster"}})
			response, err := nodeService.GetNodeLabelKeys(restrictedContext, clusterID, "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusForbidden))
			response, err = nodeService.GetNodeLabelValues(restrictedContext, clusterID, "a", "", 0, 50)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

})
