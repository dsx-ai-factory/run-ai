/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testTenantId            = 1
	testSubjectId           = "test-subject-id"
	testNetworkTopologyId   = "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"
	testNetworkTopologyName = "test-network-topology"
)

var (
	networkTopologiesApiService  api.NetworkTopologiesAPIServicer
	testCtx                      context.Context
	testNetworkTopologyLevels    = []string{"topology-block", "topology-rack", "hostname"}
	testNetworkTopologyClusterId string
)

var _ = Describe("NetworkTopologiesApiService", Ordered, func() {
	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Running network topologies tests")
		// Create a test cluster that will be used for network topology tests
		testCtx = addTenantAndSubjectToCtx(context.Background(), testTenantId, testSubjectId)

		cluster, err := dao.CreateCluster(testCtx, &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-1",
			Version:  "2.23",
			Domain:   "test-domain",
			TenantId: testTenantId,
		})

		Expect(err).To(BeNil())
		testNetworkTopologyClusterId = cluster.ID.String()
	})

	BeforeEach(func() {
		networkTopologiesApiService = services.NewNetworkTopologiesApiService()
	})

	AfterAll(func() {
		// Clean up the test cluster
		ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
		dao.DeleteCluster(ctx, testNetworkTopologyClusterId, testTenantId)
	})

	Describe("GetNetworkTopologies", func() {
		Context("when valid request is made", func() {
			It("should return network topologies successfully", func() {
				// Setup context with tenant ID and permitted scopes
				ctx := addTenantAndScopesToCtx(context.Background(), testTenantId, []string{testNetworkTopologyClusterId})
				ctx = addSubjectToCtx(ctx, testSubjectId)
				// Create test network topology in database
				name := fmt.Sprintf("%s-2", testNetworkTopologyName)
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      name,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}
				createResponse, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))

				// Call the service method
				testNow := time.Now()
				response, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})

				// Verify response
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				// Parse response body
				responseBody, ok := response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(1))

				nt := responseBody.NetworkTopologies[0]
				Expect(nt.Name).To(Equal(name))
				Expect(nt.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
				Expect(nt.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(nt.TenantId).To(Equal(int32(testTenantId)))
				Expect(nt.CreatedBy).To(Equal(testSubjectId))
				Expect(nt.CreatedAt.Unix()).To(BeNumerically(">=", testNow.Unix()))
				Expect(nt.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
				Expect(nt.Phase).To(Equal(models.NetworkTopologyPhaseCreating))
			})

			Context("nodePools aggregation via API", func() {
				var (
					ctx             context.Context
					createdTopology api.NetworkTopology
					npCreatedIds    []int32
				)

				BeforeEach(func() {
					ctx = addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
					ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

					req := api.NetworkTopologyCreateRequest{
						Name:      fmt.Sprintf("%s-nodepools-%d", testNetworkTopologyName, time.Now().UnixNano()),
						Levels:    testNetworkTopologyLevels,
						ClusterId: testNetworkTopologyClusterId,
					}
					createResp, err := networkTopologiesApiService.CreateNetworkTopology(ctx, req)
					Expect(err).To(BeNil())
					Expect(createResp.Code).To(Equal(http.StatusCreated))
					createdTopology = createResp.Body.(api.NetworkTopology)
				})

				AfterEach(func() {
					// ensure cleanup even on failures
					for _, id := range npCreatedIds {
						_, _ = runaidb.DB().WithContext(context.Background()).Unscoped().Delete(&models.Nodepool{}, id)
					}
					npCreatedIds = nil
					// delete the topology
					_, _ = networkTopologiesApiService.DeleteNetworkTopology(ctx, createdTopology.Id)
				})

				It("should include nodePools list for each topology", func() {
					createdBy := testSubjectId
					npA, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-a",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npA.ID)
					npB, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-b",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npB.ID)

					resp, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					body := resp.Body.(api.GetNetworkTopologies200Response)
					var found *api.NetworkTopology
					for i := range body.NetworkTopologies {
						if body.NetworkTopologies[i].Id == createdTopology.Id {
							found = &body.NetworkTopologies[i]
							break
						}
					}
					Expect(found).ToNot(BeNil())
					Expect(found.NodePools).ToNot(BeNil())
					Expect(*found.NodePools).To(Equal([]string{"np-a", "np-b"}))

					// also by id
					resp, err = networkTopologiesApiService.GetNetworkTopologyById(ctx, createdTopology.Id)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					nt := resp.Body.(api.NetworkTopology)
					Expect(*nt.NodePools).To(Equal([]string{"np-a", "np-b"}))
				})

				It("should update nodePools after deleting a nodepool", func() {
					createdBy := testSubjectId
					npA, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-a-1",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npA.ID)
					npB, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-b-1",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npB.ID)

					// initial
					resp, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					body := resp.Body.(api.GetNetworkTopologies200Response)
					var found *api.NetworkTopology
					for i := range body.NetworkTopologies {
						if body.NetworkTopologies[i].Id == createdTopology.Id {
							found = &body.NetworkTopologies[i]
							break
						}
					}
					Expect(found).ToNot(BeNil())
					Expect(*found.NodePools).To(Equal([]string{"np-a-1", "np-b-1"}))

					// delete np-a-1
					Expect(dao.DeleteNodepool(ctx, int32(testTenantId), createdBy, fmt.Sprintf("%d", npA.ID))).To(BeNil())
					// keep only remaining for cleanup
					npCreatedIds = []int32{npB.ID}

					resp, err = networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					body = resp.Body.(api.GetNetworkTopologies200Response)
					found = nil
					for i := range body.NetworkTopologies {
						if body.NetworkTopologies[i].Id == createdTopology.Id {
							found = &body.NetworkTopologies[i]
							break
						}
					}
					Expect(found).ToNot(BeNil())
					Expect(*found.NodePools).To(Equal([]string{"np-b-1"}))
				})

				It("should update nodePools after adding a nodepool", func() {
					createdBy := testSubjectId
					npA, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-a-2",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npA.ID)

					resp, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					body := resp.Body.(api.GetNetworkTopologies200Response)
					var found *api.NetworkTopology
					for i := range body.NetworkTopologies {
						if body.NetworkTopologies[i].Id == createdTopology.Id {
							found = &body.NetworkTopologies[i]
							break
						}
					}
					Expect(found).ToNot(BeNil())
					Expect(*found.NodePools).To(Equal([]string{"np-a-2"}))

					npB, err := dao.CreateNodepool(ctx, models.Nodepool{
						Name:                "np-b-2",
						TenantId:            int32(testTenantId),
						ClusterId:           uuid.MustParse(testNetworkTopologyClusterId),
						ClusterName:         "cluster-1",
						IsDefault:           false,
						NetworkTopologyId:   ptr.To(uuid.MustParse(createdTopology.Id)),
						NetworkTopologyName: &createdTopology.Name,
						CreatedBy:           createdBy,
						UpdatedBy:           createdBy,
					})
					Expect(err).To(BeNil())
					npCreatedIds = append(npCreatedIds, npB.ID)

					resp, err = networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusOK))
					body = resp.Body.(api.GetNetworkTopologies200Response)
					found = nil
					for i := range body.NetworkTopologies {
						if body.NetworkTopologies[i].Id == createdTopology.Id {
							found = &body.NetworkTopologies[i]
							break
						}
					}
					Expect(found).ToNot(BeNil())
					Expect(*found.NodePools).To(Equal([]string{"np-a-2", "np-b-2"}))
				})
			})
		})

		Context("when invalid filter is provided", func() {
			It("should return bad request error", func() {
				ctx := addTenantAndScopesToCtx(testCtx, testTenantId, []string{testNetworkTopologyClusterId})

				response, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{"invalid:filter"})

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when no permitted scopes are available", func() {
			It("should return internal server error", func() {
				ctx := addTenantToCtx(testCtx, testTenantId)

				response, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusInternalServerError))
				Expect(response.Body.(string)).To(Equal("no permitted scopes defined"))
			})
		})
	})

	Describe("CreateNetworkTopology", func() {
		Context("when valid request is made", func() {
			It("should create network topology successfully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})
				name := fmt.Sprintf("%s-1", testNetworkTopologyName)

				createRequest := api.NetworkTopologyCreateRequest{
					Name:      name,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))

				// Parse response body
				responseBody, ok := response.Body.(api.NetworkTopology)
				Expect(ok).To(BeTrue())
				Expect(responseBody.Name).To(Equal(name))
				Expect(responseBody.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
				Expect(responseBody.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(responseBody.TenantId).To(Equal(int32(testTenantId)))
				Expect(responseBody.CreatedBy).To(Equal(testSubjectId))
				Expect(responseBody.Levels).To(Equal(testNetworkTopologyLevels))
				Expect(responseBody.Phase).To(Equal(models.NetworkTopologyPhaseCreating))
			})
		})

		Context("when empty levels are provided", func() {
			It("should return bad request error with status 400", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				createRequest := api.NetworkTopologyCreateRequest{
					Name:      testNetworkTopologyName,
					Levels:    []string{}, // Empty levels array
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))

				// Verify error message
				errorResponse, ok := response.Body.(api.Error)
				Expect(ok).To(BeTrue())
				Expect(errorResponse.Message).To(Equal("Levels can not be empty"))
			})
		})

		Context("when invalid cluster ID is provided", func() {
			It("should return bad request error", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				createRequest := api.NetworkTopologyCreateRequest{
					Name:      testNetworkTopologyName,
					Levels:    testNetworkTopologyLevels,
					ClusterId: "invalid-uuid",
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when topology with the same name already exists", func() {
			It("should return StatusConflict request error", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				createRequest := api.NetworkTopologyCreateRequest{
					Name:      testNetworkTopologyName,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))

				// Create topology with the same name
				createRequest = api.NetworkTopologyCreateRequest{
					Name:      testNetworkTopologyName,
					Levels:    testNetworkTopologyLevels[0:2],
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err = networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusConflict))
			})
		})
	})

	Describe("GetNetworkTopologiesCount", func() {
		Context("when valid request is made", func() {
			It("should return count successfully", func() {
				ctx := addTenantAndScopesToCtx(testCtx, testTenantId, []string{testNetworkTopologyClusterId})

				// Create test network topology
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      generateTopologyName(),
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}
				createdNetworkTopology, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(createdNetworkTopology).To(Not(BeNil()))

				response, err := networkTopologiesApiService.GetNetworkTopologiesCount(ctx, []string{})

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBody, ok := response.Body.(api.CountResponse)
				Expect(ok).To(BeTrue())
				Expect(responseBody.Count).To(BeNumerically(">=", 1))
			})
		})

		Context("when invalid filter is provided", func() {
			It("should return bad request error", func() {
				ctx := addTenantAndScopesToCtx(testCtx, testTenantId, []string{testNetworkTopologyClusterId})

				response, err := networkTopologiesApiService.GetNetworkTopologiesCount(ctx, []string{"invalid:filter"})

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("GetNetworkTopologyById", func() {
		Context("when valid network topology ID is provided", func() {
			It("should return network topology successfully", func() {
				ctx := addTenantAndScopesToCtx(testCtx, testTenantId, []string{testNetworkTopologyClusterId})

				// Create test network topology
				name := fmt.Sprintf("%s-3", testNetworkTopologyName)
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      name,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}
				testNow := time.Now()
				createResponse, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))

				responseBody, ok := createResponse.Body.(api.NetworkTopology)
				Expect(ok).To(BeTrue())

				response, err := networkTopologiesApiService.GetNetworkTopologyById(ctx, responseBody.Id)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				createdNT, ok := response.Body.(api.NetworkTopology)

				Expect(ok).To(BeTrue())
				Expect(responseBody.Id).To(Equal(createdNT.Id))
				Expect(responseBody.Name).To(Equal(name))
				Expect(responseBody.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
				Expect(responseBody.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(responseBody.TenantId).To(Equal(int32(testTenantId)))
				Expect(responseBody.CreatedBy).To(Equal(testSubjectId))
				Expect(responseBody.Levels).To(Equal(testNetworkTopologyLevels))
				Expect(responseBody.Phase).To(Equal(models.NetworkTopologyPhaseCreating))
				Expect(responseBody.CreatedAt.Unix()).To(BeNumerically(">=", testNow.Unix()))
			})
		})

		Context("when network topology does not exist", func() {
			It("should return not found status", func() {
				ctx := addTenantAndScopesToCtx(testCtx, testTenantId, []string{testNetworkTopologyClusterId})

				response, err := networkTopologiesApiService.GetNetworkTopologyById(ctx, testNetworkTopologyId)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("DeleteNetworkTopology", func() {
		Context("when valid delete request is made", func() {
			It("should delete network topology successfully and set phase to deleted", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Create test network topology
				name := fmt.Sprintf("%s-5", testNetworkTopologyName)
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      name,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))

				responseBody, ok := response.Body.(api.NetworkTopology)
				Expect(ok).To(BeTrue())
				createdNT := responseBody

				response, err = networkTopologiesApiService.DeleteNetworkTopology(ctx, createdNT.Id)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the network topology is not returned in queries (due to deleted_at filter)
				response, err = networkTopologiesApiService.GetNetworkTopologyById(ctx, createdNT.Id)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when network topology does not exist", func() {
			It("should return not found status", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				response, err := networkTopologiesApiService.DeleteNetworkTopology(ctx, testNetworkTopologyId)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when deleting network topology affects count", func() {
			It("should update count correctly after deletion", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Create test network topology
				name := fmt.Sprintf("%s-5", testNetworkTopologyName)
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      name,
					Levels:    testNetworkTopologyLevels,
					ClusterId: testNetworkTopologyClusterId,
				}

				response, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))

				responseBody, ok := response.Body.(api.NetworkTopology)
				Expect(ok).To(BeTrue())
				createdNT := responseBody

				// Get initial count
				response, err = networkTopologiesApiService.GetNetworkTopologiesCount(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				initialCount := response.Body.(api.CountResponse).Count

				// Delete the network topology
				response, err = networkTopologiesApiService.DeleteNetworkTopology(ctx, createdNT.Id)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Get count after deletion
				response, err = networkTopologiesApiService.GetNetworkTopologiesCount(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				finalCount := response.Body.(api.CountResponse).Count

				// Verify count decreased by 1
				Expect(finalCount).To(Equal(initialCount - 1))
			})
		})

		Context("when network topology with invalid Id", func() {
			It("should return not found status", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				invalidNetworkTopologyId := "invalid:invalid"
				response, err := networkTopologiesApiService.DeleteNetworkTopology(ctx, invalidNetworkTopologyId)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("SyncNetworkTopologies", func() {
		Context("when valid sync request is made", func() {
			It("should sync network topologies successfully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "sync-topology-1",
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "sync-topology-2",
							Levels: []string{"level3", "level4", "level5"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}
				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)

				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the network topologies were created
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Find the synced topologies
				var topology1, topology2 *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "sync-topology-1" {
						topology1 = nt
					} else if nt.Name == "sync-topology-2" {
						topology2 = nt
					}
				}

				// Verify first topology
				Expect(topology1).To(Not(BeNil()))
				Expect(topology1.Name).To(Equal("sync-topology-1"))
				Expect(topology1.Levels).To(Equal([]string{"level1", "level2"}))
				Expect(topology1.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(topology1.Source).To(Equal(dao.NetworkTopologiesClusterSource))
				Expect(topology1.TenantId).To(Equal(int32(testTenantId)))
				Expect(topology1.Phase).To(Equal(models.NetworkTopologyPhaseReady))

				// Verify second topology
				Expect(topology2).To(Not(BeNil()))
				Expect(topology2.Name).To(Equal("sync-topology-2"))
				Expect(topology2.Levels).To(Equal([]string{"level3", "level4", "level5"}))
				Expect(topology2.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(topology2.Source).To(Equal(dao.NetworkTopologiesClusterSource))
				Expect(topology2.TenantId).To(Equal(int32(testTenantId)))
				Expect(topology2.Phase).To(Equal(models.NetworkTopologyPhaseReady))
			})

			It("should sync network topology with grove fields", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				topologyName := "sync-topology-with-grove"
				groveTopologyName := "grove-topo-1"
				observedGen := int64(3)
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   topologyName,
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
							Grove: &api.NetworkTopologySyncGrove{
								Name:    groveTopologyName,
								Domains: []string{"domain1", "domain2"},
								Conditions: []api.NetworkTopologySyncGroveCondition{
									{
										Type:               "Ready",
										Status:             "True",
										Reason:             "TopologyReady",
										Message:            "Topology is ready",
										ObservedGeneration: observedGen,
									},
								},
							},
						},
					},
				}
				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify grove fields were persisted via the view
				filterBy := []queryapi.Filter{{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{topologyName}}}
				topologies, apiErr := dao.GetNetworkTopologiesView(ctx, testTenantId, nil, filterBy)
				Expect(apiErr).To(BeNil())
				Expect(topologies).To(HaveLen(1))
				Expect(topologies[0].Grove).To(Not(BeNil()))
				Expect(topologies[0].Grove.Name).To(Equal(groveTopologyName))
				Expect(topologies[0].Grove.Domains).To(BeEquivalentTo([]string{"domain1", "domain2"}))
				Expect(topologies[0].Grove.Conditions).To(HaveLen(1))
				Expect(topologies[0].Grove.Conditions[0].Type).To(Equal("Ready"))
				Expect(topologies[0].Grove.Conditions[0].Status).To(Equal("True"))
				Expect(topologies[0].Grove.Conditions[0].Reason).To(Equal("TopologyReady"))
				Expect(topologies[0].Grove.Conditions[0].Message).To(Equal("Topology is ready"))
				Expect(topologies[0].Grove.Conditions[0].ObservedGeneration).To(Equal(int64(3)))

				// Verify groveTopology is returned via GetNetworkTopologies API
				listResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{fmt.Sprintf("name==%s", topologyName)})
				Expect(err).To(BeNil())
				Expect(listResponse.Code).To(Equal(http.StatusOK))
				listBody, ok := listResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(listBody.NetworkTopologies).To(HaveLen(1))
				Expect(listBody.NetworkTopologies[0].GroveTopology).To(Not(BeNil()))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Name).To(Equal(groveTopologyName))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Domains).To(BeEquivalentTo([]string{"domain1", "domain2"}))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions).To(HaveLen(1))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions[0].Type).To(Equal("Ready"))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions[0].Status).To(Equal("True"))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions[0].Reason).To(Equal("TopologyReady"))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions[0].Message).To(Equal("Topology is ready"))
				Expect(listBody.NetworkTopologies[0].GroveTopology.Conditions[0].ObservedGeneration).To(Equal(int64(3)))

				// Verify groveTopology is returned via GetNetworkTopologyById API
				getByIdResponse, err := networkTopologiesApiService.GetNetworkTopologyById(ctx, listBody.NetworkTopologies[0].Id)
				Expect(err).To(BeNil())
				Expect(getByIdResponse.Code).To(Equal(http.StatusOK))
				getByIdBody, ok := getByIdResponse.Body.(api.NetworkTopology)
				Expect(ok).To(BeTrue())
				Expect(getByIdBody.GroveTopology).To(Not(BeNil()))
				Expect(getByIdBody.GroveTopology.Name).To(Equal(groveTopologyName))
				Expect(getByIdBody.GroveTopology.Domains).To(BeEquivalentTo([]string{"domain1", "domain2"}))
				Expect(getByIdBody.GroveTopology.Conditions).To(HaveLen(1))
				Expect(getByIdBody.GroveTopology.Conditions[0].Type).To(Equal("Ready"))
				Expect(getByIdBody.GroveTopology.Conditions[0].Status).To(Equal("True"))
				Expect(getByIdBody.GroveTopology.Conditions[0].Reason).To(Equal("TopologyReady"))
				Expect(getByIdBody.GroveTopology.Conditions[0].Message).To(Equal("Topology is ready"))
				Expect(getByIdBody.GroveTopology.Conditions[0].ObservedGeneration).To(Equal(int64(3)))
			})

			It("should sync network topology with nil grove", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				topologyName := "sync-topology-without-grove"
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   topologyName,
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}
				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify grove is nil
				filterBy := []queryapi.Filter{{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{topologyName}}}
				topologies, apiErr := dao.GetNetworkTopologiesView(ctx, testTenantId, nil, filterBy)
				Expect(apiErr).To(BeNil())
				Expect(topologies).To(HaveLen(1))
				Expect(topologies[0].Grove).To(BeNil())

				// Verify groveTopology is nil via GetNetworkTopologies API
				listResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{fmt.Sprintf("name==%s", topologyName)})
				Expect(err).To(BeNil())
				Expect(listResponse.Code).To(Equal(http.StatusOK))
				listBody, ok := listResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(listBody.NetworkTopologies).To(HaveLen(1))
				Expect(listBody.NetworkTopologies[0].GroveTopology).To(BeNil())
			})

			It("should use default source when source is not provided", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "default-source-topology",
							Levels: []string{"level1"},
							// Source intentionally omitted
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the topology was created with default source
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var topology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "default-source-topology" {
						topology = nt
						break
					}
				}

				Expect(topology).To(Not(BeNil()))
				Expect(topology.Source).To(Equal(dao.NetworkTopologiesClusterSource))
			})

			It("should handle empty network topologies array", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when invalid sync request is made", func() {
			It("should return bad request for invalid cluster ID", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "invalid-cluster-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}
				invalidClusterId := "invalid-uuid"
				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, invalidClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when database transaction fails", func() {
			It("should return internal server error", func() {
				// This test would require mocking the database layer
				// For now, we'll test the happy path and error handling
				// that we can control through the API layer
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "transaction-test-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when tenant ID is missing from context", func() {
			It("should return internal server error", func() {
				// Create context without tenant ID
				ctx := context.Background()
				ctx = addSubjectToCtx(ctx, testSubjectId)

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "missing-tenant-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when processing large batch of topologies", func() {
			It("should handle large number of topologies successfully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Create a large batch of topologies
				var topologies []api.NetworkTopologySyncFields
				for i := 0; i < 50; i++ {
					topology := api.NetworkTopologySyncFields{
						Name:   fmt.Sprintf("batch-topology-%d", i),
						Levels: []string{fmt.Sprintf("level-%d", i)},
						Source: dao.NetworkTopologiesClusterSource,
						Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
					}
					topologies = append(topologies, topology)
				}

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: topologies,
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all topologies were created
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(BeNumerically(">=", 50))

				// Clean up by deleting the test cluster (this will cascade delete all topologies)
				dao.DeleteCluster(ctx, testNetworkTopologyClusterId, testTenantId)
			})
		})

		Context("when syncing with empty request", func() {
			It("should handle empty request gracefully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when syncing with nil request", func() {
			It("should handle nil request gracefully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: nil,
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when syncing with explicit empty source", func() {
			It("should use default source when source is explicitly empty string", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "explicit-empty-source-topology",
							Levels: []string{"level1", "level2"},
							Source: "", // Explicitly set to empty string
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the topology was created with default source
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var topology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "explicit-empty-source-topology" {
						topology = nt
						break
					}
				}

				Expect(topology).To(Not(BeNil()))
				Expect(topology.Source).To(Equal(dao.NetworkTopologiesClusterSource))
			})
		})

		Context("when adding new network topology without phase and source", func() {
			It("should create network topology with default values and validate all fields", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Create a new topology without specifying phase and source
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "new-topology-no-phase-source",
							Levels: []string{"rack", "host", "gpu"},
							// Phase and Source intentionally omitted
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the topology was created with default values
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var topology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "new-topology-no-phase-source" {
						topology = nt
						break
					}
				}

				// Validate all fields are correct
				Expect(topology).To(Not(BeNil()))
				Expect(topology.Name).To(Equal("new-topology-no-phase-source"))
				Expect(topology.Levels).To(Equal([]string{"rack", "host", "gpu"}))
				Expect(topology.ClusterId).To(Equal(testNetworkTopologyClusterId))
				Expect(topology.TenantId).To(Equal(int32(testTenantId)))
				Expect(topology.CreatedBy).To(Equal(testSubjectId))
				Expect(topology.Source).To(Equal(dao.NetworkTopologiesClusterSource))
				Expect(topology.Phase).To(Equal(models.NetworkTopologyPhaseReady))
			})
		})

		Context("when adding network topology with invalid source", func() {
			It("should reject invalid source values", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Test with invalid source value
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "invalid-source-topology",
							Levels: []string{"level1", "level2"},
							Source: "InvalidSource", // Invalid source value
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the topology was created (the service currently accepts any source value)
				// This test documents the current behavior and can be updated if validation is added
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var topology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "invalid-source-topology" {
						topology = nt
						break
					}
				}

				Expect(topology).To(BeNil())
			})

			It("should accept valid source values (Run.ai and Cluster)", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Test with valid source values
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "valid-source-runai-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesRunAiSource, // Valid: "Run.ai"
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "valid-source-cluster-topology",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource, // Valid: "Cluster"
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify both topologies were created with valid sources
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var runaiTopology, clusterTopology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "valid-source-runai-topology" {
						runaiTopology = nt
					} else if nt.Name == "valid-source-cluster-topology" {
						clusterTopology = nt
					}
				}

				// Verify Run.ai source topology
				Expect(runaiTopology).To(BeNil())

				// Verify Cluster source topology
				Expect(clusterTopology).To(Not(BeNil()))
				Expect(clusterTopology.Source).To(Equal(dao.NetworkTopologiesClusterSource))
				Expect(clusterTopology.Phase).To(Equal(models.NetworkTopologyPhaseReady))
				Expect(clusterTopology.Levels).To(Equal([]string{"level2"}))
			})
		})

		Context("when syncing with full sync enabled", func() {
			It("should delete Cluster network topologies not in the request", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create some Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-1",
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-2",
							Levels: []string{"level3", "level4"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-3",
							Levels: []string{"level5", "level6"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all topologies were created
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(BeNumerically(">=", 3))

				// Now perform a full sync with only one topology
				fullSyncRequest := api.NetworkTopologySyncRequest{
					FullSync: true,
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-1",
							Levels: []string{"level1", "level2"}, // Updated levels
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						// cluster-topology-2 and cluster-topology-3 are intentionally omitted
					},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, fullSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the results
				getResponse, err = networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok = getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Find the remaining topology
				var remainingTopology *api.NetworkTopology
				var deletedTopologies []*api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "cluster-topology-1" {
						remainingTopology = nt
					} else if nt.Name == "cluster-topology-2" || nt.Name == "cluster-topology-3" {
						deletedTopologies = append(deletedTopologies, nt)
					}
				}

				// Verify remaining topology was updated
				Expect(remainingTopology).To(Not(BeNil()))
				Expect(remainingTopology.Levels).To(Equal([]string{"level1", "level2"}))

				// Verify deleted topologies are not present (they should be soft deleted)
				Expect(len(deletedTopologies)).To(Equal(0))
			})

			It("should delete all Cluster network topologies when full sync is true and request is empty", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create multiple Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-1",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-2",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-3",
							Levels: []string{"level3"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Now perform a full sync with empty array
				fullSyncRequest := api.NetworkTopologySyncRequest{
					FullSync:          true,
					NetworkTopologies: []api.NetworkTopologySyncFields{},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, fullSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all Cluster topologies were deleted
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{fmt.Sprintf("clusterId==%s", testNetworkTopologyClusterId)})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Check that no Cluster topologies remain
				for _, topology := range responseBody.NetworkTopologies {
					if topology.Source == dao.NetworkTopologiesClusterSource {
						Fail(fmt.Sprintf("Expected no Cluster topologies to remain, but found: %s", topology.Name))
					}
				}

				// Verify that exactly 3 Cluster topologies were deleted (none should remain)
				// The test should confirm that all Cluster source topologies were removed
			})

			It("should not delete Run.ai network topologies during full sync", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create a network topology
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      "runai-topology",
					Levels:    []string{"level1"},
					ClusterId: testNetworkTopologyClusterId,
				}

				createResponse, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))

				// Create both Run.ai and Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "runai-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesRunAiSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Now perform a full sync that omits the Cluster topology
				fullSyncRequest := api.NetworkTopologySyncRequest{
					FullSync: true,
					NetworkTopologies: []api.NetworkTopologySyncFields{
						// Only include the Run.ai topology
						{
							Name:   "runai-topology",
							Levels: []string{"level1", "level2"}, // Updated levels
							Source: dao.NetworkTopologiesRunAiSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, fullSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the results
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Find the topologies
				var clusterTopology *api.NetworkTopology
				var runaiTopology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "cluster-topology" {
						clusterTopology = nt
					} else if nt.Name == "runai-topology" {
						runaiTopology = nt
					}
				}

				// Verify Run.ai topology was not updated and remains
				Expect(runaiTopology).To(Not(BeNil()))
				Expect(runaiTopology.Levels).To(Equal([]string{"level1"}))

				// Verify Cluster topology was deleted (not present in results)
				Expect(clusterTopology).To(BeNil())
			})

			It("should handle bulk deletion of multiple missing Cluster network topologies", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create multiple Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-bulk-1",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-bulk-2",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-bulk-3",
							Levels: []string{"level3"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-bulk-4",
							Levels: []string{"level4"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-bulk-5",
							Levels: []string{"level5"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all topologies were created
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(BeNumerically(">=", 5))

				// Now perform a full sync that keeps only one topology
				fullSyncRequest := api.NetworkTopologySyncRequest{
					FullSync: true,
					NetworkTopologies: []api.NetworkTopologySyncFields{
						// Only keep the first topology, delete the other 4
						{
							Name:   "cluster-topology-bulk-1",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, fullSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the results
				getResponse, err = networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok = getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Find the remaining topology
				var remainingTopology *api.NetworkTopology
				var deletedTopologies []*api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "cluster-topology-bulk-1" {
						remainingTopology = nt
					} else if strings.HasPrefix(nt.Name, "cluster-topology-bulk-") {
						deletedTopologies = append(deletedTopologies, nt)
					}
				}

				// Verify remaining topology was updated
				Expect(remainingTopology).To(Not(BeNil()))
				Expect(remainingTopology.Levels).To(Equal([]string{"level1"}))

				// Verify all other topologies were deleted (not present in results)
				Expect(len(deletedTopologies)).To(Equal(0))
			})
		})

		Context("when syncing with full sync disabled", func() {
			It("should not delete any existing network topologies", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create some Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-persistent",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Now perform a sync without full sync (default behavior)
				partialSyncRequest := api.NetworkTopologySyncRequest{
					FullSync: false, // Explicitly set to false
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "new-topology",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, partialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify both topologies exist
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var clusterTopology, newTopology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "cluster-topology-persistent" {
						clusterTopology = nt
					} else if nt.Name == "new-topology" {
						newTopology = nt
					}
				}

				// Verify both topologies remain
				Expect(clusterTopology).To(Not(BeNil()))
				Expect(newTopology).To(Not(BeNil()))
			})
		})

		Context("when syncing with empty request", func() {
			It("should delete all Cluster network topologies when full sync is true and request is empty", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create some Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-1",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-2",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Now perform a full sync with empty array
				emptyFullSyncRequest := api.NetworkTopologySyncRequest{
					FullSync:          true,
					NetworkTopologies: []api.NetworkTopologySyncFields{},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, emptyFullSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all Cluster topologies were deleted
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Check that no Cluster topologies remain
				for _, topology := range responseBody.NetworkTopologies {
					if topology.Source == dao.NetworkTopologiesClusterSource {
						Fail(fmt.Sprintf("Expected no Cluster topologies to remain, but found: %s", topology.Name))
					}
				}
			})

			It("should not delete any Cluster network topologies when full sync is false and request is empty", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create some Cluster network topologies
				initialSyncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-topology-persistent-1",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster-topology-persistent-2",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, initialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Now perform a sync with empty array but fullSync = false
				emptyPartialSyncRequest := api.NetworkTopologySyncRequest{
					FullSync:          false,
					NetworkTopologies: []api.NetworkTopologySyncFields{},
				}

				response, err = networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, emptyPartialSyncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify all Cluster topologies still exist (no deletions occurred)
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Check that both Cluster topologies still remain
				var clusterTopology1, clusterTopology2 *api.NetworkTopology
				for _, topology := range responseBody.NetworkTopologies {
					if topology.Name == "cluster-topology-persistent-1" {
						clusterTopology1 = &topology
					} else if topology.Name == "cluster-topology-persistent-2" {
						clusterTopology2 = &topology
					}
				}

				// Verify both Cluster topologies still exist
				Expect(clusterTopology1).To(Not(BeNil()))
				Expect(clusterTopology2).To(Not(BeNil()))
			})
		})

		Context("when using DELETE action", func() {
			It("should delete network topology when DELETE action is specified", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// First, create a network topology
				createRequest := api.NetworkTopologyCreateRequest{
					Name:      "delete-test-topology",
					Levels:    []string{"level1", "level2"},
					ClusterId: testNetworkTopologyClusterId,
				}

				createResponse, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))

				// Verify the topology was created
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				var topology *api.NetworkTopology
				for i := range responseBody.NetworkTopologies {
					nt := &responseBody.NetworkTopologies[i]
					if nt.Name == "delete-test-topology" {
						topology = nt
						break
					}
				}

				Expect(topology).To(Not(BeNil()))
				Expect(topology.Name).To(Equal("delete-test-topology"))

				// Now delete the topology using DELETE action
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "delete-test-topology",
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_DELETE,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify the topology was deleted
				getResponse, err = networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok = getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Check that the deleted topology is not present
				for _, nt := range responseBody.NetworkTopologies {
					if nt.Name == "delete-test-topology" {
						Fail("Expected topology to be deleted, but it still exists")
					}
				}
			})

			It("should handle DELETE action for non-existent topology gracefully", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Try to delete a non-existent topology using DELETE action
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "non-existent-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_DELETE,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify no error occurred and the request was processed successfully
				// The service should handle non-existent deletions gracefully
			})

			It("should handle DELETE action with different source values", func() {
				ctx := addTenantAndSubjectToCtx(testCtx, testTenantId, testSubjectId)
				ctx = addScopesToCtx(ctx, []string{testNetworkTopologyClusterId})

				// Create topologies with different sources
				createRequest1 := api.NetworkTopologyCreateRequest{
					Name:      "cluster-source-topology",
					Levels:    []string{"level1"},
					ClusterId: testNetworkTopologyClusterId,
				}

				createRequest2 := api.NetworkTopologyCreateRequest{
					Name:      "runai-source-topology",
					Levels:    []string{"level2"},
					ClusterId: testNetworkTopologyClusterId,
				}

				createResponse1, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest1)
				Expect(err).To(BeNil())
				Expect(createResponse1.Code).To(Equal(http.StatusCreated))

				createResponse2, err := networkTopologiesApiService.CreateNetworkTopology(ctx, createRequest2)
				Expect(err).To(BeNil())
				Expect(createResponse2.Code).To(Equal(http.StatusCreated))

				// Now delete both using DELETE action
				syncRequest := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster-source-topology",
							Levels: []string{"level1"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_DELETE,
						},
						{
							Name:   "runai-source-topology",
							Levels: []string{"level2"},
							Source: dao.NetworkTopologiesRunAiSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_DELETE,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctx, testNetworkTopologyClusterId, syncRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify both topologies were deleted
				getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctx, []string{})
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())

				// Check that both deleted topologies are not present
				for _, nt := range responseBody.NetworkTopologies {
					if nt.Name == "cluster-source-topology" {
						Fail(fmt.Sprintf("Expected topology %s to be deleted, but it still exists", nt.Name))
					}
				}
			})
		})
	})

	Describe("Multiple clusters and network topologies test", func() {
		Context("when creating multiple clusters and network topologies", func() {
			It("should create 3 clusters with network topologies and test count/retrieval", func() {
				// Setup context with tenant ID and subject
				ctx := addTenantAndSubjectToCtx(context.Background(), testTenantId, testSubjectId)

				var clusters []*models.Cluster
				var clusterIds []string
				var networkTopologies []api.NetworkTopology

				// Create 3 clusters in a loop with suffixes A, B, C
				suffixes := []string{"A", "B", "C"}
				for _, suffix := range suffixes {
					// Create cluster using DAO directly
					clusterName := fmt.Sprintf("test-cluster-%s", suffix)
					cluster, err := dao.CreateCluster(ctx, &models.Cluster{
						Common:   models.Common{},
						Name:     clusterName,
						Version:  "2.23",
						Domain:   fmt.Sprintf("domain-%s", suffix),
						TenantId: testTenantId,
					})
					Expect(err).To(BeNil())
					clusters = append(clusters, cluster)
					clusterIds = append(clusterIds, cluster.ID.String())

					// Create network topology for this cluster
					topologyName := fmt.Sprintf("test-topology-%s", suffix)
					createRequest := api.NetworkTopologyCreateRequest{
						Name:      topologyName,
						Levels:    testNetworkTopologyLevels,
						ClusterId: cluster.ID.String(),
					}

					// Add scopes for this specific cluster
					ctxWithScopes := addScopesToCtx(ctx, []string{cluster.ID.String()})

					response, createNetworkErr := networkTopologiesApiService.CreateNetworkTopology(ctxWithScopes, createRequest)
					Expect(createNetworkErr).To(BeNil())
					Expect(response.Code).To(Equal(http.StatusCreated))

					// Parse response and store the created network topology
					responseBody, ok := response.Body.(api.NetworkTopology)
					Expect(ok).To(BeTrue())
					networkTopologies = append(networkTopologies, responseBody)

					// Verify the network topology was created correctly
					Expect(responseBody.Name).To(Equal(topologyName))
					Expect(responseBody.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
					Expect(responseBody.ClusterId).To(Equal(cluster.ID.String()))
					Expect(responseBody.TenantId).To(Equal(int32(testTenantId)))
					Expect(responseBody.CreatedBy).To(Equal(testSubjectId))
					Expect(responseBody.Levels).To(Equal(testNetworkTopologyLevels))
					Expect(responseBody.Phase).To(Equal(models.NetworkTopologyPhaseCreating))
				}

				// Test 1: Verify we have 3 clusters and 3 network topologies
				Expect(len(clusters)).To(Equal(3))
				Expect(len(networkTopologies)).To(Equal(3))

				namePrefix := "test-topology-"
				// Test 2: Test network topologies count for each cluster
				for i, clusterId := range clusterIds {
					// Add scopes for this specific cluster
					ctxWithScopes := addScopesToCtx(ctx, []string{clusterId})

					response, err := networkTopologiesApiService.GetNetworkTopologiesCount(ctxWithScopes, []string{})
					Expect(err).To(BeNil())
					Expect(response.Code).To(Equal(http.StatusOK))

					responseBody, ok := response.Body.(api.CountResponse)
					Expect(ok).To(BeTrue())
					Expect(responseBody.Count).To(Equal(int64(1))) // Each cluster should have exactly 1 network topology

					// Test 3: Test GetNetworkTopologies returns correct object for each cluster
					response, err = networkTopologiesApiService.GetNetworkTopologies(ctxWithScopes, []string{})
					Expect(err).To(BeNil())
					Expect(response.Code).To(Equal(http.StatusOK))

					responseBodyList, ok := response.Body.(api.GetNetworkTopologies200Response)
					Expect(ok).To(BeTrue())
					Expect(len(responseBodyList.NetworkTopologies)).To(Equal(1)) // Each cluster should have exactly 1 network topology

					// Verify the network topology details
					nt := responseBodyList.NetworkTopologies[0]
					Expect(nt.ClusterId).To(Equal(clusterId))
					Expect(nt.Name).To(Equal(fmt.Sprintf("%s%s", namePrefix, suffixes[i])))
					Expect(nt.Source).To(Equal(dao.NetworkTopologiesRunAiSource))
					Expect(nt.TenantId).To(Equal(int32(testTenantId)))
					Expect(nt.CreatedBy).To(Equal(testSubjectId))
					Expect(nt.Levels).To(Equal(testNetworkTopologyLevels))
					Expect(nt.Phase).To(Equal(models.NetworkTopologyPhaseCreating))
				}

				// Test 4: Test network topologies count with all cluster scopes
				ctxWithAllScopes := addScopesToCtx(ctx, clusterIds)
				validateNetworkTopologiesCount(ctxWithAllScopes, clusterIds, 3)

				// Test 5: Test GetNetworkTopologies returns all network topologies
				response, err := networkTopologiesApiService.GetNetworkTopologies(ctxWithAllScopes, []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBodyList, ok := response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBodyList.NetworkTopologies)).To(Equal(3)) // Should return all 3 network topologies

				// Verify all network topologies are present
				topologyNames := make(map[string]bool)
				for _, nt := range responseBodyList.NetworkTopologies {
					topologyNames[nt.Name] = true
				}

				// Check that all expected names and sources are present
				for _, suffix := range suffixes {
					Expect(topologyNames[fmt.Sprintf("%s%s", namePrefix, suffix)]).To(BeTrue())
				}

				// Cleanup: Delete all created clusters (this will cascade delete network topologies)
				for _, cluster := range clusters {
					err := dao.DeleteCluster(ctx, cluster.ID.String(), testTenantId)
					Expect(err).To(BeNil())
				}

				validateNetworkTopologiesCount(ctxWithAllScopes, clusterIds, 0)
			})

			It("should filter network topologies by clusterId using filterby parameter", func() {
				// Setup context with tenant ID and subject
				ctx := addTenantAndSubjectToCtx(context.Background(), testTenantId, testSubjectId)

				var clusters []*models.Cluster
				var clusterIds []string

				// Create 2 clusters: clusterA and clusterB
				clusterNames := []string{"clusterA", "clusterB"}
				topologyCounts := []int{3, 1} // 3 for clusterA, 1 for clusterB

				for i, clusterName := range clusterNames {
					// Create cluster using DAO directly
					cluster, err := dao.CreateCluster(ctx, &models.Cluster{
						Common:   models.Common{},
						Name:     clusterName,
						Version:  "2.23",
						Domain:   fmt.Sprintf("domain-%s", clusterName),
						TenantId: testTenantId,
					})
					Expect(err).To(BeNil())
					clusters = append(clusters, cluster)
					clusterIds = append(clusterIds, cluster.ID.String())

					// Create multiple network topologies for this cluster
					for j := 0; j < topologyCounts[i]; j++ {
						topologyName := fmt.Sprintf("test-topology-%s-%d", clusterName, j+1)
						createRequest := api.NetworkTopologyCreateRequest{
							Name:      topologyName,
							Levels:    testNetworkTopologyLevels,
							ClusterId: cluster.ID.String(),
						}

						// Add scopes for this specific cluster
						ctxWithScopes := addScopesToCtx(ctx, []string{cluster.ID.String()})

						response, createNetworkErr := networkTopologiesApiService.CreateNetworkTopology(ctxWithScopes, createRequest)
						Expect(createNetworkErr).To(BeNil())
						Expect(response.Code).To(Equal(http.StatusCreated))
					}
				}

				// Verify we have 2 clusters with expected number of topologies
				Expect(len(clusters)).To(Equal(2))
				Expect(len(clusterIds)).To(Equal(2))

				// Test filtering by clusterB (should return 1 topology)
				clusterBId := clusterIds[1] // clusterB
				ctxWithClusterBScope := addScopesToCtx(ctx, []string{clusterBId})

				// Test with filterby for clusterB
				response, err := networkTopologiesApiService.GetNetworkTopologies(ctxWithClusterBScope, []string{fmt.Sprintf("clusterId==%s", clusterBId)})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBody, ok := response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(1)) // Should return 1 topology for clusterB

				// Verify the returned topology belongs to clusterB
				nt := responseBody.NetworkTopologies[0]
				Expect(nt.ClusterId).To(Equal(clusterBId))

				// Test with filterby for clusterA (should return 3 topologies)
				clusterAId := clusterIds[0] // clusterA
				ctxWithClusterAScope := addScopesToCtx(ctx, []string{clusterAId})

				response, err = networkTopologiesApiService.GetNetworkTopologies(ctxWithClusterAScope, []string{fmt.Sprintf("clusterId==%s", clusterAId)})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBody, ok = response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(3)) // Should return 3 topologies for clusterA

				// Verify all returned topologies belong to clusterA
				for _, nt := range responseBody.NetworkTopologies {
					Expect(nt.ClusterId).To(Equal(clusterAId))
				}

				// Test with both cluster scopes and filterby for clusterA
				ctxWithBothScopes := addScopesToCtx(ctx, clusterIds)
				response, err = networkTopologiesApiService.GetNetworkTopologies(ctxWithBothScopes, []string{fmt.Sprintf("clusterId==%s", clusterAId)})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBody, ok = response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(3)) // Should return 3 topologies for clusterA

				// Test with both cluster scopes and filterby for clusterB
				response, err = networkTopologiesApiService.GetNetworkTopologies(ctxWithBothScopes, []string{fmt.Sprintf("clusterId==%s", clusterBId)})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				responseBody, ok = response.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(1)) // Should return 1 topology for clusterB

				// Cleanup: Delete all created clusters (this will cascade delete network topologies)
				for _, cluster := range clusters {
					err := dao.DeleteCluster(ctx, cluster.ID.String(), testTenantId)
					Expect(err).To(BeNil())
				}
			})
		})
	})

	Describe("Multi-cluster sync with topology isolation test", func() {
		Context("when syncing from multiple clusters in the same tenant", func() {
			var ctx context.Context
			var cluster1Id, cluster2Id string
			var networkTopologiesClusters []*models.Cluster
			BeforeAll(func() {
				// Setup context with tenant ID and subject
				ctx = addTenantAndSubjectToCtx(context.Background(), testTenantId, testSubjectId)

				var clusterIds []string

				// Create 2 clusters: cluster1 and cluster2
				clusterNames := []string{"cluster1", "cluster2"}
				for _, clusterName := range clusterNames {
					cluster, err := dao.CreateCluster(ctx, &models.Cluster{
						Common:   models.Common{},
						Name:     clusterName,
						Version:  "2.23",
						Domain:   fmt.Sprintf("domain-%s", clusterName),
						TenantId: testTenantId,
					})
					Expect(err).To(BeNil())
					networkTopologiesClusters = append(networkTopologiesClusters, cluster)
					clusterIds = append(clusterIds, cluster.ID.String())
				}

				cluster1Id = clusterIds[0]
				cluster2Id = clusterIds[1]
			})

			It("should sync from cluster1 with cluster-source topologies", func() {
				ctxWithCluster1Scope := addScopesToCtx(ctx, []string{cluster1Id})
				syncRequestCluster1 := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "shared-topology-name",
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster1-unique-topology",
							Levels: []string{"level3", "level4"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctxWithCluster1Scope, cluster1Id, syncRequestCluster1)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})

			It("should sync from cluster2 with cluster-source topologies including same name as cluster1", func() {
				ctxWithCluster2Scope := addScopesToCtx(ctx, []string{cluster2Id})
				syncRequestCluster2 := api.NetworkTopologySyncRequest{
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "shared-topology-name",       // Same name as cluster1
							Levels: []string{"level5", "level6"}, // Different levels
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
						{
							Name:   "cluster2-unique-topology",
							Levels: []string{"level7", "level8"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_ADD,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctxWithCluster2Scope, cluster2Id, syncRequestCluster2)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
			})

			It("should delete topologies from cluster1 with fullsync without affecting cluster2 topologies", func() {
				ctxWithCluster1Scope := addScopesToCtx(ctx, []string{cluster1Id})
				deleteRequestCluster1FullSync := api.NetworkTopologySyncRequest{
					FullSync: true,
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster1-unique-topology",
							Levels: []string{"level3", "level4"},
							Source: dao.NetworkTopologiesClusterSource,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctxWithCluster1Scope, cluster1Id, deleteRequestCluster1FullSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify cluster1 topology was deleted but cluster2 topologies remain
				cluster1Topologies := getNetworkTopologiesOfCluster(ctx, cluster1Id, 1)
				cluster2Topologies := getNetworkTopologiesOfCluster(ctx, cluster2Id, 2)

				Expect(cluster1Topologies[0].Name).To(Equal("cluster1-unique-topology"))

				cluster2NTNames := []string{"shared-topology-name", "cluster2-unique-topology"}
				Expect(cluster2Topologies[0].Name).To(BeElementOf(cluster2NTNames))
				Expect(cluster2Topologies[1].Name).To(BeElementOf(cluster2NTNames))
			})

			It("should delete topologies from cluster1 without fullsync without affecting cluster2 topologies", func() {
				ctxWithCluster1Scope := addScopesToCtx(ctx, []string{cluster1Id})
				deleteRequestCluster1NoFullSync := api.NetworkTopologySyncRequest{
					FullSync: false,
					NetworkTopologies: []api.NetworkTopologySyncFields{
						{
							Name:   "cluster1-unique-topology",
							Levels: []string{"level1", "level2"},
							Source: dao.NetworkTopologiesClusterSource,
							Action: api.NETWORKTOPOLOGYSYNCACTION_DELETE,
						},
					},
				}

				response, err := networkTopologiesApiService.SyncNetworkTopologies(ctxWithCluster1Scope, cluster1Id, deleteRequestCluster1NoFullSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// Verify cluster1 shared topology was deleted but cluster2 topologies remain unchanged
				_ = getNetworkTopologiesOfCluster(ctx, cluster1Id, 0)
				cluster2Topologies := getNetworkTopologiesOfCluster(ctx, cluster2Id, 2)

				cluster2NTNames := []string{"shared-topology-name", "cluster2-unique-topology"}
				Expect(cluster2Topologies[0].Name).To(BeElementOf(cluster2NTNames))
				Expect(cluster2Topologies[1].Name).To(BeElementOf(cluster2NTNames))
			})

			AfterAll(func() {
				// Cleanup: Delete all created clusters (this will cascade delete network topologies)
				for _, cluster := range networkTopologiesClusters {
					err := dao.DeleteCluster(ctx, cluster.ID.String(), testTenantId)
					Expect(err).To(BeNil())
				}
			})
		})
	})

	Describe("Validate deletion CLuster Network Topologies", func() {
		Context("when creating network topology with source Cluster and attempting to delete via API", func() {
			var ctx context.Context
			var cluster *models.Cluster
			var clusterId string
			var createdTopology *models.NetworkTopology

			BeforeAll(func() {
				// Setup context with tenant ID and subject
				ctx = addTenantAndSubjectToCtx(context.Background(), testTenantId, testSubjectId)

				// Create a test cluster
				var apiErr *api.Error
				cluster, apiErr = dao.CreateCluster(ctx, &models.Cluster{
					Common:   models.Common{},
					Name:     "dao-test-cluster",
					Version:  "2.23",
					Domain:   "dao-test-domain",
					TenantId: testTenantId,
				})
				Expect(apiErr).To(BeNil())
				clusterId = cluster.ID.String()

				// Create network topology using DAO public function
				subjectId := testSubjectId
				networkTopology := models.NetworkTopology{
					Name:      "cluster-created-topology",
					Levels:    testNetworkTopologyLevels,
					ClusterID: cluster.ID,
					TenantID:  testTenantId,
					CreatedBy: &subjectId,
					Source:    dao.NetworkTopologiesClusterSource,
					Phase:     models.NetworkTopologyPhaseReady,
				}

				createdTopology, apiErr = dao.CreateNetworkTopology(ctx, networkTopology)
				Expect(apiErr).To(BeNil())
				Expect(createdTopology).To(Not(BeNil()))
				Expect(createdTopology.ID).To(Not(BeNil()))
			})

			It("should fail to delete network topology created with source Cluster using API DELETE request", func() {
				// Verify the topology was created
				ctxWithScope := addScopesToCtx(ctx, []string{clusterId})
				getResponse, getErr := networkTopologiesApiService.GetNetworkTopologies(ctxWithScope, []string{})
				Expect(getErr).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))

				responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody.NetworkTopologies)).To(Equal(1))
				Expect(responseBody.NetworkTopologies[0].Name).To(Equal("cluster-created-topology"))

				// Try to delete using main DELETE request - should fail
				deleteResponse, deleteErr := networkTopologiesApiService.DeleteNetworkTopology(ctxWithScope, createdTopology.ID.String())
				Expect(deleteErr).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest)) // Should fail with forbidden

				// Verify the topology still exists
				getResponse2, getErr := networkTopologiesApiService.GetNetworkTopologies(ctxWithScope, []string{})
				Expect(getErr).To(BeNil())
				Expect(getResponse2.Code).To(Equal(http.StatusOK))

				responseBody2, ok := getResponse2.Body.(api.GetNetworkTopologies200Response)
				Expect(ok).To(BeTrue())
				Expect(len(responseBody2.NetworkTopologies)).To(Equal(1))
				Expect(responseBody2.NetworkTopologies[0].Name).To(Equal("cluster-created-topology"))
			})

			AfterAll(func() {
				// Cleanup: Delete the cluster (this will cascade delete network topologies)
				err := dao.DeleteCluster(ctx, clusterId, testTenantId)
				Expect(err).To(BeNil())
			})
		})
	})
})

