/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"k8s.io/utils/ptr"

	clusters_metadata_cache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao/cache/clusters/metadata"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	pubsub "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	pubsubmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/mocks"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	cluster_installation "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation"
	cluster_installation_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation/mocks"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"

	"github.com/onsi/ginkgo/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	cluster_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cacheTesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"

	cd "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	cdmock "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"

	"github.com/pkg/errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	smocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services"
	csc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services/cluster-state-calculator"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	idm_client_api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	idm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/mocks"
	org_unit_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/mocks"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/mocks"
	tm_client_api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	tm_client_mock "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/mocks"

	ext_api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"

	"github.com/stretchr/testify/mock"
)

type Mocks struct {
	clusterStateCalculator      *csc_mocks.ClusterStateCalculator
	tenantSettingsClient        *bc_mocks.TenantSettingsClient
	idmClient                   *idm_client.IIdm
	orgUnitClient               *org_unit_client.IOrgUnitClient
	tmInternalApi               *tm_client_mock.InternalApi
	tmTenantsApi                *tm_client_mock.TenantsApi
	clusterDeleter              *cdmock.ClusterDeleter
	pubsub                      *pubsubmocks.PubSuber
	orgClient                   *org_unit_service_client_mock.IOrgUnitClient
	cacheMock                   *cacheTesting.MockCacher
	clusterAccessControlCreator *smocks.ClusterAccessControlManager
}

const (
	tenantId        = 1
	anotherTenantId = "619"
	fullUrl         = "tenant.mock-q7z3k.example.com"
	testSuite       = "cluster-service"
)

var (
	mocks              *Mocks
	clustersApiService api.ClustersAPIServicer
	ctx                context.Context

	clusters []api.DisplayedCluster
)

func GetApiClient() []idm_client_api.Client {
	attributeMap := make(map[string]interface{})

	attributeMap["runai.client_type"] = "workloads_auth"
	attributeMap["runai.cluster_id"] = "Any"
	clientArray := []idm_client_api.Client{
		{
			Id:         "Any",
			Attributes: attributeMap,
		},
	}
	return clientArray
}

func GetInstallerServiceAccount() *idm_client_api.ServiceAccountPostResponse {
	return &idm_client_api.ServiceAccountPostResponse{
		Id:           "Any",
		Name:         "ApplicationInstaller",
		ClientSecret: "Any_Secret",
	}
}

func GetInstallerServiceAccountForTenantId() *idm_client_api.ServiceAccountPostResponse {
	return &idm_client_api.ServiceAccountPostResponse{
		Id:           "Any",
		Name:         "ApplicationInstaller",
		ClientSecret: "Tenant_Secret",
	}
}

func prepareMocks() {
	prepareMocksExt(false)
}

func prepareMocksExt(mockClusterDeleter bool) {
	runai_config.Set(&config.Service{PubsubInitialized: true})
	mocks = &Mocks{
		clusterStateCalculator:      &csc_mocks.ClusterStateCalculator{},
		tenantSettingsClient:        &bc_mocks.TenantSettingsClient{},
		idmClient:                   &idm_client.IIdm{},
		orgUnitClient:               &org_unit_client.IOrgUnitClient{},
		tmInternalApi:               &tm_client_mock.InternalApi{},
		tmTenantsApi:                &tm_client_mock.TenantsApi{},
		clusterDeleter:              cdmock.NewClusterDeleter(ginkgo.GinkgoT()),
		pubsub:                      &pubsubmocks.PubSuber{},
		orgClient:                   &org_unit_service_client_mock.IOrgUnitClient{},
		clusterAccessControlCreator: &smocks.ClusterAccessControlManager{},
	}
	pubsub.SetMock(mocks.pubsub)
	mocks.clusterStateCalculator.On("CalculateState", mock.Anything).Return(string(api.CLUSTERSTATE_CONNECTED))
	mocks.pubsub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mocks.orgUnitClient.On("CountDepartments", mock.Anything, mock.Anything).Return(2, nil)
	mocks.orgUnitClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mocks.clusterAccessControlCreator.On("CreateAccessControl", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tenantIdStr := strconv.Itoa(tenantId)
	ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "user-1")
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{tenantIdStr},
	})
	ctx = addAuditToCtx(ctx)

	var clusterDeleter cd.ClusterDeleter
	if mockClusterDeleter {
		clusterDeleter = mocks.clusterDeleter
	} else {
		clusterDeleter = cd.NewClusterDeleter(authorization_client.NewAuthorizationClient(), mocks.orgUnitClient, nil)
	}

	helmActions := &cluster_installation_mocks.HelmActions{}
	clusterInstaller := cluster_installation.NewClusterInstallService(helmActions)
	// Setup service
	nodepoolService := services.NewNodePoolsApiService(mocks.tenantSettingsClient, mocks.orgUnitClient)
	clustersApiService = services.NewClustersApiService(mocks.clusterStateCalculator, mocks.idmClient,
		clusterDeleter,
		clusterInstaller,
		&tm_client_api.APIClient{
			InternalAPI: mocks.tmInternalApi,
			TenantsAPI:  mocks.tmTenantsApi,
		}, nodepoolService,
		mocks.orgUnitClient,
		mocks.clusterAccessControlCreator)

	mocks.orgUnitClient.On("ClusterDeleted", mock.Anything, mock.Anything).Return(nil)

	// Setup empty clusters list
	clusters = []api.DisplayedCluster{}
}

