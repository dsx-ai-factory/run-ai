/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/settings"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	policiesClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	policiesApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/middleware"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
)

const (
	urlNimServices = NimServicesAPIPath
)

// sample NIMService spec with all mandatory fields set
var sampleNimServiceSpec = api.NimServiceSpec{
	NgcAuthSecret: convutils.Ref("ngc-secret"),
	Image:         convutils.Ref("nvidia/nim:latest"),
	ServingPort: &api.NimServiceServingPort{
		Port: convutils.Ref(int32(8000)),
	},
	ModelStore: &api.NimServiceSpecAllOfModelStore{
		NimCache: &api.NimCache{
			Name: convutils.Ref("nim-cache"),
		},
	},
	PriorityClass: convutils.Ref("high"),
	Category:      convutils.Ref("inference"),
}

var _ = DescribeTable("removeEmptyExposeRouter",
	func(manifest map[string]interface{}, routerShouldRemain bool) {
		removeEmptyExposeRouter(manifest)

		spec, _ := manifest["spec"].(map[string]interface{})
		expose, _ := spec["expose"].(map[string]interface{})
		if routerShouldRemain {
			Expect(expose).To(HaveKey("router"))
		} else if expose != nil {
			Expect(expose).NotTo(HaveKey("router"))
		}
	},
	Entry("strips an empty router", map[string]interface{}{
		"spec": map[string]interface{}{
			"expose": map[string]interface{}{
				"service": map[string]interface{}{"port": int32(8000)},
				"router":  map[string]interface{}{},
			},
		},
	}, false),
	Entry("keeps a populated router", map[string]interface{}{
		"spec": map[string]interface{}{
			"expose": map[string]interface{}{
				"router": map[string]interface{}{"hostDomainName": "example.com"},
			},
		},
	}, true),
	Entry("is a no-op when expose is missing", map[string]interface{}{
		"spec": map[string]interface{}{},
	}, false),
	Entry("is a no-op when spec is missing", map[string]interface{}{}, false),
)

