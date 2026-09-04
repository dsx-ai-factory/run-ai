/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"time"

	"github.com/jinzhu/copier"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

var _ = Describe("asset compliance tests", func() {
	var policyMockUpdateTime time.Time
	var randoms *testutils.RandomValues

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		org_unit.SetMock(org_unit.NewSimpleMock(randoms))

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

		checker.InitClustersCache(testing.GetDefaultClustersCacheConfig())
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
	})

	Context("list assets to include compliance info - environment", func() {
		It("one asset should be ok, one not compliant", func() {
			workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
					EnvironmentVariables: *policyService.NewNullableInstancesRules(&policyService.InstancesRules{
						Instances: *policyService.NewNullableItemRules(&policyService.ItemRules{
							CanAdd: *policyService.NewNullableBool(convutils.Ref(false)),
						}),
					}),
					Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
						RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
							Max: *policyService.NewNullableInt32(convutils.Ref(int32(40))),
						}),
						RunAsGid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
							Min: *policyService.NewNullableInt32(convutils.Ref(int32(50))),
						}),
					}),
				}),
			}

			trainingPolicy := policyService.TrainingPolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
					Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
						RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
							Min: *policyService.NewNullableInt32(convutils.Ref(int32(40))),
						}),
						RunAsGid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
							Max: *policyService.NewNullableInt32(convutils.Ref(int32(50))),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:         policyMockUpdateTime,
					EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
					EffectiveTrainingPolicyV2:  ConvertToTrainingPolicy(&trainingPolicy),
				},
			})

			goodForWorkspaceBadForTraining := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(39)),
				},
				Compliance: &api.ComplianceInfo{},
			}
			goodForTrainingBadForWorkspace := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
						{
							Name:  convutils.Ref("varname"),
							Value: convutils.Ref("varvalue"),
						},
					},
					RunAsUid: convutils.Ref(int64(51)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			assetsToPatch[goodForWorkspaceBadForTraining.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForWorkspaceBadForTraining.Meta,
				Spec:           &goodForWorkspaceBadForTraining.Spec,
				ComplianceInfo: &goodForWorkspaceBadForTraining.Compliance,
			}
			assetsToPatch[goodForTrainingBadForWorkspace.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForTrainingBadForWorkspace.Meta,
				Spec:           &goodForTrainingBadForWorkspace.Spec,
				ComplianceInfo: &goodForTrainingBadForWorkspace.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}

			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)
			patcher, err := NewComplianceInfoPatcher(context.Background(), int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect((*assetsToPatch[goodForWorkspaceBadForTraining.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForWorkspace := *assetsToPatch[goodForTrainingBadForWorkspace.Meta.Id].ComplianceInfo
			Expect(complianceBadForWorkspace.Compliance).To(Equal(false))
			Expect(len(complianceBadForWorkspace.Reason)).To(Equal(2))

			// we cannot depend on the order of compliance in the reasons array so look specifically for the fields
			// we expect to find there, and then make sure their compliance error is correct
			runAsUidIndex := -1
			environmentVarIndex := -1
			if *complianceBadForWorkspace.Reason[0].Field == "security.runAsUid" {
				runAsUidIndex = 0
			} else {
				runAsUidIndex = 1
			}
			if *complianceBadForWorkspace.Reason[0].Field == "environmentVariables.instances" {
				environmentVarIndex = 0
			} else {
				environmentVarIndex = 1
			}
			Expect(runAsUidIndex).NotTo(Equal(-1))
			Expect(environmentVarIndex).NotTo(Equal(-1))

			Expect(complianceBadForWorkspace.Reason[runAsUidIndex].Details).To(ContainSubstring("must be no greater than 40"))
			Expect(complianceBadForWorkspace.Reason[environmentVarIndex].Details).To(ContainSubstring("the administrator prohibited adding new items"))

			patcher, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeTraining, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect((*assetsToPatch[goodForTrainingBadForWorkspace.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForTraining := *assetsToPatch[goodForWorkspaceBadForTraining.Meta.Id].ComplianceInfo
			Expect(complianceBadForTraining.Compliance).To(Equal(false))
			Expect(complianceBadForTraining.Reason[0].Details).To(ContainSubstring("must be no less than 40"))
			Expect(*complianceBadForTraining.Reason[0].Field).To(Equal("security.runAsUid"))
		})
	})

	//  compliance info for assets does not check connections. the reason is that the information in the asset
	//  is incomplete, so we leave this validation of ports to the submission of the workload itself where all
	//  the information is available.
	Context("list assets to include compliance info - ports", func() {
		It("one asset should be ok, one not compliant", func() {
			workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
					Ports: *policyService.NewNullablePortsRules(&policyService.PortsRules{
						Attributes: *policyService.NewNullablePortRules(&policyService.PortRules{
							Container: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
								Min: *policyService.NewNullableInt32(convutils.Ref(int32(8000))),
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(9000))),
							}),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:         policyMockUpdateTime,
					EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
				},
			})

			nonCompliant := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					Connections: []api.Connection{
						{
							Name: "hello",
							InternalToolInfo: &api.InternalToolInfo{
								ContainerPort: 4000,
							},
						},
					},
				},
				Compliance: &api.ComplianceInfo{},
			}

			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			assetsToPatch[nonCompliant.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &nonCompliant.Meta,
				Spec:           &nonCompliant.Spec,
				ComplianceInfo: &nonCompliant.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}

			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)
			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect((*assetsToPatch[nonCompliant.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
		})
	})

	Context("list assets to include compliance info - distributed", func() {
		It("one asset should be ok, one not compliant", func() {
			distributedPolicy := policyService.DistributedPolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableDistributedPolicyRulesV2(&policyService.DistributedPolicyRulesV2{
					Master: *policyService.NewNullableReplicaRulesV2(&policyService.ReplicaRulesV2{
						Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
							RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
								Min: *policyService.NewNullableInt32(convutils.Ref(int32(100))),
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(200))),
							}),
						}),
					}),
					Worker: *policyService.NewNullableDistributedPolicyRulesV2Worker(&policyService.DistributedPolicyRulesV2Worker{
						Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
							RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
								Min: *policyService.NewNullableInt32(convutils.Ref(int32(150))),
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(250))),
							}),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:           policyMockUpdateTime,
					EffectiveDistributedPolicyV2: ConvertToDistributedPolicy(&distributedPolicy),
				},
			})

			goodForMasterBadForWorker := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-master",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(100)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			goodForWorkerBadForMaster := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-worker",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(250)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			goodForBoth := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-worker",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(200)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			assetsToPatch[goodForMasterBadForWorker.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForMasterBadForWorker.Meta,
				Spec:           &goodForMasterBadForWorker.Spec,
				ComplianceInfo: &goodForMasterBadForWorker.Compliance,
			}
			assetsToPatch[goodForWorkerBadForMaster.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForWorkerBadForMaster.Meta,
				Spec:           &goodForWorkerBadForMaster.Spec,
				ComplianceInfo: &goodForWorkerBadForMaster.Compliance,
			}
			assetsToPatch[goodForBoth.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForBoth.Meta,
				Spec:           &goodForBoth.Spec,
				ComplianceInfo: &goodForBoth.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}

			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeDistributed, string(openapi.REPLICATYPE_WORKER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect((*assetsToPatch[goodForBoth.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[goodForWorkerBadForMaster.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForMaster := *assetsToPatch[goodForMasterBadForWorker.Meta.Id].ComplianceInfo
			Expect(complianceBadForMaster.Compliance).To(Equal(false))
			Expect(*complianceBadForMaster.Reason[0].Field).To(Equal("security.runAsUid"))
			Expect(complianceBadForMaster.Reason[0].Details).To(ContainSubstring("must be no less than 150"))

			patcher, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeDistributed, string(openapi.REPLICATYPE_MASTER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect((*assetsToPatch[goodForBoth.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[goodForMasterBadForWorker.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForWorker := *assetsToPatch[goodForWorkerBadForMaster.Meta.Id].ComplianceInfo
			Expect(complianceBadForWorker.Compliance).To(Equal(false))
			Expect(*complianceBadForWorker.Reason[0].Field).To(Equal("security.runAsUid"))
			Expect(complianceBadForWorker.Reason[0].Details).To(ContainSubstring("must be no greater than 200"))
		})
	})

	Context("list assets to include compliance info - distributed", func() {
		It("non related fields should not disqualify the asset", func() {
			distributedPolicy := policyService.DistributedPolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableDistributedPolicyRulesV2(&policyService.DistributedPolicyRulesV2{
					Master: *policyService.NewNullableReplicaRulesV2(&policyService.ReplicaRulesV2{
						Image: *policyService.NewNullableStringRules(&policyService.StringRules{
							Required: *policyService.NewNullableBool(convutils.Ref(true)),
						}),
						Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
							HostIpc: *policyService.NewNullableBooleanRules(&policyService.BooleanRules{
								Required: *policyService.NewNullableBool(convutils.Ref(true)),
							}),
						}),
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							HostPath: *policyService.NewNullableHostPathsRules(&policyService.HostPathsRules{
								Attributes: *policyService.NewNullableHostPathRules(&policyService.HostPathRules{
									Path: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
							Nfs: *policyService.NewNullableNfssRules(&policyService.NfssRules{
								Attributes: *policyService.NewNullableNfsRules(&policyService.NfsRules{
									MountPath: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
						Compute: *policyService.NewNullableSupersetRulesAllOfCompute(&policyService.SupersetRulesAllOfCompute{
							CpuCoreLimit: *policyService.NewNullableNumberRules(&policyService.NumberRules{
								Required: *policyService.NewNullableBool(convutils.Ref(true)),
							}),
						}),
					}),
					Worker: *policyService.NewNullableDistributedPolicyRulesV2Worker(&policyService.DistributedPolicyRulesV2Worker{
						ImagePullPolicy: *policyService.NewNullableImagePullPolicyRules(&policyService.ImagePullPolicyRules{
							Required: *policyService.NewNullableBool(convutils.Ref(true)),
						}),
						Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
							HostNetwork: *policyService.NewNullableBooleanRules(&policyService.BooleanRules{
								Required: *policyService.NewNullableBool(convutils.Ref(true)),
							}),
						}),
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							HostPath: *policyService.NewNullableHostPathsRules(&policyService.HostPathsRules{
								Attributes: *policyService.NewNullableHostPathRules(&policyService.HostPathRules{
									MountPath: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
							Nfs: *policyService.NewNullableNfssRules(&policyService.NfssRules{
								Attributes: *policyService.NewNullableNfsRules(&policyService.NfsRules{
									Path: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
						Compute: *policyService.NewNullableSupersetRulesAllOfCompute(&policyService.SupersetRulesAllOfCompute{
							CpuCoreRequest: *policyService.NewNullableNumberRules(&policyService.NumberRules{
								Required: *policyService.NewNullableBool(convutils.Ref(true)),
							}),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:           policyMockUpdateTime,
					EffectiveDistributedPolicyV2: ConvertToDistributedPolicy(&distributedPolicy),
				},
			})

			environmentAsset := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "environment",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec:       api.EnvironmentAssetSpec{},
				Compliance: &api.ComplianceInfo{},
			}

			computeAsset := api.ComputeAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "compute",
					Kind: api.ASSETKIND_COMPUTE,
				},
				Spec:       api.ComputeAssetSpec{},
				Compliance: &api.ComplianceInfo{},
			}

			hostPathAsset := api.HostPathAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "host-path",
					Kind: api.ASSETKIND_HOST_PATH,
				},
				Spec:       api.HostPathSpec{},
				Compliance: &api.ComplianceInfo{},
			}

			nfsAsset := api.NfsAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "nfs",
					Kind: api.ASSETKIND_NFS,
				},
				Spec:       api.NfsAssetSpec{},
				Compliance: &api.ComplianceInfo{},
			}

			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)

			assetsToPatch[environmentAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &environmentAsset.Meta,
				Spec:           &environmentAsset.Spec,
				ComplianceInfo: &environmentAsset.Compliance,
			}
			assetsToPatch[computeAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &computeAsset.Meta,
				Spec:           &computeAsset.Spec,
				ComplianceInfo: &computeAsset.Compliance,
			}
			assetsToPatch[hostPathAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &hostPathAsset.Meta,
				Spec:           &hostPathAsset.Spec,
				ComplianceInfo: &hostPathAsset.Compliance,
			}
			assetsToPatch[nfsAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &nfsAsset.Meta,
				Spec:           &nfsAsset.Spec,
				ComplianceInfo: &nfsAsset.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}

			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeDistributed, string(openapi.REPLICATYPE_WORKER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect(environmentAsset.Compliance.Compliance).To(Equal(false))
			Expect(computeAsset.Compliance.Compliance).To(Equal(false))
			Expect(hostPathAsset.Compliance.Compliance).To(Equal(false))
			Expect(nfsAsset.Compliance.Compliance).To(Equal(false))

			Expect(len(environmentAsset.Compliance.Reason)).To(Equal(2))
			if convutils.Deref(environmentAsset.Compliance.Reason[0].Field, "") == "imagePullPolicy" {
				Expect(*environmentAsset.Compliance.Reason[1].Field).To(Equal("security.hostNetwork"))
			} else {
				Expect(*environmentAsset.Compliance.Reason[0].Field).To(Equal("security.hostNetwork"))
				Expect(*environmentAsset.Compliance.Reason[1].Field).To(Equal("imagePullPolicy"))
			}

			Expect(computeAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("compute.cpuCoreRequest"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))
			Expect(hostPathAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("storage.hostPath.attributes.mountPath"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))
			Expect(nfsAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("storage.nfs.attributes.path"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))

			patcher, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeDistributed, string(openapi.REPLICATYPE_MASTER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect(environmentAsset.Compliance.Compliance).To(Equal(false))
			Expect(computeAsset.Compliance.Compliance).To(Equal(false))
			Expect(hostPathAsset.Compliance.Compliance).To(Equal(false))
			Expect(nfsAsset.Compliance.Compliance).To(Equal(false))

			if convutils.Deref(environmentAsset.Compliance.Reason[0].Field, "") == "image" {
				Expect(*environmentAsset.Compliance.Reason[1].Field).To(Equal("security.hostIpc"))
			} else {
				Expect(*environmentAsset.Compliance.Reason[0].Field).To(Equal("security.hostIpc"))
				Expect(*environmentAsset.Compliance.Reason[1].Field).To(Equal("image"))
			}

			Expect(computeAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("compute.cpuCoreLimit"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))
			Expect(hostPathAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("storage.hostPath.attributes.path"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))
			Expect(nfsAsset.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("storage.nfs.attributes.mountPath"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))
		})
	})

	Context("list assets to include compliance info - compute", func() {
		It("one asset should be ok, one not compliant, but only relevant fields should be taken into account", func() {
			goodAssetId := uuid.NewString()

			effectivePolicyPayload := policyService.WorkspacePolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
					Image: *policyService.NewNullableStringRules(&policyService.StringRules{
						Required: *policyService.NewNullableBool(convutils.Ref(true)),
					}),
					Security: *policyService.NewNullableSupersetRulesAllOfSecurity(&policyService.SupersetRulesAllOfSecurity{
						HostIpc: *policyService.NewNullableBooleanRules(&policyService.BooleanRules{
							Required: *policyService.NewNullableBool(convutils.Ref(true)),
						}),
					}),
					Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
						HostPath: *policyService.NewNullableHostPathsRules(&policyService.HostPathsRules{
							Attributes: *policyService.NewNullableHostPathRules(&policyService.HostPathRules{
								Path: *policyService.NewNullableStringRules(&policyService.StringRules{
									Required: *policyService.NewNullableBool(convutils.Ref(true)),
								}),
							}),
						}),
					}),
					Compute: *policyService.NewNullableSupersetRulesAllOfCompute(&policyService.SupersetRulesAllOfCompute{
						CpuCoreLimit: *policyService.NewNullableNumberRules(&policyService.NumberRules{
							Min: *policyService.NewNullableFloat32(convutils.Ref(float32(1))),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:         policyMockUpdateTime,
					EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&effectivePolicyPayload),
				},
			})

			goodAsset := api.ComputeAsset{
				Meta: api.AssetMeta{
					Id:   goodAssetId,
					Name: "good",
					Kind: api.ASSETKIND_COMPUTE,
				},
				Spec: api.ComputeAssetSpec{
					CpuCoreLimit: convutils.Ref(2.0),
				},
				Compliance: &api.ComplianceInfo{},
			}
			badAsset := api.ComputeAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "bad",
					Kind: api.ASSETKIND_COMPUTE,
				},
				Spec: api.ComputeAssetSpec{
					CpuCoreLimit: convutils.Ref(0.0),
				},
				Compliance: &api.ComplianceInfo{},
			}
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodAsset.Meta,
				Spec:           &goodAsset.Spec,
				ComplianceInfo: &goodAsset.Compliance,
			}
			assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &badAsset.Meta,
				Spec:           &badAsset.Spec,
				ComplianceInfo: &badAsset.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}

			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())
			Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

			Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Field:   convutils.Ref("compute.cpuCoreLimit"),
					Details: "must be no less than 1",
					Rule:    convutils.Ref(api.POLICYRULEENUM_MIN),
				},
			}))
		})

		Context("list assets to include compliance info - defaultFrom", func() {
			It("calculated fields must take into account", func() {
				effectivePolicyPayload := policyService.WorkspacePolicyDefaultsAndRulesV2{
					Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
						Compute: *policyService.NewNullableSupersetRulesAllOfCompute(&policyService.SupersetRulesAllOfCompute{
							CpuCoreLimit: *policyService.NewNullableNumberRules(&policyService.NumberRules{
								Max: *policyService.NewNullableFloat32(convutils.Ref(float32(10))),
								DefaultFrom: *policyService.NewNullableDefaultFromRule(&policyService.DefaultFromRule{
									Field:  *policyService.NewNullableString(convutils.Ref("compute.cpuCoreRequest")),
									Factor: *policyService.NewNullableFloat64(convutils.Ref(4.0)),
								}),
							}),
						}),
					}),
				}

				rulesAsJson, _ := gojsonq.LoadJson(effectivePolicyPayload.Rules.Get())
				orderWorker, circleErr := policy_compliance.NewDefaultsFromOrder(rulesAsJson)
				Expect(circleErr).To(BeNil())

				effectivePolicyPayload.SetDefaultFromOrder(policyService.DefaultFromOrder{
					Worker: convutils.Ref(map[string][]string(orderWorker)),
				})

				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt:         policyMockUpdateTime,
						EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&effectivePolicyPayload),
					},
				})

				badAsset := api.ComputeAsset{
					Meta: api.AssetMeta{
						Id:   uuid.NewString(),
						Name: "bad",
						Kind: api.ASSETKIND_COMPUTE,
					},
					Spec: api.ComputeAssetSpec{
						CpuCoreRequest: convutils.Ref(18.0),
					},
					Compliance: &api.ComplianceInfo{},
				}

				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
				}

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Field:   convutils.Ref("compute.cpuCoreLimit"),
						Details: "must be no greater than 10",
						Rule:    convutils.Ref(api.POLICYRULEENUM_MAX),
					},
				}))
			})
		})

		Context("list assets to include compliance info - hostPath", func() {
			It("one asset should be ok, one not compliant", func() {
				goodAssetId := uuid.NewString()
				badAssetId := uuid.NewString()

				effectivePolicyPayload := policyService.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: *policyService.NewNullableWorkspacePolicyDefaultsV2(&policyService.WorkspacePolicyDefaultsV2{
						Storage: *policyService.NewNullableSupersetDefaultsAllOfStorage(&policyService.SupersetDefaultsAllOfStorage{
							HostPath: *policyService.NewNullableHostPathsDefaults(&policyService.HostPathsDefaults{
								Instances: []policyService.HostPathInstance{
									{
										Name: *policyService.NewNullableString(&goodAssetId),
									},
								},
							}),
						}),
					}),
					Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							HostPath: *policyService.NewNullableHostPathsRules(&policyService.HostPathsRules{
								Attributes: *policyService.NewNullableHostPathRules(&policyService.HostPathRules{
									ReadOnly: *policyService.NewNullableBooleanRules(&policyService.BooleanRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
					}),
				}

				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt:         policyMockUpdateTime,
						EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&effectivePolicyPayload),
					},
				})

				goodAsset := api.HostPathAsset{
					Meta: api.AssetMeta{
						Id:   goodAssetId,
						Name: "good",
						Kind: api.ASSETKIND_HOST_PATH,
					},
					Spec: api.HostPathSpec{
						Path:     convutils.Ref("/data"),
						ReadOnly: convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}
				badAsset := api.HostPathAsset{
					Meta: api.AssetMeta{
						Id:   badAssetId,
						Name: "bad",
						Kind: api.ASSETKIND_HOST_PATH,
					},
					Spec: api.HostPathSpec{
						Path: convutils.Ref("/data"),
					},
					Compliance: &api.ComplianceInfo{},
				}
				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &goodAsset.Meta,
					Spec:           &goodAsset.Spec,
					ComplianceInfo: &goodAsset.Compliance,
				}
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
				}

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				err = patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				// bad assets cannot be imposed so even though both good and bad are in the list, only good is marked
				// as imposed
				Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Imposed).To(Equal(true))
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Imposed).To(Equal(false))

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Field:   convutils.Ref("storage.hostPath.attributes.readOnly"),
						Details: "value must be provided for required fields",
						Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
					},
				}))
			})
		})

		Context("list assets to include compliance info - NFS", func() {
			It("one asset should be ok, one not compliant", func() {
				trainingPolicy := policyService.TrainingPolicyDefaultsAndRulesV2{
					Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							Nfs: *policyService.NewNullableNfssRules(&policyService.NfssRules{
								Attributes: *policyService.NewNullableNfsRules(&policyService.NfsRules{
									ReadOnly: *policyService.NewNullableBooleanRules(&policyService.BooleanRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
					}),
				}

				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt:        policyMockUpdateTime,
						EffectiveTrainingPolicyV2: ConvertToTrainingPolicy(&trainingPolicy),
					},
				})

				goodAsset := api.NfsAsset{
					Meta: api.AssetMeta{
						Id:   uuid.NewString(),
						Name: "good",
						Kind: api.ASSETKIND_NFS,
					},
					Spec: api.NfsAssetSpec{
						Path:     convutils.Ref("/data"),
						ReadOnly: convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}
				badAsset := api.NfsAsset{
					Meta: api.AssetMeta{
						Id:   uuid.NewString(),
						Name: "bad",
						Kind: api.ASSETKIND_NFS,
					},
					Spec: api.NfsAssetSpec{
						Path: convutils.Ref("/data"),
					},
					Compliance: &api.ComplianceInfo{},
				}
				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &goodAsset.Meta,
					Spec:           &goodAsset.Spec,
					ComplianceInfo: &goodAsset.Compliance,
				}
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
				}

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeTraining, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				err = patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Field:   convutils.Ref("storage.nfs.attributes.readOnly"),
						Details: "value must be provided for required fields",
						Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
					},
				}))
			})
		})

		Context("list assets to include compliance info - S3", func() {
			It("one asset should be ok, one not compliant", func() {
				workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
					Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							S3: *policyService.NewNullableS3sRules(&policyService.S3sRules{
								Attributes: *policyService.NewNullableS3Rules(&policyService.S3Rules{
									Bucket: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
					}),
				}

				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt:         policyMockUpdateTime,
						EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
					},
				})

				existingName := "existing-dept-1"
				nameValidationQuery, _ := scope.GetNameValidationQuery(context.Background(), existingName, api.SCOPE_DEPARTMENT, nil, &randoms.DepartmentId, &randoms.ClusterId, nil, api.ASSETKIND_ACCESS_KEY, nil)

				existingAccessKey, err := dao.CreateAccessKey(context.Background(), api.AccessKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:         "existing-dept-1",
						Scope:        api.SCOPE_DEPARTMENT,
						ClusterId:    convutils.Ref(randoms.ClusterId),
						DepartmentId: convutils.Ref(randoms.DepartmentId),
					},
					Spec: api.AccessKeyCreationSpec{
						ExistingSecretName: convutils.Ref("dept-1-exist"),
					},
				}, 0, nameValidationQuery)
				Expect(err).To(BeNil())

				var existingAccessKeyCreReq api.AccessKeyCreationRequest
				copier.Copy(&existingAccessKeyCreReq, &existingAccessKey)

				missingName := "missing-dept-1"
				nameValidationQuery, _ = scope.GetNameValidationQuery(context.Background(), missingName, api.SCOPE_DEPARTMENT, nil, &randoms.DepartmentId, &randoms.ClusterId, nil, api.ASSETKIND_ACCESS_KEY, nil)

				missingAccessKey, err := dao.CreateAccessKey(context.Background(), api.AccessKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:         "missing-dept-1",
						Scope:        api.SCOPE_DEPARTMENT,
						ClusterId:    convutils.Ref(randoms.ClusterId),
						DepartmentId: convutils.Ref(randoms.DepartmentId),
					},
					Spec: api.AccessKeyCreationSpec{
						ExistingSecretName: convutils.Ref("dept-1-missing"),
					},
				}, 0, nameValidationQuery)
				Expect(err).To(BeNil())

				goodAsset := api.S3Asset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "good",
						Kind:      api.ASSETKIND_S3,
					},
					Spec: api.S3AssetSpec{
						Path:   convutils.Ref("/data"),
						Bucket: convutils.Ref("s3-bucket"),
					},
					Compliance: &api.ComplianceInfo{},
				}

				badAsset := api.S3Asset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "bad",
						Kind:      api.ASSETKIND_S3,
					},
					Spec: api.S3AssetSpec{
						Path: convutils.Ref("/data"),
					},
					Compliance: &api.ComplianceInfo{},
				}

				existingInCluster := api.S3Asset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "s3-existing-dept-1",
						Kind:      api.ASSETKIND_S3,
					},
					Spec: api.S3AssetSpec{
						Path:             convutils.Ref("/data"),
						Bucket:           convutils.Ref("s3-bucket"),
						AccessKeyAssetId: convutils.Ref(existingAccessKey.Meta.Id),
					},
					Compliance: &api.ComplianceInfo{},
				}

				missingInCluster := api.S3Asset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "s3-missing-dept-1",
						Kind:      api.ASSETKIND_S3,
					},
					Spec: api.S3AssetSpec{
						Path:             convutils.Ref("/data"),
						Bucket:           convutils.Ref("s3-bucket"),
						AccessKeyAssetId: convutils.Ref(missingAccessKey.Meta.Id),
					},
					Compliance: &api.ComplianceInfo{},
				}

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithAccessKey(randoms.ClusterId, &existingAccessKeyCreReq, nil))

				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &goodAsset.Meta,
					Spec:           &goodAsset.Spec,
					ComplianceInfo: &goodAsset.Compliance,
				}
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
				}
				assetsToPatch[existingInCluster.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &existingInCluster.Meta,
					Spec:           &existingInCluster.Spec,
					ComplianceInfo: &existingInCluster.Compliance,
				}
				assetsToPatch[missingInCluster.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &missingInCluster.Meta,
					Spec:           &missingInCluster.Spec,
					ComplianceInfo: &missingInCluster.Compliance,
				}

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				err = patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))
				Expect((*assetsToPatch[existingInCluster.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[missingInCluster.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Field:   convutils.Ref("storage.s3.attributes.bucket"),
						Details: "value must be provided for required fields",
						Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
					},
				}))

				Expect((*assetsToPatch[missingInCluster.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Details: "This data source can't be selected because its credentials' status is \"Issues found\". Contact your administrator",
					},
				}))
			})
		})

		Context("list assets to include compliance info - Git", func() {
			It("one asset should be ok, one not compliant", func() {
				workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
					Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
						Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
							Git: *policyService.NewNullableGitsRules(&policyService.GitsRules{
								Attributes: *policyService.NewNullableGitRules(&policyService.GitRules{
									Revision: *policyService.NewNullableStringRules(&policyService.StringRules{
										Required: *policyService.NewNullableBool(convutils.Ref(true)),
									}),
								}),
							}),
						}),
					}),
				}

				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt:         policyMockUpdateTime,
						EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
					},
				})

				nameValidationQuery, _ := scope.GetNameValidationQuery(context.Background(), "existing-dept-1", api.SCOPE_DEPARTMENT, nil, &randoms.DepartmentId, &randoms.ClusterId, nil, api.ASSETKIND_PASSWORD, nil)
				existingPassword, err := dao.CreatePassword(context.Background(), api.PasswordCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:         "existing-dept-1",
						Scope:        api.SCOPE_DEPARTMENT,
						ClusterId:    convutils.Ref(randoms.ClusterId),
						DepartmentId: convutils.Ref(randoms.DepartmentId),
					},
					Spec: api.PasswordCreationSpec{
						ExistingSecretName: convutils.Ref("dept-1-exist"),
					},
				}, 0, nameValidationQuery)
				Expect(err).To(BeNil())

				var existingPasswordCreReq api.PasswordCreationRequest
				copier.Copy(&existingPasswordCreReq, &existingPassword)

				nameValidationQuery, _ = scope.GetNameValidationQuery(context.Background(), "missing-dept-1", api.SCOPE_DEPARTMENT, nil, &randoms.DepartmentId, &randoms.ClusterId, nil, api.ASSETKIND_PASSWORD, nil)
				missingPassword, err := dao.CreatePassword(context.Background(), api.PasswordCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:         "missing-dept-1",
						Scope:        api.SCOPE_DEPARTMENT,
						ClusterId:    convutils.Ref(randoms.ClusterId),
						DepartmentId: convutils.Ref(randoms.DepartmentId),
					},
					Spec: api.PasswordCreationSpec{
						ExistingSecretName: convutils.Ref("dept-1-missing"),
					},
				}, 0, nameValidationQuery)
				Expect(err).To(BeNil())

				goodAsset := api.GitAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "good",
						Kind:      api.ASSETKIND_GIT,
					},
					Spec: api.GitAssetSpec{
						Path:     convutils.Ref("/data"),
						Revision: convutils.Ref("1.0.0"),
					},
					Compliance: &api.ComplianceInfo{},
				}

				badAsset := api.GitAsset{
					Meta: api.AssetMeta{
						Id:   uuid.NewString(),
						Name: "bad",
						Kind: api.ASSETKIND_GIT,
					},
					Spec: api.GitAssetSpec{
						Path: convutils.Ref("/data"),
					},
					Compliance: &api.ComplianceInfo{},
				}

				existingInCluster := api.GitAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						ClusterId: convutils.Ref(randoms.ClusterId),
						Name:      "git-existing-dept-1",
						Kind:      api.ASSETKIND_GIT,
					},
					Spec: api.GitAssetSpec{
						Path:         convutils.Ref("/data"),
						Revision:     convutils.Ref("1.0.0"),
						CredentialId: convutils.Ref(existingPassword.Meta.Id),
					},
					Compliance: &api.ComplianceInfo{},
				}

				missingInCluster := api.GitAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						Name:      "git-missing-dept-1",
						ClusterId: convutils.Ref(randoms.ClusterId),
						Kind:      api.ASSETKIND_GIT,
					},
					Spec: api.GitAssetSpec{
						Path:         convutils.Ref("/data"),
						Revision:     convutils.Ref("1.0.0"),
						CredentialId: convutils.Ref(missingPassword.Meta.Id),
					},
					Compliance: &api.ComplianceInfo{},
				}

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(randoms.ClusterId, &existingPasswordCreReq, nil))

				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &goodAsset.Meta,
					Spec:           &goodAsset.Spec,
					ComplianceInfo: &goodAsset.Compliance,
				}
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
				}
				assetsToPatch[existingInCluster.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &existingInCluster.Meta,
					Spec:           &existingInCluster.Spec,
					ComplianceInfo: &existingInCluster.Compliance,
				}
				assetsToPatch[missingInCluster.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &missingInCluster.Meta,
					Spec:           &missingInCluster.Spec,
					ComplianceInfo: &missingInCluster.Compliance,
				}

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				err = patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				Expect((*assetsToPatch[goodAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				Expect((*assetsToPatch[existingInCluster.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
				Expect((*assetsToPatch[missingInCluster.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))

				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Field:   convutils.Ref("storage.git.attributes.revision"),
						Details: "value must be provided for required fields",
						Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
					},
				}))

				Expect((*assetsToPatch[missingInCluster.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
					{
						Details: "This data source can't be selected because its credentials' status is \"Issues found\". Contact your administrator",
					},
				}))
			})
		})
	})

	Context("list assets to include compliance info - distributed inference", func() {
		It("should validate compliance for leader and worker replicas separately", func() {
			distributedInferencePolicy := policyService.DistributedInferencePolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableDistributedInferenceRulesV2(&policyService.DistributedInferenceRulesV2{
					Leader: *policyService.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyService.DistributedInferenceLeaderWorkerRulesV2{
						Security: *policyService.NewNullableDistributedInferenceLeaderWorkerRulesV2Security(&policyService.DistributedInferenceLeaderWorkerRulesV2Security{
							RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
								Min: *policyService.NewNullableInt32(convutils.Ref(int32(100))),
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(200))),
							}),
						}),
					}),
					Worker: *policyService.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyService.DistributedInferenceLeaderWorkerRulesV2{
						Security: *policyService.NewNullableDistributedInferenceLeaderWorkerRulesV2Security(&policyService.DistributedInferenceLeaderWorkerRulesV2Security{
							RunAsUid: *policyService.NewNullableIntegerRulesOptional(&policyService.IntegerRulesOptional{
								Min: *policyService.NewNullableInt32(convutils.Ref(int32(150))),
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(250))),
							}),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:                    policyMockUpdateTime,
					EffectiveDistributedInferencePolicyV2: ConvertToDistributedInferencePolicy(&distributedInferencePolicy),
				},
			})

			goodForLeaderBadForWorker := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-leader",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(100)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			goodForWorkerBadForLeader := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-worker",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(250)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			goodForBoth := api.EnvironmentAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good-for-both",
					Kind: api.ASSETKIND_ENVIRONMENT,
				},
				Spec: api.EnvironmentAssetSpec{
					RunAsUid: convutils.Ref(int64(175)),
				},
				Compliance: &api.ComplianceInfo{},
			}

			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			assetsToPatch[goodForLeaderBadForWorker.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForLeaderBadForWorker.Meta,
				Spec:           &goodForLeaderBadForWorker.Spec,
				ComplianceInfo: &goodForLeaderBadForWorker.Compliance,
			}
			assetsToPatch[goodForWorkerBadForLeader.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForWorkerBadForLeader.Meta,
				Spec:           &goodForWorkerBadForLeader.Spec,
				ComplianceInfo: &goodForWorkerBadForLeader.Compliance,
			}
			assetsToPatch[goodForBoth.Meta.Id] = response_patching.ResponsePatchingPtrs{
				Meta:           &goodForBoth.Meta,
				Spec:           &goodForBoth.Spec,
				ComplianceInfo: &goodForBoth.Compliance,
			}

			permittedScopes := auth.PermittedScopes{}
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), DistributedInference, string(openapi.REPLICATYPE_LEADER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect((*assetsToPatch[goodForBoth.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[goodForLeaderBadForWorker.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForLeader := *assetsToPatch[goodForWorkerBadForLeader.Meta.Id].ComplianceInfo
			Expect(complianceBadForLeader.Compliance).To(Equal(false))
			Expect(*complianceBadForLeader.Reason[0].Field).To(Equal("security.runAsUid"))
			Expect(complianceBadForLeader.Reason[0].Details).To(ContainSubstring("must be no greater than 200"))

			patcher, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), DistributedInference, string(openapi.REPLICATYPE_WORKER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())

			err = patcher.PatchComplianceInfo(context.Background())
			Expect(err).To(BeNil())

			Expect((*assetsToPatch[goodForBoth.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[goodForWorkerBadForLeader.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))

			complianceBadForWorker := *assetsToPatch[goodForLeaderBadForWorker.Meta.Id].ComplianceInfo
			Expect(complianceBadForWorker.Compliance).To(Equal(false))
			Expect(*complianceBadForWorker.Reason[0].Field).To(Equal("security.runAsUid"))
			Expect(complianceBadForWorker.Reason[0].Details).To(ContainSubstring("must be no less than 150"))
		})
	})

	Context("validation of complyToReplicaType parameter", func() {
		It("should require complyToReplicaType for Distributed workload type", func() {
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			permittedScopes := auth.PermittedScopes{}
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			_, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), Distributed, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("copmlyToReplicaType is mandatory when complyToWorkloadType=Distributed"))
		})

		It("should require complyToReplicaType for DistributedInference workload type", func() {
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			permittedScopes := auth.PermittedScopes{}
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			_, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), DistributedInference, "", assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("copmlyToReplicaType is mandatory when complyToWorkloadType=DistributedInference"))
		})

		It("should reject complyToReplicaType for non-distributed workload types", func() {
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			permittedScopes := auth.PermittedScopes{}
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			_, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, string(openapi.REPLICATYPE_WORKER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("copmlyToReplicaType is not valid when complyToWorkloadType=Workspace"))

			_, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeTraining, string(openapi.REPLICATYPE_LEADER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("copmlyToReplicaType is not valid when complyToWorkloadType=Training"))

			_, err = NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), auth.WorkloadTypeInference, string(openapi.REPLICATYPE_WORKER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("copmlyToReplicaType is not valid when complyToWorkloadType=Inference"))
		})

		It("should accept LEADER replica type for DistributedInference", func() {
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
			permittedScopes := auth.PermittedScopes{}
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:                    policyMockUpdateTime,
					EffectiveDistributedInferencePolicyV2: ConvertToDistributedInferencePolicy(&policyService.DistributedInferencePolicyDefaultsAndRulesV2{}),
				},
			})

			patcher, err := NewComplianceInfoPatcher(context.Background(),
				int32(randoms.ProjectId), DistributedInference, string(openapi.REPLICATYPE_LEADER), assetsToPatch, &permittedScopes, assetClusterStatus)
			Expect(err).To(BeNil())
			Expect(patcher).NotTo(BeNil())
			Expect(patcher.workloadType).To(Equal(DistributedInference))
			Expect(patcher.replicaType).To(Equal(string(openapi.REPLICATYPE_LEADER)))
		})
	})

})
