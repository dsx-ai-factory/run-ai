/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	testHost             = "http://localhost"
	urlWorkloads         = "/api/v2/workloads"
	urlInternalWorkloads = "/internal/v2/workloads"
	batchUrl             = "/internal/workloads/batch"
)

var sampleGvk = api.Gvk{
	Group:   convutils.Ref("apps"),
	Version: "v1",
	Kind:    "Deployment",
}

var sampleManifestSpec = `{
	"apiVersion": "apps/v1",
	"kind": "Deployment",
	"metadata": {
		"name": "my-app"
	},
	"spec": {
		"replicas": 1,
		"selector": {
			"matchLabels": {
				"app": "my-app"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "my-app"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "app",
						"image": "nginx:latest"
					}
				]
			}
		}
	}
}
`

var prioPremCombinations = []PriorityPreemptCombo{
	{
		Priority:       "high",
		Preemptibility: nil,
		Expected:       api.PREEMPTIBILITY_NON_PREEMPTIBLE,
	},
	{
		Priority:       "high",
		Preemptibility: convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE),
		Expected:       api.PREEMPTIBILITY_PREEMPTIBLE,
	},
	{
		Priority:       "low",
		Preemptibility: nil,
		Expected:       api.PREEMPTIBILITY_PREEMPTIBLE,
	},
	{
		Priority:       "low",
		Preemptibility: convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE),
		Expected:       api.PREEMPTIBILITY_NON_PREEMPTIBLE,
	},
}

var defaultPrioritiesAndCategories = workloadsApi.GetAvailablePrioritiesAndCategories200Response{
	Priorities: []workloadsApi.ClusterPriority{
		{Name: "high", Value: 120},
		{Name: "medium", Value: 80},
		{Name: "low", Value: 40},
		{Name: "default-priority", Value: 0},
	},
	Categories: []workloadsApi.WorkloadCategory{
		{Name: "train"},
		{Name: "workspace"},
		{Name: "inference"},
		{Name: "default-category"},
	},
}

type PriorityPreemptCombo struct {
	Priority       string
	Preemptibility *api.Preemptibility
	Expected       api.Preemptibility
}

func (c PriorityPreemptCombo) ToString() string {
	result := "priority: " + c.Priority + " + preemptibility: "

	if c.Preemptibility != nil {
		result += string(*c.Preemptibility)
	} else {
		result += "unspecified"
	}

	result += " -> expecting " + string(c.Expected)

	return result
}