var _ = Describe("NimServicesImpl", func() {
	var (
		router chi.Router

		randoms    *testutils.RandomValues
		oldCluster *testutils.RandomValues

		defaultWorkloadsMock *workloads.SimpleMock
		orgUnitMock          *org_unit.SimpleMock

		nimServicesImpl *NimServicesImpl
	)

	BeforeEach(func() {
		setting_pkg.Init(nil, nil)
		setting_pkg.Register[*settings.YamlSubmission]()

		randoms = testutils.NewRandomValues()
		oldCluster = randoms.CloneWithDifferentCluster()

		k8sUtil := NewK8sResourceUtil()
		workloadsImpl := &WorkloadsV2Impl{
			k8sUtil: k8sUtil,
		}
		nimServicesImpl = NewNimServicesImpl(workloadsImpl)

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		// Setup org unit mock with default project
		orgUnitMock = org_unit.NewSimpleMock().
			WithRandoms(randoms).
			WithRandoms(oldCluster)
		org_unit.SetMock(orgUnitMock)

		mockClient := clusterservice.NewSimpleMock()

		displayedCluster := &clustersApi.DisplayedCluster{
			Uuid: randoms.ClusterId,
		}
		displayedCluster.Domain.Set(convutils.Ref("https://example.com"))
		displayedCluster.Version.Set(convutils.Ref(testutils.DefaultClusterVersionForTesting))
		clusterStatus := clustersApi.NewClusterDisplayedStatus()
		ingressClass := "nginx"
		clusterStatus.Config.Set(&clustersApi.ClusterReportedConfig{
			IngressClass: *clustersApi.NewNullableString(&ingressClass),
		})
		displayedCluster.Status.Set(clusterStatus)
		mockClient.WithCluster(displayedCluster.Uuid, displayedCluster)
		mockClient.WithDefaultNodePool(randoms.ClusterId)
		clustersClient.SetMock(mockClient)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)).
			WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, "2.22.0")))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")
		setting_pkg.Init(nil, nil)
		Expect(setting_pkg.Register[*settings.YamlSubmission]()).To(BeNil())

		// Setup default workloads client mock
		defaultWorkloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
			WithPrioritiesAndCategories(oldCluster.ClusterId, defaultPrioritiesAndCategories)

		// Add NIMService workload type for tests
		readyPhase := "Ready"
		nimServiceWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:       "NIMService",
			Group:      "apps.nvidia.com",
			CategoryId: "default-category",
			PriorityId: "default-priority",
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1alpha1"),
					Phase:     &readyPhase,
				},
			},
		}

		// Add Pod workload type for regular workload tests
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

		defaultWorkloadsMock.WithWorkloadType(nimServiceWorkloadType)
		defaultWorkloadsMock.WithWorkloadType(podWorkloadType)
		workloads.SetMock(defaultWorkloadsMock)

		dao.RestartUpdateCountForTest()

		router = createNimServicesRouterForTest(nimServicesImpl)

		// Cluster domain is required for NIMService creation, for ingress
		err := os.Setenv("CLUSTER_DOMAIN", "https://example.com")
		Expect(err).To(BeNil())

		policiesClient.SetMock(policiesClient.SimpleMockClientFactory{
			SimpleMockClient: policiesClient.SimpleMockClient{
				EffectiveUpdatedAt:        time.Now(),
				EffectiveNimServicePolicy: &policiesApi.NimServicePolicyDefaultsAndRules{},
			},
		})
	})

	AfterEach(func() {
		cleanTables()
	})

	Context("successful creation", func() {
		DescribeTable("creates a default NIM service workload", func(ctx SpecContext, domain *string) {
			cluster, err := clustersClient.Get().GetClusterByUuid(ctx, randoms.ClusterId)
			Expect(err).To(BeNil())
			cluster.Domain.Set(domain)

			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-1",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, payload := createNimServiceForTest(router, randoms, &toCreate)

			// Verify the workload was created
			Expect(id).NotTo(BeEmpty())

			Expect(payload).NotTo(BeNil())
			Expect(payload.Metadata.Name).To(Equal("nim-service-1"))
			Expect(payload.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
			Expect(*payload.Metadata.Priority).To(Equal(*sampleNimServiceSpec.PriorityClass))
			Expect(*payload.Metadata.Category).To(Equal(*sampleNimServiceSpec.Category))

			Expect(payload.DesiredPhase).NotTo(BeEmpty())

			Expect(payload.Spec).To(Equal(sampleNimServiceSpec))
			if domain == nil {
				Expect(payload.Metadata.Configuration.Networking).To(BeNil())
			} else {
				Expect(payload.Metadata.Configuration.Networking).NotTo(BeNil())
			}
		},
			Entry("with a cluster domain", convutils.Ref("https://example.com")),
			Entry("without a cluster domain", nil),
		)

		It("preemptibility is correct", func() {
			for idx, combination := range prioPremCombinations {
				toCreate := api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      fmt.Sprintf("nim-service-%d", idx),
						ProjectId: randoms.ProjectIdStr,
					},
					Spec: sampleNimServiceSpec,
				}

				toCreate.Spec.PriorityClass = convutils.Ref(combination.Priority)
				toCreate.Spec.Preemptibility = combination.Preemptibility

				id, payload := createNimServiceForTest(router, randoms, &toCreate)

				// Verify the workload was created
				Expect(id).NotTo(BeEmpty())

				Expect(payload).NotTo(BeNil())
				Expect(*payload.Metadata.Priority).To(Equal(combination.Priority))
				Expect(*payload.Metadata.Preemptibility).To(Equal(combination.Expected))
			}
		})

		It("should create a NIM service with generated name when UseGivenNameAsPrefix is true", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:                 "nim-prefix",
					UseGivenNameAsPrefix: true,
					ProjectId:            randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, payload := createNimServiceForTest(router, randoms, &toCreate)

			// Verify the workload was created
			Expect(id).NotTo(BeEmpty())

			Expect(payload).NotTo(BeNil())
			// Name should start with the prefix
			Expect(payload.Metadata.Name).To(HavePrefix("nim-prefix-"))
			// Name should be longer than just the prefix (has suffix)
			Expect(len(payload.Metadata.Name)).To(BeNumerically(">", len("nim-prefix-")))
			Expect(payload.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
			Expect(*payload.Metadata.Priority).To(Equal(*sampleNimServiceSpec.PriorityClass))
			Expect(*payload.Metadata.Category).To(Equal(*sampleNimServiceSpec.Category))

			Expect(payload.DesiredPhase).NotTo(BeEmpty())

			Expect(payload.Spec).To(Equal(sampleNimServiceSpec))
		})

		It("should create multiple NIM services with different generated names when UseGivenNameAsPrefix is true", func() {
			toCreate1 := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:                 "nim-multi",
					UseGivenNameAsPrefix: true,
					ProjectId:            randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			toCreate2 := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:                 "nim-multi",
					UseGivenNameAsPrefix: true,
					ProjectId:            randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id1, payload1 := createNimServiceForTest(router, randoms, &toCreate1)
			id2, payload2 := createNimServiceForTest(router, randoms, &toCreate2)

			// Verify both workloads were created
			Expect(id1).NotTo(BeEmpty())
			Expect(id2).NotTo(BeEmpty())

			// Both names should start with the same prefix
			Expect(payload1.Metadata.Name).To(HavePrefix("nim-multi-"))
			Expect(payload2.Metadata.Name).To(HavePrefix("nim-multi-"))

			// But the generated names should be different
			Expect(payload1.Metadata.Name).NotTo(Equal(payload2.Metadata.Name))
		})

		It("should create NIM service on old cluster (< 2.25) with legacy ingress in NIM CR and no Configuration.Networking", func() {
			oldClusterVersion := "2.24.0"

			oldDisplayedCluster := &clustersApi.DisplayedCluster{
				Uuid: oldCluster.ClusterId,
			}
			oldDisplayedCluster.Domain.Set(convutils.Ref("https://example.com"))
			oldDisplayedCluster.Version.Set(convutils.Ref(oldClusterVersion))

			mockClient := clusterservice.NewSimpleMock()

			mainDisplayedCluster := &clustersApi.DisplayedCluster{
				Uuid: randoms.ClusterId,
			}
			mainDisplayedCluster.Domain.Set(convutils.Ref("https://example.com"))
			mainDisplayedCluster.Version.Set(convutils.Ref(testutils.DefaultClusterVersionForTesting))
			mockClient.WithCluster(mainDisplayedCluster.Uuid, mainDisplayedCluster)
			mockClient.WithDefaultNodePool(randoms.ClusterId)
			mockClient.WithCluster(oldDisplayedCluster.Uuid, oldDisplayedCluster)
			mockClient.WithDefaultNodePool(oldCluster.ClusterId)
			clustersClient.SetMock(mockClient)

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)).
				WithCluster(oldCluster.ClusterId, clusterInfoOf(oldCluster, oldClusterVersion)))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

			readyPhase := "Ready"
			wlMock := workloads.NewSimpleMock().
				WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
				WithPrioritiesAndCategories(oldCluster.ClusterId, defaultPrioritiesAndCategories)
			wlMock.WithWorkloadType(workloadsApi.WorkloadTypeConfig{
				Name:       "NIMService",
				Group:      "apps.nvidia.com",
				CategoryId: "default-category",
				PriorityId: "default-priority",
				ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
					{
						ClusterId: convutils.Ref(oldCluster.ClusterId),
						Version:   convutils.Ref("v1alpha1"),
						Phase:     &readyPhase,
					},
				},
			})
			workloads.SetMock(wlMock)

			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-old-cluster",
					ProjectId: oldCluster.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, payload := createNimServiceForTest(router, oldCluster, &toCreate)
			Expect(id).NotTo(BeEmpty())
			Expect(payload.Metadata.Name).To(Equal("nim-old-cluster"))

			workloadsRouter := createRouterForTest()
			updates := getUpdatesForTest(workloadsRouter, oldCluster, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(BeNumerically(">=", 1))
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil())
			Expect(update.Configuration.Networking).To(BeNil())

			spec, ok := update.WorkloadManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose, ok := spec["expose"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			ingress, ok := expose["ingress"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(ingress["enabled"]).To(Equal(true))
		})

		It("omits the empty expose router from the synced NIMService manifest", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-no-router",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())

			workloadsRouter := createRouterForTest()
			updates := getUpdatesForTest(workloadsRouter, randoms, 0, 50, emptyClusterReports)
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil())

			spec, ok := update.WorkloadManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose, ok := spec["expose"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(expose).To(HaveKey("service"))
			Expect(expose).NotTo(HaveKey("router"))
		})

		It("sets Configuration.Networking from servingPort (expose or auth) and returns it in workload updates", func() {
			specWithExpose := sampleNimServiceSpec
			specWithExpose.ServingPort = &api.NimServiceServingPort{
				Port:              convutils.Ref(int32(8000)),
				ExposeExternally:  convutils.Ref(true),
				ExposedUrl:        convutils.Ref("https://my-nim.example.com"),
				AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
			}
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-networking-test",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: specWithExpose,
			}

			id, payload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(id).NotTo(BeEmpty())
			Expect(payload.Metadata.Name).To(Equal("nim-networking-test"))

			workloadsRouter := createRouterForTest()
			updates := getUpdatesForTest(workloadsRouter, randoms, 0, 50, emptyClusterReports)
			Expect(len(updates.Workloads)).To(BeNumerically(">=", 1))
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil(), "workload update should contain networking config")
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
			Expect(*update.Configuration.NetworkingEndpoints).To(HaveLen(1))
			endpoint := (*update.Configuration.NetworkingEndpoints)[0]
			Expect(endpoint.Services).NotTo(BeNil())
			Expect(*endpoint.Services).To(HaveLen(1))
			Expect((*endpoint.Services)[0].Ref).NotTo(BeNil())
			Expect((*endpoint.Services)[0].Ref.Name).To(Equal("nim-networking-test"))
			Expect((*endpoint.Services)[0].Ref.Port).To(Equal(int32(8000)))
			Expect(endpoint.ExternalAccess).NotTo(BeNil())
			Expect(endpoint.ExternalAccess.Ingress).NotTo(BeNil())
			spec, ok := (*endpoint.ExternalAccess.Ingress)["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			rules, ok := spec["rules"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(rules).To(HaveLen(1))
			rule, ok := rules[0].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(rule["host"]).To(Equal("my-nim.example.com"))
		})
	})

	Context("composer failure", func() {
		It("should return error when composer fails", func() {
			// Create invalid spec that will cause composer to fail
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-invalid",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					// Missing mandatory image field
					Image: nil,
					ServingPort: &api.NimServiceServingPort{ // so spec won't be empty
						Port: convutils.Ref(int32(8000)),
					},
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("failed to compose NIMService spec"))
		})
	})

	Context("servingPort validation", func() {
		It("returns BadRequest when servingPort has authorization but no external exposure", func() {
			specWithAuthNoExpose := sampleNimServiceSpec
			specWithAuthNoExpose.ServingPort = &api.NimServiceServingPort{
				Port:              convutils.Ref(int32(8000)),
				ExposeExternally:  convutils.Ref(false),
				AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
			}
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-auth-no-expose",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: specWithAuthNoExpose,
			}

			err := failToCreateNimServiceForTest(router, randoms, &toCreate, http.StatusBadRequest)
			Expect(err.Message).To(ContainSubstring("servingPort: authorization requires external exposure"))
		})

		It("returns BadRequest on update when servingPort has authorization but no external exposure", func() {
			id, _ := createNimServiceForTest(router, randoms, &api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-auth-no-expose",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			})

			resp := tryToUpdateNimServiceForTest(router, randoms, id, &api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					ServingPort: &api.NimServiceServingPort{
						Port:              convutils.Ref(int32(8000)),
						ExposeExternally:  convutils.Ref(false),
						AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
					},
				},
			})

			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("servingPort: authorization requires external exposure"))
		})
	})

	Context("org unit failure", func() {
		It("should return error when GetProjectById fails", func() {
			// Use a different project ID that's not in the mock to trigger error
			invalidProjectId := "99988877"

			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-1",
					ProjectId: invalidProjectId,
				},
				Spec: sampleNimServiceSpec,
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusNotFound))
			Expect(createResp.Body.String()).To(ContainSubstring("project does not exist"))
		})
	})

	Context("policy validation", func() {
		var policy policiesApi.NimServicePolicyDefaultsAndRules

		BeforeEach(func() {
			policy = policiesApi.NimServicePolicyDefaultsAndRules{
				Defaults: *policiesApi.NewNullableNimServiceDefaults(&policiesApi.NimServiceDefaults{
					Autoscaling: *policiesApi.NewNullableNimServiceDefaultsAllOfAutoscaling(&policiesApi.NimServiceDefaultsAllOfAutoscaling{
						Metric:      *policiesApi.NewNullableString(convutils.Ref("default_metric")),
						MinReplicas: *policiesApi.NewNullableInt32(convutils.Ref(int32(3))),
						MaxReplicas: *policiesApi.NewNullableInt32(convutils.Ref(int32(5))),
					}),
				}),
				Rules: *policiesApi.NewNullableNimServiceRules(&policiesApi.NimServiceRules{
					Autoscaling: *policiesApi.NewNullableNimServiceRulesAllOfAutoscaling(&policiesApi.NimServiceRulesAllOfAutoscaling{
						Metric: *policiesApi.NewNullableStringRules(&policiesApi.StringRules{
							Options: []policiesApi.StringOption{
								{Value: "default_metric"},
								{Value: "metric_1"},
								{Value: "metric_2"},
							},
						}),
						MinReplicas: *policiesApi.NewNullableIntegerRules(&policiesApi.IntegerRules{
							CanEdit: *policiesApi.NewNullableBool(convutils.Ref(false)),
						}),
						MaxReplicas: *policiesApi.NewNullableIntegerRules(&policiesApi.IntegerRules{
							Max: *policiesApi.NewNullableInt32(convutils.Ref(int32(10))),
						}),
					}),
				}),
			}

			policiesClient.SetMock(policiesClient.SimpleMockClientFactory{
				SimpleMockClient: policiesClient.SimpleMockClient{
					EffectiveUpdatedAt:        time.Now(),
					EffectiveNimServicePolicy: &policy,
				},
			})
		})

		It("mutate policy + missing field", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-1",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			err := failToCreateNimServiceForTest(router, randoms, &toCreate, http.StatusBadRequest)
			Expect(err.Message).To(ContainSubstring("failed to compose NIMService spec"))
			Expect(err.Details).To(ContainSubstring("metricThreshold is mandatory"))
		})

		It("mutate policy", func() {
			var modifiedSpec api.NimServiceSpec
			convutils.CopySpecs(&modifiedSpec, sampleNimServiceSpec)

			modifiedSpec.Autoscaling = &api.NimServiceSpecAllOfAutoscaling{
				MetricThreshold: convutils.Ref(int32(6)),
				MaxReplicas:     convutils.Ref(int32(7)),
			}

			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-1",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: modifiedSpec,
			}

			var expectedSpec api.NimServiceSpec
			convutils.CopySpecs(&expectedSpec, sampleNimServiceSpec)

			expectedSpec.Autoscaling = &api.NimServiceSpecAllOfAutoscaling{
				Metric:          policy.Defaults.Get().Autoscaling.Get().Metric.Get(),
				MinReplicas:     policy.Defaults.Get().Autoscaling.Get().MinReplicas.Get(),
				MaxReplicas:     toCreate.Spec.Autoscaling.MaxReplicas,
				MetricThreshold: toCreate.Spec.Autoscaling.MetricThreshold,
			}

			_, payload := createNimServiceForTest(router, randoms, &toCreate)
			testutils.Expect(payload.Spec).ToEqual(expectedSpec)
		})

		It("validate policy", func() {
			var modifiedSpec api.NimServiceSpec
			convutils.CopySpecs(&modifiedSpec, sampleNimServiceSpec)

			modifiedSpec.Autoscaling = &api.NimServiceSpecAllOfAutoscaling{
				Metric:          convutils.Ref("xxx"),
				MetricThreshold: convutils.Ref(int32(900)),
				MaxReplicas:     convutils.Ref(int32(1000)),
			}

			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-1",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: modifiedSpec,
			}

			err := failToCreateNimServiceForTest(router, randoms, &toCreate, http.StatusBadRequest)
			Expect(err.Message).To(ContainSubstring("The workload request does not comply with the policy set by the administrator"))

			foundIssues := make(map[string]string)
			for _, issue := range err.ComplianceIssues {
				if issue.Field != "" {
					foundIssues[issue.Field] = issue.Details
				}
			}

			expectedIssues := map[string]string{
				"autoscaling.maxReplicas": "must be no greater than 10",
				"autoscaling.metric":      "the value of this field must be one of: default_metric, metric_1, metric_2",
			}

			Expect(foundIssues).To(Equal(expectedIssues))
		})
	})

	DescribeTable("preserves unsupported cluster version errors for GPU mode", func(update, fromPolicy bool) {
		displayed := clusterInfoOf(randoms, "3.0.13")
		var status clustersApi.ClusterDisplayedStatus
		Expect(json.Unmarshal([]byte(`{"dependencies":{"required":{"gpu-stack":{"available":true,"components":{"nvidia":{"available":true}}}}}}`), &status)).To(Succeed())
		displayed.Status.Set(&status)
		clustersClient.SetMock(clusterservice.NewSimpleMock().WithCluster(randoms.ClusterId, &displayed).WithDefaultNodePool(randoms.ClusterId))

		spec := sampleNimServiceSpec
		spec.Compute = &api.NimServiceSpecAllOfCompute{}
		spec.Compute.GpuMemoryPortionRequest = convutils.Ref(0.25)
		request := api.NimServiceCreateRequest{
			Metadata: api.NimServiceMetadataCreateParams{Name: "unavailable-mode", ProjectId: randoms.ProjectIdStr},
			Spec:     spec,
		}
		var id string
		var created *api.NimServiceResponse
		if update {
			// The existing allocation is valid without an explicitly selected mode.
			id, created = createNimServiceForTest(router, randoms, &request)
		}

		if fromPolicy {
			var policy policiesApi.NimServicePolicyDefaultsAndRules
			Expect(json.Unmarshal([]byte(`{"defaults":{"compute":{"gpuComputeMode":"sm-sharing"}}}`), &policy)).To(Succeed())
			policiesClient.SetMock(policiesClient.SimpleMockClientFactory{
				SimpleMockClient: policiesClient.SimpleMockClient{
					EffectiveUpdatedAt:        time.Now(),
					EffectiveNimServicePolicy: &policy,
				},
			})
		} else {
			request.Spec.Compute.GpuComputeMode = convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING)
		}

		var failure api.Error
		if update {
			response := tryToUpdateNimServiceForTest(router, randoms, id, &api.NimServiceUpdateRequest{Spec: request.Spec})
			Expect(response.Code).To(Equal(http.StatusBadRequest), response.Body.String())
			Expect(json.Unmarshal(response.Body.Bytes(), &failure)).To(Succeed())
			get, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), nil)
			Expect(err).To(BeNil())
			get.Header = *headerForAdmin(randoms)
			response = httptest.NewRecorder()
			router.ServeHTTP(response, get)
			Expect(response.Code).To(Equal(http.StatusOK))
			var unchanged api.NimServiceResponse
			Expect(json.Unmarshal(response.Body.Bytes(), &unchanged)).To(Succeed())
			testutils.Expect(unchanged.Spec).ToEqual(created.Spec)
		} else {
			failure = *SubmissionErrorToApiError(failToCreateNimServiceForTest(router, randoms, &request, http.StatusBadRequest))
		}
		Expect(failure.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(failure.Message).To(Equal("failed to compose NIMService spec"))
		Expect(failure.Details).To(ContainSubstring("compute.gpuComputeMode requires cluster version 3.0.14 or later"))
	},
		Entry("create with explicit mode", false, false),
		Entry("create with policy mode", false, true),
		Entry("update with explicit mode", true, false),
	)

	It("persists and recomposes GPU mode through create, GET and update", func() {
		displayed := clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)
		var status clustersApi.ClusterDisplayedStatus
		Expect(json.Unmarshal([]byte(`{"dependencies":{"required":{"gpu-stack":{"available":true,"components":{"nvidia":{"available":true}}}}}}`), &status)).To(Succeed())
		displayed.Status.Set(&status)
		clustersClient.SetMock(clusterservice.NewSimpleMock().WithCluster(randoms.ClusterId, &displayed).WithDefaultNodePool(randoms.ClusterId))
		spec := sampleNimServiceSpec
		spec.Compute = &api.NimServiceSpecAllOfCompute{
			GpuMemoryPortionRequest: convutils.Ref(0.25),
			GpuComputeMode:          convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING),
		}
		request := api.NimServiceCreateRequest{Metadata: api.NimServiceMetadataCreateParams{Name: "mode-model", UseGivenNameAsPrefix: true, ProjectId: randoms.ProjectIdStr}, Spec: spec}
		id, created := createNimServiceForTest(router, randoms, &request)
		Expect(created.Spec.Compute.GpuComputeMode).To(Equal(convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING)))
		Expect(created.Metadata.Name).To(HavePrefix("mode-model-"))
		get := func() api.NimServiceResponse {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), nil)
			Expect(err).To(BeNil())
			req.Header = *headerForAdmin(randoms)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.String()).NotTo(ContainSubstring(`"computeMode"`))
			var result api.NimServiceResponse
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(Succeed())
			return result
		}
		loaded := get()
		Expect(loaded.Spec.Compute.GpuComputeMode).To(Equal(convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING)))
		updates := getUpdatesForTest(createRouterForTest(), randoms, 0, 50, emptyClusterReports)
		var manifest map[string]interface{}
		for _, update := range updates.Workloads {
			if update.Id == id {
				manifest = update.WorkloadManifest
			}
		}
		Expect(manifest).NotTo(BeNil())
		annotations := manifest["spec"].(map[string]interface{})["annotations"].(map[string]interface{})
		Expect(annotations).To(HaveKeyWithValue("nvidia.com/container."+created.Metadata.Name+"-ctr.gpu-compute.mode", "sm-sharing"))
		loaded.Spec.Compute.GpuComputeMode = convutils.Ref(api.GPUCOMPUTEMODE_TIME_SLICING)
		updated := updateNimServiceForTest(router, randoms, id, &api.NimServiceUpdateRequest{Spec: loaded.Spec})
		Expect(updated.Spec.Compute.GpuComputeMode).To(Equal(convutils.Ref(api.GPUCOMPUTEMODE_TIME_SLICING)))
		loaded = get()
		Expect(loaded.Spec.Compute.GpuComputeMode).To(Equal(convutils.Ref(api.GPUCOMPUTEMODE_TIME_SLICING)))
		updates = getUpdatesForTest(createRouterForTest(), randoms, 0, 50, emptyClusterReports)
		for _, update := range updates.Workloads {
			if update.Id == id {
				manifest = update.WorkloadManifest
			}
		}
		annotations = manifest["spec"].(map[string]interface{})["annotations"].(map[string]interface{})
		Expect(annotations).To(HaveKeyWithValue("nvidia.com/container."+created.Metadata.Name+"-ctr.gpu-compute.mode", "time-slicing"))
	})

	DescribeTable("uses ordinary NIM update handling for mode and CPU fields", func(computeJSON, policyJSON, wantMode string, wantCPU *float64) {
		displayed := clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)
		var status clustersApi.ClusterDisplayedStatus
		Expect(json.Unmarshal([]byte(`{"dependencies":{"required":{"gpu-stack":{"available":true,"components":{"nvidia":{"available":true}}}}}}`), &status)).To(Succeed())
		displayed.Status.Set(&status)
		clustersClient.SetMock(clusterservice.NewSimpleMock().WithCluster(randoms.ClusterId, &displayed).WithDefaultNodePool(randoms.ClusterId))
		spec := sampleNimServiceSpec
		spec.Compute = &api.NimServiceSpecAllOfCompute{
			GpuMemoryPortionRequest: convutils.Ref(0.25),
			GpuComputeMode:          convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING),
			CpuCoreRequest:          convutils.Ref(1.0),
		}
		id, created := createNimServiceForTest(router, randoms, &api.NimServiceCreateRequest{
			Metadata: api.NimServiceMetadataCreateParams{Name: "ordinary-update", ProjectId: randoms.ProjectIdStr}, Spec: spec,
		})

		var policy policiesApi.NimServicePolicyDefaultsAndRules
		Expect(json.Unmarshal([]byte(policyJSON), &policy)).To(Succeed())
		policiesClient.SetMock(policiesClient.SimpleMockClientFactory{
			SimpleMockClient: policiesClient.SimpleMockClient{EffectiveUpdatedAt: time.Now(), EffectiveNimServicePolicy: &policy},
		})
		request := api.NimServiceUpdateRequest{Spec: api.NimServiceSpec{Image: convutils.Ref("nvidia/nim:updated")}}
		Expect(json.Unmarshal([]byte(computeJSON), &request.Spec.Compute)).To(Succeed())
		updated := updateNimServiceForTest(router, randoms, id, &request)
		Expect(updated.Spec.Image).To(Equal(request.Spec.Image))
		Expect(updated.Spec.Compute.CpuCoreRequest).To(Equal(wantCPU))
		if wantMode == "" {
			Expect(updated.Spec.Compute.GpuComputeMode).To(BeNil())
		} else {
			Expect(updated.Spec.Compute.GpuComputeMode).To(Equal(convutils.Ref(api.GpuComputeMode(wantMode))))
		}
		get, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), nil)
		Expect(err).ToNot(HaveOccurred())
		get.Header = *headerForAdmin(randoms)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, get)
		Expect(response.Code).To(Equal(http.StatusOK))
		var loaded api.NimServiceResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &loaded)).To(Succeed())
		Expect(loaded.Spec).To(Equal(updated.Spec))

		updates := getUpdatesForTest(createRouterForTest(), randoms, 1, 50, emptyClusterReports)
		Expect(updates.Workloads).To(HaveLen(1))
		Expect(updates.Workloads[0].Id).To(Equal(id))
		nativeSpec := updates.Workloads[0].WorkloadManifest["spec"].(map[string]interface{})
		annotations, _ := nativeSpec["annotations"].(map[string]interface{})
		key := "nvidia.com/container." + created.Metadata.Name + "-ctr.gpu-compute.mode"
		if wantMode == "" {
			Expect(annotations).NotTo(HaveKey(key))
		} else {
			Expect(annotations).To(HaveKeyWithValue(key, wantMode))
		}
	},
		Entry("replacement compute does not backfill policy defaults", `{"gpuMemoryPortionRequest":0.25}`, `{"defaults":{"compute":{"gpuComputeMode":"sm-sharing","cpuCoreRequest":8}}}`, "", (*float64)(nil)),
		Entry("legacy memory replacement has the same default behavior", `{"gpuPortionRequest":0.25}`, `{"defaults":{"compute":{"gpuComputeMode":"sm-sharing","cpuCoreRequest":8}}}`, "", (*float64)(nil)),
		Entry("unrelated edit preserves saved compute after policy changes", `null`, `{"defaults":{"compute":{"gpuComputeMode":"time-slicing","cpuCoreRequest":8}}}`, "sm-sharing", convutils.Ref(1.0)),
		Entry("updates do not introduce field-specific locks", `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"time-slicing","cpuCoreRequest":2}`, `{"defaults":{"compute":{"gpuComputeMode":"sm-sharing","cpuCoreRequest":8}},"rules":{"compute":{"gpuComputeMode":{"canEdit":false},"cpuCoreRequest":{"canEdit":false}}}}`, "time-slicing", convutils.Ref(2.0)),
		Entry("omitted fields follow existing required-rule behavior", `{"gpuMemoryPortionRequest":0.25}`, `{"rules":{"compute":{"gpuComputeMode":{"required":true},"cpuCoreRequest":{"required":true}}}}`, "", (*float64)(nil)),
		Entry("explicit fields follow existing blocked-rule behavior", `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"time-slicing","cpuCoreRequest":2}`, `{"rules":{"compute":{"gpuComputeMode":{"blocked":true},"cpuCoreRequest":{"blocked":true}}}}`, "time-slicing", convutils.Ref(2.0)),
	)

	DescribeTable("persists only the resolved GPU mode alongside submitted NIM fields", func(computeJSON, modeSource, wantMode string) {
		displayed := clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)
		var status clustersApi.ClusterDisplayedStatus
		Expect(json.Unmarshal([]byte(`{"dependencies":{"required":{"gpu-stack":{"available":true,"components":{"nvidia":{"available":true}}}}}}`), &status)).To(Succeed())
		displayed.Status.Set(&status)
		clustersClient.SetMock(clusterservice.NewSimpleMock().WithCluster(randoms.ClusterId, &displayed).WithDefaultNodePool(randoms.ClusterId))

		policyCompute := map[string]interface{}{"cpuCoreRequest": 2, "gpuMemoryPortionRequest": 0.25}
		policyMap := map[string]interface{}{"defaults": map[string]interface{}{"compute": policyCompute, "replicas": 3}}
		if modeSource == "policy" || modeSource == "locked-policy" {
			policyCompute["gpuComputeMode"] = "sm-sharing"
		}
		if modeSource == "locked-policy" {
			policyMap["rules"] = map[string]interface{}{"compute": map[string]interface{}{
				"gpuComputeMode": map[string]interface{}{"canEdit": false},
			}}
		}
		var policy policiesApi.NimServicePolicyDefaultsAndRules
		Expect(convutils.CopySpecs(&policy, policyMap)).To(Succeed())
		policiesClient.SetMock(policiesClient.SimpleMockClientFactory{
			SimpleMockClient: policiesClient.SimpleMockClient{EffectiveUpdatedAt: time.Now(), EffectiveNimServicePolicy: &policy},
		})

		spec := sampleNimServiceSpec
		Expect(json.Unmarshal([]byte(computeJSON), &spec.Compute)).To(Succeed())
		var expected api.NimServiceSpec
		Expect(convutils.CopySpecs(&expected, spec)).To(Succeed())
		if wantMode != "" {
			if expected.Compute == nil {
				expected.Compute = &api.NimServiceSpecAllOfCompute{}
			}
			expected.Compute.GpuComputeMode = convutils.Ref(api.GpuComputeMode(wantMode))
		}
		if expected.Compute != nil {
			if expected.Compute.GpuPortionRequest != nil {
				expected.Compute.GpuMemoryPortionRequest = expected.Compute.GpuPortionRequest
			} else {
				expected.Compute.GpuPortionRequest = expected.Compute.GpuMemoryPortionRequest
			}
		}
		request := api.NimServiceCreateRequest{
			Metadata: api.NimServiceMetadataCreateParams{Name: "mode-persistence", ProjectId: randoms.ProjectIdStr}, Spec: spec,
		}
		id, created := createNimServiceForTest(router, randoms, &request)
		expectedCPU := 2.0
		if spec.Compute != nil && spec.Compute.CpuCoreRequest != nil {
			expectedCPU = *spec.Compute.CpuCoreRequest
		}
		Expect(created.Spec.Compute.CpuCoreRequest).To(Equal(convutils.Ref(expectedCPU)))
		Expect(created.Spec.Replicas).To(Equal(convutils.Ref(int32(3))))

		updates := getUpdatesForTest(createRouterForTest(), randoms, 0, 50, emptyClusterReports)
		var manifest map[string]interface{}
		for _, update := range updates.Workloads {
			if update.Id == id {
				manifest = update.WorkloadManifest
			}
		}
		Expect(manifest).NotTo(BeNil())
		nativeSpec := manifest["spec"].(map[string]interface{})
		Expect(nativeSpec["replicas"]).To(BeNumerically("==", 3))
		resources := nativeSpec["resources"].(map[string]interface{})
		Expect(resources["requests"]).To(HaveKeyWithValue("cpu", fmt.Sprintf("%g", expectedCPU)))
		modeAnnotation := "nvidia.com/container." + created.Metadata.Name + "-ctr.gpu-compute.mode"
		if wantMode == "" {
			Expect(nativeSpec["annotations"]).NotTo(HaveKey(modeAnnotation))
		} else {
			Expect(nativeSpec["annotations"]).To(HaveKeyWithValue(modeAnnotation, wantMode))
		}

		get := func() api.NimServiceResponse {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header = *headerForAdmin(randoms)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result api.NimServiceResponse
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(Succeed())
			return result
		}
		Expect(get().Spec).To(Equal(expected))
		if spec.Compute != nil {
			expected.Replicas = convutils.Ref(int32(2))
			updated := updateNimServiceForTest(router, randoms, id, &api.NimServiceUpdateRequest{Spec: api.NimServiceSpec{Replicas: expected.Replicas}})
			Expect(updated.Spec).To(Equal(expected))
			Expect(get().Spec).To(Equal(expected))
		}
	},
		Entry("explicit mode with new memory name", `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`, "explicit", "sm-sharing"),
		Entry("policy mode with deprecated memory name", `{"gpuPortionRequest":0.25}`, "policy", "sm-sharing"),
		Entry("explicit mode overrides policy default", `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"time-slicing"}`, "policy", "time-slicing"),
		Entry("implicit mode stays omitted", `{"gpuMemoryPortionRequest":0.25}`, "implicit", ""),
		Entry("mode is saved when compute was omitted", `null`, "policy", "sm-sharing"),
		Entry("compute stays omitted without a configured mode", `null`, "implicit", ""),
		Entry("explicit CPU value remains submitted", `{"gpuPortionRequest":0.25,"cpuCoreRequest":1}`, "policy", "sm-sharing"),
		Entry("locked policy mode is saved", `{"gpuMemoryPortionRequest":0.25}`, "locked-policy", "sm-sharing"),
	)

	Context("GetNimService", func() {
		var createdWorkloadId string
		var workloadsRouter chi.Router

		BeforeEach(func() {
			// Create a nim service workload for testing GetNimService
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-service-get-test",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			var err error
			createdWorkloadId, _ = createNimServiceForTest(router, randoms, &toCreate)
			Expect(err).To(BeNil())
			Expect(createdWorkloadId).NotTo(BeEmpty())

			workloadsRouter = createRouterForTest()
		})

		Context("successful retrieval", func() {
			It("should retrieve nim service by ID", func() {
				// First get the original workload via WorkloadV2 API
				getWorkloadReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2/workloads/%s", testHost, createdWorkloadId), nil)
				Expect(err).To(BeNil())
				getWorkloadReq.Header = *headerForAdmin(randoms)

				getWorkloadResp := httptest.NewRecorder()
				workloadsRouter.ServeHTTP(getWorkloadResp, getWorkloadReq)
				Expect(getWorkloadResp.Code).To(Equal(http.StatusOK))

				var originalWorkload api.WorkloadV2GetResponse
				jsonErr := json.Unmarshal(getWorkloadResp.Body.Bytes(), &originalWorkload)
				Expect(jsonErr).To(BeNil())

				// Now get via NIM service API
				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, createdWorkloadId), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusOK))

				var result api.NimServiceResponse
				jsonErr = json.Unmarshal(getResp.Body.Bytes(), &result)
				Expect(jsonErr).To(BeNil())

				// Validate that all workload fields are the same
				Expect(result.Metadata).To(Equal(originalWorkload.Workload.Metadata))
				Expect(result.DesiredPhase).To(Equal(originalWorkload.Workload.DesiredPhase))

				// Validate that the returned spec is the same as the creation request spec
				Expect(result.Spec).To(Equal(sampleNimServiceSpec))
			})

			It("should retrieve nim service created with generated name", func() {
				// Create a nim service with UseGivenNameAsPrefix
				toCreate := api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:                 "nim-generated",
						UseGivenNameAsPrefix: true,
						ProjectId:            randoms.ProjectIdStr,
					},
					Spec: sampleNimServiceSpec,
				}

				generatedWorkloadId, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
				Expect(generatedWorkloadId).NotTo(BeEmpty())
				Expect(createdPayload.Metadata.Name).To(HavePrefix("nim-generated-"))

				// Now retrieve it via NIM service API
				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, generatedWorkloadId), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusOK))

				var result api.NimServiceResponse
				jsonErr := json.Unmarshal(getResp.Body.Bytes(), &result)
				Expect(jsonErr).To(BeNil())

				// Validate that the name matches the created one
				Expect(result.Metadata.Name).To(Equal(createdPayload.Metadata.Name))
				Expect(result.Metadata.Name).To(HavePrefix("nim-generated-"))
				Expect(result.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))

				// Validate that the returned spec is the same as the creation request spec
				Expect(result.Spec).To(Equal(sampleNimServiceSpec))
			})

			It("should retrieve nim service with all spec fields populated", func() {
				// Create a more comprehensive nim service with all fields
				comprehensiveSpec := api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:comprehensive"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(9000)),
					},
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
				}

				toCreate := api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "nim-service-comprehensive",
						ProjectId: randoms.ProjectIdStr,
					},
					Spec: comprehensiveSpec,
				}

				workloadId, _ := createNimServiceForTest(router, randoms, &toCreate)

				// First get the original workload via WorkloadV2 API
				getWorkloadReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2/workloads/%s", testHost, workloadId), nil)
				Expect(err).To(BeNil())
				getWorkloadReq.Header = *headerForAdmin(randoms)

				getWorkloadResp := httptest.NewRecorder()
				workloadsRouter.ServeHTTP(getWorkloadResp, getWorkloadReq)
				Expect(getWorkloadResp.Code).To(Equal(http.StatusOK))

				var originalWorkload api.WorkloadV2GetResponse
				jsonErr := json.Unmarshal(getWorkloadResp.Body.Bytes(), &originalWorkload)
				Expect(jsonErr).To(BeNil())

				// Now get via NIM service API
				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, workloadId), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusOK))

				var result api.NimServiceResponse
				jsonErr = json.Unmarshal(getResp.Body.Bytes(), &result)
				Expect(jsonErr).To(BeNil())

				// Validate that all workload fields are the same
				Expect(result.Metadata).To(Equal(originalWorkload.Workload.Metadata))
				Expect(result.DesiredPhase).To(Equal(originalWorkload.Workload.DesiredPhase))

				// Validate that the returned spec is the same as the creation request spec
				Expect(result.Spec).To(Equal(comprehensiveSpec))
			})
		})

		Context("error cases", func() {
			It("should return 404 for non-existent workload", func() {
				// Use a valid UUID format for the non-existent ID
				nonExistentId := "550e8400-e29b-41d4-a716-446655440000"

				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, nonExistentId), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusNotFound))
			})

			It("should return 400 for workload not created via nim-services route", func() {
				// Create a regular workload (not via nim-services)
				regularWorkload := createRegularWorkloadForTest(randoms)

				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, regularWorkload.Metadata.Id), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusBadRequest))
				Expect(getResp.Body.String()).To(ContainSubstring("workload was not created via nim-services route"))
			})
		})
	})

	Context("Multi-Node and LWS Validation", func() {
		// Helper function to configure cluster with LWS availability
		configureClusterLWS := func(lwsAvailable bool) {
			displayedCluster := clustersApi.DisplayedCluster{
				Uuid: randoms.ClusterId,
			}
			displayedCluster.Domain.Set(convutils.Ref("https://example.com"))
			displayedCluster.Version.Set(convutils.Ref(testutils.DefaultClusterVersionForTesting))

			// Set up GPU stack and LWS
			nvidiaComponent := map[string]clustersApi.ClusterDependencyStatus{
				"nvidia": {
					Available: true,
				},
			}
			status := &clustersApi.ClusterDisplayedStatus{
				Dependencies: &clustersApi.ClusterDependenciesStatus{
					Required: map[string]clustersApi.ClusterDependencyStatus{
						"gpu-stack": {
							Available:  true,
							Components: &nvidiaComponent,
						},
					},
					Optional: map[string]clustersApi.ClusterDependencyStatus{
						"lws": {
							Available: lwsAvailable,
						},
					},
				},
			}
			displayedCluster.Status.Set(status)

			mockClient := clusterservice.NewSimpleMock()
			mockClient.WithCluster(displayedCluster.Uuid, &displayedCluster)
			mockClient.WithDefaultNodePool(randoms.ClusterId)
			clustersClient.SetMock(mockClient)

			// Reinitialize cluster capabilities checker
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, displayedCluster))
			checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")
		}

		BeforeEach(func() {
			// By default, configure cluster with LWS enabled
			configureClusterLWS(true)
		})

		Context("successful multi-node creation with LWS installed", func() {
			It("should create multi-node service when LWS is available", func() {
				multiNodeSpec := sampleNimServiceSpec
				multiNodeSpec.MultiNode = &api.NimServiceMultiNode{
					Workers: convutils.Ref(int32(1)),
				}
				multiNodeSpec.Compute = &api.NimServiceSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(8)),
				}

				request := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "multi-node-test",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: multiNodeSpec,
				}

				id, result := createNimServiceForTest(router, randoms, request)
				Expect(id).NotTo(BeEmpty())
				Expect(result.Spec.MultiNode).NotTo(BeNil())
				Expect(result.Spec.MultiNode.Workers).To(Equal(convutils.Ref(int32(1))))
			})
		})

		Context("fail multi-node creation without LWS  installed", func() {
			It("should return error when LWS is not installed", func() {
				// Override default: set LWS as NOT available
				configureClusterLWS(false)

				multiNodeSpec := sampleNimServiceSpec
				multiNodeSpec.MultiNode = &api.NimServiceMultiNode{
					Workers: convutils.Ref(int32(1)),
				}

				request := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "multi-node-no-lws",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: multiNodeSpec,
				}

				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(request))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)

				Expect(createResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createResp.Body.String()).To(ContainSubstring("LeaderWorkerSet (LWS) is not installed"))
			})
		})

		Context("successful single-node creation without LWS  installed", func() {
			It("should succeed when multiNode is not set at all", func() {
				// Override default: set LWS as NOT available
				configureClusterLWS(false)

				// No multiNode field at all
				request := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "no-multinode-field",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: sampleNimServiceSpec,
				}

				id, result := createNimServiceForTest(router, randoms, request)
				Expect(id).NotTo(BeEmpty())
				Expect(result).NotTo(BeNil())
			})
		})

		Context("create and retrieve multi-node service", func() {
			It("should preserve multiNode spec on retrieval", func() {
				// Use default cluster setup with LWS enabled

				multiNodeSpec := sampleNimServiceSpec
				multiNodeSpec.MultiNode = &api.NimServiceMultiNode{
					Workers: convutils.Ref(int32(2)),
				}
				multiNodeSpec.Compute = &api.NimServiceSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(4)),
				}

				createRequest := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "multinode-retrieve-test",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: multiNodeSpec,
				}

				id, _ := createNimServiceForTest(router, randoms, createRequest)

				// Retrieve the service
				getReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), nil)
				Expect(err).To(BeNil())
				getReq.Header = *headerForAdmin(randoms)

				getResp := httptest.NewRecorder()
				router.ServeHTTP(getResp, getReq)

				Expect(getResp.Code).To(Equal(http.StatusOK))

				var result api.NimServiceResponse
				jsonErr := json.Unmarshal(getResp.Body.Bytes(), &result)
				Expect(jsonErr).To(BeNil())

				// Verify multiNode spec is preserved
				Expect(result.Spec.MultiNode).NotTo(BeNil())
				Expect(result.Spec.MultiNode.Workers).To(Equal(convutils.Ref(int32(2))))
				Expect(result.Spec.Compute.GpuDevicesRequest).To(Equal(convutils.Ref(int32(4))))
			})
		})

		Context("multi-node with zero GPU devices", func() {
			It("should default to tensor parallelism of 1 when gpuDevicesRequest is 0", func() {
				// Use default cluster setup with LWS enabled

				multiNodeSpec := sampleNimServiceSpec
				multiNodeSpec.MultiNode = &api.NimServiceMultiNode{
					Workers: convutils.Ref(int32(1)),
				}
				multiNodeSpec.Compute = &api.NimServiceSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(0)), // 0 GPUs should default tensorParallelism to 1
				}

				request := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "multinode-zero-gpu-test",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: multiNodeSpec,
				}

				id, result := createNimServiceForTest(router, randoms, request)
				Expect(id).NotTo(BeEmpty())
				Expect(result.Spec.MultiNode).NotTo(BeNil())
				Expect(result.Spec.MultiNode.Workers).To(Equal(convutils.Ref(int32(1))))
				Expect(result.Spec.Compute.GpuDevicesRequest).To(Equal(convutils.Ref(int32(0))))
			})
		})

		Context("validation errors", func() {
			It("should return error when multiNode is set but workers is nil", func() {
				// Use default cluster setup with LWS enabled

				multiNodeSpec := sampleNimServiceSpec
				multiNodeSpec.MultiNode = &api.NimServiceMultiNode{
					Workers: nil, // workers is mandatory when multiNode is set
				}

				request := &api.NimServiceCreateRequest{
					Metadata: api.NimServiceMetadataCreateParams{
						Name:      "multinode-missing-workers",
						ProjectId: fmt.Sprintf("%d", randoms.ProjectId),
					},
					Spec: multiNodeSpec,
				}

				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(request))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)

				Expect(createResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createResp.Body.String()).To(ContainSubstring("multiNode: workers is mandatory"))
			})
		})
	})

	Context("isOldFlowNimService", func() {
		It("returns true when manifest has spec.expose.ingress.enabled=true", func() {
			manifest := map[string]interface{}{
				"spec": map[string]interface{}{
					"expose": map[string]interface{}{
						"ingress": map[string]interface{}{
							"enabled": true,
						},
					},
				},
			}
			Expect(isOldFlowNimService(manifest)).To(BeTrue())
		})

		It("returns false when manifest has spec.expose.ingress.enabled=false", func() {
			manifest := map[string]interface{}{
				"spec": map[string]interface{}{
					"expose": map[string]interface{}{
						"ingress": map[string]interface{}{
							"enabled": false,
						},
					},
				},
			}
			Expect(isOldFlowNimService(manifest)).To(BeFalse())
		})

		It("returns false when manifest has no expose section", func() {
			manifest := map[string]interface{}{
				"spec": map[string]interface{}{},
			}
			Expect(isOldFlowNimService(manifest)).To(BeFalse())
		})

		It("returns false when manifest has no ingress section", func() {
			manifest := map[string]interface{}{
				"spec": map[string]interface{}{
					"expose": map[string]interface{}{},
				},
			}
			Expect(isOldFlowNimService(manifest)).To(BeFalse())
		})

		It("returns false when manifest has no spec section", func() {
			manifest := map[string]interface{}{}
			Expect(isOldFlowNimService(manifest)).To(BeFalse())
		})

		It("returns false for nil manifest", func() {
			Expect(isOldFlowNimService(nil)).To(BeFalse())
		})
	})

	Context("updateRequestHasAuthFields", func() {
		It("returns true when authorizationType is AUTHENTICATED_USERS", func() {
			spec := api.NimServiceSpec{
				ServingPort: &api.NimServiceServingPort{
					AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
				},
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeTrue())
		})

		It("returns true when authorizationType is AUTHORIZED_USERS", func() {
			spec := api.NimServiceSpec{
				ServingPort: &api.NimServiceServingPort{
					AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
					AuthorizedUsers:   &[]string{"user1"},
				},
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeTrue())
		})

		It("returns true when authorizationType is AUTHORIZED_GROUPS", func() {
			spec := api.NimServiceSpec{
				ServingPort: &api.NimServiceServingPort{
					AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS),
					AuthorizedGroups:  &[]string{"group1"},
				},
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeTrue())
		})

		It("returns false when authorizationType is PUBLIC (no real auth required)", func() {
			spec := api.NimServiceSpec{
				ServingPort: &api.NimServiceServingPort{
					AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC),
				},
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeFalse())
		})

		It("returns false when servingPort has no authorizationType", func() {
			spec := api.NimServiceSpec{
				ServingPort: &api.NimServiceServingPort{
					Port: convutils.Ref(int32(8000)),
				},
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeFalse())
		})

		It("returns false when servingPort is nil", func() {
			spec := api.NimServiceSpec{
				Image: convutils.Ref("nvidia/nim:latest"),
			}
			Expect(updateRequestHasAuthFields(spec)).To(BeFalse())
		})
	})

	Context("UpdateNimServiceSpec", func() {
		It("should reject update with auth fields for old-flow NIM service", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-old-flow-test",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			permittedScopes, _ := authClient.ParsePermittedScopes(setContext(randoms.TenantIdStr, adminUser, []string{}))
			qRecord, dbErr := dao.DbGetWorkload(setContext(randoms.TenantIdStr, adminUser, []string{}), id, permittedScopes)
			Expect(dbErr).To(BeNil())

			spec, ok := qRecord.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose, ok := spec["expose"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose["ingress"] = map[string]interface{}{"enabled": true}
			Expect(dao.DbUpsertWorkload(setContext(randoms.TenantIdStr, adminUser, []string{}), qRecord)).To(BeNil())

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					ServingPort: &api.NimServiceServingPort{
						AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
					},
				},
			}
			updateResp := tryToUpdateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("cannot set authorization fields on a NIM service that was created with ingress managed by the NIM operator"))
		})

		It("should allow update without auth fields for old-flow NIM service", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-old-flow-non-auth",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			permittedScopes, _ := authClient.ParsePermittedScopes(setContext(randoms.TenantIdStr, adminUser, []string{}))
			qRecord, dbErr := dao.DbGetWorkload(setContext(randoms.TenantIdStr, adminUser, []string{}), id, permittedScopes)
			Expect(dbErr).To(BeNil())

			spec, ok := qRecord.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose, ok := spec["expose"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose["ingress"] = map[string]interface{}{"enabled": true}
			Expect(dao.DbUpsertWorkload(setContext(randoms.TenantIdStr, adminUser, []string{}), qRecord)).To(BeNil())

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}
			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:v2")))
		})

		It("should keep ingress enabled in NIM CR when updating old-flow NIM on >= 2.25 cluster", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-old-flow-ingress-keep",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			ctx := setContext(randoms.TenantIdStr, adminUser, []string{})
			permittedScopes, _ := authClient.ParsePermittedScopes(ctx)
			qRecord, dbErr := dao.DbGetWorkload(ctx, id, permittedScopes)
			Expect(dbErr).To(BeNil())

			spec, ok := qRecord.RawManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose, ok := spec["expose"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			expose["ingress"] = map[string]interface{}{"enabled": true}
			qRecord.Configuration = nil
			Expect(dao.DbUpsertWorkload(ctx, qRecord)).To(BeNil())

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}
			updateNimServiceForTest(router, randoms, id, &toUpdate)

			workloadsRouter := createRouterForTest()
			updates := getUpdatesForTest(workloadsRouter, randoms, 0, 50, emptyClusterReports)
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil())
			Expect(update.Configuration.Networking).To(BeNil())

			updatedSpec, ok := update.WorkloadManifest["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			updatedExpose, hasExpose := updatedSpec["expose"].(map[string]interface{})
			Expect(hasExpose).To(BeTrue())
			updatedIngress, hasIngress := updatedExpose["ingress"].(map[string]interface{})
			Expect(hasIngress).To(BeTrue())
			Expect(updatedIngress["enabled"]).To(Equal(true))
		})

		It("should update Configuration.Networking when updating new-flow NIM on >= 2.25 cluster", func() {
			specWithExpose := sampleNimServiceSpec
			specWithExpose.ServingPort = &api.NimServiceServingPort{
				Port:             convutils.Ref(int32(8000)),
				ExposeExternally: convutils.Ref(true),
				ExposedUrl:       convutils.Ref("https://nim-original.example.com"),
			}
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-net-update",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: specWithExpose,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					ServingPort: &api.NimServiceServingPort{
						Port:             convutils.Ref(int32(8000)),
						ExposeExternally: convutils.Ref(true),
						ExposedUrl:       convutils.Ref("https://nim-updated.example.com"),
					},
				},
			}
			updateNimServiceForTest(router, randoms, id, &toUpdate)

			workloadsRouter := createRouterForTest()
			updates := getUpdatesForTest(workloadsRouter, randoms, 0, 50, emptyClusterReports)
			var update *api.WorkloadV2Update
			for i := range updates.Workloads {
				if updates.Workloads[i].Id == id {
					update = &updates.Workloads[i]
					break
				}
			}
			Expect(update).NotTo(BeNil())
			Expect(update.Configuration.Networking).NotTo(BeNil())
			Expect(update.Configuration.NetworkingEndpoints).NotTo(BeNil())
		})

		It("should update nim service with spec changes", func() {
			// Create initial NIM service
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-1",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("high"),
					Category:      convutils.Ref("inference"),
					Image:         convutils.Ref("nvidia/nim:latest"),
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(8000)),
					},
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
				},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:latest")))

			// Update with new metadata and spec
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:v2")))

			// Get the workload to verify persistence
			getNimService, err := nimServicesImpl.GetNimService(setContext(randoms.TenantIdStr, adminUser, []string{}), id)
			Expect(err).To(BeNil())
			Expect(getNimService.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:v2")))
			// Getting fields that were part of the create request but not part of the update, to verify they are not changed
			Expect(getNimService.Spec.NgcAuthSecret).To(Equal(convutils.Ref("ngc-secret")))
			Expect(getNimService.Spec.ServingPort.Port).To(Equal(convutils.Ref(int32(8000))))
			Expect(getNimService.Spec.ModelStore.NimCache.Name).To(Equal(convutils.Ref("nim-cache")))
		})

		It("should update only priority and keep category and spec unchanged", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-2",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.PriorityClass = convutils.Ref("high")
			toCreate.Spec.Category = convutils.Ref("inference")

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.PriorityClass).To(Equal(convutils.Ref("high")))
			Expect(createdPayload.Spec.Category).To(Equal(convutils.Ref("inference")))
			Expect(*createdPayload.Metadata.Priority).To(Equal("high"))
			Expect(*createdPayload.Metadata.Category).To(Equal("inference"))

			// Update only priority
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("low"),
				},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			// Check both spec and metadata fields
			Expect(updatedPayload.Spec.PriorityClass).To(Equal(convutils.Ref("low")))
			Expect(updatedPayload.Spec.Category).To(Equal(convutils.Ref("inference")))
			Expect(*updatedPayload.Metadata.Priority).To(Equal("low"))
			Expect(*updatedPayload.Metadata.Category).To(Equal("inference"))
			Expect(updatedPayload.Spec.Image).To(Equal(sampleNimServiceSpec.Image))
		})

		It("should update only category and keep priority and spec unchanged", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-3",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.PriorityClass = convutils.Ref("high")
			toCreate.Spec.Category = convutils.Ref("inference")

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.PriorityClass).To(Equal(convutils.Ref("high")))
			Expect(createdPayload.Spec.Category).To(Equal(convutils.Ref("inference")))
			Expect(*createdPayload.Metadata.Priority).To(Equal("high"))
			Expect(*createdPayload.Metadata.Category).To(Equal("inference"))

			// Update only category
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Category: convutils.Ref("workspace"),
				},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			// Check both spec and metadata fields
			Expect(updatedPayload.Spec.PriorityClass).To(Equal(convutils.Ref("high")))
			Expect(updatedPayload.Spec.Category).To(Equal(convutils.Ref("workspace")))
			Expect(*updatedPayload.Metadata.Priority).To(Equal("high"))
			Expect(*updatedPayload.Metadata.Category).To(Equal("workspace"))
			Expect(updatedPayload.Spec.Image).To(Equal(sampleNimServiceSpec.Image))
		})

		It("should merge spec at root level and replace entire field", func() {
			// Create NIM service with compute - CPU
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-4",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("high"),
					Category:      convutils.Ref("inference"),
					Image:         convutils.Ref("nvidia/nim:latest"),
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(8000)),
					},
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
					Compute: &api.NimServiceSpecAllOfCompute{
						CpuCoreRequest: convutils.Ref(float64(2)),
					},
				},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Compute.CpuCoreRequest).To(Equal(convutils.Ref(float64(2))))

			// Update: replace CPU with Memory
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Compute: &api.NimServiceSpecAllOfCompute{
						CpuMemoryRequest: convutils.Ref("100M"),
					},
				},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)

			// Verify CPU is gone, Memory is present. This shows that Compute field is replaced entirely, not merged.
			Expect(updatedPayload.Spec.Compute.CpuCoreRequest).To(BeNil())
			Expect(updatedPayload.Spec.Compute.CpuMemoryRequest).To(Equal(convutils.Ref("100M")))
		})

		It("should distinguish between empty array (remove) and nil (keep unchanged)", func() {
			// Create NIM service with 2 environment variables
			toCreate1 := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-5a",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("high"),
					Category:      convutils.Ref("inference"),
					Image:         convutils.Ref("nvidia/nim:latest"),
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(8000)),
					},
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{Name: convutils.Ref("VAR1"), Value: convutils.Ref("value1")},
						{Name: convutils.Ref("VAR2"), Value: convutils.Ref("value2")},
					},
				},
			}

			id1, createdPayload1 := createNimServiceForTest(router, randoms, &toCreate1)
			Expect(len(*createdPayload1.Spec.EnvironmentVariables)).To(Equal(2))

			// Update with empty array - should remove env vars
			emptyEnvVars := []api.EnvironmentVariable{}
			toUpdate1 := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					EnvironmentVariables: &emptyEnvVars,
				},
			}

			updatedPayload1 := updateNimServiceForTest(router, randoms, id1, &toUpdate1)
			Expect(updatedPayload1.Spec.EnvironmentVariables).To(BeNil())

			// Create another NIM service with 2 environment variables
			toCreate2 := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-5b",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("high"),
					Category:      convutils.Ref("inference"),
					Image:         convutils.Ref("nvidia/nim:latest"),
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(8000)),
					},
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{Name: convutils.Ref("VAR1"), Value: convutils.Ref("value1")},
						{Name: convutils.Ref("VAR2"), Value: convutils.Ref("value2")},
					},
				},
			}

			id2, createdPayload2 := createNimServiceForTest(router, randoms, &toCreate2)
			Expect(len(*createdPayload2.Spec.EnvironmentVariables)).To(Equal(2))

			// Update with nil (field not included) - should keep env vars
			toUpdate2 := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
					// EnvironmentVariables is not specified (nil), should remain unchanged
				},
			}

			updatedPayload2 := updateNimServiceForTest(router, randoms, id2, &toUpdate2)
			Expect(len(*updatedPayload2.Spec.EnvironmentVariables)).To(Equal(2))
			Expect((*updatedPayload2.Spec.EnvironmentVariables)[0].Name).To(Equal(convutils.Ref("VAR1")))
			Expect((*updatedPayload2.Spec.EnvironmentVariables)[1].Name).To(Equal(convutils.Ref("VAR2")))
		})

		It("should distinguish between empty object (remove) and nil (keep unchanged)", func() {
			// Create NIM service with autoscaling configuration
			toCreate1 := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-6a",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("high"),
					Category:      convutils.Ref("inference"),
					Image:         convutils.Ref("nvidia/nim:latest"),
					NgcAuthSecret: convutils.Ref("ngc-secret"),
					ServingPort: &api.NimServiceServingPort{
						Port: convutils.Ref(int32(8000)),
					},
					ModelStore: &api.NimServiceSpecAllOfModelStore{
						NimCache: &api.NimCache{
							Name: convutils.Ref("nim-cache"),
						},
					},
					Autoscaling: &api.NimServiceSpecAllOfAutoscaling{
						MinReplicas:     convutils.Ref(int32(1)),
						MaxReplicas:     convutils.Ref(int32(5)),
						Metric:          convutils.Ref("cpu"),
						MetricThreshold: convutils.Ref(int32(80)),
					},
				},
			}

			id1, createdPayload1 := createNimServiceForTest(router, randoms, &toCreate1)
			Expect(createdPayload1.Spec.Autoscaling).ToNot(BeNil())
			Expect(createdPayload1.Spec.Autoscaling.MinReplicas).To(Equal(convutils.Ref(int32(1))))
			Expect(createdPayload1.Spec.Autoscaling.MaxReplicas).To(Equal(convutils.Ref(int32(5))))

			// Update with empty object - should remove autoscaling
			emptyAutoscaling := api.NimServiceSpecAllOfAutoscaling{}
			toUpdate1 := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Autoscaling: &emptyAutoscaling,
				},
			}

			updatedPayload1 := updateNimServiceForTest(router, randoms, id1, &toUpdate1)
			Expect(updatedPayload1.Spec.Autoscaling).To(BeNil())
		})

		It("should reject fieldsToRemove with a non-existent field name", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-invalid-remove-field",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			// "multiNodexx" is not a valid NimServiceRemovableFields value (typo of "multiNode")
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{"multiNodexx"},
			}

			updateResp := tryToUpdateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest), "Response body: %s", updateResp.Body.String())
			Expect(updateResp.Body.String()).To(ContainSubstring("multiNodexx"))
		})

		It("should remove compute section using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-compute",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Compute = &api.NimServiceSpecAllOfCompute{
				CpuCoreRequest: convutils.Ref(float64(2)),
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Compute).ToNot(BeNil())

			// Use fieldsToRemove to remove compute (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_COMPUTE},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Compute).To(BeNil())
			// Other fields should remain unchanged
			Expect(updatedPayload.Spec.Image).To(Equal(sampleNimServiceSpec.Image))
		})

		It("should remove security section using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-security",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Security = &api.NimServiceSpecAllOfSecurity{
				RunAsUid: convutils.Ref(int64(1000)),
				RunAsGid: convutils.Ref(int64(1000)),
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Security).ToNot(BeNil())

			// Use fieldsToRemove to remove security (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_SECURITY},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Security).To(BeNil())
		})

		It("should remove annotations using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-annotations",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Annotations = &[]api.Annotation{
				{Name: convutils.Ref("key1"), Value: convutils.Ref("value1")},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Annotations).ToNot(BeNil())

			// Use fieldsToRemove to remove annotations (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_ANNOTATIONS},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Annotations).To(BeNil())
		})

		It("should remove labels using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-labels",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Labels = &[]api.Label{
				{Name: convutils.Ref("label1"), Value: convutils.Ref("value1")},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Labels).ToNot(BeNil())

			// Use fieldsToRemove to remove labels (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_LABELS},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Labels).To(BeNil())
		})

		It("should remove environmentVariables using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-envvars",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.EnvironmentVariables = &[]api.EnvironmentVariable{
				{Name: convutils.Ref("VAR1"), Value: convutils.Ref("value1")},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.EnvironmentVariables).ToNot(BeNil())

			// Use fieldsToRemove to remove environmentVariables (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_ENVIRONMENT_VARIABLES},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.EnvironmentVariables).To(BeNil())
		})

		It("should remove nodePools using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-nodepools",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.NodePools = &[]string{"pool1", "pool2"}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.NodePools).ToNot(BeNil())

			// Use fieldsToRemove to remove nodePools (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_NODE_POOLS},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.NodePools).To(BeNil())
		})

		It("should remove imagePullSecrets using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-imagepullsecrets",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.ImagePullSecrets = &[]api.ImagePullSecret{{Name: "secret1"}}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.ImagePullSecrets).ToNot(BeNil())

			// Use fieldsToRemove to remove imagePullSecrets (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_IMAGE_PULL_SECRETS},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.ImagePullSecrets).To(BeNil())
		})

		It("should remove autoscaling using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-autoscaling-field",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Autoscaling = &api.NimServiceSpecAllOfAutoscaling{
				MinReplicas:     convutils.Ref(int32(1)),
				MaxReplicas:     convutils.Ref(int32(5)),
				Metric:          convutils.Ref("cpu"),
				MetricThreshold: convutils.Ref(int32(80)),
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Autoscaling).ToNot(BeNil())

			// Use fieldsToRemove to remove autoscaling (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec:           api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_AUTOSCALING},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Autoscaling).To(BeNil())
		})

		It("should remove multiple sections using fieldsToRemove", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-multiple",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Compute = &api.NimServiceSpecAllOfCompute{
				CpuCoreRequest: convutils.Ref(float64(2)),
			}
			toCreate.Spec.EnvironmentVariables = &[]api.EnvironmentVariable{
				{Name: convutils.Ref("VAR1"), Value: convutils.Ref("value1")},
			}
			toCreate.Spec.Annotations = &[]api.Annotation{
				{Name: convutils.Ref("key1"), Value: convutils.Ref("value1")},
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Compute).ToNot(BeNil())
			Expect(createdPayload.Spec.EnvironmentVariables).ToNot(BeNil())
			Expect(createdPayload.Spec.Annotations).ToNot(BeNil())

			// Remove all three using fieldsToRemove (must provide non-empty spec)
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{Image: sampleNimServiceSpec.Image},
				FieldsToRemove: []api.NimServiceRemovableFields{
					api.NIMSERVICEREMOVABLEFIELDS_COMPUTE,
					api.NIMSERVICEREMOVABLEFIELDS_ENVIRONMENT_VARIABLES,
					api.NIMSERVICEREMOVABLEFIELDS_ANNOTATIONS,
				},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Compute).To(BeNil())
			Expect(updatedPayload.Spec.EnvironmentVariables).To(BeNil())
			Expect(updatedPayload.Spec.Annotations).To(BeNil())
			// Other fields should remain unchanged
			Expect(updatedPayload.Spec.Image).To(Equal(sampleNimServiceSpec.Image))
			Expect(updatedPayload.Spec.ModelStore).ToNot(BeNil())
		})

		It("should allow removing fields while adding others in same update", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-remove-and-add",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}
			toCreate.Spec.Compute = &api.NimServiceSpecAllOfCompute{
				CpuCoreRequest: convutils.Ref(float64(2)),
			}

			id, createdPayload := createNimServiceForTest(router, randoms, &toCreate)
			Expect(createdPayload.Spec.Compute).ToNot(BeNil())
			Expect(createdPayload.Spec.EnvironmentVariables).To(BeNil())

			// Remove compute and add environment variables in a single update
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{Name: convutils.Ref("NEW_VAR"), Value: convutils.Ref("new_value")},
					},
				},
				FieldsToRemove: []api.NimServiceRemovableFields{api.NIMSERVICEREMOVABLEFIELDS_COMPUTE},
			}

			updatedPayload := updateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updatedPayload.Spec.Compute).To(BeNil())
			Expect(updatedPayload.Spec.EnvironmentVariables).ToNot(BeNil())
			Expect(len(*updatedPayload.Spec.EnvironmentVariables)).To(Equal(1))
			Expect(*(*updatedPayload.Spec.EnvironmentVariables)[0].Name).To(Equal("NEW_VAR"))
		})

		It("should return 404 when workload does not exist", func() {
			fakeId := "00000000-0000-0000-0000-000000000000"
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}

			updateResp := tryToUpdateNimServiceForTest(router, randoms, fakeId, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 when workload was not created via nim-services route", func() {
			// Create a regular workload (not via nim-services)
			regularWorkload := createRegularWorkloadForTest(randoms)

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}

			updateResp := tryToUpdateNimServiceForTest(router, randoms, regularWorkload.Metadata.Id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("workload was not created via nim-services route"))
		})

		It("should fail when updating with invalid priority", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-8",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("invalid"),
					Image:         convutils.Ref("nvidia/nim:v2"),
				},
			}

			updateResp := tryToUpdateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("priority 'invalid' is not available"))
		})

		It("should fail when updating with invalid category", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-9",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Category: convutils.Ref("invalid"),
					Image:    convutils.Ref("nvidia/nim:v2"),
				},
			}

			updateResp := tryToUpdateNimServiceForTest(router, randoms, id, &toUpdate)
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body.String()).To(ContainSubstring("category 'invalid' is not available"))
		})

		It("should handle multiple updates to same nim service", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-10",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			// First update
			toUpdate1 := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v2"),
				},
			}
			updatedPayload1 := updateNimServiceForTest(router, randoms, id, &toUpdate1)
			Expect(updatedPayload1.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:v2")))

			// Second update
			toUpdate2 := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					Image: convutils.Ref("nvidia/nim:v3"),
				},
			}
			updatedPayload2 := updateNimServiceForTest(router, randoms, id, &toUpdate2)
			Expect(updatedPayload2.Spec.Image).To(Equal(convutils.Ref("nvidia/nim:v3")))
		})

		It("should propagate updates to cluster", func() {
			toCreate := api.NimServiceCreateRequest{
				Metadata: api.NimServiceMetadataCreateParams{
					Name:      "nim-update-test-11",
					ProjectId: randoms.ProjectIdStr,
				},
				Spec: sampleNimServiceSpec,
			}

			id, _ := createNimServiceForTest(router, randoms, &toCreate)

			// Create a workloads router to check updates
			workloadsRouter := createRouterForTest()

			// Agent syncs creation (update count 1)
			updates := getUpdatesForTest(workloadsRouter, randoms, 0, 50, emptyClusterReports)
			Expect(updates.ResponseInfo.HighestUpdate).To(Equal(int64(1)))

			// Update the NIM service
			toUpdate := api.NimServiceUpdateRequest{
				Spec: api.NimServiceSpec{
					PriorityClass: convutils.Ref("medium"),
					Image:         convutils.Ref("nvidia/nim:v2"),
				},
			}
			updateNimServiceForTest(router, randoms, id, &toUpdate)

			// Agent receives update #2 (manifest + networking config in single update)
			updates = getUpdatesForTest(workloadsRouter, randoms, 1, 50, emptyClusterReports)
			Expect(updates.ResponseInfo.HighestUpdate).To(Equal(int64(2)))
			Expect(len(updates.Workloads)).To(Equal(1))
			Expect(updates.Workloads[0].UpdateCount).To(Equal(int64(2)))
			Expect(updates.Workloads[0].Id).To(Equal(id))
		})
	})
})

