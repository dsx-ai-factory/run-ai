/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("calculate effective test", func() {
	stringRules := &api.StringRules{
		Required: convutils.Ref(true),
	}
	imagePullPolicyRules := &api.ImagePullPolicyRules{
		Required: stringRules.Required,
	}
	integerRules := &api.IntegerRules{
		Required: convutils.Ref(true),
	}
	booleanRules := &api.BooleanRules{
		Required: convutils.Ref(true),
	}
	numberRules := &api.NumberRules{
		Required: convutils.Ref(true),
	}
	itemRules := &api.ItemRules{
		CanAdd: convutils.Ref(true),
	}

	Context("fields from parent", func() {
		It("we should decide on the right field", func() {
			ht := HierarchyTree{}

			rules := api.WorkspacePolicyRulesV2{
				Command:         stringRules,
				Args:            stringRules,
				Image:           stringRules,
				ImagePullPolicy: imagePullPolicyRules,
				Probes: &api.ProbesRules{
					Readiness: &api.ProbeRules{
						InitialDelaySeconds: integerRules,
					},
				},
				PodAffinity: &api.PodAffinityRules{
					Key: stringRules,
				},
				EnvironmentVariables: &api.InstancesRules{
					Instances: itemRules,
				},
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsNonRoot:       booleanRules,
					SupplementalGroups: stringRules,
					HostIpc:            booleanRules,
					HostNetwork:        booleanRules,
				},
				Compute: &api.SupersetRulesAllOfCompute{
					CpuCoreRequest:    numberRules,
					LargeShmRequest:   booleanRules,
					GpuDevicesRequest: integerRules,
					GpuPortionLimit:   numberRules,
					ExtendedResources: &api.ExtendedResourcesRules{
						Attributes: &api.ExtendedResourceRules{
							Quantity: stringRules,
						},
						Instances: itemRules,
					},
				},
				Storage: &api.SupersetRulesAllOfStorage{
					Pvc: &api.PvcsRules{
						Attributes: &api.PvcRules{
							ClaimName: stringRules,
							ClaimInfo: &api.ClaimInfoRules{
								StorageClass: stringRules,
								Size:         stringRules,
							},
						},
						Instances: itemRules,
					},
				},
			}

			rulesMap, _ := gojsonq.LoadJson(rules)

			actualRulesFromParent := ht.DetermineSourceOfRules(rulesMap, rulesMap)

			expectedRulesFromParent := RulesFromParentPerField{
				"args":                                 false,
				"command":                              false,
				"environmentVariables":                 false,
				"image":                                false,
				"podAffinity.key":                      false,
				"imagePullPolicy":                      false,
				"probes.readiness.initialDelaySeconds": false,

				"compute.cpuCoreRequest":             true,
				"compute.extendedResources":          true,
				"compute.extendedResources.quantity": false,
				"compute.gpuDevicesRequest":          true,
				"compute.gpuPortionLimit":            true,
				"compute.largeShmRequest":            true,

				"security.hostIpc":            true,
				"security.supplementalGroups": true,
				"security.hostNetwork":        true,
				"security.runAsNonRoot":       true,

				"storage.pvc":                        true,
				"storage.pvc.claimName":              false,
				"storage.pvc.claimInfo.size":         false,
				"storage.pvc.claimInfo.storageClass": false,
			}

			testutils.Expect(actualRulesFromParent).ToEqual(expectedRulesFromParent)
		})
	})

	Context("system rules inherited through an intermediate policy", func() {
		It("allows the current policy to override the system rule", func() {
			ht := HierarchyTree{}

			systemRules := api.TrainingPolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Privileged: &api.BooleanRules{
						CanEdit: convutils.Ref(false),
						SourceOfRule: &api.SourceOfRule{
							Scope: api.SCOPETYPE_SYSTEM,
						},
					},
				},
			}
			projectRules := api.TrainingPolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Privileged: &api.BooleanRules{
						CanEdit: convutils.Ref(true),
					},
				},
			}

			systemRulesMap, err := gojsonq.LoadJson(systemRules)
			Expect(err).To(BeNil())
			projectRulesMap, err := gojsonq.LoadJson(projectRules)
			Expect(err).To(BeNil())

			clusterEffectiveRules, err := ht.MergeRules(dbmodels.JsonBMap(systemRulesMap).Clone(), nil, nil, map[string]interface{}{
				"scope":     api.SCOPETYPE_CLUSTER,
				"clusterId": "cluster-1",
			})
			Expect(err).To(BeNil())

			rulesFromParent := ht.DetermineSourceOfRules(projectRulesMap, clusterEffectiveRules)
			testutils.Expect(rulesFromParent["security.privileged"]).ToEqual(false)

			projectEffectiveRules, err := ht.MergeRules(clusterEffectiveRules, projectRulesMap, rulesFromParent, map[string]interface{}{
				"scope":     api.SCOPETYPE_PROJECT,
				"projectId": 1111,
			})
			Expect(err).To(BeNil())

			privilegedRule := convutils.MapMe(convutils.MapMe(projectEffectiveRules["security"])["privileged"])
			testutils.Expect(privilegedRule["canEdit"]).ToEqual(true)
			testutils.Expect(convutils.MapMe(privilegedRule["sourceOfRule"])["scope"]).ToEqual(api.SCOPETYPE_PROJECT)
		})

		DescribeTable("keeps an explicitly authored ancestor security rule authoritative", func(sourceOfRule api.SourceOfRule) {
			ht := HierarchyTree{}

			parentRules := api.TrainingPolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Privileged: &api.BooleanRules{
						CanEdit:      convutils.Ref(false),
						SourceOfRule: &sourceOfRule,
					},
				},
			}
			projectRules := api.TrainingPolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Privileged: &api.BooleanRules{
						CanEdit: convutils.Ref(true),
					},
				},
			}

			parentRulesMap, err := gojsonq.LoadJson(parentRules)
			Expect(err).To(BeNil())
			projectRulesMap, err := gojsonq.LoadJson(projectRules)
			Expect(err).To(BeNil())

			rulesFromParent := ht.DetermineSourceOfRules(projectRulesMap, parentRulesMap)
			testutils.Expect(rulesFromParent["security.privileged"]).ToEqual(true)

			projectEffectiveRules, err := ht.MergeRules(parentRulesMap, projectRulesMap, rulesFromParent, map[string]interface{}{
				"scope":     api.SCOPETYPE_PROJECT,
				"projectId": 1111,
			})
			Expect(err).To(BeNil())

			privilegedRule := convutils.MapMe(convutils.MapMe(projectEffectiveRules["security"])["privileged"])
			testutils.Expect(privilegedRule["canEdit"]).ToEqual(false)
			testutils.Expect(convutils.MapMe(privilegedRule["sourceOfRule"])["scope"]).ToEqual(sourceOfRule.Scope)
		},
			Entry("tenant rule", api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}),
			Entry("cluster rule", api.SourceOfRule{Scope: api.SCOPETYPE_CLUSTER, ClusterId: convutils.Ref("cluster-1")}),
			Entry("department rule", api.SourceOfRule{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("department-1")}),
		)
	})

	Context("fields from parent - distributed", func() {
		It("we should decide on the right field", func() {
			ht := HierarchyTree{}

			rules := api.DistributedPolicyRulesV2{
				Worker: &api.DistributedPolicyRulesV2Worker{
					Command: stringRules,
					Image:   stringRules,
					Probes: &api.ProbesRules{
						Readiness: &api.ProbeRules{
							InitialDelaySeconds: integerRules,
						},
					},
					EnvironmentVariables: &api.InstancesRules{
						Instances: itemRules,
					},
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsNonRoot: booleanRules,
						HostIpc:      booleanRules,
					},
					Compute: &api.SupersetRulesAllOfCompute{
						CpuCoreRequest:    numberRules,
						LargeShmRequest:   booleanRules,
						GpuDevicesRequest: integerRules,
						GpuPortionLimit:   numberRules,
					},
					Storage: &api.SupersetRulesAllOfStorage{
						Pvc: &api.PvcsRules{
							Attributes: &api.PvcRules{
								ClaimInfo: &api.ClaimInfoRules{
									StorageClass: stringRules,
									Size:         stringRules,
								},
							},
							Instances: itemRules,
						},
					},
				},
				Master: &api.ReplicaRulesV2{
					Args:            stringRules,
					Image:           stringRules,
					ImagePullPolicy: imagePullPolicyRules,
					PodAffinity: &api.PodAffinityRules{
						Key: stringRules,
					},
					EnvironmentVariables: &api.InstancesRules{
						Instances: itemRules,
					},
					Security: &api.SupersetRulesAllOfSecurity{
						SupplementalGroups: stringRules,
						HostNetwork:        booleanRules,
					},
					Compute: &api.SupersetRulesAllOfCompute{
						CpuCoreRequest: numberRules,
						ExtendedResources: &api.ExtendedResourcesRules{
							Attributes: &api.ExtendedResourceRules{
								Quantity: stringRules,
							},
							Instances: itemRules,
						},
					},
					Storage: &api.SupersetRulesAllOfStorage{
						Pvc: &api.PvcsRules{
							Attributes: &api.PvcRules{
								ClaimName: stringRules,
							},
							Instances: itemRules,
						},
					},
				},
			}

			rulesMap, _ := gojsonq.LoadJson(rules)

			actualRulesFromParent := ht.DetermineSourceOfRules(rulesMap, rulesMap)

			expectedRulesFromParent := RulesFromParentPerField{
				"worker.environmentVariables":                 false,
				"worker.probes.readiness.initialDelaySeconds": false,
				"worker.storage.pvc.claimInfo.size":           false,
				"worker.image":                                false,
				"worker.compute.gpuDevicesRequest":            true,
				"worker.compute.gpuPortionLimit":              true,
				"worker.compute.cpuCoreRequest":               true,
				"worker.compute.largeShmRequest":              true,
				"worker.command":                              false,
				"worker.security.hostIpc":                     true,
				"worker.security.runAsNonRoot":                true,
				"worker.storage.pvc":                          true,
				"worker.storage.pvc.claimInfo.storageClass":   false,

				"master.imagePullPolicy":                    false,
				"master.environmentVariables":               false,
				"master.podAffinity.key":                    false,
				"master.args":                               false,
				"master.image":                              false,
				"master.compute.cpuCoreRequest":             true,
				"master.compute.extendedResources.quantity": false,
				"master.compute.extendedResources":          true,
				"master.storage.pvc":                        true,
				"master.storage.pvc.claimName":              false,
				"master.security.supplementalGroups":        true,
				"master.security.hostNetwork":               true,
			}

			testutils.Expect(actualRulesFromParent).ToEqual(expectedRulesFromParent)
		})
	})

	Context("imposed assets - enrich rules with locked items", func() {
		It("should enrich rules with imposed asset locked rules for regular workload", func() {
			// Test that enrichRulesWithImposed properly merges locked rules
			ht := HierarchyTree{
				treeId: TreeId{TenantId: 1, WorkloadType: WorkloadTypeWorkspace},
			}

			assetId := "asset-456"
			imposedRules := map[string]interface{}{
				"nfs": map[string]interface{}{
					"instances": map[string]interface{}{
						"locked": []interface{}{assetId},
					},
				},
			}

			result, err := ht.enrichRulesWithImposed(nil, imposedRules, nil)

			// Verify result structure
			expectedResult := map[string]interface{}{
				"storage": map[string]interface{}{
					"nfs": map[string]interface{}{
						"instances": map[string]interface{}{
							"locked": []interface{}{assetId},
						},
					},
				},
			}

			testutils.Expect(result).ToEqual(expectedResult)
			testutils.Expect(err).ToEqual((*api.Error)(nil))
		})

		It("should enrich rules for distributed workload - worker storage", func() {
			ht := HierarchyTree{
				treeId: TreeId{TenantId: 1, WorkloadType: WorkloadTypeDistributed},
			}

			assetId := "asset-789"
			imposedRules := map[string]interface{}{
				"nfs": map[string]interface{}{
					"instances": map[string]interface{}{
						"locked": []interface{}{assetId},
					},
				},
			}

			result, err := ht.enrichRulesWithImposed(nil, imposedRules, nil)

			// Verify structure: worker.storage.nfs.instances.locked
			expectedResult := map[string]interface{}{
				"worker": map[string]interface{}{
					"storage": map[string]interface{}{
						"nfs": map[string]interface{}{
							"instances": map[string]interface{}{
								"locked": []interface{}{assetId},
							},
						},
					},
				},
			}

			testutils.Expect(result).ToEqual(expectedResult)
			testutils.Expect(err).ToEqual((*api.Error)(nil))
		})

		It("should enrich rules for distributed inference - leader storage", func() {
			ht := HierarchyTree{
				treeId: TreeId{TenantId: 1, WorkloadType: WorkloadTypeDistributedInference},
			}

			assetId := "asset-999"
			imposedMasterRules := map[string]interface{}{
				"pvc": map[string]interface{}{
					"instances": map[string]interface{}{
						"locked": []interface{}{assetId},
					},
				},
			}

			result, err := ht.enrichRulesWithImposed(nil, nil, imposedMasterRules)

			// Verify structure: leader.storage.pvc.instances.locked
			expectedResult := map[string]interface{}{
				"leader": map[string]interface{}{
					"storage": map[string]interface{}{
						"pvc": map[string]interface{}{
							"instances": map[string]interface{}{
								"locked": []interface{}{assetId},
							},
						},
					},
				},
			}

			testutils.Expect(result).ToEqual(expectedResult)
			testutils.Expect(err).ToEqual((*api.Error)(nil))
		})

		It("should merge imposed rules with existing rules", func() {
			ht := HierarchyTree{
				treeId: TreeId{TenantId: 1, WorkloadType: WorkloadTypeWorkspace},
			}

			existingRules := map[string]interface{}{
				"storage": map[string]interface{}{
					"existingPvc": map[string]interface{}{
						"instances": map[string]interface{}{
							"canAdd": false,
						},
					},
				},
			}

			assetId := "new-asset-123"
			imposedRules := map[string]interface{}{
				"nfs": map[string]interface{}{
					"instances": map[string]interface{}{
						"locked": []interface{}{assetId},
					},
				},
			}

			result, err := ht.enrichRulesWithImposed(existingRules, imposedRules, nil)

			// Should have both existing PVC rules and new NFS locked rules
			testutils.Expect(err).ToEqual((*api.Error)(nil))

			// Verify the structure manually
			storageMap := convutils.MapMe(result["storage"])

			// Check existing PVC rules are preserved
			existingPvcMap := convutils.MapMe(storageMap["existingPvc"])
			existingPvcInstancesMap := convutils.MapMe(existingPvcMap["instances"])
			testutils.Expect(existingPvcInstancesMap["canAdd"]).ToEqual(false)

			// Check new NFS locked rules are added
			nfsMap := convutils.MapMe(storageMap["nfs"])
			nfsInstancesMap := convutils.MapMe(nfsMap["instances"])
			testutils.Expect(nfsInstancesMap["locked"]).ToEqual([]interface{}{assetId})
		})

		It("should return nil when no imposed rules provided", func() {
			ht := HierarchyTree{
				treeId: TreeId{TenantId: 1, WorkloadType: WorkloadTypeWorkspace},
			}

			result, err := ht.enrichRulesWithImposed(nil, nil, nil)

			testutils.Expect(result).ToEqual(nil)
			testutils.Expect(err).ToEqual((*api.Error)(nil))
		})
	})
})