var _ = Describe("WorkloadsV2Impl", func() {
	var (
		router chi.Router

		randoms    *testutils.RandomValues
		oldCluster *testutils.RandomValues

		defaultWorkloadsMock *workloads.SimpleMock

		sampleManifest map[string]interface{}
		csclient       *clusterservice.SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		oldCluster = randoms.CloneWithDifferentCluster()

		jsonErr := json.Unmarshal([]byte(sampleManifestSpec), &sampleManifest)
		Expect(jsonErr).To(BeNil())

		k8sUtil := NewK8sResourceUtil()
		impl = &WorkloadsV2Impl{
			k8sUtil: k8sUtil,
		}

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		org_unit.SetMock(org_unit.NewSimpleMock().
			WithRandoms(randoms).
			WithRandoms(oldCluster))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)).
			WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, "2.22.0")))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		setting_pkg.Init(nil, nil)
		setting_pkg.Register[*settings.YamlSubmission]()
		// Setup default workloads client mock for existing tests
		defaultWorkloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
			WithPrioritiesAndCategories(oldCluster.ClusterId, defaultPrioritiesAndCategories)

		// Add default workload types for common test cases
		readyPhase := "Ready"
		jobWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Job",
			Group:      "batch",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		podWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Pod",
			Group:      "",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}
		deploymentWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "Deployment",
			Group:      "apps",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
				{
					ClusterId: convutils.Ref(oldCluster.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}

		defaultWorkloadsMock.
			WithWorkloadType(jobWorkloadType).
			WithWorkloadType(podWorkloadType).
			WithWorkloadType(deploymentWorkloadType)
		workloads.SetMock(defaultWorkloadsMock)

		policies.SetMock(policies.SimpleMockClientFactory{})

		csclient = clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		dao.RestartUpdateCountForTest()

		router = createRouterForTest()
	})

	AfterEach(func() {
		cleanTables()
	})

	Context("workload action history", func() {
		It("records create, update, and delete once", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			request := &api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "action-history", ProjectId: randoms.ProjectIdStr,
					Priority: convutils.Ref("high"), Category: convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, request)
			updateWorkloadForTest(router, randoms, id, &api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{Priority: convutils.Ref("high"), Category: convutils.Ref("train")},
				Manifest: sampleManifest,
			})
			deleteWorkloadForTest(router, randoms, id)

			actions := recorder.recorded()
			Expect(actions).To(HaveLen(3))
			Expect([]workloadsApi.WorkloadActionType{actions[0].Action, actions[1].Action, actions[2].Action}).To(Equal([]workloadsApi.WorkloadActionType{
				workloadsApi.WORKLOADACTIONTYPE_CREATE,
				workloadsApi.WORKLOADACTIONTYPE_UPDATE,
				workloadsApi.WORKLOADACTIONTYPE_DELETE,
			}))
			for _, action := range actions {
				Expect(action.WorkloadID).To(Equal(id))
				Expect(action.WorkloadName).To(Equal("action-history"))
				Expect(action.ClusterID).To(Equal(randoms.ClusterId))
				Expect(action.AcceptedAt).NotTo(BeZero())
			}
		})

		It("skips dry-run creates", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			_, submissionErrors := impl.CreateWorkloadV2WithDryRun(
				setContext(randoms.TenantIdStr, adminUser, nil),
				api.WorkloadV2CreateRequest{
					Metadata: api.WorkloadV2MetadataCreateParams{
						Name: "dry-run-action-history", ProjectId: randoms.ProjectIdStr,
						Priority: convutils.Ref("high"), Category: convutils.Ref("train"),
					},
					Manifest: sampleManifest,
				},
				CreateWorkloadV2Options{DryRun: true, RecordAction: recordAction},
			)
			Expect(submissionErrors).To(BeEmpty())
			Expect(recorder.recorded()).To(BeEmpty())
		})

		It("does not duplicate v1 actions handled through internal manager paths", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			id, _ := createInternalWorkloadForTest(router, randoms, &api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "v1-translated", ProjectId: randoms.ProjectIdStr,
					Priority: convutils.Ref("high"), Category: convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			})
			updateResponse := tryToUpdateInternalWorkloadForTest(router, randoms, id, &api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{Priority: convutils.Ref("high"), Category: convutils.Ref("train")},
				Manifest: sampleManifest,
			})
			Expect(updateResponse.Code).To(Equal(http.StatusAccepted), updateResponse.Body.String())
			ctx := setContext(randoms.TenantIdStr, adminUser, nil)
			response, apiErr := impl.ProcessBatch(ctx, api.WorkloadsV2Batch{Ids: []string{id}, Action: api.WORKLOADSV2BATCHACTION_DELETE})
			Expect(apiErr).To(BeNil())
			Expect(response.Succeeded).To(ConsistOf(id))
			Expect(recorder.recorded()).To(BeEmpty())
		})
	})

	Context("networking input validation", func() {
		for _, operation := range []string{"create", "dry-run", "update"} {
			Context(operation, func() {
				DescribeTable("rejects malformed networking before queueing",
					func(networking, message string) {
						method, path := http.MethodPost, urlWorkloads
						metadata := map[string]interface{}{"projectId": randoms.ProjectIdStr}
						var id string
						if operation == "dry-run" {
							path += "?dryRun=true"
						} else if operation == "update" {
							id, _ = createWorkloadForTest(router, randoms, &api.WorkloadV2CreateRequest{
								Metadata: api.WorkloadV2MetadataCreateParams{
									ProjectId: randoms.ProjectIdStr,
									Category:  convutils.Ref("default-category"),
									Priority:  convutils.Ref("default-priority"),
								},
								Manifest: sampleManifest,
							})
							method, path = http.MethodPut, urlWorkloads+"/"+id
							metadata = map[string]interface{}{}
						}
						ctx := setContext(randoms.TenantIdStr, adminUser, nil)
						before, err := dao.DbListUpdates(ctx, randoms.ClusterId, 0, 50)
						Expect(err).To(BeNil())
						// Keep JSON null entries intact through the actual HTTP decoder.
						metadata["configuration"] = map[string]interface{}{"networking": json.RawMessage(networking)}
						req := httptest.NewRequest(method, testHost+path, makePayload(map[string]interface{}{
							"metadata": metadata, "manifest": sampleManifest,
						}))
						req.Header = *headerForAdmin(randoms)
						resp := httptest.NewRecorder()
						router.ServeHTTP(resp, req)

						Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
						Expect(resp.Body.String()).To(ContainSubstring(message))
						after, err := dao.DbListUpdates(ctx, randoms.ClusterId, 0, 50)
						Expect(err).To(BeNil())
						Expect(after).To(Equal(before))
						if id != "" {
							Expect(getWorkloadForTest(router, randoms, id).Workload.Metadata.Configuration.Networking).To(BeNil())
						}
					},
					Entry("null endpoint", `[null]`, "networking[0] must contain a service or externalAccess"),
					Entry("null service", `[{"services":[null]}]`, "networking[0].services[0] must contain spec or ref"),
					Entry("service with both spec and ref", `[{"services":[{"spec":{"name":"one","selector":{"app":"one"},"ports":[{"port":80}]},"ref":{"name":"one","port":80}}]}]`, "networking[0].services[0] must not contain both spec and ref"),
					Entry("multiple services", `[{"services":[{"ref":{"name":"one","port":80}},{"ref":{"name":"two","port":80}}]}]`, "networking[0].services must contain at most one service"),
				)
			})
		}
	})

	Context("gvk policy is skipped for workloads translated from v1", func() {
		// A gvk policy that only allows image ubuntu/redhat. The sample
		// Deployment uses nginx:latest, so the policy fails whenever it runs.
		const failingImagePolicyJSON = `{
			"id": "translated-from-v1-skip-test",
			"effectiveRules": [
				{
					"field": "image",
					"specSelector": "allContainers",
					"enforcements": [ { "options": ["ubuntu", "redhat"] } ]
				}
			],
			"breakdown": {
				"fields": {
					"image": [
						{ "specSelectors": ["allContainers"], "specRef": "image", "formatter": { "builtin": "KeyValue" } }
					]
				},
				"specSelectors": {
					"allContainers": { "query": "(.spec.template.spec.containers[]?, .spec.template.spec.initContainers[]?) | select(.)" }
				}
			}
		}`

		BeforeEach(func() {
			gvkPolicy := &policyApi.GetGvkPolicyResp{}
			Expect(json.Unmarshal([]byte(failingImagePolicyJSON), gvkPolicy)).To(BeNil())

			policyMock := policies.SimpleMockClientFactory{}
			policyMock.WithGvkPolicy(
				randoms.ProjectIdStr,
				policyApi.GVK{Group: *policyApi.NewNullableString(convutils.Ref("apps")), Kind: "Deployment", Version: "v1"},
				gvkPolicy,
			)
			policies.SetMock(policyMock)
		})

		It("applies the policy when the workload is not translated from v1", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "not-translated-from-v1",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: sampleManifest, // nginx:latest violates the policy
			}

			resp := tryToCreateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest), "expected the policy to be applied: %s", resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("field image failed options validation"))
		})

		It("skips the policy when the workload is created via internal endpoint", func() {
			var translatedManifest map[string]interface{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &translatedManifest)).To(BeNil())
			translatedManifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				ConvertedFromNativeLabel: "true",
			}

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "translated-from-v1",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: translatedManifest, // same violating image, but policy is skipped
			}

			// Creation succeeds via internal endpoint even though the image violates the policy,
			// because internal endpoint sets TranslatedFromV1=true which makes ApplyPolicy skip it.
			id, created := createInternalWorkloadForTest(router, randoms, &request)
			Expect(id).ToNot(BeEmpty())
			Expect(created).ToNot(BeNil())

			// The persisted record reflects that it was translated from v1.
			ctx := setContext(randoms.TenantIdStr, adminUser, nil)
			stored, getErr := dao.DbGetWorkload(ctx, id, nil)
			Expect(getErr).To(BeNil())
			Expect(stored.TranslatedFromV1).To(BeTrue())
		})

		It("applies the policy when client forges the converted-from-native label on public endpoint", func() {
			var forgedManifest map[string]interface{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &forgedManifest)).To(BeNil())
			forgedManifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
				ConvertedFromNativeLabel: "true",
			}

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "forged-label-attempt",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: forgedManifest, // nginx:latest violates the policy
			}

			// Creation fails because the public endpoint strips the forged label
			// and applies the policy normally.
			resp := tryToCreateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest), "expected the policy to be applied despite forged label: %s", resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("field image failed options validation"))
		})
	})

	It("deep copies networking config before persisting", func() {
		serviceType := api.SERVICETYPE_NODE_PORT
		protocol := api.SERVICEPROTOCOL_TCP
		targetPort := int32(8080)
		nodePort := int32(30080)
		portName := "http"
		selector := map[string]string{"app": "before"}
		authType := api.AUTHORIZATIONTYPE_AUTHORIZED_USERS
		forceJSON := true
		authorizedUsers := []string{"alice"}
		ingress := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "before",
			},
			"spec": map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{"host": "before.example.com"},
				},
			},
		}
		services := []api.NetworkingService{{
			Spec: &api.ServiceSpec{
				Name:     "svc",
				Type:     &serviceType,
				Selector: &selector,
				Ports: []api.ServicePort{{
					Name:       &portName,
					Port:       80,
					TargetPort: &targetPort,
					NodePort:   &nodePort,
					Protocol:   &protocol,
				}},
			},
			Authorization: &api.Workloadv2Authorization{
				AuthorizationType: &authType,
				AuthorizedUsers:   &authorizedUsers,
				ForceJsonResponse: &forceJSON,
			},
		}}
		networking := []api.NetworkingEndpoint{{
			ExternalAccess: &api.ExternalAccess{Ingress: &ingress},
			Services:       &services,
		}}

		cloned := cloneNetworkingConfig(&networking)

		(*networking[0].ExternalAccess.Ingress)["metadata"].(map[string]interface{})["name"] = "after"
		(*networking[0].ExternalAccess.Ingress)["spec"].(map[string]interface{})["rules"].([]interface{})[0].(map[string]interface{})["host"] = "after.example.com"
		(*(*networking[0].Services)[0].Spec.Selector)["app"] = "after"
		(*networking[0].Services)[0].Spec.Ports[0].Port = 443
		*(*networking[0].Services)[0].Spec.Ports[0].Name = "https"
		(*(*networking[0].Services)[0].Authorization.AuthorizedUsers)[0] = "bob"
		*(*networking[0].Services)[0].Authorization.ForceJsonResponse = false

		clonedEndpoint := (*cloned)[0]
		Expect((*clonedEndpoint.ExternalAccess.Ingress)["metadata"].(map[string]interface{})["name"]).To(Equal("before"))
		Expect((*clonedEndpoint.ExternalAccess.Ingress)["spec"].(map[string]interface{})["rules"].([]interface{})[0].(map[string]interface{})["host"]).To(Equal("before.example.com"))
		Expect(*(*clonedEndpoint.Services)[0].Spec.Selector).To(Equal(map[string]string{"app": "before"}))
		Expect((*clonedEndpoint.Services)[0].Spec.Ports[0].Port).To(Equal(int32(80)))
		Expect(*(*clonedEndpoint.Services)[0].Spec.Ports[0].Name).To(Equal("http"))
		Expect((*(*clonedEndpoint.Services)[0].Authorization.AuthorizedUsers)[0]).To(Equal("alice"))
		Expect(*(*clonedEndpoint.Services)[0].Authorization.ForceJsonResponse).To(BeTrue())
	})

	Context("workload lifecycle", func() {
		It("should populate all updates to the cluster and return the workload in get", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Mnnvl: convutils.Ref(api.MNNVLMODE_NONE),
					},
				},
				Manifest: sampleManifest,
			}

			expectedConfiguration := api.WorkloadV2Configuration{
				Mnnvl: convutils.Ref(api.MNNVLMODE_NONE),
			}
			expectedMeta := requestToResponseMeta(&toCreate.Metadata, randoms, adminUser, sampleGvk, &expectedConfiguration)

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			testutils.Expect(payload.Metadata).ToEqual(expectedMeta)

			// update should be marked as queue, because the agent did not saw it yet.
			getPayload := getWorkloadForTest(router, randoms, id)
			testutils.Expect(getPayload.Workload.Metadata).ToEqual(expectedMeta)
			testutils.Expect(getPayload.Workload.RawManifest).ToEqual(sampleManifest)
			Expect(len(getPayload.Workload.WorkloadManifest)).NotTo(Equal(0))
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_QUEUED))

			expectedUpdate := metaToUpdate(expectedMeta, id, int64(1), api.DESIREDPHASE_RUNNING, getPayload.Workload.WorkloadManifest)

			// expecting to get one update with the created payload. status should be applying
			// because the agent received this update and expected to process it
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(updates.SystemInfo.HighestUpdate).To(Equal(int64(1)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0]).To(Equal(expectedUpdate))

			getPayload = getWorkloadForTest(router, randoms, id)
			testutils.Expect(getPayload.Workload.Metadata).ToEqual(expectedMeta)
			testutils.Expect(getPayload.Workload.RawManifest).ToEqual(sampleManifest)
			Expect(len(getPayload.Workload.WorkloadManifest)).NotTo(Equal(0))
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))

			// re-run the updates request but this time send cluster report
			syncReport := api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 1,
					},
				},
			}

			updates = getUpdatesForTest(router, randoms, 0, 50, syncReport)
			Expect(updates.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0]).To(Equal(expectedUpdate))

			actualSyncReport := defaultWorkloadsMock.SyncIssuesByCluster[randoms.ClusterId]
			Expect(actualSyncReport).NotTo(Equal(nil))
			Expect(len(actualSyncReport.Updates)).To(Equal(1))
			Expect(*actualSyncReport.Updates[0].Id).To(Equal(id))

			// now the status changed to synced because agent reported processing of this workload
			getPayload = getWorkloadForTest(router, randoms, id)
			testutils.Expect(getPayload.Workload.Metadata).ToEqual(expectedMeta)
			testutils.Expect(getPayload.Workload.RawManifest).ToEqual(sampleManifest)
			Expect(len(getPayload.Workload.WorkloadManifest)).NotTo(Equal(0))
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_SYNCED))

			deletedPayload := deleteWorkloadForTest(router, randoms, id)
			testutils.Expect(deletedPayload.Metadata).ToEqual(expectedMeta)
			Expect(deletedPayload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))

			// cluster get the deletion update.
			expectedUpdate = metaToUpdate(expectedMeta, id, int64(2), api.DESIREDPHASE_DELETED, nil)

			updates = getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(2, 2)))
			Expect(updates.SystemInfo.HighestUpdate).To(Equal(int64(2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0]).To(Equal(expectedUpdate))

			getPayload = getWorkloadForTest(router, randoms, id)
			testutils.Expect(getPayload.Workload.Metadata).ToEqual(expectedMeta)
			testutils.Expect(getPayload.Workload.RawManifest).ToEqual(sampleManifest)
			Expect(len(getPayload.Workload.WorkloadManifest)).NotTo(Equal(0))
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_DELETING))

			// cluster reports about successfully deleting the records
			// record should be deleted and cluster receives no further updates
			updates = getUpdatesForTest(router, randoms, 0, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 2,
					},
				},
			})
			Expect(updates.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(0, 0)))
			Expect(updates.SystemInfo.HighestUpdate).To(Equal(int64(2)))
			Expect(len(updates.Workloads)).To(Equal(0))

			// make sure record is deleted
			verifyWorkloadNotFound(router, randoms, id)

			// expect one event have been sent to workloads
			notifications := defaultWorkloadsMock.ByWorkloadName[toCreate.Metadata.Name]
			Expect(len(notifications)).To(Equal(1))
			Expect(notifications[0].Phase).To(Equal(workloadsApi.PHASE_CREATING))
		})

		It("persists Configuration.Networking on create and returns it in updates", func() {
			workloadName := "networking-test"
			port := int32(8000)
			authType := api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{
							{
								Services: &[]api.NetworkingService{
									{
										Ref: &api.ServiceRef{Name: workloadName, Port: port},
										Authorization: &api.Workloadv2Authorization{
											AuthorizationType: &authType,
										},
									},
								},
							},
						},
					},
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())
			Expect(payload.Metadata.Name).To(Equal(workloadName))
			Expect(payload.Metadata.Configuration.Networking).NotTo(BeNil())
			Expect(*payload.Metadata.Configuration.Networking).To(HaveLen(1))

			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(BeNumerically(">=", 1))
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.Networking.Services).NotTo(BeNil())
			Expect(*update.Configuration.Networking.Services).To(HaveLen(1))
			Expect((*update.Configuration.Networking.Services)[0].Ref).NotTo(BeNil())
			Expect((*update.Configuration.Networking.Services)[0].Ref.Name).To(Equal(workloadName))
			Expect((*update.Configuration.Networking.Services)[0].Ref.Port).To(Equal(port))
			Expect((*update.Configuration.Networking.Services)[0].Authorization).NotTo(BeNil())
			Expect(*(*update.Configuration.Networking.Services)[0].Authorization.AuthorizationType).To(Equal(api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS))
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
			Expect(*update.Configuration.NetworkingEndpoints).To(HaveLen(1))
			Expect((*update.Configuration.NetworkingEndpoints)[0].Services).NotTo(BeNil())
			Expect(*(*update.Configuration.NetworkingEndpoints)[0].Services).To(HaveLen(1))
			Expect((*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Ref).NotTo(BeNil())
			Expect((*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Ref.Name).To(Equal(workloadName))
			Expect((*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Ref.Port).To(Equal(port))
			Expect((*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Authorization).NotTo(BeNil())
			Expect(*(*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Authorization.AuthorizationType).To(Equal(api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS))
		})

		It("binds created Service selectors to workload pod templates using the server-generated workload ID", func() {
			configureDeploymentKarta(defaultWorkloadsMock, randoms.ClusterId)

			manifest := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &manifest)).To(Succeed())
			labels, found, err := unstructured.NestedStringMap(manifest, "spec", "template", "metadata", "labels")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			labels[networkingWorkloadIDLabel] = "caller-value"
			Expect(unstructured.SetNestedStringMap(manifest, labels, "spec", "template", "metadata", "labels")).To(Succeed())

			serviceType := api.SERVICETYPE_NODE_PORT
			nodePort := int32(30080)
			selector := map[string]string{
				"app":                     "my-app",
				networkingWorkloadIDLabel: "caller-value",
			}
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "networking-created-service",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Spec: &api.ServiceSpec{
									Name:     "networking-created-service",
									Type:     &serviceType,
									Selector: &selector,
									Ports:    []api.ServicePort{{Port: 8080, NodePort: &nodePort}},
								},
							}},
						}},
					},
				},
				Manifest: manifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())

			effectiveLabels, found, err := unstructured.NestedStringMap(
				payload.WorkloadManifest,
				"spec", "template", "metadata", "labels",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(effectiveLabels).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
			Expect(effectiveLabels).To(HaveKeyWithValue("app", "my-app"))

			Expect(payload.Metadata.Configuration.Networking).NotTo(BeNil())
			createdService := (*(*payload.Metadata.Configuration.Networking)[0].Services)[0].Spec
			Expect(createdService).NotTo(BeNil())
			Expect(*createdService.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
			Expect(*createdService.Selector).To(HaveKeyWithValue("app", "my-app"))
			Expect(*createdService.Type).To(Equal(api.SERVICETYPE_NODE_PORT))
			Expect(*createdService.Ports[0].NodePort).To(Equal(nodePort))

			update := findWorkloadUpdate(getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports), id)
			Expect(update).NotTo(BeNil())
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
			updateService := (*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Spec
			Expect(*updateService.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))

			updatedManifest := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &updatedManifest)).To(Succeed())
			updatedLabels, found, err := unstructured.NestedStringMap(updatedManifest, "spec", "template", "metadata", "labels")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			updatedLabels[networkingWorkloadIDLabel] = "updated-caller-value"
			Expect(unstructured.SetNestedStringMap(updatedManifest, updatedLabels, "spec", "template", "metadata", "labels")).To(Succeed())

			updatedSelector := map[string]string{
				"app":                     "my-app",
				networkingWorkloadIDLabel: "updated-caller-value",
			}
			updatedPayload := updateWorkloadForTest(router, randoms, id, &api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Configuration: &api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Spec: &api.ServiceSpec{
									Name:     "networking-created-service",
									Type:     &serviceType,
									Selector: &updatedSelector,
									Ports:    []api.ServicePort{{Port: 8080, NodePort: &nodePort}},
								},
							}},
						}},
					},
				},
				Manifest: updatedManifest,
			})

			updatedEffectiveLabels, found, err := unstructured.NestedStringMap(
				updatedPayload.WorkloadManifest,
				"spec", "template", "metadata", "labels",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(updatedEffectiveLabels).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
			updatedService := (*(*updatedPayload.Metadata.Configuration.Networking)[0].Services)[0].Spec
			Expect(*updatedService.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))

			networkingOnlySelector := map[string]string{
				"app":                     "my-app",
				networkingWorkloadIDLabel: "networking-only-caller-value",
			}
			networkingOnlyConfig := []api.NetworkingEndpoint{{
				Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
					Name:     "networking-created-service",
					Type:     &serviceType,
					Selector: &networkingOnlySelector,
					Ports:    []api.ServicePort{{Port: 8080, NodePort: &nodePort}},
				}}},
			}}
			ctx := setContext(randoms.TenantIdStr, adminUser, nil)
			Expect(impl.UpdateWorkloadNetworkingConfig(ctx, id, &networkingOnlyConfig)).To(BeNil())
			stored, dbErr := dao.DbGetWorkload(ctx, id, nil)
			Expect(dbErr).To(BeNil())
			storedService := (*(*stored.Configuration.Networking)[0].Services)[0].Spec
			Expect(*storedService.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
		})

		It("removes a protected identity supplied by a workload without networking", func() {
			manifest := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &manifest)).To(Succeed())
			labels, found, err := unstructured.NestedStringMap(manifest, "spec", "template", "metadata", "labels")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			labels[networkingWorkloadIDLabel] = "victim-workload-id"
			Expect(unstructured.SetNestedStringMap(manifest, labels, "spec", "template", "metadata", "labels")).To(Succeed())

			id, payload := createWorkloadForTest(router, randoms, &api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "networking-label-spoof-without-service",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			})
			Expect(id).NotTo(BeEmpty())

			effectiveLabels, found, err := unstructured.NestedStringMap(
				payload.WorkloadManifest,
				"spec", "template", "metadata", "labels",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(effectiveLabels).To(HaveKeyWithValue("app", "my-app"))
			Expect(effectiveLabels).NotTo(HaveKey(networkingWorkloadIDLabel))
		})

		It("applies the same binding to redirected native submissions", func() {
			configureDeploymentKarta(defaultWorkloadsMock, randoms.ClusterId)

			manifest := map[string]interface{}{}
			Expect(json.Unmarshal([]byte(sampleManifestSpec), &manifest)).To(Succeed())
			manifestMetadata := manifest["metadata"].(map[string]interface{})
			manifestMetadata["labels"] = map[string]interface{}{ConvertedFromNativeLabel: "true"}

			serviceType := api.SERVICETYPE_CLUSTER_IP
			selector := map[string]string{"app": "my-app"}
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "native-networking-created-service",
					ProjectId: randoms.ProjectIdStr,
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
								Name:     "native-networking-created-service",
								Type:     &serviceType,
								Selector: &selector,
								Ports:    []api.ServicePort{{Port: 8080}},
							}}},
						}},
					},
				},
				Manifest: manifest,
			}

			id, payload := createInternalWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())
			effectiveLabels, found, err := unstructured.NestedStringMap(
				payload.WorkloadManifest,
				"spec", "template", "metadata", "labels",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(effectiveLabels).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
			createdService := (*(*payload.Metadata.Configuration.Networking)[0].Services)[0].Spec
			Expect(*createdService.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, id))
		})

		It("round-trips labeled create services across create, readback, update, and agent updates", func() {
			configureDeploymentKarta(defaultWorkloadsMock, randoms.ClusterId)

			workloadName := "labeled-networking"
			selector := map[string]string{"app": workloadName}
			nimLabels := map[string]string{
				"run.ai/nim-workload":                 "true",
				"run.ai/nim-workload-metrics-service": "true",
				"ordinary":                            "nim-value",
			}
			vllmLabels := map[string]string{
				"run.ai/vllm-workload":                 "true",
				"run.ai/vllm-workload-metrics-service": "true",
				"shared":                               "metrics",
			}
			networking := []api.NetworkingEndpoint{
				{Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
					Name: workloadName + "-nim", Selector: &selector, Labels: nimLabels,
					Ports: []api.ServicePort{{Port: 8000}},
				}}}},
				{Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
					Name: workloadName + "-metrics", Selector: &selector, Labels: vllmLabels,
					Ports: []api.ServicePort{{Port: 9090}},
				}}}},
			}
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: workloadName, ProjectId: randoms.ProjectIdStr, Category: convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{Networking: &networking},
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)
			nimLabels["ordinary"] = "mutated-after-create"
			vllmLabels["shared"] = "mutated-after-create"

			readback := getWorkloadForTest(router, randoms, id)
			Expect(readback.Workload.Metadata.Configuration.Networking).NotTo(BeNil())
			Expect(*readback.Workload.Metadata.Configuration.Networking).To(HaveLen(2))
			Expect((*(*readback.Workload.Metadata.Configuration.Networking)[0].Services)[0].Spec.Labels).To(
				HaveKeyWithValue("ordinary", "nim-value"))
			Expect((*(*readback.Workload.Metadata.Configuration.Networking)[1].Services)[0].Spec.Labels).To(
				HaveKeyWithValue("shared", "metrics"))

			replacementLabels := map[string]string{
				"run.ai/vllm-workload":                 "true",
				"run.ai/vllm-workload-metrics-service": "true",
				"ordinary":                             "updated",
			}
			replacement := []api.NetworkingEndpoint{{Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
				Name: workloadName + "-updated", Selector: &selector, Labels: replacementLabels,
				Ports: []api.ServicePort{{Port: 8080}},
			}}}}}
			updateWorkloadForTest(router, randoms, id, &api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Configuration: &api.WorkloadV2Configuration{Networking: &replacement},
				},
				Manifest: sampleManifest,
			})
			replacementLabels["ordinary"] = "mutated-after-update"

			readback = getWorkloadForTest(router, randoms, id)
			Expect(*readback.Workload.Metadata.Configuration.Networking).To(HaveLen(1))
			Expect((*(*readback.Workload.Metadata.Configuration.Networking)[0].Services)[0].Spec.Labels).To(
				Equal(map[string]string{
					"run.ai/vllm-workload":                 "true",
					"run.ai/vllm-workload-metrics-service": "true",
					"ordinary":                             "updated",
				}))

			update := findWorkloadUpdate(getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports), id)
			Expect(update).NotTo(BeNil())
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
			Expect((*(*update.Configuration.NetworkingEndpoints)[0].Services)[0].Spec.Labels).To(
				HaveKeyWithValue("ordinary", "updated"))
		})

		It("propagates Configuration.Networking changes on update", func() {
			workloadName := "networking-update-test"
			authType := api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS
			createPort := int32(8000)
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Ref:           &api.ServiceRef{Name: workloadName, Port: createPort},
								Authorization: &api.Workloadv2Authorization{AuthorizationType: &authType},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())

			// Mirror the inference submitter: an update that carries a recomposed configuration must
			// replace the create-time configuration. Here the service port changes 8000 -> 9000.
			updatedPort := int32(9000)
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Configuration: &api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Ref:           &api.ServiceRef{Name: workloadName, Port: updatedPort},
								Authorization: &api.Workloadv2Authorization{AuthorizationType: &authType},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}
			updateWorkloadForTest(router, randoms, id, &toUpdate)

			update := findWorkloadUpdate(getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports), id)
			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.Networking.Services).NotTo(BeNil())
			Expect(*update.Configuration.Networking.Services).To(HaveLen(1))
			Expect((*update.Configuration.Networking.Services)[0].Ref).NotTo(BeNil())
			Expect((*update.Configuration.Networking.Services)[0].Ref.Port).To(Equal(updatedPort))
		})

		It("clears create-time Configuration when an explicit empty configuration is sent", func() {
			workloadName := "networking-clear-test"
			authType := api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS
			port := int32(8000)
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Ref:           &api.ServiceRef{Name: workloadName, Port: port},
								Authorization: &api.Workloadv2Authorization{AuthorizationType: &authType},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())

			// An explicit (non-nil) empty configuration fully replaces the previous one, so the
			// create-time networking must be cleared (not retained).
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{Configuration: &api.WorkloadV2Configuration{}},
				Manifest: sampleManifest,
			}
			updateWorkloadForTest(router, randoms, id, &toUpdate)

			update := findWorkloadUpdate(getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports), id)
			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Networking).To(BeNil())
		})

		It("preserves create-time Configuration on a config-less update", func() {
			workloadName := "networking-preserve-test"
			authType := api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS
			port := int32(8000)
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Ref:           &api.ServiceRef{Name: workloadName, Port: port},
								Authorization: &api.Workloadv2Authorization{AuthorizationType: &authType},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())

			// A manifest-only update carries no configuration (all sub-fields nil), so the create-time
			// networking must survive unchanged.
			toUpdate := api.WorkloadV2UpdateRequest{Manifest: sampleManifest}
			updateWorkloadForTest(router, randoms, id, &toUpdate)

			update := findWorkloadUpdate(getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports), id)
			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.Networking.Services).NotTo(BeNil())
			Expect(*update.Configuration.Networking.Services).To(HaveLen(1))
			Expect((*update.Configuration.Networking.Services)[0].Ref).NotTo(BeNil())
			Expect((*update.Configuration.Networking.Services)[0].Ref.Port).To(Equal(port))
		})

		DescribeTable("requires ownership enforcement for direct networking submissions", func(dryRun bool) {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.26.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "networking-version-gate", ProjectId: randoms.ProjectIdStr,
					Configuration: api.WorkloadV2Configuration{Networking: &[]api.NetworkingEndpoint{{
						Services: &[]api.NetworkingService{{Ref: &api.ServiceRef{Name: "networking-version-gate", Port: 8080}}},
					}}},
				},
				Manifest: sampleManifest,
			}
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s?dryRun=%t", testHost, urlWorkloads, dryRun), makePayload(request))
			Expect(err).NotTo(HaveOccurred())
			req.Header = *headerForAdmin(randoms)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("requires cluster version 3.0.0 and above for networking ownership validation"))
		}, Entry("create", false), Entry("dry run", true))

		It("allows old-cluster workloads without networking but rejects adding networking on update", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.26.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")
			id, _ := createWorkloadForTest(router, randoms, &api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "networking-update-gate", ProjectId: randoms.ProjectIdStr,
					Priority: convutils.Ref("high"), Category: convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			})
			request := api.WorkloadV2UpdateRequest{Manifest: sampleManifest, Metadata: api.WorkloadV2MetadataUpdateParams{
				Configuration: &api.WorkloadV2Configuration{Networking: &[]api.NetworkingEndpoint{{
					Services: &[]api.NetworkingService{{Ref: &api.ServiceRef{Name: "networking-update-gate", Port: 8080}}},
				}}},
			}}
			resp := tryToUpdateWorkloadForTest(router, randoms, id, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("networking ownership validation"))
			request.Metadata.Configuration.Networking = &[]api.NetworkingEndpoint{}
			updateWorkloadForTest(router, randoms, id, &request)
		})

		It("rejects metadata.configuration.networking for cluster < 2.25", func() {
			// Use a cluster version that supports V2 workloads (>= 2.23) but not networking (< 2.25)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)).
				WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, "2.24.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			workloadName := "networking-old-cluster"
			port := int32(8000)
			authType := api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: oldCluster.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{
							{
								Services: &[]api.NetworkingService{
									{
										Ref: &api.ServiceRef{Name: workloadName, Port: port},
										Authorization: &api.Workloadv2Authorization{
											AuthorizationType: &authType,
										},
									},
								},
							},
						},
					},
				},
				Manifest: sampleManifest,
			}

			createResp := tryToCreateWorkloadForTest(router, randoms, &toCreate)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest), "create with networking on cluster < 2.25 should be rejected: %s", createResp.Body.String())
			Expect(createResp.Body.String()).To(ContainSubstring("metadata.configuration.networking is supported for cluster version 2.25.0 and above"))
		})

		It("rejects native endpoint-only networking fields for clusters below endpoint-array networking support", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)).
				WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, "2.25.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			workloadName := "networking-rich-old-agent"
			serviceType := api.SERVICETYPE_NODE_PORT
			nodePort := int32(30080)
			selector := map[string]string{"app": workloadName}
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: oldCluster.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Spec: &api.ServiceSpec{
									Name:     workloadName,
									Type:     &serviceType,
									Selector: &selector,
									Ports: []api.ServicePort{{
										Port:     8000,
										NodePort: &nodePort,
									}},
								},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}

			createResp := tryToCreateInternalWorkloadForTest(router, oldCluster, &toCreate)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest), "create with endpoint-only fields on cluster below endpoint-array networking support should be rejected: %s", createResp.Body.String())
			Expect(createResp.Body.String()).To(ContainSubstring("metadata.configuration.networking includes values that require cluster version 2.26.0 and above"))
			Expect(createResp.Body.String()).To(ContainSubstring("type"))
		})

		It("accepts native agent-networking-compatible config for clusters below endpoint-array networking support", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.25.0")).
				WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, "2.22.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			workloadName := "networking-compatible-cluster"
			port := int32(8000)
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      workloadName,
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Networking: &[]api.NetworkingEndpoint{{
							Services: &[]api.NetworkingService{{
								Ref: &api.ServiceRef{Name: workloadName, Port: port},
							}},
						}},
					},
				},
				Manifest: sampleManifest,
			}

			id, _ := createInternalWorkloadForTest(router, randoms, &toCreate)

			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}

			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.Networking.Services).NotTo(BeNil())
			Expect((*update.Configuration.Networking.Services)[0].Ref).NotTo(BeNil())
			Expect((*update.Configuration.Networking.Services)[0].Ref.Name).To(Equal(workloadName))
			Expect((*update.Configuration.Networking.Services)[0].Ref.Port).To(Equal(port))
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
			Expect(*update.Configuration.NetworkingEndpoints).To(HaveLen(1))
		})

		It("persists PVC configuration on create and returns it in updates and GET", func() {
			pvcName := "data"
			pvcsConfig := &api.WorkloadV2UpdateConfigurationPvcsConfig{CleanupOnCompletion: convutils.Ref(true)}
			accessMode := api.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE
			pvcs := []api.WorkloadV2UpdateConfigurationPvc{
				{
					Metadata: api.WorkloadV2UpdateConfigurationPvcMetadata{Name: pvcName},
					Spec: api.WorkloadV2UpdateConfigurationPvcSpec{
						AccessModes: []api.WorkloadV2UpdateConfigurationPvcAccessMode{accessMode},
						Resources: api.WorkloadV2UpdateConfigurationPvcResources{
							Requests: api.WorkloadV2UpdateConfigurationPvcResourcesRequests{Storage: "1Gi"},
						},
					},
				},
			}
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "pvcs-test",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
					Configuration: api.WorkloadV2Configuration{
						Pvcs:       &pvcs,
						PvcsConfig: pvcsConfig,
					},
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())
			Expect(payload.Metadata.Configuration.Pvcs).NotTo(BeNil())
			Expect(*payload.Metadata.Configuration.Pvcs).To(HaveLen(1))
			Expect((*payload.Metadata.Configuration.Pvcs)[0].Metadata.Name).To(Equal(pvcName))
			Expect(payload.Metadata.Configuration.PvcsConfig).To(Equal(pvcsConfig))

			getPayload := getWorkloadForTest(router, randoms, id)
			Expect(getPayload.Workload.Metadata.Configuration.Pvcs).NotTo(BeNil())
			Expect(*getPayload.Workload.Metadata.Configuration.Pvcs).To(HaveLen(1))
			Expect((*getPayload.Workload.Metadata.Configuration.Pvcs)[0].Metadata.Name).To(Equal(pvcName))
			Expect((*getPayload.Workload.Metadata.Configuration.Pvcs)[0].Spec.AccessModes).To(ConsistOf(accessMode))
			Expect((*getPayload.Workload.Metadata.Configuration.Pvcs)[0].Spec.Resources.Requests.Storage).To(Equal("1Gi"))
			Expect(getPayload.Workload.Metadata.Configuration.PvcsConfig).To(Equal(pvcsConfig))

			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(BeNumerically(">=", 1))
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil(), "workload update should be found")
			Expect(update.Configuration.Pvcs).NotTo(BeNil())
			Expect(*update.Configuration.Pvcs).To(HaveLen(1))
			Expect((*update.Configuration.Pvcs)[0].Metadata.Name).To(Equal(pvcName))
			Expect((*update.Configuration.Pvcs)[0].Spec.AccessModes).To(ConsistOf(accessMode))
			Expect((*update.Configuration.Pvcs)[0].Spec.Resources.Requests.Storage).To(Equal("1Gi"))
			Expect(update.Configuration.PvcsConfig).To(Equal(pvcsConfig))
		})

	})

	Context("sync error", func() {
		It("should return correct status", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			expectedMeta := requestToResponseMeta(&toCreate.Metadata, randoms, adminUser, sampleGvk, nil)

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			testutils.Expect(payload.Metadata).ToEqual(expectedMeta)

			// agent pulls update #1
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(1)))

			getPayload := getWorkloadForTest(router, randoms, id)
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))
			Expect(getPayload.SyncInfo.Message).To(Equal(""))

			// agent reports sync issue
			updates = getUpdatesForTest(router, randoms, 1, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 1,
						Message:     convutils.Ref("sync issue with update 1"),
					},
				},
			})
			Expect(len(updates.Workloads)).To(Equal(0))

			// make sure the report from the cluster is reflected
			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_FAILED))
			Expect(getPayload.SyncInfo.Message).To(Equal("sync issue with update 1"))

			// now attempt to delete the record
			deletedPayload := deleteWorkloadForTest(router, randoms, id)
			testutils.Expect(deletedPayload.Metadata).ToEqual(expectedMeta)
			Expect(deletedPayload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))

			// the cluster receives the update
			updates = getUpdatesForTest(router, randoms, 1, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(2, 2)))
			Expect(updates.SystemInfo.HighestUpdate).To(Equal(int64(2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(2)))
			Expect(updates.Workloads[0].DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(updates.Workloads[0].WorkloadManifest).To(BeNil())

			// the status is changes to "deleting"
			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_DELETING))

			// cluster reports about issues with the deletion
			updates = getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 2,
						Message:     convutils.Ref("sync issue with update 2"),
					},
				},
			})
			Expect(len(updates.Workloads)).To(Equal(0))

			// make sure record is deleted, even though cluster could not delete it
			verifyWorkloadNotFound(router, randoms, id)
		})
	})

	Context("multiple name", func() {
		It("should return correct status", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			expectedMeta := requestToResponseMeta(&toCreate.Metadata, randoms, adminUser, sampleGvk, nil)

			_, payload := createWorkloadForTest(router, randoms, &toCreate)
			testutils.Expect(payload.Metadata).ToEqual(expectedMeta)

			Eventually(func() int {
				// wait for the creation event to reach workloads service mock
				if defaultWorkloadsMock.ByWorkloadName[toCreate.Metadata.Name] != nil {
					// once event arrives, we expect workloads controller to reject similar name
					resp := tryToCreateWorkloadForTest(router, randoms, &toCreate)
					return resp.Code
				}
				// the precondition that the event reaches workloads service not happened yet
				return http.StatusPreconditionFailed
			}, "500ms", "100ms").To(Equal(http.StatusConflict))
		})
	})

	Context("delayed report", func() {
		It("should be ignored", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			expectedMeta := requestToResponseMeta(&toCreate.Metadata, randoms, adminUser, sampleGvk, nil)

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			testutils.Expect(payload.Metadata).ToEqual(expectedMeta)

			// agent pulls update #1
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(1)))

			getPayload := getWorkloadForTest(router, randoms, id)
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))
			Expect(getPayload.SyncInfo.Message).To(Equal(""))

			// now attempt to delete the record
			deletedPayload := deleteWorkloadForTest(router, randoms, id)
			testutils.Expect(deletedPayload.Metadata).ToEqual(expectedMeta)
			Expect(deletedPayload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))

			// the cluster receives the update
			updates = getUpdatesForTest(router, randoms, 1, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(2, 2)))
			Expect(updates.SystemInfo.HighestUpdate).To(Equal(int64(2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(2)))
			Expect(updates.Workloads[0].DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(updates.Workloads[0].WorkloadManifest).To(BeNil())

			// the status is changes to "deleting"
			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_DELETING))

			// agent reports sync issue with the creation, which is no longer relevant
			updates = getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 1,
						Message:     convutils.Ref("sync issue with update 1"),
					},
				},
			})
			Expect(len(updates.Workloads)).To(Equal(0))

			// status remains the same
			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_DELETING))
			Expect(getPayload.SyncInfo.Message).To(Equal(""))

			// cluster reports about issues with the deletion
			updates = getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 2,
						Message:     convutils.Ref("sync issue with update 2"),
					},
				},
			})
			Expect(len(updates.Workloads)).To(Equal(0))

			// make sure record is deleted, even though cluster could not delete it
			verifyWorkloadNotFound(router, randoms, id)
		})
	})

	Context("simultaneous updates", func() {
		It("each workload is treated properly", func() {
			toCreate1 := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			toCreate2 := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic2",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			expectedMeta1 := requestToResponseMeta(&toCreate1.Metadata, randoms, adminUser, sampleGvk, nil)
			expectedMeta2 := requestToResponseMeta(&toCreate2.Metadata, randoms, adminUser, sampleGvk, nil)

			id1, payload1 := createWorkloadForTest(router, randoms, &toCreate1)
			testutils.Expect(payload1.Metadata).ToEqual(expectedMeta1)

			id2, payload2 := createWorkloadForTest(router, randoms, &toCreate2)
			testutils.Expect(payload2.Metadata).ToEqual(expectedMeta2)

			// agent receives both updates
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 2)))
			Expect(len(updates.Workloads)).To(Equal(2))

			getPayload1 := getWorkloadForTest(router, randoms, id1)
			Expect(getPayload1.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))
			Expect(getPayload1.SyncInfo.Message).To(Equal(""))

			getPayload2 := getWorkloadForTest(router, randoms, id2)
			Expect(getPayload1.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))
			Expect(getPayload1.SyncInfo.Message).To(Equal(""))

			// agent reports back
			updates = getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id1,
						UpdateCount: 1,
						Message:     convutils.Ref("sync issue with update 1"),
					},
					{
						Id:          id2,
						UpdateCount: 2,
					},
				},
			})
			Expect(len(updates.Workloads)).To(Equal(0))

			// status of workload 1 changes to Synced with issues
			getPayload1 = getWorkloadForTest(router, randoms, id1)
			Expect(getPayload1.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload1.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_FAILED))
			Expect(getPayload1.SyncInfo.Message).To(Equal("sync issue with update 1"))

			// status of workload 2 is synced, no issues
			getPayload2 = getWorkloadForTest(router, randoms, id2)
			Expect(getPayload2.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			Expect(getPayload2.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_SYNCED))
			Expect(getPayload2.SyncInfo.Message).To(Equal(""))
		})
	})

	Context("invoke API - old cluster", func() {
		It("should fail with not implemented", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: oldCluster.ProjectIdStr,
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("This API is incompatible with the version of cluster"))
		})
	})

	Context("bulk delete test", func() {
		It("should populate all updates to the cluster and return the result", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			//
			//   create 5 workloads
			//
			ids := make([]string, 0)

			for i := 1; i <= 5; i++ {
				toCreate.Metadata.Name = fmt.Sprintf("generic-%d", i)
				id, _ := createWorkloadForTest(router, randoms, &toCreate)
				ids = append(ids, id)
			}

			idsSet := datastructs.NewSet(ids...)

			// expecting to get creation updates about all those workloads
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(Equal(5))

			// now we want to delete workloads 1 3 and 5
			request := api.WorkloadsV2Batch{
				Ids: []string{
					ids[0],
					ids[2],
					ids[4],
				},
				Action: api.WORKLOADSV2BATCHACTION_DELETE,
			}

			deleteReq, err := http.NewRequest("POST", batchUrl, makePayload(request))
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(randoms)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusOK))

			var result api.BatchResponse
			jsonErr := json.Unmarshal(deleteResp.Body.Bytes(), &result)
			Expect(jsonErr).To(BeNil())

			Expect(len(result.Failed)).To(Equal(0))
			Expect(len(result.Succeeded)).To(Equal(3))

			for _, id := range result.Succeeded {
				Expect(idsSet.Has(id)).To(Equal(true))
			}

			// now we expect 3 deletion updates to be sent to the agent
			updates = getUpdatesForTest(router, randoms, 5, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(Equal(3))

			for _, workload := range updates.Workloads {
				Expect(idsSet.Has(workload.Id)).To(Equal(true))
				Expect(workload.DesiredPhase).To(Equal(api.DESIREDPHASE_DELETED))
			}
		})
	})

	Context("invoke API - non available priority", func() {
		It("should fail", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("na"),
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("priority 'na' is not available in this cluster"))
		})
	})

	Context("invoke API - non available category", func() {
		It("should fail", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "generic1",
					ProjectId: randoms.ProjectIdStr,
					Category:  convutils.Ref("na"),
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("category 'na' is not available in this cluster"))
		})
	})

	Context("GetWorkloadName", func() {
		It("should return API name when provided and not used as prefix", func() {
			metadata := &api.WorkloadV2MetadataCreateParams{
				Name:                 "api-workload",
				UseGivenNameAsPrefix: false,
				ProjectId:            "test-project",
			}
			manifest := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "yaml-workload",
				},
			}

			name := impl.GetWorkloadName(context.TODO(), metadata, manifest)
			Expect(name).To(Equal("api-workload"))
		})

		It("should generate API name with suffix when UseGivenNameAsPrefix is true", func() {
			metadata := &api.WorkloadV2MetadataCreateParams{
				Name:                 "api-prefix",
				UseGivenNameAsPrefix: true,
				ProjectId:            "test-project",
			}
			manifest := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "yaml-workload",
				},
			}

			name := impl.GetWorkloadName(context.TODO(), metadata, manifest)
			Expect(name).To(HavePrefix("api-prefix-"))
			Expect(len(name)).To(BeNumerically(">", len("api-prefix-")))
		})

		It("should return YAML name when API name is empty", func() {
			metadata := &api.WorkloadV2MetadataCreateParams{
				Name:      "",
				ProjectId: "test-project",
			}
			manifest := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "yaml-workload",
				},
			}

			name := impl.GetWorkloadName(context.TODO(), metadata, manifest)
			Expect(name).To(Equal("yaml-workload"))
		})

		It("should return YAML generateName as prefix when API name is empty", func() {
			metadata := &api.WorkloadV2MetadataCreateParams{
				Name:      "",
				ProjectId: "test-project",
			}
			manifest := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"generateName": "yaml-prefix-",
				},
			}

			name := impl.GetWorkloadName(context.TODO(), metadata, manifest)
			Expect(name).To(ContainSubstring("yaml-prefix-"))
		})

		It("should prioritize YAML name over generateName when both exist", func() {
			metadata := &api.WorkloadV2MetadataCreateParams{
				Name:      "",
				ProjectId: "test-project",
			}
			manifest := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":         "yaml-workload",
					"generateName": "yaml-prefix-",
				},
			}

			name := impl.GetWorkloadName(context.TODO(), metadata, manifest)
			Expect(name).To(Equal("yaml-workload"))
		})
	})

	Context("Cluster ID and Project ID validation", func() {
		It("should fail when both projectId and clusterId are empty", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "test-workload",
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("either projectId or clusterId must be provided"))
		})

		It("should fail when both projectId and clusterId are provided", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("projectId and clusterId are mutually exclusive"))
		})

		It("should succeed when only projectId is provided", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusAccepted))
		})

		It("should succeed when only clusterId is provided with namespace in manifest", func() {
			manifestWithNamespace := make(map[string]interface{})
			for k, v := range sampleManifest {
				manifestWithNamespace[k] = v
			}
			metadata := manifestWithNamespace["metadata"].(map[string]interface{})
			metadata["namespace"] = fmt.Sprintf("runai-%s", randoms.ProjectIdStr)

			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ClusterId: randoms.ClusterId,
				},
				Manifest: manifestWithNamespace,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusAccepted))

			var result api.WorkloadV2
			err = json.Unmarshal(createResp.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
		})

		It("should fail when clusterId is provided without namespace in manifest", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ClusterId: randoms.ClusterId,
				},
				Manifest: sampleManifest,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("namespace must be specified in the manifest"))
		})

		It("should fail when clusterId is provided with invalid namespace", func() {
			manifestWithInvalidNamespace := make(map[string]interface{})
			for k, v := range sampleManifest {
				manifestWithInvalidNamespace[k] = v
			}
			metadata := manifestWithInvalidNamespace["metadata"].(map[string]interface{})
			metadata["namespace"] = "invalid-namespace"

			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ClusterId: randoms.ClusterId,
				},
				Manifest: manifestWithInvalidNamespace,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlWorkloads), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusNotFound))
			Expect(createResp.Body.String()).To(ContainSubstring("no project found for cluster"))
		})
	})

	Context("priority VS preemptible", func() {
		It("should populate all updates to the cluster and return the workload in get", func() {
			for idx, combination := range prioPremCombinations {
				toCreate := api.WorkloadV2CreateRequest{
					Metadata: api.WorkloadV2MetadataCreateParams{
						Name:           fmt.Sprintf("workload-%d", idx),
						ProjectId:      randoms.ProjectIdStr,
						Category:       convutils.Ref("train"),
						Priority:       convutils.Ref(combination.Priority),
						Preemptibility: combination.Preemptibility,
					},
					Manifest: sampleManifest,
				}

				_, payload := createWorkloadForTest(router, randoms, &toCreate)
				Expect(payload.Metadata.Preemptibility).NotTo(BeNil())
				Expect(*payload.Metadata.Preemptibility).To(Equal(combination.Expected), combination.ToString())

				manifest, err := gojsonq.LoadJson(payload.WorkloadManifest)
				Expect(err).To(BeNil())

				metadata := convutils.MapMe(manifest["metadata"])
				labels := convutils.MapMe(metadata["labels"])
				Expect(labels[constants.LabelPreemptibilityName]).To(Equal(string(combination.Expected)), combination.ToString())
			}
		})
	})

	Context("workload update", func() {
		It("should update workload and propagate changes through sync workflow", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-1",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(payload.Metadata.Priority).To(Equal(convutils.Ref("high")))
			Expect(payload.Metadata.Category).To(Equal(convutils.Ref("train")))

			// Agent syncs creation (update count 1)
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(1)))

			getPayload := getWorkloadForTest(router, randoms, id)
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))

			// Update the workload with new priority and category
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("medium"),
					Category: convutils.Ref("workspace"),
				},
				Manifest: sampleManifest,
			}

			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("medium")))
			Expect(updatedPayload.Metadata.Category).To(Equal(convutils.Ref("workspace")))

			// Agent receives update #2 with new manifest
			updates = getUpdatesForTest(router, randoms, 1, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(2, 2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(2)))

			// Verify sync status transitions
			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_APPLYING))

			// Agent reports sync success
			updates = getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 2,
					},
				},
			})

			getPayload = getWorkloadForTest(router, randoms, id)
			Expect(getPayload.SyncInfo.Status).To(Equal(api.WORKLOADV2SYNCSTATUS_SYNCED))
		})

		It("should update only priority and keep category unchanged", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-2",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(payload.Metadata.Priority).To(Equal(convutils.Ref("high")))
			Expect(payload.Metadata.Category).To(Equal(convutils.Ref("train")))
			Expect(payload.Metadata.Preemptibility).To(Equal(convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE)))

			// Update with priority only
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("low"),
				},
				Manifest: sampleManifest,
			}

			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("low")))
			Expect(updatedPayload.Metadata.Category).To(Equal(convutils.Ref("train")))
			Expect(updatedPayload.Metadata.Preemptibility).To(Equal(convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE)))
		})

		It("should update only category and keep priority unchanged", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-3",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(payload.Metadata.Priority).To(Equal(convutils.Ref("high")))
			Expect(payload.Metadata.Category).To(Equal(convutils.Ref("train")))

			// Update with category only
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Category: convutils.Ref("inference"),
				},
				Manifest: sampleManifest,
			}

			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("high")))
			Expect(updatedPayload.Metadata.Category).To(Equal(convutils.Ref("inference")))
		})

		It("should update manifest without changing priority or category", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-4",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			// Create modified manifest
			modifiedManifest := make(map[string]interface{})
			for k, v := range sampleManifest {
				modifiedManifest[k] = v
			}
			spec := modifiedManifest["spec"].(map[string]interface{})
			template := spec["template"].(map[string]interface{})
			templateSpec := template["spec"].(map[string]interface{})
			containers := templateSpec["containers"].([]interface{})
			container := containers[0].(map[string]interface{})
			container["image"] = "nginx:1.21"

			// Update with modified manifest only
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: modifiedManifest,
			}

			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("high")))
			Expect(updatedPayload.Metadata.Category).To(Equal(convutils.Ref("train")))

			// Verify manifest changed
			getPayload := getWorkloadForTest(router, randoms, id)
			rawSpec := getPayload.Workload.RawManifest["spec"].(map[string]interface{})
			rawTemplate := rawSpec["template"].(map[string]interface{})
			rawTemplateSpec := rawTemplate["spec"].(map[string]interface{})
			rawContainers := rawTemplateSpec["containers"].([]interface{})
			rawContainer := rawContainers[0].(map[string]interface{})
			Expect(rawContainer["image"]).To(Equal("nginx:1.21"))

			// Verify new update is sent to cluster
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(BeNumerically(">=", 1))
		})

		It("should fail when updating with non-available priority", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-5",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			// Attempt to update with invalid priority
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("na"),
				},
				Manifest: sampleManifest,
			}

			updateResp := tryToUpdateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("priority 'na' is not available"))
		})

		It("should fail when updating with non-available category", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-6",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			// Attempt to update with invalid category
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Category: convutils.Ref("na"),
				},
				Manifest: sampleManifest,
			}

			updateResp := tryToUpdateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("category 'na' is not available"))
		})

		It("should fail when workload does not exist", func() {
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("high"),
				},
				Manifest: sampleManifest,
			}

			fakeId := "00000000-0000-0000-0000-000000000000"
			updateResp := tryToUpdateWorkloadForTest(router, randoms, fakeId, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusNotFound))
		})

		It("should fail when updated manifest has different GVK", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-8",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: sampleManifest, // Deployment (apps/v1)
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			// Create a Job manifest (different GVK)
			jobManifest := map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "my-job",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "job-container",
									"image": "busybox:latest",
								},
							},
							"restartPolicy": "Never",
						},
					},
				},
			}

			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: jobManifest,
			}

			updateResp := tryToUpdateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject update of v1-cast workload (Job + workloadKind=Training)", func() {
			jobManifest := map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "v1-cast-training",
					"labels": map[string]interface{}{
						ConvertedFromNativeLabel:    "true",
						constants.LabelWorkloadKind: "Training",
					},
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "trainer",
									"image": "busybox:latest",
								},
							},
							"restartPolicy": "Never",
						},
					},
				},
			}

			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "v1-cast-training",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: jobManifest,
			}
			// Use internal endpoint so TranslatedFromV1 is set to true
			id, _ := createInternalWorkloadForTest(router, randoms, &toCreate)

			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: jobManifest,
			}
			updateResp := tryToUpdateInternalWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusForbidden))
			Expect(updateResp.Body.String()).To(ContainSubstring("immutable"))
			Expect(updateResp.Body.String()).To(ContainSubstring("Training"))
		})

		It("should allow update of v1-cast workload whose workloadKind is not in the immutable allowlist", func() {
			jobManifest := map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "v1-cast-other",
					"labels": map[string]interface{}{
						ConvertedFromNativeLabel:    "true",
						constants.LabelWorkloadKind: "Inference",
					},
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "inference",
									"image": "busybox:latest",
								},
							},
							"restartPolicy": "Never",
						},
					},
				},
			}

			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "v1-cast-other",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: jobManifest,
			}
			// Use internal endpoint so TranslatedFromV1 is set to true
			id, _ := createInternalWorkloadForTest(router, randoms, &toCreate)

			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: jobManifest,
			}
			updateResp := tryToUpdateInternalWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusAccepted))
		})

		It("should reject a public update of an internally created workload", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "internal-create-public-update",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: sampleManifest,
			}
			id, _ := createInternalWorkloadForTest(router, randoms, &toCreate)

			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: sampleManifest,
			}
			updateResp := tryToUpdateWorkloadForTest(router, randoms, id, &toUpdate)

			Expect(updateResp.Code).To(Equal(http.StatusForbidden))
			Expect(updateResp.Body.String()).To(ContainSubstring("cannot update workload from a different API type"))
		})

		It("should reject an internal update of a publicly created workload", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "public-create-internal-update",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: sampleManifest,
			}
			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{},
				Manifest: sampleManifest,
			}
			updateResp := tryToUpdateInternalWorkloadForTest(router, randoms, id, &toUpdate)

			Expect(updateResp.Code).To(Equal(http.StatusForbidden))
			Expect(updateResp.Body.String()).To(ContainSubstring("cannot update workload from a different API type"))
		})

		It("should recalculate preemptibility when priority changes", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-10",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, payload := createWorkloadForTest(router, randoms, &toCreate)
			Expect(payload.Metadata.Preemptibility).To(Equal(convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE)))

			// Check manifest labels
			manifest1, err := gojsonq.LoadJson(payload.WorkloadManifest)
			Expect(err).To(BeNil())
			metadata1 := convutils.MapMe(manifest1["metadata"])
			labels1 := convutils.MapMe(metadata1["labels"])
			Expect(labels1[constants.LabelPreemptibilityName]).To(Equal(string(api.PREEMPTIBILITY_NON_PREEMPTIBLE)))

			// Update to low priority
			toUpdate := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("low"),
				},
				Manifest: sampleManifest,
			}

			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("low")))
			Expect(updatedPayload.Metadata.Preemptibility).To(Equal(convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE)))

			// Check manifest labels
			manifest2, err := gojsonq.LoadJson(updatedPayload.WorkloadManifest)
			Expect(err).To(BeNil())
			metadata2 := convutils.MapMe(manifest2["metadata"])
			labels2 := convutils.MapMe(metadata2["labels"])
			Expect(labels2[constants.LabelPreemptibilityName]).To(Equal(string(api.PREEMPTIBILITY_PREEMPTIBLE)))
		})

		It("should handle multiple updates to same workload", func() {
			toCreate := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "update-test-11",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
					Category:  convutils.Ref("train"),
				},
				Manifest: sampleManifest,
			}

			id, _ := createWorkloadForTest(router, randoms, &toCreate)

			// Agent syncs update #1
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(1, 1)))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(1)))

			// Report sync success for update #1
			getUpdatesForTest(router, randoms, 1, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 1,
					},
				},
			})

			// First update
			toUpdate1 := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("medium"),
				},
				Manifest: sampleManifest,
			}
			updateWorkloadForTest(router, randoms, id, &toUpdate1)

			// Agent syncs update #2
			updates = getUpdatesForTest(router, randoms, 1, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(2, 2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(2)))

			// Report sync success for update #2
			getUpdatesForTest(router, randoms, 2, 50, api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:          id,
						UpdateCount: 2,
					},
				},
			})

			// Second update
			toUpdate2 := api.WorkloadV2UpdateRequest{
				Metadata: api.WorkloadV2MetadataUpdateParams{
					Priority: convutils.Ref("low"),
				},
				Manifest: sampleManifest,
			}
			updatedPayload := updateWorkloadForTest(router, randoms, id, &toUpdate2)
			Expect(updatedPayload.Metadata.Priority).To(Equal(convutils.Ref("low")))

			// Agent receives only latest update (#3)
			updates = getUpdatesForTest(router, randoms, 2, 50, emptyClusterReports)
			Expect(updates.ResponseInfo).To(Equal(responseInfoForTest(3, 3)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(3)))
		})
	})
})