func createNimServicesRouterForTest(nimServicesImpl *NimServicesImpl) chi.Router {
	// Create a service that wraps our implementation
	nimServicesService := &NimServicesService{impl: nimServicesImpl}
	nimServicesController := api.NewNVIDIANIMAPIController(nimServicesService)

	router := api.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(runaictx.ContextMiddleware)
		// Add the request context middleware to capture request path and body
		router.Use(middleware.RequestContextMiddleware)

		api.AddRoutes(router, nimServicesController)
	})

	return router
}

func createNimServiceForTest(router chi.Router, randoms *testutils.RandomValues, request *api.NimServiceCreateRequest) (string, *api.NimServiceResponse) {
	createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(request))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	// Debug: print response if not expected status
	if createResp.Code != http.StatusAccepted {
		fmt.Printf("Expected 202, got %d. Response body: %s\n", createResp.Code, createResp.Body.String())
	}

	Expect(createResp.Code).To(Equal(http.StatusAccepted), "Response body: %s", createResp.Body.String())

	var result api.NimServiceResponse
	jsonErr := json.Unmarshal(createResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	id := sanitizeMeta(&result.Metadata)
	Expect(result.Metadata.DeletedAt).To(BeNil())

	return id, &result
}

func updateNimServiceForTest(router chi.Router, randoms *testutils.RandomValues, id string, request *api.NimServiceUpdateRequest) *api.NimServiceResponse {
	updateReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), makePayload(request))
	Expect(err).To(BeNil())
	updateReq.Header = *headerForAdmin(randoms)

	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	Expect(updateResp.Code).To(Equal(http.StatusAccepted), "Response body: %s", updateResp.Body.String())

	var result api.NimServiceResponse
	jsonErr := json.Unmarshal(updateResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	sanitizeMeta(&result.Metadata)

	return &result
}