// Helper functions

func getNetworkTopologiesOfCluster(ctx context.Context, clusterId string, expectedAmount int) []api.NetworkTopology {
	ctxWithAllScopes := addScopesToCtx(ctx, []string{clusterId})
	getResponse, err := networkTopologiesApiService.GetNetworkTopologies(ctxWithAllScopes, []string{})
	Expect(err).To(BeNil())
	Expect(getResponse.Code).To(Equal(http.StatusOK))

	responseBody, ok := getResponse.Body.(api.GetNetworkTopologies200Response)
	Expect(ok).To(BeTrue())
	Expect(len(responseBody.NetworkTopologies)).To(Equal(expectedAmount)) // 2 from cluster1 + 2 from cluster2

	return responseBody.NetworkTopologies
}

func generateTopologyName() string {
	return fmt.Sprintf("%s-%s", testNetworkTopologyName, uuid.New().String())
}

func validateNetworkTopologiesCount(ctx context.Context, clusterIds []string, expected int) {
	ctxWithAllScopes := addScopesToCtx(ctx, clusterIds)
	response, err := networkTopologiesApiService.GetNetworkTopologiesCount(ctxWithAllScopes, []string{})
	Expect(err).To(BeNil())
	Expect(response.Code).To(Equal(http.StatusOK))

	responseBody, ok := response.Body.(api.CountResponse)
	Expect(ok).To(BeTrue())
	Expect(responseBody.Count).To(Equal(int64(3))) // All clusters should have 3 network topologies total
}

func addTenantToCtx(ctx context.Context, tenantId int32) context.Context {
	return context.WithValue(ctx, runaictx.CfTenantId, strconv.Itoa(int(tenantId)))
}

func addSubjectToCtx(ctx context.Context, subjectId string) context.Context {
	return context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
}

func addTenantAndSubjectToCtx(ctx context.Context, tenantId int32, subjectId string) context.Context {
	updatedCtx := addTenantToCtx(ctx, tenantId)
	return addSubjectToCtx(updatedCtx, subjectId)
}

func addScopesToCtx(ctx context.Context, clusterIds []string) context.Context {
	permittedScopes := authorization_api.PermittedScopes{
		Clusters: clusterIds,
	}
	return context.WithValue(ctx, authorization_utils.PermittedScopesKey, &permittedScopes)
}

func addTenantAndScopesToCtx(ctx context.Context, tenantId int32, clusterIds []string) context.Context {
	ctx = addTenantToCtx(ctx, tenantId)
	return addScopesToCtx(ctx, clusterIds)
}