const urlRunState = "/api/v2/workloads/%s/runState"

func tryRunStateForTest(router chi.Router, randoms *testutils.RandomValues, id string, req api.WorkloadRunStateRequest) *httptest.ResponseRecorder {
	patchReq, err := http.NewRequest("PATCH",
		fmt.Sprintf("%s"+urlRunState, testHost, id),
		makePayload(req))
	Expect(err).To(BeNil())
	patchReq.Header = *headerForAdmin(randoms)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, patchReq)
	return resp
}

func runStateForTest(router chi.Router, randoms *testutils.RandomValues, id string, req api.WorkloadRunStateRequest) *api.WorkloadV2 {
	resp := tryRunStateForTest(router, randoms, id, req)
	Expect(resp.Code).To(Equal(http.StatusAccepted), "runState failed: %s", resp.Body.String())

	var result api.WorkloadV2
	Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
	sanitizeMeta(&result.Metadata)
	return &result
}

func suspendDefinitionForBatchJob() *kartav1alpha1.SuspendDefinition {
	return &kartav1alpha1.SuspendDefinition{
		SuspendActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "true"},
		},
		ResumeActions: []kartav1alpha1.SuspendAction{
			{Path: ".spec.suspend", Value: "false"},
		},
	}
}

func strPtr(s string) *string { return &s }

