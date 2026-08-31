/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/utils"

	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

const versionOfClusterWithNoStatusInfo = "2.15.0"

var _ = Describe("PVC compliance tests", func() {
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

	Context("list assets to include compliance info - pvc", func() {
		It("one asset should be ok, one not compliant to policy, two existing - one ok, one missing in cluster", func() {
			workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
					Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
						Pvc: *policyService.NewNullablePvcsRules(&policyService.PvcsRules{
							Attributes: *policyService.NewNullablePvcRules(&policyService.PvcRules{
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
					EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
				},
			})
			clusterPvc := testing.BuildRandomClusterPVCAsset(randoms)
			departmentPvc := testing.BuildRandomDepartmentPVCAsset(randoms)
			projectPvc := testing.BuildRandomProjectPVCAsset(randoms)

			goodAsset := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "good",
					Kind: api.ASSETKIND_PVC,
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ReadOnly:  convutils.Ref(false),
					ClaimName: convutils.Ref("good-claim"),
				},
				Compliance: &api.ComplianceInfo{},
			}
			var goodAssetCreReq api.PvcCreationRequest
			copier.Copy(&goodAssetCreReq, &goodAsset)

			badAsset := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:   uuid.NewString(),
					Name: "bad",
					Kind: api.ASSETKIND_PVC,
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ClaimName: convutils.Ref("bad-claim"),
				},
				Compliance: &api.ComplianceInfo{},
			}

			var badAssetCreReq api.PvcCreationRequest
			copier.Copy(&badAssetCreReq, &badAsset)

			existInClusterCluster := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "exist-in-cluster-cluster",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_CLUSTER,
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: clusterPvc.Spec.ExistingPvc,
					ClaimName:   clusterPvc.Spec.ClaimName,
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}
			existInClusterDept := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:           uuid.NewString(),
					Name:         "exist-in-cluster-dept",
					Kind:         api.ASSETKIND_PVC,
					Scope:        api.SCOPE_DEPARTMENT,
					DepartmentId: convutils.Ref(randoms.DepartmentId),
					ClusterId:    convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: departmentPvc.Spec.ExistingPvc,
					ClaimName:   departmentPvc.Spec.ClaimName,
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}
			existInClusterProj := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "exist-in-cluster-proj",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_PROJECT,
					ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: projectPvc.Spec.ExistingPvc,
					ClaimName:   projectPvc.Spec.ClaimName,
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}
			missingInClusterCluster := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "missing-in-cluster-cluster",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_CLUSTER,
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: convutils.Ref(false),
					ClaimName:   convutils.Ref("pvc-cluster-missing"),
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}
			missingInClusterDept := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:           uuid.NewString(),
					Name:         "missing-in-cluster-dept",
					Kind:         api.ASSETKIND_PVC,
					Scope:        api.SCOPE_DEPARTMENT,
					DepartmentId: convutils.Ref(randoms.DepartmentId),
					ClusterId:    convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: convutils.Ref(false),
					ClaimName:   convutils.Ref("pvc-dept-missing"),
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}
			missingInClusterProj := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "missing-in-cluster-proj",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_PROJECT,
					ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					ExistingPvc: convutils.Ref(true),
					ClaimName:   convutils.Ref("pvc-proj-missing"),
					Path:        convutils.Ref("/data"),
					ReadOnly:    convutils.Ref(false),
				},
				Compliance: &api.ComplianceInfo{},
			}

			pvcWithStandardStorageClass := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "good",
					Kind:      api.ASSETKIND_PVC,
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ReadOnly:  convutils.Ref(false),
					ClaimName: convutils.Ref("good-claim"),
					ClaimInfo: &api.ClaimInfo{
						StorageClass: convutils.Ref("standard"),
					},
				},
				Compliance: &api.ComplianceInfo{},
			}

			var pvcWithStandardStorageClassCreReq api.PvcCreationRequest
			copier.Copy(&pvcWithStandardStorageClassCreReq, &pvcWithStandardStorageClass)

			pvcWithNonStandardStorageClass := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "good",
					Kind:      api.ASSETKIND_PVC,
					ClusterId: convutils.Ref(randoms.ClusterId),
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ReadOnly:  convutils.Ref(false),
					ClaimName: convutils.Ref("good-claim"),
					ClaimInfo: &api.ClaimInfo{
						StorageClass: convutils.Ref("non-standard"),
					},
				},
				Compliance: &api.ComplianceInfo{},
			}

			var pvcWithNonStandardStorageClassCreReq api.PvcCreationRequest
			copier.Copy(&pvcWithNonStandardStorageClassCreReq, &pvcWithNonStandardStorageClass)

			standardPermission := &k8sObjectsApi.StorageClassPermissions{}
			standardPermission.AllowedForWorkloads.Set(convutils.Ref(true))
			standardStorageClass := k8sObjectsApi.StorageClassConfiguration{
				Id:      uuid.NewString(),
				Name:    "standard",
				ScopeId: randoms.ClusterId,
			}
			standardStorageClass.Permissions.Set(standardPermission)

			nonStandardPermission := &k8sObjectsApi.StorageClassPermissions{}
			nonStandardPermission.AllowedForWorkloads.Set(convutils.Ref(false))
			nonStandardStorageClass := k8sObjectsApi.StorageClassConfiguration{
				Id:      uuid.NewString(),
				Name:    "non-standard",
				ScopeId: randoms.ClusterId,
			}
			nonStandardStorageClass.Permissions.Set(nonStandardPermission)

			trackerMock := k8sObjectsTracker.NewSimpleMockClient().
				WithPvc(randoms.ClusterId, &clusterPvc, nil).
				WithPvc(randoms.ClusterId, &departmentPvc, nil).
				WithPvc(randoms.ClusterId, &projectPvc, nil).
				WithPvc(randoms.ClusterId, &goodAssetCreReq, nil).
				WithPvc(randoms.ClusterId, &badAssetCreReq, nil).
				WithPvc(randoms.ClusterId, &pvcWithStandardStorageClassCreReq, nil).
				WithPvc(randoms.ClusterId, &pvcWithNonStandardStorageClassCreReq, nil)

			trackerMock.WithStorageClassConfig(standardStorageClass.Name, standardStorageClass)
			trackerMock.WithStorageClassConfig(nonStandardStorageClass.Name, nonStandardStorageClass)

			k8sObjectsTracker.SetMock(trackerMock)

			assetsSpec := make(map[string]api.PvcAssetSpec)
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)

			for _, asset := range []*api.PvcAsset{
				&goodAsset, &badAsset,
				&existInClusterCluster, &existInClusterDept, &existInClusterProj,
				&missingInClusterCluster, &missingInClusterDept, &missingInClusterProj,
				&pvcWithStandardStorageClass, &pvcWithNonStandardStorageClass,
			} {
				assetsToPatch[asset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &asset.Meta,
					Spec:           &asset.Spec,
					ComplianceInfo: &asset.Compliance,
				}
				assetsSpec[asset.Meta.Id] = asset.Spec
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
					Field:   convutils.Ref("storage.pvc.attributes.readOnly"),
					Details: "value must be provided for required fields",
					Rule:    convutils.Ref(api.POLICYRULEENUM_REQUIRED),
				},
			}))

			Expect((*assetsToPatch[pvcWithStandardStorageClass.Meta.Id].ComplianceInfo).Compliance).To(Equal(true))
			Expect((*assetsToPatch[pvcWithNonStandardStorageClass.Meta.Id].ComplianceInfo).Compliance).To(Equal(false))
			Expect((*assetsToPatch[pvcWithNonStandardStorageClass.Meta.Id].ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{
				{
					Details: "The underlying storage class of the PVC is forbidden for running workloads",
				},
			}))

			for _, pvcAsset := range assetsToPatch {
				if strings.Contains(pvcAsset.Meta.Name, "exist-") {
					Expect((*pvcAsset.ComplianceInfo).Compliance).To(Equal(true), pvcAsset.Meta.Name)
				}
				if strings.Contains(pvcAsset.Meta.Name, "missing-") {
					message := "This data source can't be selected because its status is \"Issues found\". Contact your administrator"
					Expect((*pvcAsset.ComplianceInfo).Compliance).To(Equal(false), pvcAsset.Meta.Name)
					Expect((*pvcAsset.ComplianceInfo).Reason).To(Equal([]api.ComplianceInfoReason{{Details: message}}), pvcAsset.Meta.Name)
				}
			}
		})
	})

	Context("pvc no matching attributes", func() {
		It("one asset should be ok, the other does not", func() {
			workspacePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:         policyMockUpdateTime,
					EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
				},
			})

			goodAsset := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "good",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_PROJECT,
					ClusterId: convutils.Ref(randoms.ClusterId),
					ProjectId: convutils.Ref(int32(randoms.ProjectId)),
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ReadOnly:  convutils.Ref(false),
					ClaimName: convutils.Ref("good-claim"),
					ClaimInfo: &api.ClaimInfo{
						StorageClass: convutils.Ref("standard"),
						AddedAttrValues: []api.PvcAddedAttrValue{
							{
								Key:   "retries",
								Value: "5",
							},
						},
					},
				},
				Compliance: &api.ComplianceInfo{},
			}
			var goodAssetCreReq api.PvcCreationRequest
			copier.Copy(&goodAssetCreReq, &goodAsset)

			badAsset := api.PvcAsset{
				Meta: api.AssetMeta{
					Id:        uuid.NewString(),
					Name:      "bad",
					Kind:      api.ASSETKIND_PVC,
					Scope:     api.SCOPE_PROJECT,
					ClusterId: convutils.Ref(randoms.ClusterId),
					ProjectId: convutils.Ref(int32(randoms.ProjectId)),
				},
				Spec: api.PvcAssetSpec{
					Path:      convutils.Ref("/data"),
					ClaimName: convutils.Ref("bad-claim"),
					ClaimInfo: &api.ClaimInfo{
						StorageClass: convutils.Ref("standard"),
						AddedAttrValues: []api.PvcAddedAttrValue{
							{
								Key:   "retries",
								Value: "55",
							},
						},
					},
				},
				Compliance: &api.ComplianceInfo{},
			}

			var badAssetCreReq api.PvcCreationRequest
			copier.Copy(&badAssetCreReq, &badAsset)

			standardStorageClass := k8sObjectsApi.StorageClassConfiguration{
				Id:        uuid.NewString(),
				Name:      "standard",
				ScopeType: string(api.SCOPETYPE_CLUSTER),
				ScopeId:   randoms.ClusterId,
			}
			standardStorageClass.Permissions.Set(&k8sObjectsApi.StorageClassPermissions{
				AllowedForWorkloads: *k8sObjectsApi.NewNullableBool(convutils.Ref(true)),
			})

			standardStorageClass.Customization.Set(&k8sObjectsApi.StorageClassCustomization{
				Attributes: []k8sObjectsApi.PvcAddedAttribute{
					{
						Key:      "retries",
						Format:   *k8sObjectsApi.NewNullablePvcAddedAttributeFormat(convutils.Ref(k8sObjectsApi.PVCADDEDATTRIBUTEFORMAT_NUMBER)),
						MaxValue: *k8sObjectsApi.NewNullableFloat32(convutils.Ref(float32(10.0))),
					},
				},
			})

			trackerMock := k8sObjectsTracker.NewSimpleMockClient().
				WithPvc(randoms.ClusterId, &goodAssetCreReq, nil).
				WithPvc(randoms.ClusterId, &badAssetCreReq, nil)

			trackerMock.WithStorageClassConfig(standardStorageClass.Name, standardStorageClass)

			k8sObjectsTracker.SetMock(trackerMock)

			assetsSpec := make(map[string]api.PvcAssetSpec)
			assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)

			for _, asset := range []*api.PvcAsset{
				&goodAsset, &badAsset,
			} {
				assetsToPatch[asset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &asset.Meta,
					Spec:           &asset.Spec,
					ComplianceInfo: &asset.Compliance,
				}
				assetsSpec[asset.Meta.Id] = asset.Spec
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
			Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason[0].Details).To(Equal("attribute 'retries': value must be no greater then 10"))
		})

		Context("list assets to include compliance info - pvc - old cluster", func() {
			It("only message for existing pvcs", func() {
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, versionOfClusterWithNoStatusInfo, false)))

				clusterPvc := testing.BuildRandomClusterPVCAsset(randoms)
				departmentPvc := testing.BuildRandomDepartmentPVCAsset(randoms)
				projectPvc := testing.BuildRandomProjectPVCAsset(randoms)

				existInClusterCluster := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						Name:      "exist-in-cluster-cluster",
						Kind:      api.ASSETKIND_PVC,
						Scope:     api.SCOPE_CLUSTER,
						ClusterId: convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: clusterPvc.Spec.ExistingPvc,
						ClaimName:   clusterPvc.Spec.ClaimName,
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
				}
				existInClusterDept := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:           uuid.NewString(),
						Name:         "exist-in-cluster-dept",
						Kind:         api.ASSETKIND_PVC,
						Scope:        api.SCOPE_DEPARTMENT,
						DepartmentId: convutils.Ref(randoms.DepartmentId),
						ClusterId:    convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: departmentPvc.Spec.ExistingPvc,
						ClaimName:   departmentPvc.Spec.ClaimName,
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}
				existInClusterProj := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						Name:      "exist-in-cluster-proj",
						Kind:      api.ASSETKIND_PVC,
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
						ClusterId: convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: projectPvc.Spec.ExistingPvc,
						ClaimName:   projectPvc.Spec.ClaimName,
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}
				missingInClusterCluster := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						Name:      "missing-in-cluster-cluster",
						Kind:      api.ASSETKIND_PVC,
						Scope:     api.SCOPE_CLUSTER,
						ClusterId: convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: convutils.Ref(false),
						ClaimName:   convutils.Ref("pvc-cluster-missing"),
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
				}
				missingInClusterDept := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:           uuid.NewString(),
						Name:         "missing-in-cluster-dept",
						Kind:         api.ASSETKIND_PVC,
						Scope:        api.SCOPE_DEPARTMENT,
						DepartmentId: convutils.Ref(randoms.DepartmentId),
						ClusterId:    convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: convutils.Ref(false),
						ClaimName:   convutils.Ref("pvc-dept-missing"),
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}
				missingInClusterProj := api.PvcAsset{
					Meta: api.AssetMeta{
						Id:        uuid.NewString(),
						Name:      "missing-in-cluster-proj",
						Kind:      api.ASSETKIND_PVC,
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
						ClusterId: convutils.Ref(randoms.ClusterId),
					},
					Spec: api.PvcAssetSpec{
						ExistingPvc: convutils.Ref(true),
						ClaimName:   convutils.Ref("pvc-proj-missing"),
						Path:        convutils.Ref("/data"),
						ReadOnly:    convutils.Ref(false),
					},
					Compliance: &api.ComplianceInfo{},
				}

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithPvc(randoms.ClusterId, &clusterPvc, nil).
					WithPvc(randoms.ClusterId, &departmentPvc, nil).
					WithPvc(randoms.ClusterId, &projectPvc, nil))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.9.0", false)))

				assetsSpec := make(map[string]api.PvcAssetSpec)
				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)

				for _, asset := range []*api.PvcAsset{
					&existInClusterCluster, &existInClusterDept, &existInClusterProj,
					&missingInClusterCluster, &missingInClusterDept, &missingInClusterProj,
				} {
					assetsToPatch[asset.Meta.Id] = response_patching.ResponsePatchingPtrs{
						Meta:           &asset.Meta,
						Spec:           &asset.Spec,
						ComplianceInfo: &asset.Compliance,
					}
					assetsSpec[asset.Meta.Id] = asset.Spec
				}

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

				permittedScopes := auth.PermittedScopes{}

				assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(assetsToPatch, &permittedScopes)

				patcher, err := NewComplianceInfoPatcher(context.Background(),
					int32(randoms.ProjectId), auth.WorkloadTypeWorkspace, "", assetsToPatch, &permittedScopes, assetClusterStatus)
				Expect(err).To(BeNil())

				err = patcher.PatchComplianceInfo(context.Background())
				Expect(err).To(BeNil())

				for _, pvcAsset := range assetsToPatch {
					// In old clusters we don't check the existence of the PVCs in the cluster, so we expect them all to be complying
					Expect((*pvcAsset.ComplianceInfo).Compliance).To(Equal(true), pvcAsset.Meta.Name)
				}
			})
		})

		Context("list assets to include compliance info - data volume", func() {
			It("should be complied when the status is one of 'no issues found' and 'no status'", func() {
				goodAssetId := uuid.NewString()
				badAssetId := uuid.NewString()

				goodAsset := api.DatasourceListResponseEntry{
					Meta: api.AssetMeta{
						Id:   goodAssetId,
						Kind: api.ASSETKIND_DATA_VOLUME,
						Name: "good",
					},
					Spec: api.DatasourceListResponseAssetSpec{
						DataVolume: &api.DataVolumeAssetSpec{
							Id: utils.Ref(uuid.NewString()),
						},
					},
					Status: &api.AssetClusterStatusInfo{
						Status: api.ASSETSYNCSTATUSENUM_NO_STATUS,
					},
				}

				badAsset := api.DatasourceListResponseEntry{
					Meta: api.AssetMeta{
						Id:   badAssetId,
						Kind: api.ASSETKIND_DATA_VOLUME,
						Name: "bad",
					},
					Spec: api.DatasourceListResponseAssetSpec{
						DataVolume: &api.DataVolumeAssetSpec{
							Id: utils.Ref(uuid.NewString()),
						},
					},
					Status: &api.AssetClusterStatusInfo{
						Status: api.ASSETSYNCSTATUSENUM_FAILED,
					},
				}

				assetsToPatch := make(map[string]response_patching.ResponsePatchingPtrs)
				assetsToPatch[goodAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &goodAsset.Meta,
					Spec:           &goodAsset.Spec,
					ComplianceInfo: &goodAsset.Compliance,
					Status:         &goodAsset.Status,
				}
				assetsToPatch[badAsset.Meta.Id] = response_patching.ResponsePatchingPtrs{
					Meta:           &badAsset.Meta,
					Spec:           &badAsset.Spec,
					ComplianceInfo: &badAsset.Compliance,
					Status:         &badAsset.Status,
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
				Expect((*assetsToPatch[badAsset.Meta.Id].ComplianceInfo).Reason).To(Equal(
					[]api.ComplianceInfoReason{
						{
							Details: fmt.Sprintf("The data volume status is %s", badAsset.Status.Status),
						},
					}))
			})
		})
	})
})