var _ = Describe("ApiClustersService", Ordered, func() {
	var allNodepoolsBefore map[int32]bool

	BeforeAll(func() {
		// get all nodepools from before
		nodepools, err := dao.GetNodepools(context.Background(), int32(tenantId), authorization_api.PermittedScopes{System: ptr.To(true)}, 0, 0, "", "id", nil)
		Expect(err).ShouldNot(HaveOccurred())
		for _, np := range nodepools {
			allNodepoolsBefore[np.ID] = true
		}
	})

	AfterAll(func() {
		// get all nodepools after tests
		nodepools, err := dao.GetNodepools(context.Background(), int32(tenantId), authorization_api.PermittedScopes{System: ptr.To(true)}, 0, 0, "", "id", nil)
		Expect(err).ShouldNot(HaveOccurred())

		// delete all new nodepools created during tests
		for _, np := range nodepools {
			if _, exists := allNodepoolsBefore[np.ID]; !exists {
				err := dao.DeleteNodepool(context.Background(), tenantId, "", strconv.Itoa(int(np.ID)))
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	})

	Describe("Service tests", func() {
		BeforeEach(func() {
			err := cache.ClearCache()
			Expect(err).To(BeNil())
			// Setup empty mocks
			prepareMocks()
		})
		AfterEach(func() {
			// Returning to zero state
			deleteAllClusters(ctx, clustersApiService)
		})

		Describe("GetClustersCount", Ordered, func() {
			var otherCtx context.Context
			BeforeEach(func() {
				otherCtx = context.WithValue(context.Background(), runaictx.CfTenantId, anotherTenantId)
				otherCtx = context.WithValue(otherCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{anotherTenantId},
				})
				otherCtx = addAuditToCtx(otherCtx)

				prepareMocks()
				// Need to clean db before each test to ensure there are no clusters with/out subdomains
				// that mess up the expected count
				deleteCtxClusters(otherCtx, clustersApiService)
			})

			expectClustersCount := func(actual []api.ClustersCount, expected []api.ClustersCount) {
				Expect(actual[0]).To(Equal(expected[0]))
				Expect(actual[1]).To(Equal(expected[1]))
			}

			It("Should return zero counters when there are no clusters", func() {
				expectedCount := []api.ClustersCount{
					{Key: "true", Count: 0},
					{Key: "false", Count: 0},
					{Key: "unknown", Count: 0},
				}
				resp, err := clustersApiService.GetClustersCount(otherCtx, ext_api.GroupBySubdomain)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusOK))
				clusterCount, ok := resp.Body.([]api.ClustersCount)
				Expect(ok).To(BeTrue())
				expectClustersCount(clusterCount, expectedCount)
			})

			It("Should fail when using invalid group by", func() {
				resp, err := clustersApiService.GetClustersCount(otherCtx, "invalid")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
			})

			It("Successful counting", func() {
				resp, err := clustersApiService.GetClustersCount(otherCtx, ext_api.GroupBySubdomain)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusOK))
				clusterCount, ok := resp.Body.([]api.ClustersCount)
				Expect(ok).To(BeTrue())

				expectedCount := []api.ClustersCount{
					{Key: "true", Count: 2},
					{Key: "false", Count: 3},
					{Key: "unknown", Count: 1},
				}
				mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
				mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// create clusters with subdomains configuration
				clustersSubdomain := []bool{false, true, false, true, false}
				for i, subdomain := range clustersSubdomain {
					cluster := createCluster(otherCtx, clustersApiService, clusterNameGenerator.Generate(testSuite), fmt.Sprintf("cluster-%v", i), "")
					updateClusterSubdomain(otherCtx, cluster, subdomain)
				}
				// add one more cluster without status
				createCluster(otherCtx, clustersApiService, clusterNameGenerator.Generate(testSuite), fmt.Sprintf("cluster-%v", len(clustersSubdomain)), "")

				countResp, err := clustersApiService.GetClustersCount(otherCtx, ext_api.GroupBySubdomain)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(countResp.Code).To(Equal(http.StatusOK))
				clusterCount, ok = countResp.Body.([]api.ClustersCount)
				Expect(ok).To(BeTrue())
				expectClustersCount(clusterCount, expectedCount)
			})
		})
		Describe("CreateCluster", func() {
			It("Should create the cluster correctly", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
				}

				Expect(clusters[0].Name).To(Equal(cluster1Name))

				// Check that the cluster was created
				resp, err := clustersApiService.GetClusters(ctx, "full", []string{}, false)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				Expect(resp.Body).To(HaveLen(1))

				displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedClusters).To(HaveLen(1))
				Expect(displayedClusters[0].Name).To(Equal(cluster1Name))
				Expect(*displayedClusters[0].Domain).To(Equal("cluster-1"))
			})

			It("Should create the cluster correctly ignoring tenant ID", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createClusterWithTenantId(ctx, clustersApiService, cluster1Name, "cluster-1", "", 1009),
				}
				Expect(clusters[0].Name).To(Equal(cluster1Name))
				// Check that the cluster was created
				resp, err := clustersApiService.GetClusters(ctx, "full", []string{}, false)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				Expect(resp.Body).To(HaveLen(1))

				displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedClusters).To(HaveLen(1))
				Expect(displayedClusters[0].Name).To(Equal(cluster1Name))
				Expect(*displayedClusters[0].Domain).To(Equal("cluster-1"))
				tenantId, apiErr := services.ParseTenantId(ctx)
				Expect(apiErr).To(BeNil())
				Expect(displayedClusters[0].TenantId).To(Equal(int32(tenantId)))
			})

			It("Should fail when CSP Cloud Operator is trying to create a cluster for with an invalid tenant ID", func() {
				mocks.tmTenantsApi.On("GetTenantByIdInternal", mock.Anything, mock.Anything).Return(tm_client_api.ApiGetTenantByIdInternalRequest{})
				mocks.tmTenantsApi.On("GetTenantByIdInternalExecute", mock.Anything).Return(nil, &http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				newCtx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
				newCtx = addAuditToCtx(newCtx)
				resp, _ := clustersApiService.CreateCluster(newCtx, api.ClusterCreationRequest{
					Name:     cluster1Name,
					Domain:   "cluster-1",
					Version:  "",
					TenantId: 1009,
				})
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
			})

			It("Should succeed when CSP Cloud Operator is creating tenant using a valid tenant ID", func() {
				var createdTenantId int32 = 1009
				myTrue := true
				mocks.tmTenantsApi.On("GetTenantByIdInternal", mock.Anything, mock.Anything).Return(tm_client_api.ApiGetTenantByIdInternalRequest{})
				mocks.tmTenantsApi.On("GetTenantByIdInternalExecute", mock.Anything).
					Return(&tm_client_api.Tenant{
						Id: int32(createdTenantId),
					}, nil, nil).Maybe()
				newCtx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
				newCtx = context.WithValue(newCtx, runaictx.CfTenantId, "1")
				newCtx = context.WithValue(newCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{"1"},
					System:  &myTrue,
				})
				newCtx = addAuditToCtx(newCtx)

				cluster1Name := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createClusterWithTenantId(newCtx, clustersApiService, cluster1Name, "cluster-1", "", createdTenantId),
				}
				Expect(clusters[0].Name).To(Equal(cluster1Name))
				// Check that the cluster was created
				resp, err := clustersApiService.GetClusters(newCtx, "full", []string{}, false)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				found := false
				for _, cluster := range displayedClusters {
					if cluster.Name == cluster1Name {
						Expect(cluster.TenantId).To(Equal(int32(createdTenantId)))
						found = true
					}
				}
				Expect(found).To(BeTrue())
			})

			It("Should fail when creating cluster with same name for tenant", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				cluster, _ := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusCreated))

				cluster, _ = clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusConflict))
			})
			It("Should succeed when cluster with the same name is deleted", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				cluster, _ := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusCreated))
				uuid := cluster.Body.(api.DisplayedCluster).Uuid
				_ = dao.DeleteCluster(ctx, uuid, tenantId)
				cluster, _ = clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusCreated))
			})
			It("Should succeed  when creating cluster with same name for different tenant", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				cluster, _ := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusCreated))
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				randIdValues := rng.Intn(60) + 2000
				tenantIdStr := strconv.Itoa(randIdValues)
				newTenantCtx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
				newTenantCtx = context.WithValue(newTenantCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{tenantIdStr},
				})
				newTenantCtx = addAuditToCtx(newTenantCtx)
				cluster, _ = clustersApiService.CreateCluster(newTenantCtx, api.ClusterCreationRequest{
					Name:   cluster1Name,
					Domain: "myDomain",
				})
				Expect(cluster.Code).To(Equal(http.StatusCreated))
			})

			It("CreateCluster - validate default nodepool is created as expected with default values", func() {
				cluster1Name := clusterNameGenerator.Generate("default-nodepool")
				clusterResp, err := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
					Name:     cluster1Name,
					Domain:   "cluster-1",
					TenantId: tenantId,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(clusterResp.Code).To(Equal(http.StatusCreated))
				createdCluster := clusterResp.Body.(api.DisplayedCluster)

				// Validate in DB that default nodepool is created correctly
				nodepools, err := dao.GetNodepools(ctx, tenantId, *mustPermittedScopes(createdCluster.Uuid), 0, 0, "", "id", nil)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(nodepools).To(HaveLen(1))
				defaultNodepool := nodepools[0]
				Expect(defaultNodepool.Name).To(Equal(services.NodepoolDefaultName))
				Expect(defaultNodepool.AllowAdjustableWeights).To((BeFalse()))
				Expect(defaultNodepool.AllowQuota).To((BeTrue()))
				Expect(defaultNodepool.AllowRanks).To((BeTrue()))
				Expect(defaultNodepool.AllowCpuQuota).To((BeFalse()))
			})
		})

		Describe("UpdateCluster", func() {
			createdTenantId := int32(1010)
			It("Should update the cluster correctly", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				clusterUpdateName := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
				}
				resp, err := clustersApiService.UpdateCluster(ctx, clusters[0].Uuid, api.ClusterUpdateRequest{
					Name: clusterUpdateName,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				// Check that the cluster was updated
				resp, err = clustersApiService.GetClusterByUuid(ctx, clusters[0].Uuid, "full")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedCluster.Name).To(Equal(clusterUpdateName))
			})
			It("Should fail when updating cluster with name that exists name for tenant", func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				cluster2Name := clusterNameGenerator.Generate(testSuite)

				clusters = []api.DisplayedCluster{
					createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
					createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", ""),
				}
				resp, _ := clustersApiService.UpdateCluster(ctx, clusters[0].Uuid, api.ClusterUpdateRequest{
					Name: cluster2Name,
				})

				Expect(resp.Code).To(Equal(http.StatusConflict))
			})
			It("CSP Should update the cluster correctly", func() {
				mocks.tmTenantsApi.On("GetTenantByIdInternal", mock.Anything, mock.Anything).Return(tm_client_api.ApiGetTenantByIdInternalRequest{})
				mocks.tmTenantsApi.On("GetTenantByIdInternalExecute", mock.Anything).
					Return(&tm_client_api.Tenant{
						Id: int32(createdTenantId),
					}, nil, nil).Maybe()

				myTrue := true
				newCtx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
				newCtx = context.WithValue(newCtx, runaictx.CfTenantId, "1")
				newCtx = context.WithValue(newCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{"1"},
					System:  &myTrue,
				})
				newCtx = addAuditToCtx(newCtx)
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				clusterUpdateName := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createClusterWithTenantId(newCtx, clustersApiService, cluster1Name, "cluster-1", "", createdTenantId),
				}
				resp, err := clustersApiService.UpdateCluster(newCtx, clusters[0].Uuid, api.ClusterUpdateRequest{
					Name: clusterUpdateName,
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				// Check that the cluster was updated
				resp, err = clustersApiService.GetClusterByUuid(newCtx, clusters[0].Uuid, "full")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedCluster.Name).To(Equal(clusterUpdateName))
			})
		})

		Describe("UpdateClusterLiveness", func() {
			Context("Happy flow", func() {
				It("Should update the cluster liveness correctly", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
					}
					now := time.Now()
					resp, err := clustersApiService.UpdateClusterLiveness(ctx, clusters[0].Uuid, api.ClusterLivenessSyncRequest{
						TimeTag: now.UnixMilli(),
					})
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(204))

					// Check that the cluster was updated
					resp, err = clustersApiService.GetClusterByUuid(ctx, clusters[0].Uuid, "full")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(200))

					displayedCluster, ok := resp.Body.(api.DisplayedCluster)
					Expect(ok).To(BeTrue())
					Expect(displayedCluster.LastLiveness.Unix()).To(Equal(now.Unix()))
				})
			})

			Context("Bad flow", func() {
				Context("When the cluster does not exist", func() {
					It("Should return 204", func() {
						nonExistingUuid := uuid.New().String()
						resp, err := clustersApiService.UpdateClusterLiveness(ctx, nonExistingUuid, api.ClusterLivenessSyncRequest{})
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resp.Code).To(Equal(204))
					})
				})
			})
		})

		Describe("UpdateClusterInfo", func() {
			BeforeEach(func() {
				cluster1Name := clusterNameGenerator.Generate(testSuite)
				cluster2Name := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
					createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", ""),
				}
			})
			It("Should fail update the cluster info when updating internal workload clients fails", func() {
				now := time.Now()
				err := errors.Errorf("Error getting clients")
				mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
				mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(err)

				resp, err := clustersApiService.UpdateClusterInfo(ctx, clusters[0].Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand-a": {
								Ready:                   true,
								UnreadyThresholdCrossed: false,
								LastTransitionTime:      now,
								Reasons:                 []string{},
							},
						},
						Dependencies: api.ClusterDependenciesStatus{
							Required: map[string]api.ClusterDependencyStatus{
								"required-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"required-component": {
											Available: true,
										},
									},
								},
							},
							Optional: map[string]api.ClusterDependencyStatus{
								"optional-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"optional-component": {
											Available: true,
										},
									},
								},
							},
						},
					},
					Domain:  "cluster-1",
					Version: "2.20",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(500))
				mocks.idmClient.AssertCalled(GinkgoT(), "GetInternalClientByFilter", mock.Anything, mock.Anything)
				mocks.idmClient.AssertCalled(GinkgoT(), "AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything)
			})
			It("Should update the cluster info without updating internal workload clients", func() {
				now := time.Now()
				resp, err := clustersApiService.UpdateClusterInfo(ctx, clusters[0].Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand-a": {
								Ready:                   true,
								UnreadyThresholdCrossed: false,
								LastTransitionTime:      now,
								Reasons:                 []string{},
							},
						},
						Dependencies: api.ClusterDependenciesStatus{
							Required: map[string]api.ClusterDependencyStatus{
								"required-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"required-component": {
											Available: true,
										},
									},
								},
							},
							Optional: map[string]api.ClusterDependencyStatus{
								"optional-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"optional-component": {
											Available: true,
										},
									},
								},
							},
						},
					},
					Domain:  "cluster-1",
					Version: "",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))
				mocks.idmClient.AssertNotCalled(GinkgoT(), "GetInternalClientByFilter", mock.Anything, mock.Anything)
				mocks.idmClient.AssertNotCalled(GinkgoT(), "AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything)
			})
			It("Should update the cluster info correctly", func() {
				mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
				mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				now := time.Now()
				resp, err := clustersApiService.UpdateClusterInfo(ctx, clusters[0].Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand-a": {
								Ready:                   true,
								UnreadyThresholdCrossed: false,
								LastTransitionTime:      now,
								Reasons:                 []string{},
							},
						},
						Dependencies: api.ClusterDependenciesStatus{
							Required: map[string]api.ClusterDependencyStatus{
								"required-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"required-component": {
											Available: true,
										},
									},
								},
							},
							Optional: map[string]api.ClusterDependencyStatus{
								"optional-dependency": {
									Available: true,
									Reason:    "",
									Components: map[string]api.ClusterDependencyStatus{
										"optional-component": {
											Available: true,
										},
									},
								},
							},
						},
					},
					Domain:  "cluster-1",
					Version: "1.0.0",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				resp, err = clustersApiService.UpdateClusterInfo(ctx, clusters[1].Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand-a": {
								Ready: false,
							},
						},
					},
					Domain:  "cluster-2",
					Version: "1.0.0",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				// Check that the clusters were updated
				resp, err = clustersApiService.GetClusters(ctx, "full", []string{}, false)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedClusters).To(HaveLen(2))

				clusterMap := make(map[string]api.DisplayedCluster)
				for _, cluster := range displayedClusters {
					clusterMap[cluster.Name] = cluster
				}

				Expect(clusterMap).To(HaveKey(clusters[0].Name))
				Expect(clusterMap[clusters[0].Name].Status.Operands).To(HaveLen(1))
				Expect(clusterMap[clusters[0].Name].Status.Operands["operand-a"].Ready).To(BeTrue())
				Expect(clusterMap[clusters[0].Name].Status.Operands["operand-a"].UnreadyThresholdCrossed).To(BeFalse())
				Expect(clusterMap[clusters[0].Name].Status.Operands["operand-a"].LastTransitionTime.Unix()).To(Equal(now.Unix()))
				Expect(clusterMap[clusters[0].Name].Status.Operands["operand-a"].Reasons).To(BeEmpty())

				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Required).To(HaveLen(1))
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Required["required-dependency"].Available).To(BeTrue())
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Required["required-dependency"].Reason).To(BeEmpty())
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Required["required-dependency"].Components).To(HaveLen(1))
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Required["required-dependency"].Components["required-component"].Available).To(BeTrue())

				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Optional).To(HaveLen(1))
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Optional["optional-dependency"].Available).To(BeTrue())
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Optional["optional-dependency"].Reason).To(BeEmpty())
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Optional["optional-dependency"].Components).To(HaveLen(1))
				Expect(clusterMap[clusters[0].Name].Status.Dependencies.Optional["optional-dependency"].Components["optional-component"].Available).To(BeTrue())

				Expect(clusterMap).To(HaveKey(clusters[1].Name))
				Expect(clusterMap[clusters[1].Name].Status.Operands).To(HaveLen(1))
				Expect(clusterMap[clusters[1].Name].Status.Operands["operand-a"].Ready).To(BeFalse())
				Expect(clusterMap[clusters[1].Name].Status.Operands["operand-a"].UnreadyThresholdCrossed).To(BeFalse())
				Expect(clusterMap[clusters[1].Name].Status.Operands["operand-a"].LastTransitionTime.IsZero()).To(BeTrue())
				Expect(clusterMap[clusters[1].Name].Status.Operands["operand-a"].Reasons).To(BeEmpty())

				Expect(clusterMap[clusters[1].Name].Status.Dependencies.Required).To(BeEmpty())
				Expect(clusterMap[clusters[1].Name].Status.Dependencies.Optional).To(BeEmpty())
				mocks.idmClient.AssertCalled(GinkgoT(), "GetInternalClientByFilter", mock.Anything, mock.Anything)
				mocks.idmClient.AssertCalled(GinkgoT(), "AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything)
			})

			Context("Backward compatibility", func() {
				When("The cluster dependencies status is not provided under the status field", func() {
					It("Should update the cluster info using the deprecated dependencies field", func() {
						cluster1Name := clusterNameGenerator.Generate(testSuite)
						mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
						mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)
						clusters = []api.DisplayedCluster{
							createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
						}
						mocks.idmClient.On("UpdateInternalWorkloadClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)
						resp, err := clustersApiService.UpdateClusterInfo(ctx, clusters[0].Uuid, api.ClusterInfoSyncRequest{
							Dependencies: &api.ClusterDependenciesStatus{
								Required: map[string]api.ClusterDependencyStatus{
									"required-dependency": {
										Available: true,
										Reason:    "",
										Components: map[string]api.ClusterDependencyStatus{
											"required-component": {
												Available: true,
											},
										},
									},
								},
							},
						})
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resp.Code).To(Equal(204))

						// Check that the cluster was updated
						resp, err = clustersApiService.GetClusterByUuid(ctx, clusters[0].Uuid, "full")
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resp.Code).To(Equal(200))

						displayedCluster, ok := resp.Body.(api.DisplayedCluster)
						Expect(ok).To(BeTrue())

						Expect(displayedCluster.Status.Dependencies.Required).To(HaveLen(1))
						Expect(displayedCluster.Status.Dependencies.Required["required-dependency"].Available).To(BeTrue())
						Expect(displayedCluster.Status.Dependencies.Required["required-dependency"].Reason).To(BeEmpty())
						Expect(displayedCluster.Status.Dependencies.Required["required-dependency"].Components).To(HaveLen(1))
						Expect(displayedCluster.Status.Dependencies.Required["required-dependency"].Components["required-component"].Available).To(BeTrue())

						Expect(displayedCluster.Status.Dependencies.Optional).To(BeEmpty())
					})
				})
			})
		})

		Describe("GetClusterInstallInfoByUuid", Ordered, func() {
			BeforeEach(func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				clusters = []api.DisplayedCluster{
					createCluster(ctx, clustersApiService, clusterName, "cluster-1", ""),
				}
			})

			It("should return correct installation information with cluster url", func() {
				mocks.idmClient.On("MintInstallerBootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return("Any_Secret", nil)

				curCtx := context.WithValue(ctx, utils.FullHostKey, fullUrl)
				resp, err := clustersApiService.GetClusterInstallInfoByUuid(curCtx, clusters[0].Uuid, "2.16.3", "https://test-url.runai")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				installationInfo, ok := resp.Body.(api.ClusterInstallationInfoResponse)
				expectedInstallationStr := `helm repo add runai  --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=` + fullUrl + ` \
--set controlPlane.bootstrapToken=Any_Secret \
--set cluster.uid=` + clusters[0].Uuid + ` \
--set cluster.url=https://test-url.runai --version="2.16.3" --create-namespace`
				Expect(ok).To(BeTrue())
				Expect(installationInfo.InstallationStr).To(Equal(expectedInstallationStr))
				Expect(installationInfo.RepositoryName).To(Equal("runai"))
				Expect(installationInfo.BootstrapToken).To(Equal("Any_Secret"))
			})

			It("should return correct installation information with null cluster url and no hostFullName in context", func() {
				mocks.idmClient.On("MintInstallerBootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return("Any_Secret", nil)

				resp, err := clustersApiService.GetClusterInstallInfoByUuid(ctx, clusters[0].Uuid, "2.17.2", "")

				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				appConfig := runai_config.Get[*config.App]()
				expectedInstallationStr := `helm repo add runai  --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=` + appConfig.Host + ` \
--set controlPlane.bootstrapToken=Any_Secret \
--set cluster.uid=` + clusters[0].Uuid + ` \
--set cluster.url=` + appConfig.Host + ` --version="2.17.2" --create-namespace`
				installationInfo, ok := resp.Body.(api.ClusterInstallationInfoResponse)
				Expect(ok).To(BeTrue())
				Expect(installationInfo.InstallationStr).To(Equal(expectedInstallationStr))
				Expect(installationInfo.RepositoryName).To(Equal("runai"))
				Expect(installationInfo.BootstrapToken).To(Equal("Any_Secret"))
			})

			It("should return correct installation information for latest env", func() {
				mocks.idmClient.On("MintInstallerBootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return("Any_Secret", nil)
				fullUrl := "tenant.mock-q7z3k.example.com"
				curCtx := context.WithValue(ctx, utils.FullHostKey, fullUrl)
				resp, err := clustersApiService.GetClusterInstallInfoByUuid(curCtx, clusters[0].Uuid, ">0.0.0-0", "")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				expectedInstallationStr := `helm repo add runai  --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=` + fullUrl + ` \
--set controlPlane.bootstrapToken=Any_Secret \
--set cluster.uid=` + clusters[0].Uuid + ` \
--set cluster.url=` + fullUrl + ` --version=">0.0.0-0" --create-namespace`
				installationInfo, ok := resp.Body.(api.ClusterInstallationInfoResponse)
				Expect(ok).To(BeTrue())
				Expect(installationInfo.InstallationStr).To(Equal(expectedInstallationStr))
				Expect(installationInfo.RepositoryName).To(Equal("runai"))
				Expect(installationInfo.BootstrapToken).To(Equal("Any_Secret"))
			})

			It("failure when no client returned with error", func() {
				mocks.idmClient.On("MintInstallerBootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("error"))

				// Act
				resp, err := clustersApiService.GetClusterInstallInfoByUuid(ctx, clusters[0].Uuid, "test-version", "")

				// Assert
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(500))
			})

			It("CSP should return correct installation information with cluster url", func() {
				mocks.idmClient.On("MintInstallerBootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return("Tenant_Secret", nil)
				mocks.tmInternalApi.On("GetTenantDomain", mock.Anything, mock.Anything).Return(tm_client_api.ApiGetTenantDomainRequest{})
				mocks.tmInternalApi.On("GetTenantDomainExecute", mock.Anything).Return("tenant.domain", nil, nil)

				myTrue := true
				curCtx := context.WithValue(ctx, utils.FullHostKey, fullUrl)
				curCtx = context.WithValue(curCtx, runaictx.CfIsManagementTenant, "true")
				curCtx = context.WithValue(curCtx, runaictx.CfTenantId, "1")
				curCtx = context.WithValue(curCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{"1"},
					System:  &myTrue,
				})

				resp, err := clustersApiService.GetClusterInstallInfoByUuid(curCtx, clusters[0].Uuid, "2.16.3", "https://test-url.runai")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				installationInfo, ok := resp.Body.(api.ClusterInstallationInfoResponse)
				expectedInstallationStr := `helm repo add runai  --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=tenant.domain \
--set controlPlane.bootstrapToken=Tenant_Secret \
--set cluster.uid=` + clusters[0].Uuid + ` \
--set cluster.url=https://test-url.runai --version="2.16.3" --create-namespace`
				Expect(ok).To(BeTrue())
				Expect(installationInfo.InstallationStr).To(Equal(expectedInstallationStr))
				Expect(installationInfo.RepositoryName).To(Equal("runai"))
				Expect(installationInfo.BootstrapToken).To(Equal("Tenant_Secret"))
			})
		})

		Describe("GetClusters", func() {
			Context("When there are no clusters", func() {
				It("Should return empty list", func() {
					resp, err := clustersApiService.GetClusters(ctx, "full", []string{}, false)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(200))
					Expect(resp.Body).To(BeEmpty())
				})
			})

			Context("When there are clusters", func() {
				It("Should return the clusters correctly", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					cluster2Name := clusterNameGenerator.Generate(testSuite)
					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
						createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", ""),
					}
					resp, err := clustersApiService.GetClusters(ctx, "full", []string{}, false)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(200))
					Expect(resp.Body).To(HaveLen(2))

					displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
					Expect(ok).To(BeTrue())
					Expect(displayedClusters).To(HaveLen(2))

					clusterNames := make(map[string]bool)
					clusterDomains := make(map[string]bool)

					for _, cluster := range displayedClusters {
						clusterNames[cluster.Name] = true
						clusterDomains[*cluster.Domain] = true
					}

					Expect(clusterNames).To(HaveKey(cluster1Name))
					Expect(clusterNames).To(HaveKey(cluster2Name))
					Expect(clusterDomains).To(HaveKey("cluster-1"))
					Expect(clusterDomains).To(HaveKey("cluster-2"))
				})

				Context("When includeRequestedForDelete is true", func() {
					It("Should return all clusters including those marked for deletion", func() {
						// Enable async cluster deletion
						runai_config.Set(&config.Service{PubsubInitialized: true})
						runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: true})

						prepareMocks()

						deleteCtxClusters(ctx, clustersApiService)
						cluster1Name := clusterNameGenerator.Generate(testSuite)
						cluster2Name := clusterNameGenerator.Generate(testSuite)
						clusters = []api.DisplayedCluster{
							createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
							createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", ""),
						}

						// Mark one cluster for deletion
						resp, err := clustersApiService.DeleteCluster(ctx, clusters[0].Uuid, true)
						Expect(err).ShouldNot(HaveOccurred())

						// Get clusters with includeRequestedForDelete=true
						resp, err = clustersApiService.GetClusters(ctx, "full", []string{}, true)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resp.Code).To(Equal(200))
						Expect(resp.Body).To(HaveLen(2))

						displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
						Expect(ok).To(BeTrue())
						Expect(displayedClusters).To(HaveLen(2))

						clusterNames := make(map[string]bool)
						clusterDomains := make(map[string]bool)

						for _, cluster := range displayedClusters {
							clusterNames[cluster.Name] = true
							clusterDomains[*cluster.Domain] = true
						}

						Expect(clusterNames).To(HaveKey(cluster1Name))
						Expect(clusterNames).To(HaveKey(cluster2Name))
						Expect(clusterDomains).To(HaveKey("cluster-1"))
						Expect(clusterDomains).To(HaveKey("cluster-2"))
					})

					It("Should return clusters that are not marked for deletion", func() {
						deleteCtxClusters(ctx, clustersApiService)
						cluster1Name := clusterNameGenerator.Generate(testSuite)
						cluster2Name := clusterNameGenerator.Generate(testSuite)
						clusters = []api.DisplayedCluster{
							createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
							createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", ""),
						}

						// Mark one cluster for deletion
						resp, err := clustersApiService.DeleteCluster(ctx, clusters[0].Uuid, false)
						Expect(err).ShouldNot(HaveOccurred())

						// Get clusters with includeRequestedForDelete=false
						resp, err = clustersApiService.GetClusters(ctx, "full", []string{}, false)
						Expect(err).ShouldNot(HaveOccurred())
						Expect(resp.Code).To(Equal(200))
						Expect(resp.Body).To(HaveLen(1))

						displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
						Expect(ok).To(BeTrue())
						Expect(displayedClusters).To(HaveLen(1))
						Expect(displayedClusters[0].Name).To(Equal(cluster2Name))
						Expect(*displayedClusters[0].Domain).To(Equal("cluster-2"))
					})
				})
			})
		})

		Describe("GetClustersMetadata", func() {
			Context("When there are no clusters", func() {
				It("Should return empty list", func() {
					resp, err := clustersApiService.GetClustersMinimalData(ctx, []string{}, 0, 0)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(200))

					payload := resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(0))
				})
			})

			Context("When there are clusters", func() {
				It("Should return the clusters correctly - pagination", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					cluster2Name := clusterNameGenerator.Generate(testSuite)
					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "1.1"),
						createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", "2.2"),
					}
					resp, err := clustersApiService.GetClustersMinimalData(ctx, []string{}, 0, 0)
					Expect(err).To(BeNil())

					payload := resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(2))

					clusterVersions := make(map[string]string)

					for _, cluster := range payload.Clusters {
						clusterVersions[cluster.Name] = convutils.Deref(cluster.Version, "")
					}

					Expect(clusterVersions).To(HaveKey(cluster1Name))
					Expect(clusterVersions).To(HaveKey(cluster2Name))
					Expect(clusterVersions[cluster1Name]).To(Equal("1.1"))
					Expect(clusterVersions[cluster2Name]).To(Equal("2.2"))

					//
					// get them page by page
					//
					versionsPaged := make(map[string]string)

					resp, err = clustersApiService.GetClustersMinimalData(ctx, []string{}, 0, 1)
					Expect(err).To(BeNil())

					payload = resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(1))
					versionsPaged[payload.Clusters[0].Name] = convutils.Deref(payload.Clusters[0].Version, "")

					resp, err = clustersApiService.GetClustersMinimalData(ctx, []string{}, 1, 100)
					Expect(err).To(BeNil())
					payload = resp.Body.(api.GetClustersMinimalData200Response)

					Expect(payload.Clusters).To(HaveLen(1))
					versionsPaged[payload.Clusters[0].Name] = convutils.Deref(payload.Clusters[0].Version, "")

					Expect(versionsPaged).To(Equal(clusterVersions))
				})
			})

			Context("When filtering by name", func() {
				It("Should successfully filter cluster by name when there are multiple clusters", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					cluster2Name := clusterNameGenerator.Generate(testSuite)
					cluster3Name := clusterNameGenerator.Generate(testSuite)

					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "1.1"),
						createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", "2.2"),
						createCluster(ctx, clustersApiService, cluster3Name, "cluster-3", "3.3"),
					}
					updateClusterIngressClass(ctx, clusters[1], "nginx")

					// Filter by exact name match
					filterBy := []string{fmt.Sprintf("name==%s", cluster2Name)}
					resp, err := clustersApiService.GetClustersMinimalData(ctx, filterBy, 0, 0)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))

					payload := resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(1))
					Expect(payload.Clusters[0].Name).To(Equal(cluster2Name))
					Expect(convutils.Deref(payload.Clusters[0].Version, "")).To(Equal("2.2"))
					Expect(convutils.Deref(payload.Clusters[0].Domain, "")).To(Equal("cluster-2"))
					Expect(payload.Clusters[0].IngressClass).ToNot(BeNil())
					Expect(*payload.Clusters[0].IngressClass).To(Equal("nginx"))
				})

				It("Should return empty list when filtering by non-existing name", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					cluster2Name := clusterNameGenerator.Generate(testSuite)

					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "1.1"),
						createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", "2.2"),
					}

					// Filter by non-existing name
					nonExistingName := "non-existing-cluster-name"
					filterBy := []string{fmt.Sprintf("name==%s", nonExistingName)}
					resp, err := clustersApiService.GetClustersMinimalData(ctx, filterBy, 0, 0)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))

					payload := resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(0))
				})

				It("Should filter clusters using contains operator", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite) + "-prod"
					cluster2Name := clusterNameGenerator.Generate(testSuite) + "-dev"
					cluster3Name := clusterNameGenerator.Generate(testSuite) + "-prod"

					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "1.1"),
						createCluster(ctx, clustersApiService, cluster2Name, "cluster-2", "2.2"),
						createCluster(ctx, clustersApiService, cluster3Name, "cluster-3", "3.3"),
					}

					// Filter by name containing "prod"
					filterBy := []string{"name=@prod"}
					resp, err := clustersApiService.GetClustersMinimalData(ctx, filterBy, 0, 0)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))

					payload := resp.Body.(api.GetClustersMinimalData200Response)
					Expect(payload.Clusters).To(HaveLen(2))

					// Verify all returned clusters have "prod" in their name
					for _, cluster := range payload.Clusters {
						Expect(cluster.Name).To(ContainSubstring("prod"))
					}
				})

				It("Should return bad request for invalid filter format", func() {
					cluster1Name := clusterNameGenerator.Generate(testSuite)
					clusters = []api.DisplayedCluster{
						createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "1.1"),
					}

					// Invalid filter format (missing operator)
					filterBy := []string{"name-invalid"}
					resp, err := clustersApiService.GetClustersMinimalData(ctx, filterBy, 0, 0)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})

		Describe("GetClusterMinimalDataById", func() {
			Context("When cluster does not exist", func() {
				It("Should return 500 error", func() {
					nonExistentId := uuid.New().String()
					resp, err := clustersApiService.GetClusterMinimalDataById(ctx, nonExistentId)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusNotFound))
				})
			})

			Context("When cluster exists", func() {
				It("Should return the cluster minimal data correctly", func() {
					clusterName := clusterNameGenerator.Generate(testSuite)
					cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-domain", "1.5.0")
					updateClusterIngressClass(ctx, cluster, "nginx")

					resp, err := clustersApiService.GetClusterMinimalDataById(ctx, cluster.Uuid)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))

					minimalData, ok := resp.Body.(api.ClustersMinimalData)
					Expect(ok).To(BeTrue())
					Expect(minimalData.Uuid).To(Equal(cluster.Uuid))
					Expect(minimalData.Name).To(Equal(clusterName))
					Expect(minimalData.TenantId).To(Equal(int32(tenantId)))
					Expect(convutils.Deref(minimalData.Domain, "")).To(Equal("cluster-domain"))
					Expect(convutils.Deref(minimalData.Version, "")).To(Equal("1.5.0"))
					Expect(minimalData.IngressClass).ToNot(BeNil())
					Expect(*minimalData.IngressClass).To(Equal("nginx"))
				})

				It("Should return cluster minimal data without version when version is empty", func() {
					clusterName := clusterNameGenerator.Generate(testSuite)
					cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-domain", "")

					resp, err := clustersApiService.GetClusterMinimalDataById(ctx, cluster.Uuid)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))

					minimalData, ok := resp.Body.(api.ClustersMinimalData)
					Expect(ok).To(BeTrue())
					Expect(minimalData.Uuid).To(Equal(cluster.Uuid))
					Expect(minimalData.Name).To(Equal(clusterName))
					Expect(minimalData.Version).To(BeNil())
				})

				It("Should include correct state from cluster state calculator", func() {
					clusterName := clusterNameGenerator.Generate(testSuite)
					cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-domain", "1.0.0")

					resp, err := clustersApiService.GetClusterMinimalDataById(ctx, cluster.Uuid)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))

					minimalData, ok := resp.Body.(api.ClustersMinimalData)
					Expect(ok).To(BeTrue())
					Expect(minimalData.State).To(Equal(api.CLUSTERSTATE_CONNECTED))
				})
			})
		})

		Describe("Scope validation for clusters-minimal functions", func() {
			It("GetClustersMinimalData should return forbidden when no permitted scopes exist", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				createCluster(ctx, clustersApiService, clusterName, "cluster-1", "1.0.0")

				// Create context without permitted scopes
				emptyCtx := context.WithValue(context.Background(), runaictx.CfTenantId, strconv.Itoa(tenantId))
				emptyCtx = context.WithValue(emptyCtx, runaictx.CfSubjectID, "user-1")
				emptyCtx = context.WithValue(emptyCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{})
				emptyCtx = addAuditToCtx(emptyCtx)

				resp, err := clustersApiService.GetClustersMinimalData(emptyCtx, []string{}, 0, 0)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusForbidden))
			})

			It("GetClustersMinimalData should succeed with valid tenant scope", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				createCluster(ctx, clustersApiService, clusterName, "cluster-1", "1.0.0")

				resp, err := clustersApiService.GetClustersMinimalData(ctx, []string{}, 0, 0)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusOK))

				payload := resp.Body.(api.GetClustersMinimalData200Response)
				Expect(payload.Clusters).To(HaveLen(1))
			})

			It("GetClusterMinimalDataById should succeed with valid cluster scope", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-1", "1.0.0")

				// Create context with specific cluster scope
				clusterScopeCtx := context.WithValue(context.Background(), runaictx.CfTenantId, strconv.Itoa(tenantId))
				clusterScopeCtx = context.WithValue(clusterScopeCtx, runaictx.CfSubjectID, "user-1")
				clusterScopeCtx = context.WithValue(clusterScopeCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Clusters: []string{cluster.Uuid},
				})
				clusterScopeCtx = addAuditToCtx(clusterScopeCtx)

				resp, err := clustersApiService.GetClusterMinimalDataById(clusterScopeCtx, cluster.Uuid)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusOK))

				minimalData, ok := resp.Body.(api.ClustersMinimalData)
				Expect(ok).To(BeTrue())
				Expect(minimalData.Uuid).To(Equal(cluster.Uuid))
			})
		})

		Describe("GetClusterName", func() {
			It("Should return the cluster name correctly", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-1", "")

				resp, err := clustersApiService.GetClusterName(ctx, cluster.Uuid)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				getClusterName200Resp, ok := resp.Body.(api.GetClusterName200Response)
				Expect(ok).To(BeTrue())
				Expect(getClusterName200Resp.Name).To(Equal(clusterName))
			})
		})

		Describe("GetClusterUrl", func() {
			It("Should return the cluster url correctly", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-1", "")

				resp, err := clustersApiService.GetClusterUrl(ctx, cluster.Uuid)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				getClusterUrl200Resp, ok := resp.Body.(api.GetClusterUrl200Response)
				Expect(ok).To(BeTrue())
				Expect(getClusterUrl200Resp.Url).To(Equal("cluster-1"))
			})

			It("Should return the empty cluster url correctly", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "", "")

				resp, err := clustersApiService.GetClusterUrl(ctx, cluster.Uuid)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				getClusterUrl200Resp, ok := resp.Body.(api.GetClusterUrl200Response)
				Expect(ok).To(BeTrue())
				Expect(getClusterUrl200Resp.Url).To(Equal(""))
			})
		})

		Describe("GetClusterByUuid", func() {
			It("Should return the cluster correctly", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-1", "")

				resp, err := clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "full")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedCluster.Name).To(Equal(clusterName))
				Expect(*displayedCluster.Domain).To(Equal("cluster-1"))

				resp, err = clustersApiService.GetClusterInternal(ctx, cluster.Uuid)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok = resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedCluster.Name).To(Equal(clusterName))
				Expect(*displayedCluster.Domain).To(Equal("cluster-1"))
			})
			It("CSP Should return the cluster correctly", func() {
				var createdTenantId int32 = 1011
				myTrue := true
				mocks.tmTenantsApi.On("GetTenantByIdInternal", mock.Anything, mock.Anything).Return(tm_client_api.ApiGetTenantByIdInternalRequest{})
				mocks.tmTenantsApi.On("GetTenantByIdInternalExecute", mock.Anything).
					Return(&tm_client_api.Tenant{
						Id: int32(createdTenantId),
					}, nil, nil).Maybe()
				newCtx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
				newCtx = context.WithValue(newCtx, runaictx.CfTenantId, "1")
				newCtx = context.WithValue(newCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{"1"},
					System:  &myTrue,
				})
				newCtx = addAuditToCtx(newCtx)
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createClusterWithTenantId(newCtx, clustersApiService, clusterName, "cluster-1", "", createdTenantId)
				resp, err := clustersApiService.GetClusterByUuid(newCtx, cluster.Uuid, "full")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())
				Expect(displayedCluster.Name).To(Equal(clusterName))
				Expect(displayedCluster.TenantId).To(Equal(int32(createdTenantId)))
				Expect(*displayedCluster.Domain).To(Equal("cluster-1"))
			})
		})

		Describe("Cache tests", func() {
			It("cache is nil ", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-10", "0.0.0-Whatever")
				err := clusters_metadata_cache.SetClusterMetadataOnCache(ctx, cluster.TenantId, cluster.Uuid, nil)
				resp, err := clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "metadata")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())

				Expect(displayedCluster.Uuid).To(Equal(cluster.Uuid))
				Expect(displayedCluster.Name).To(Equal(cluster.Name))
				Expect(displayedCluster.TenantId).To(Equal(cluster.TenantId))
				apiErr := dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
				Expect(apiErr).To(BeNil())
			})

			It("Should return the BasicVerbosityClusterFields from the cache", func() {
				err := cache.ClearCache()
				Expect(err).To(BeNil())

				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-1", "0.0.0-Whatever")

				resp, err := clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "metadata")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				emptyGetter := func() (*models.Cluster, error) {
					return nil, nil
				}

				storedCluster, err := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, cluster.TenantId, cluster.Uuid, emptyGetter)
				Expect(storedCluster).ToNot(BeNil()) // made sure the cluster is in the cache

				// modify the data and save in cache
				storedCluster.Version = "0.0.0-modifiedVersion"
				err = clusters_metadata_cache.SetClusterMetadataOnCache(ctx, cluster.TenantId, cluster.Uuid, storedCluster)
				resp, err = clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "metadata")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				displayedCluster, ok := resp.Body.(api.DisplayedCluster)
				Expect(ok).To(BeTrue())

				Expect(displayedCluster.Uuid).To(Equal(storedCluster.ID.String()))
				Expect(displayedCluster.Name).To(Equal(storedCluster.Name))
				Expect(displayedCluster.TenantId).To(Equal(storedCluster.TenantId))
				Expect(displayedCluster.CreatedAt.String()).To(Equal(storedCluster.CreatedAt.String()))
				Expect(*displayedCluster.Version).To(Equal(storedCluster.Version))
				apiErr := dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
				Expect(apiErr).To(BeNil())
			})
			It("Should invalidate when updating version", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-3", "0.0.0-Whatever")
				mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
				mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				resp, err := clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "metadata")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				resp, err = clustersApiService.UpdateClusterInfo(ctx, cluster.Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Operands: map[string]api.ClusterReportedStatusOperandsValue{
							"operand-a": {
								Ready: false,
							},
						},
					},
					Domain:  "cluster-3",
					Version: "1.0.0-changed",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				emptyGetter := func() (*models.Cluster, error) {
					return nil, nil
				}
				storedCluster, err := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, cluster.TenantId, cluster.Uuid, emptyGetter)
				Expect(storedCluster).To(BeNil())
				apiErr := dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
				Expect(apiErr).To(BeNil())
			})
			It("Should invalidate when updating status (dependencies)", func() {
				clusterName := clusterNameGenerator.Generate(testSuite)
				cluster := createCluster(ctx, clustersApiService, clusterName, "cluster-4", "1.0.0")
				mocks.idmClient.On("GetInternalClientByFilter", mock.Anything, mock.Anything).Return(GetApiClient(), nil)
				mocks.idmClient.On("AppendRedirectUriToInternalClient", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// First update - cache gets populated
				resp, err := clustersApiService.UpdateClusterInfo(ctx, cluster.Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Dependencies: api.ClusterDependenciesStatus{
							Optional: map[string]api.ClusterDependencyStatus{
								"lws": {
									Available: false,
									Reason:    "Not installed",
								},
							},
						},
					},
					Domain:  "cluster-4",
					Version: "1.0.0",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				// Get cluster to populate cache
				resp, err = clustersApiService.GetClusterByUuid(ctx, cluster.Uuid, "metadata")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))

				// Second update with changed dependencies - should invalidate cache
				resp, err = clustersApiService.UpdateClusterInfo(ctx, cluster.Uuid, api.ClusterInfoSyncRequest{
					Status: &api.ClusterReportedStatus{
						Dependencies: api.ClusterDependenciesStatus{
							Optional: map[string]api.ClusterDependencyStatus{
								"lws": {
									Available: true,
									Reason:    "",
								},
							},
						},
					},
					Domain:  "cluster-4",
					Version: "1.0.0", // Same version
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(204))

				// Verify cache was invalidated (emptyGetter returns nil, so if cache was invalidated, result should be nil)
				emptyGetter := func() (*models.Cluster, error) {
					return nil, nil
				}
				storedCluster, err := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, cluster.TenantId, cluster.Uuid, emptyGetter)
				Expect(storedCluster).To(BeNil())

				apiErr := dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
				Expect(apiErr).To(BeNil())
			})
		})
	})
	Describe("Delete cluster tests", func() {
		It("Should delete cluster when asyncClusterDeletion.enabled: true and force=false", func() {
			runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: true})
			prepareMocksExt(true)
			cluster1Name := clusterNameGenerator.Generate(testSuite)
			cluster := createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "2.21.33")
			mocks.clusterDeleter.On("AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("NonAsyncDelete", mock.Anything, mock.Anything).Return(nil).Maybe()
			resp, err := clustersApiService.DeleteCluster(ctx, cluster.Uuid, false)
			Expect(err).Should(BeNil())
			Expect(resp.Code).To(Equal(202))
			mocks.clusterDeleter.AssertCalled(GinkgoT(), "AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "ForceDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "NonAsyncDelete", mock.Anything, mock.Anything)
			dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
		})
		It("Should delete cluster when asyncClusterDeletion.enabled: true and force=true", func() {
			runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: true})
			prepareMocksExt(true)
			cluster1Name := clusterNameGenerator.Generate(testSuite)
			cluster := createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "2.21.33")
			mocks.clusterDeleter.On("AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("NonAsyncDelete", mock.Anything, mock.Anything).Return(nil).Maybe()
			resp, err := clustersApiService.DeleteCluster(ctx, cluster.Uuid, true)
			Expect(err).Should(BeNil())
			Expect(resp.Code).To(Equal(202))
			mocks.clusterDeleter.AssertCalled(GinkgoT(), "ForceDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "NonAsyncDelete", mock.Anything, mock.Anything)
			dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
		})
		It("Should delete cluster when asyncClusterDeletion.enabled: false and force=(true or false)", func() {
			runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: false})
			prepareMocksExt(true)
			cluster1Name := clusterNameGenerator.Generate(testSuite)
			cluster := createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "2.21.33")
			mocks.clusterDeleter.On("AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("NonAsyncDelete", mock.Anything, mock.Anything).Return(nil).Maybe()
			resp, err := clustersApiService.DeleteCluster(ctx, cluster.Uuid, true)
			Expect(err).Should(BeNil())
			Expect(resp.Code).To(Equal(204))
			mocks.clusterDeleter.AssertCalled(GinkgoT(), "NonAsyncDelete", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "ForceDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything)

			resp, err = clustersApiService.DeleteCluster(ctx, cluster.Uuid, false)
			Expect(err).Should(BeNil())
			Expect(resp.Code).To(Equal(204))
			mocks.clusterDeleter.AssertCalled(GinkgoT(), "NonAsyncDelete", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "ForceDeleteCluster", mock.Anything, mock.Anything)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything)
			dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
		})
		It("Should delete cluster when error occurred", func() {
			runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: true})
			prepareMocksExt(true)
			cluster1Name := clusterNameGenerator.Generate(testSuite)
			cluster := createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", "2.21.33")
			mocks.clusterDeleter.On("AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything).Return(api.NewError(http.StatusInternalServerError, "err", errors.New("err"))).Maybe()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			mocks.clusterDeleter.On("NonAsyncDelete", mock.Anything, mock.Anything).Return(nil).Maybe()
			resp, err := clustersApiService.DeleteCluster(ctx, cluster.Uuid, false)
			Expect(err).Should(BeNil())
			Expect(resp.Code).To(Equal(500))
			mocks.clusterDeleter.AssertCalled(GinkgoT(), "AttemptToGracefullyDeleteCluster", mock.Anything, mock.Anything)
			dao.DeleteCluster(ctx, cluster.Uuid, tenantId)
		})
	})
})