var _ = Describe("SetWorkloadRunState", func() {
	var (
		randoms       *testutils.RandomValues
		router        chi.Router
		workloadsMock *workloads.SimpleMock
	)

	BeforeEach(func() {
		cleanTables()
		dao.RestartUpdateCountForTest()
		cache.ClearStore[componentDefCacheEntry]()

		randoms = testutils.NewRandomValues()

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(randoms))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		setting_pkg.Init(nil, nil)
		setting_pkg.Register[*settings.YamlSubmission]()

		workloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
			WithWorkloadType(batchJobWorkloadType(randoms.ClusterId, suspendDefinitionForBatchJob()))
		workloads.SetMock(workloadsMock)

		policies.SetMock(policies.SimpleMockClientFactory{})

		csclient := clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		router = createRouterForTest()
	})

	createRunningJob := func(name string) string {
		manifestJSON := fmt.Sprintf(`{
			"apiVersion": "batch/v1",
			"kind": "Job",
			"metadata": {"name": %q},
			"spec": {"template": {"spec": {"containers": [{"name": "c", "image": "img:latest"}]}}}
		}`, name)
		var manifestMap map[string]interface{}
		Expect(json.Unmarshal([]byte(manifestJSON), &manifestMap)).To(Succeed())
		req := &api.WorkloadV2CreateRequest{
			Metadata: api.WorkloadV2MetadataCreateParams{
				Name:      name,
				ProjectId: randoms.ProjectIdStr,
			},
			Manifest: manifestMap,
		}
		id, _ := createWorkloadForTest(router, randoms, req)
		// A freshly created workload is registered in the mock with actual phase Creating; a "running"
		// job should be reported by the workloads service as Running so it is suspendable.
		workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_RUNNING)
		return id
	}

	Context("suspend", func() {
		It("records an accepted suspend", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			id := createRunningJob("job-suspend-1")
			recorder.reset()

			result := runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			Expect(result.DesiredPhase).To(Equal(api.DESIREDPHASE_STOPPED))
			actions := recorder.recorded()
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].Action).To(Equal(workloadsApi.WORKLOADACTIONTYPE_SUSPEND))
			Expect(actions[0].WorkloadID).To(Equal(id))
		})

		It("should apply spec.suspend=true patch to both manifests", func() {
			id := createRunningJob("job-suspend-2")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			stored := getWorkloadForTest(router, randoms, id)
			rawManifest := stored.Workload.RawManifest
			Expect(rawManifest).NotTo(BeNil())
			specMap, ok := rawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "spec should be a map")
			Expect(specMap["suspend"]).To(Equal(true))
		})

		It("should suspend an MPIJob v2beta1 through spec.runPolicy.suspend", func() {
			kubeflowSuspendDefinition := &kartav1alpha1.SuspendDefinition{
				SuspendActions: []kartav1alpha1.SuspendAction{
					{Path: ".spec.runPolicy.suspend", Value: "true"},
				},
				ResumeActions: []kartav1alpha1.SuspendAction{
					{Path: ".spec.runPolicy.suspend", Value: "false"},
				},
			}
			cache.ClearStore[componentDefCacheEntry]()
			workloadsMock.WorkloadTypes = []workloadsApi.WorkloadTypeConfig{
				workloadTypeWithComponentDefinition(
					"kubeflow.org",
					"v2beta1",
					"MPIJob",
					randoms.ClusterId,
					&kartav1alpha1.ComponentDefinition{Name: "mpijob", SuspendDefinition: kubeflowSuspendDefinition},
				),
			}

			manifest := map[string]interface{}{
				"apiVersion": "kubeflow.org/v2beta1",
				"kind":       "MPIJob",
				"metadata":   map[string]interface{}{"name": "mpi-v2beta1-suspend"},
				"spec": map[string]interface{}{
					"runPolicy": map[string]interface{}{},
				},
			}
			request := &api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "mpi-v2beta1-suspend",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: manifest,
			}
			id, _ := createWorkloadForTest(router, randoms, request)
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_RUNNING)

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			stored := getWorkloadForTest(router, randoms, id)
			spec, ok := stored.Workload.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			runPolicy, ok := spec["runPolicy"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(runPolicy["suspend"]).To(Equal(true))
		})

		It("does not record a rejected suspend", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			id := createRunningJob("job-suspend-dup")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			// Simulate the cluster reporting the workload as Stopped after the suspend.
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)
			recorder.reset()

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("not permitted"))
			Expect(recorder.recorded()).To(BeEmpty())
		})

		It("should reject suspending a workload whose actual phase is Failed", func() {
			id := createRunningJob("job-suspend-failed")
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_FAILED)

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("Failed"))
		})

		It("should return 400 when workload type has no suspendDefinition", func() {
			// Mutate the existing mock in place (rather than replacing it) so that createRunningJob's
			// SetWorkloadPhase and the handler's GetWorkload target the same mock instance.
			workloadsMock.WorkloadTypes = []workloadsApi.WorkloadTypeConfig{
				{
					Name:  "Job",
					Group: "batch",
					ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
						{ClusterId: &randoms.ClusterId, Version: strPtr("v1"), Phase: strPtr("Ready")},
					},
				},
			}

			id := createRunningJob("job-no-suspend-def")
			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("does not support suspend"))
		})

		It("should return 404 for a non-existent workload ID", func() {
			resp := tryRunStateForTest(router, randoms, "00000000-0000-0000-0000-000000000000",
				api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("resume", func() {
		It("records an accepted resume", func() {
			recorder := &capturingWorkloadActionRecorder{}
			impl.workloadActionRecorder = recorder
			id := createRunningJob("job-resume-1")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			// Simulate the cluster reporting the workload as Stopped after the suspend.
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)
			recorder.reset()
			result := runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})

			Expect(result.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
			actions := recorder.recorded()
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].Action).To(Equal(workloadsApi.WORKLOADACTIONTYPE_RESUME))
			Expect(actions[0].WorkloadID).To(Equal(id))
		})

		It("should apply spec.suspend=false patch on resume", func() {
			id := createRunningJob("job-resume-2")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)
			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})

			stored := getWorkloadForTest(router, randoms, id)
			rawManifest := stored.Workload.RawManifest
			Expect(rawManifest).NotTo(BeNil())
			specMap, ok := rawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "spec should be a map")
			Expect(specMap["suspend"]).To(Equal(false))
		})

		It("should reject resuming a workload already in Running state", func() {
			id := createRunningJob("job-resume-running")
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_RUNNING)

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("permitted only for stopped workloads"))
		})

		It("should reject resuming a workload whose actual phase is Failed", func() {
			id := createRunningJob("job-resume-failed")
			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_FAILED)

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("Failed"))
		})

		It("should resume a workload suspended cluster-side even when DesiredPhase is still Running", func() {
			// The workload was never suspended through our API, so DesiredPhase stays Running, but the
			// cluster reports it as Stopped (e.g. suspended by the runtime-rules).
			id := createRunningJob("job-resume-cluster-side")
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)

			result := runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			Expect(result.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
		})
	})

	Context("unknown run state value", func() {
		It("should return 400 for an unrecognized run state", func() {
			id := createRunningJob("job-unknown-state")
			// The router validates the enum before reaching the handler, so the error message
			// comes from the request parser, not from the switch-default in SetWorkloadRunState.
			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: "Hibernate"})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("Hibernate"))
		})
	})

	Context("suspend definition backend error", func() {
		It("should return 503 when the backend fails to return the suspend definition", func() {
			id := createRunningJob("job-backend-error")

			workloads.SetMock(workloadTypesErrorClient{
				SimpleMock: workloads.NewSimpleMock().
					WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
					WithWorkloadType(workloadsApi.WorkloadTypeConfig{
						Name:  "Job",
						Group: "batch",
						ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
							{ClusterId: &randoms.ClusterId, Version: strPtr("v1"), Phase: strPtr("Ready")},
						},
					}),
				err: workloadsApi.NewError(http.StatusServiceUnavailable, "backend unavailable"),
			})

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable))
			Expect(resp.Body.String()).To(ContainSubstring("failed to fetch component definition"))
		})
	})

	Context("cluster version check", func() {
		It("should return 400 when cluster version is below 2.26", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.25.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			id := createRunningJob("job-old-cluster-suspend")
			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("2.26"))
		})

		It("should return 400 on resume when cluster version is below 2.26", func() {
			id := createRunningJob("job-old-cluster-resume")
			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.25.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("2.26"))
		})

		It("should succeed when cluster version is exactly 2.26", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, "2.26.0")))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			id := createRunningJob("job-exact-version")
			result := runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			Expect(result.DesiredPhase).To(Equal(api.DESIREDPHASE_STOPPED))
		})
	})

	Context("resume with no suspendDefinition", func() {
		It("should return 400 when workload type has no suspendDefinition on resume", func() {
			id := createRunningJob("job-no-resume-def")
			// Suspend with a definition first so DesiredPhase becomes Stopped.
			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			// Remove the definition and attempt resume.
			workloads.SetMock(workloads.NewSimpleMock().
				WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
				WithWorkloadType(workloadsApi.WorkloadTypeConfig{
					Name:  "Job",
					Group: "batch",
					ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
						{ClusterId: &randoms.ClusterId, Version: strPtr("v1"), Phase: strPtr("Ready")},
					},
				}))
			resp := tryRunStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("does not support suspend"))
		})
	})

	Context("effective manifest patching", func() {
		It("should apply patches to both manifests so a full suspend→resume cycle produces correct state", func() {
			// If only RawManifest were patched, a subsequent anyworkload-controller reconcile using
			// EffectiveManifest would push the old (non-suspended) spec back.
			id := createRunningJob("job-effective-patch")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})
			suspended := getWorkloadForTest(router, randoms, id)
			specAfterSuspend, ok := suspended.Workload.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(specAfterSuspend["suspend"]).To(Equal(true))

			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)
			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_RESUME})
			resumed := getWorkloadForTest(router, randoms, id)
			specAfterResume, ok := resumed.Workload.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(specAfterResume["suspend"]).To(Equal(false))
		})
	})

	Context("optimistic locking", func() {
		It("should increment update_count after a successful run-state change", func() {
			id := createRunningJob("job-oc")

			before := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			var beforeCount int64
			for _, u := range before.Workloads {
				if u.Id == id {
					beforeCount = u.UpdateCount
					break
				}
			}
			Expect(beforeCount).To(Equal(int64(1)))

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			after := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			var afterCount int64
			for _, u := range after.Workloads {
				if u.Id == id {
					afterCount = u.UpdateCount
					break
				}
			}
			Expect(afterCount).To(Equal(beforeCount + 1))
		})
	})

	Context("multi-patch suspend definition", func() {
		It("should apply all patches in the suspend action list", func() {
			multiPatchDef := &kartav1alpha1.SuspendDefinition{
				SuspendActions: []kartav1alpha1.SuspendAction{
					{Path: ".spec.suspend", Value: "true"},
					{Path: ".metadata.annotations.suspended", Value: `"true"`},
				},
				ResumeActions: []kartav1alpha1.SuspendAction{
					{Path: ".spec.suspend", Value: "false"},
					{Path: ".metadata.annotations.suspended", Value: `"false"`},
				},
			}
			cache.ClearStore[componentDefCacheEntry]()
			workloadsMock.WorkloadTypes = []workloadsApi.WorkloadTypeConfig{
				batchJobWorkloadType(randoms.ClusterId, multiPatchDef),
			}

			id := createRunningJob("job-multi-patch")

			runStateForTest(router, randoms, id, api.WorkloadRunStateRequest{Value: api.RUNSTATE_SUSPEND})

			stored := getWorkloadForTest(router, randoms, id)
			rawManifest := stored.Workload.RawManifest
			Expect(rawManifest).NotTo(BeNil())

			specMap, ok := rawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(specMap["suspend"]).To(Equal(true))

			metaMap, ok := rawManifest["metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			annotations, ok := metaMap["annotations"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(annotations["suspended"]).To(Equal("true"))
		})
	})
})

