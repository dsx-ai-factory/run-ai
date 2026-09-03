/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"

	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
)

var _ = Describe("source of rules", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		tree.Restart()

		backendMock = NewBackendMockClient(2, 1)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})
	})

	Context("get effective policy - workspace", func() {
		It("make sure all fields have source of rules", func() {
			//
			//   apply a policy to the tenant
			//
			tenantPolicy := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Args: &api.StringRules{
							Required: convutils.Ref(true),
						},
						EnvironmentVariables: &api.InstancesRules{
							Instances: &api.ItemRules{
								CanAdd: convutils.Ref(false),
							},
						},
						Storage: &api.SupersetRulesAllOfStorage{
							HostPath: &api.HostPathsRules{
								Attributes: &api.HostPathRules{
									Path: &api.StringRules{
										Required: convutils.Ref(true),
									},
									MountPath: &api.StringRules{
										Required: convutils.Ref(true),
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false),
								},
							},
						},
						Ports: &api.PortsRules{
							Attributes: &api.PortRules{
								Container: &api.IntegerRules{
									Min: convutils.Ref(int32(8000)),
								},
							},
							Instances: &api.ItemRules{
								CanAdd: convutils.Ref(false),
							},
						},
					},
				},
			}

			response := applyPolicyForTest(router, urlPolicyWorkspace, makePayload(tenantPolicy), 2, http.MethodPatch)

			var addedPolicy api.WorkspacePolicyV2
			err := json.Unmarshal(response.Body.Bytes(), &addedPolicy)
			Expect(err).To(BeNil())

			Expect(addedPolicy.Effective.Rules).NotTo(BeNil())

			expectedSourceOfRules := api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}

			rules := addedPolicy.Effective.Rules
			Expect(*rules.Args.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.EnvironmentVariables.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Ports.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Ports.Attributes.Container.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Storage.HostPath.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Storage.HostPath.Attributes.Path.SourceOfRule).To(Equal(expectedSourceOfRules))
		})
	})

	Context("get effective policy - distributed", func() {
		It("make sure all fields have source of rules", func() {
			//
			//   apply a policy to the tenant
			//
			tenantPolicy := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Worker: &api.DistributedPolicyRulesV2Worker{
							Args: &api.StringRules{
								Required: convutils.Ref(true),
							},
							Storage: &api.SupersetRulesAllOfStorage{
								HostPath: &api.HostPathsRules{
									Attributes: &api.HostPathRules{
										Path: &api.StringRules{
											Required: convutils.Ref(true),
										},
										MountPath: &api.StringRules{
											Required: convutils.Ref(true),
										},
									},
									Instances: &api.ItemRules{
										CanAdd: convutils.Ref(false),
									},
								},
							},
						},
						Master: &api.ReplicaRulesV2{
							EnvironmentVariables: &api.InstancesRules{
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false),
								},
							},
							Ports: &api.PortsRules{
								Attributes: &api.PortRules{
									Container: &api.IntegerRules{
										Min: convutils.Ref(int32(8000)),
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false),
								},
							},
						},
					},
				},
			}

			response := applyPolicyForTest(router, urlPolicyDistributed, makePayload(tenantPolicy), 2, http.MethodPatch)

			var addedPolicy api.DistributedPolicyV2
			err := json.Unmarshal(response.Body.Bytes(), &addedPolicy)
			Expect(err).To(BeNil())

			Expect(addedPolicy.Effective.Rules).NotTo(BeNil())

			expectedSourceOfRules := api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}

			rules := addedPolicy.Effective.Rules
			Expect(*rules.Worker.Args.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Master.EnvironmentVariables.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Master.Ports.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Master.Ports.Attributes.Container.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Worker.Storage.HostPath.Instances.SourceOfRule).To(Equal(expectedSourceOfRules))
			Expect(*rules.Worker.Storage.HostPath.Attributes.Path.SourceOfRule).To(Equal(expectedSourceOfRules))
		})
	})
})