func createClusterWithTenantId(ctx context.Context, clustersApiService api.ClustersAPIServicer, name string, domain string, version string, tenantId int32) api.DisplayedCluster {
	GinkgoHelper()

	resp, err := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
		Name:     name,
		Domain:   domain,
		Version:  version,
		TenantId: tenantId,
	})
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.Code).To(Equal(201))

	displayedCluster, ok := resp.Body.(api.DisplayedCluster)
	Expect(ok).To(BeTrue())
	return displayedCluster
}

func createCluster(ctx context.Context, clustersApiService api.ClustersAPIServicer, name string, domain string, version string) api.DisplayedCluster {
	GinkgoHelper()

	resp, err := clustersApiService.CreateCluster(ctx, api.ClusterCreationRequest{
		Name:    name,
		Domain:  domain,
		Version: version,
	})
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.Code).To(Equal(201))

	displayedCluster, ok := resp.Body.(api.DisplayedCluster)
	Expect(ok).To(BeTrue())
	return displayedCluster
}

func updateClusterSubdomain(ctx context.Context, cluster api.DisplayedCluster, subdomain bool) {
	resp, err := clustersApiService.UpdateClusterInfo(ctx, cluster.Uuid, api.ClusterInfoSyncRequest{
		Status: &api.ClusterReportedStatus{
			Config: &api.ClusterReportedConfig{
				SubdomainEnabled: &subdomain,
			},
		},
		Domain:  *cluster.Domain,
		Version: "1.0.0",
	})
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.Code).To(Equal(http.StatusNoContent))
}