func batchJobComponentDef(suspendDef *kartav1alpha1.SuspendDefinition) *kartav1alpha1.ComponentDefinition {
	return &kartav1alpha1.ComponentDefinition{
		Name:              "Job",
		SuspendDefinition: suspendDef,
	}
}

func batchJobWorkloadType(clusterId string, suspendDef *kartav1alpha1.SuspendDefinition) workloadsApi.WorkloadTypeConfig {
	return workloadTypeWithComponentDefinition("batch", "v1", "Job", clusterId, batchJobComponentDef(suspendDef))
}

func workloadTypeWithComponentDefinition(group, version, kind, clusterId string, componentDef *kartav1alpha1.ComponentDefinition) workloadsApi.WorkloadTypeConfig {
	karta := kartav1alpha1.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: *componentDef,
			},
		},
	}
	kartaBytes, err := json.Marshal(karta)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var kartaMap map[string]interface{}
	ExpectWithOffset(1, json.Unmarshal(kartaBytes, &kartaMap)).To(Succeed())

	return workloadsApi.WorkloadTypeConfig{
		Name:  kind,
		Group: group,
		Kartas: []workloadsApi.WorkloadTypeKartasWithVersionInner{
			{Version: strPtr(version), Karta: kartaMap},
		},
		ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
			{ClusterId: &clusterId, Version: strPtr(version), Phase: strPtr("Ready")},
		},
	}
}