func tryToUpdateNimServiceForTest(router chi.Router, randoms *testutils.RandomValues, id string, request *api.NimServiceUpdateRequest) *httptest.ResponseRecorder {
	updateReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s%s/%s", testHost, urlNimServices, id), makePayload(request))
	Expect(err).To(BeNil())
	updateReq.Header = *headerForAdmin(randoms)

	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)

	return updateResp
}

func failToCreateNimServiceForTest(router chi.Router, randoms *testutils.RandomValues, request *api.NimServiceCreateRequest, expectedError int) *api.SubmissionErrorV2 {
	createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlNimServices), makePayload(request))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	Expect(createResp.Code).To(Equal(expectedError))

	var result api.SubmissionErrorV2
	jsonErr := json.Unmarshal(createResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

// Helper function to create a regular workload (not via nim-services)
func createRegularWorkloadForTest(randoms *testutils.RandomValues) *api.WorkloadV2 {
	// Create a workload via HTTP request with different path to simulate regular workload
	manifest := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": "regular-pod",
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"name":  "test-container",
					"image": "nginx:latest",
				},
			},
		},
	}

	request := api.WorkloadV2CreateRequest{
		Metadata: api.WorkloadV2MetadataCreateParams{
			Name:      "regular-workload",
			ProjectId: randoms.ProjectIdStr,
		},
		Manifest: manifest,
	}

	// Create via regular workloads endpoint (not nim-services)
	workloadsRouter := createRouterForTest()
	createReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2/workloads", testHost), makePayload(request))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	workloadsRouter.ServeHTTP(createResp, createReq)
	Expect(createResp.Code).To(Equal(http.StatusAccepted))

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(createResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}