func updateClusterIngressClass(ctx context.Context, cluster api.DisplayedCluster, ingressClass string) {
	resp, err := clustersApiService.UpdateClusterInfo(ctx, cluster.Uuid, api.ClusterInfoSyncRequest{
		Status: &api.ClusterReportedStatus{
			Config: &api.ClusterReportedConfig{
				IngressClass: &ingressClass,
			},
		},
		Domain:  *cluster.Domain,
		Version: convutils.Deref(cluster.Version, "1.0.0"),
	})
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.Code).To(Equal(http.StatusNoContent))
}

func deleteAllClusters(ctx context.Context, clustersApiService api.ClustersAPIServicer) {
	GinkgoHelper()
	myTrue := true
	newCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	newCtx = context.WithValue(newCtx, runaictx.CfTenantId, "1")
	newCtx = context.WithValue(newCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{"1"},
		System:  &myTrue,
	})
	newCtx = addAuditToCtx(newCtx)
	resp, _ := clustersApiService.GetClusters(newCtx, "metadata", []string{}, false)
	displayedClusters, _ := resp.Body.([]api.DisplayedCluster)

	for i := 0; i < len(displayedClusters); i++ {
		err := dao.DeleteCluster(newCtx, displayedClusters[i].Uuid, displayedClusters[i].TenantId)
		Expect(err).Should(BeNil())
	}
}

func deleteCtxClusters(ctx context.Context, clustersApiService api.ClustersAPIServicer) {
	GinkgoHelper()
	resp, _ := clustersApiService.GetClusters(ctx, "metadata", []string{}, true)
	displayedClusters, _ := resp.Body.([]api.DisplayedCluster)

	for i := 0; i < len(displayedClusters); i++ {
		err := dao.DeleteCluster(ctx, displayedClusters[i].Uuid, displayedClusters[i].TenantId)
		Expect(err).Should(BeNil())
	}
}