func configureDeploymentKarta(workloadsMock *workloads.SimpleMock, clusterID string) {
	deploymentType := workloadTypeWithComponentDefinition(
		"apps",
		"v1",
		"Deployment",
		clusterID,
		&kartav1alpha1.ComponentDefinition{
			Name: "Deployment",
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: convutils.Ref(".spec.template"),
			},
		},
	)
	for index := range workloadsMock.WorkloadTypes {
		if workloadsMock.WorkloadTypes[index].Name == "Deployment" &&
			workloadsMock.WorkloadTypes[index].Group == "apps" {
			workloadsMock.WorkloadTypes[index].Kartas = deploymentType.Kartas
		}
	}
}

type workloadTypesErrorClient struct {
	*workloads.SimpleMock
	err *workloadsApi.Error
}

func (c workloadTypesErrorClient) ListWorkloadTypes(_ context.Context, _ []string) (*workloadsApi.ListWorkloadTypes200Response, *workloadsApi.Error) {
	return nil, c.err
}

// patchManifestSuspendActions unit tests — no DB required, pure function tests.
var _ = Describe("patchManifestSuspendActions", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("returns nil for nil manifest", func() {
		result, err := patchManifestSuspendActions(ctx, nil, *batchJobComponentDef(suspendDefinitionForBatchJob()), true)
		Expect(err).To(BeNil())
		Expect(result).To(BeNil())
	})

	It("applies suspend actions from the component definition", func() {
		manifest := dbmodels.JsonBMap{}
		result, err := patchManifestSuspendActions(ctx, manifest, *batchJobComponentDef(suspendDefinitionForBatchJob()), true)
		Expect(err).To(BeNil())
		spec, ok := result["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(spec["suspend"]).To(Equal(true))
	})

	It("applies resume actions from the component definition", func() {
		manifest := dbmodels.JsonBMap{
			"spec": map[string]interface{}{"suspend": true},
		}
		result, err := patchManifestSuspendActions(ctx, manifest, *batchJobComponentDef(suspendDefinitionForBatchJob()), false)
		Expect(err).To(BeNil())
		spec := result["spec"].(map[string]interface{})
		Expect(spec["suspend"]).To(Equal(false))
	})

	It("applies multiple suspend actions", func() {
		componentDef := batchJobComponentDef(&kartav1alpha1.SuspendDefinition{
			SuspendActions: []kartav1alpha1.SuspendAction{
				{Path: ".spec.suspend", Value: "true"},
				{Path: ".metadata.annotations.suspended", Value: `"true"`},
			},
		})
		result, err := patchManifestSuspendActions(ctx, dbmodels.JsonBMap{}, *componentDef, true)
		Expect(err).To(BeNil())
		spec := result["spec"].(map[string]interface{})
		Expect(spec["suspend"]).To(Equal(true))
		meta := result["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations["suspended"]).To(Equal("true"))
	})

	It("returns error when SuspendDefinition is missing", func() {
		_, err := patchManifestSuspendActions(ctx, dbmodels.JsonBMap{}, kartav1alpha1.ComponentDefinition{}, true)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("does not support suspend/resume"))
	})

	It("returns error for an invalid JSON value", func() {
		componentDef := batchJobComponentDef(&kartav1alpha1.SuspendDefinition{
			SuspendActions: []kartav1alpha1.SuspendAction{
				{Path: ".spec.suspend", Value: "not-valid-json"},
			},
		})
		_, err := patchManifestSuspendActions(ctx, dbmodels.JsonBMap{}, *componentDef, true)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("failed to apply suspend/resume patch"))
	})

	It("does not mutate the original manifest", func() {
		original := dbmodels.JsonBMap{
			"spec": map[string]interface{}{"replicas": float64(3)},
		}
		_, patchErr := patchManifestSuspendActions(ctx, original, *batchJobComponentDef(suspendDefinitionForBatchJob()), true)
		Expect(patchErr).To(BeNil())
		origSpec := original["spec"].(map[string]interface{})
		Expect(origSpec).NotTo(HaveKey("suspend"))
		Expect(origSpec["replicas"]).To(Equal(float64(3)))
	})
})

var _ = Describe("ProcessBatch", func() {
	var (
		randoms       *testutils.RandomValues
		router        chi.Router
		workloadsMock *workloads.SimpleMock
	)

	BeforeEach(func() {
		cleanTables()
		dao.RestartUpdateCountForTest()

		randoms = testutils.NewRandomValues()

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(randoms))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		setting_pkg.Init(nil, nil)
		setting_pkg.Register[*settings.YamlSubmission]()

		workloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
			WithWorkloadType(batchJobWorkloadType(randoms.ClusterId, suspendDefinitionForBatchJob()))
		workloads.SetMock(workloadsMock)

		policies.SetMock(policies.SimpleMockClientFactory{})

		csclient := clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		router = createRouterForTest()
	})

	createJobForBatch := func(name string) string {
		manifestJSON := fmt.Sprintf(`{
			"apiVersion": "batch/v1",
			"kind": "Job",
			"metadata": {"name": %q},
			"spec": {"template": {"spec": {"containers": [{"name": "c", "image": "img:latest"}]}}}
		}`, name)
		var manifestMap map[string]interface{}
		Expect(json.Unmarshal([]byte(manifestJSON), &manifestMap)).To(Succeed())
		req := &api.WorkloadV2CreateRequest{
			Metadata: api.WorkloadV2MetadataCreateParams{
				Name:      name,
				ProjectId: randoms.ProjectIdStr,
			},
			Manifest: manifestMap,
		}
		id, _ := createWorkloadForTest(router, randoms, req)
		// A freshly created workload is registered in the mock with actual phase Creating; a "running"
		// job should be reported by the workloads service as Running so it is suspendable.
		workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_RUNNING)
		return id
	}

	batchForTest := func(action api.WorkloadsV2BatchAction, ids []string) *httptest.ResponseRecorder {
		req, err := http.NewRequest("POST", batchUrl, makePayload(api.WorkloadsV2Batch{
			Action: action,
			Ids:    ids,
		}))
		Expect(err).To(BeNil())
		req.Header = *headerForAdmin(randoms)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		return resp
	}

	Context("suspend action", func() {
		It("should suspend a workload and set DesiredPhase=Stopped", func() {
			id := createJobForBatch("batch-suspend-1")

			resp := batchForTest(api.WORKLOADSV2BATCHACTION_SUSPEND, []string{id})
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(ConsistOf(id))
			Expect(result.Failed).To(BeEmpty())

			workload := getWorkloadForTest(router, randoms, id)
			Expect(workload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_STOPPED))
		})

		It("should return failures for non-existent workload IDs", func() {
			nonExistentId := "00000000-0000-0000-0000-000000000000"
			resp := batchForTest(api.WORKLOADSV2BATCHACTION_SUSPEND, []string{nonExistentId})
			Expect(resp.Code).To(Equal(http.StatusOK))

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(BeEmpty())
			Expect(result.Failed).To(HaveLen(1))
			Expect(result.Failed[0].Id).To(Equal(nonExistentId))
		})

		It("should handle mixed valid and non-existent IDs", func() {
			id := createJobForBatch("batch-suspend-mixed")
			nonExistentId := "00000000-0000-0000-0000-000000000001"

			resp := batchForTest(api.WORKLOADSV2BATCHACTION_SUSPEND, []string{id, nonExistentId})
			Expect(resp.Code).To(Equal(http.StatusOK))

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(ConsistOf(id))
			Expect(result.Failed).To(HaveLen(1))
			Expect(result.Failed[0].Id).To(Equal(nonExistentId))
		})
	})

	Context("resume action", func() {
		It("should resume a suspended workload and set DesiredPhase=Running", func() {
			id := createJobForBatch("batch-resume-1")

			// suspend first
			suspendResp := batchForTest(api.WORKLOADSV2BATCHACTION_SUSPEND, []string{id})
			Expect(suspendResp.Code).To(Equal(http.StatusOK), suspendResp.Body.String())

			// Simulate the cluster reporting the workload as Stopped after the suspend.
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_STOPPED)

			// then resume
			resp := batchForTest(api.WORKLOADSV2BATCHACTION_RESUME, []string{id})
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(ConsistOf(id))
			Expect(result.Failed).To(BeEmpty())

			workload := getWorkloadForTest(router, randoms, id)
			Expect(workload.Workload.DesiredPhase).To(Equal(api.DESIREDPHASE_RUNNING))
		})

		It("should report failure when resuming an already Running workload", func() {
			// Workload is created in Running phase — no suspend first.
			id := createJobForBatch("batch-resume-already-running")
			workloadsMock.SetWorkloadPhase(id, workloadsApi.PHASE_RUNNING)

			resp := batchForTest(api.WORKLOADSV2BATCHACTION_RESUME, []string{id})
			Expect(resp.Code).To(Equal(http.StatusOK))

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(BeEmpty())
			Expect(result.Failed).To(HaveLen(1))
			Expect(result.Failed[0].Id).To(Equal(id))
		})
	})

	Context("delete action", func() {
		It("should delete workloads and set DesiredPhase=Deleted", func() {
			id1 := createJobForBatch("batch-delete-1")
			id2 := createJobForBatch("batch-delete-2")

			resp := batchForTest(api.WORKLOADSV2BATCHACTION_DELETE, []string{id1, id2})
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var result api.BatchResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			Expect(result.Succeeded).To(ConsistOf(id1, id2))
			Expect(result.Failed).To(BeEmpty())

			// verify both are marked deleted via updates
			updates := getUpdatesForTest(router, randoms, 0, 50, emptyClusterReports)
			deletedPhases := map[string]api.DesiredPhase{}
			for _, u := range updates.Workloads {
				deletedPhases[u.Id] = u.DesiredPhase
			}
			Expect(deletedPhases[id1]).To(Equal(api.DESIREDPHASE_DELETED))
			Expect(deletedPhases[id2]).To(Equal(api.DESIREDPHASE_DELETED))
		})
	})

	Context("unknown action", func() {
		It("should return 400 for an unrecognized batch action", func() {
			id := createJobForBatch("batch-unknown-action")
			resp := batchForTest("fly", []string{id})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
