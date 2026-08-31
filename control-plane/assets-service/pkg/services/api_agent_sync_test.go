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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/copier"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"k8s.io/utils/pointer"

	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

const (
	urlGetCredentialsUpdates = "/api/v1/asset/credentials/updates"
	urlGetDatasourceUpdates  = "/api/v1/asset/datasource/updates"

	salesforceUrl          = "https://www.nvidia.com/en-eu/support/enterprise/#contact-us"
	issuesFoundCredentials = "Issues were found while propagating the credential to the credentials' scope"
	issuesFoundDatasources = "Issues were found while propagating the data source to the data source's scope"
	noIssuesCredMessage    = "No issues were found while propagating the credential to the credentials' scope"

	notReadyMessage        = "This data source can't be selected yet, because it is being synchronized to the cluster."
	syncFailedMessage      = "This data source can't be selected because it could not synchronize to the cluster. Contact your administrator"
	issuesFoundMessage     = "This data source can't be selected because its status is \"Issues found\". Contact your administrator"
	beingDeletedMessage    = "This data source can't be selected, because it is being deleted."
	noClusterUpdateMessage = "Status of the asset not yet updated from the cluster"

	clusterNotSupportingAgentSync = "2.20.0"
)

var _ = Describe("AgentSync", func() {
	var randoms, anotherNewCluster, anotherOldCluster *testutils.RandomValues

	router := CreateRouter()

	Describe("Test agent sync", func() {
		BeforeEach(func() {
			dbTesting.CleanTables()

			dbTesting.RestartUpdateCounter()

			randoms = testutils.NewRandomValues()
			anotherNewCluster = randoms.CloneWithDifferentCluster()
			anotherOldCluster = anotherNewCluster.CloneWithDifferentCluster()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherNewCluster, anotherOldCluster)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherNewCluster, anotherOldCluster))
			authorization.SetMock(authMock)

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
				},
			})

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
				WithCluster(anotherNewCluster.ClusterId, testing.BuildDisplayedCluster(anotherNewCluster, testing.UnitTestDefaultClusterVersion, false)).
				WithCluster(anotherOldCluster.ClusterId, testing.BuildDisplayedCluster(anotherOldCluster, clusterNotSupportingAgentSync, false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		Context("sync handles all kind of credential assets", func() {
			It("handles all kinds of synced credentials", func() {
				// to be synced
				passwordReq := buildRandomProjectPassword(randoms)
				password, _, _ := createPassword(&randoms.ClusterId, passwordReq, router, headerForAdmin(randoms))

				accessKeyDeptReq := buildRandomDepartmentAccessKey(randoms)
				accessKeyDept, _, _ := createAccessKey(&randoms.ClusterId, accessKeyDeptReq, router, headerForAdmin(randoms))

				accessKeyProjReq := buildRandomProjectAccessKey(randoms)
				accessKeyProj, _, _ := createAccessKey(&randoms.ClusterId, accessKeyProjReq, router, headerForAdmin(randoms))

				dockerRegistryReq := buildRandomProjectDockerRegistry(randoms)
				dockerRegistryReq.Spec = api.DockerRegistryCreationSpec{
					User:     convutils.Ref("uuuu"),
					Password: convutils.Ref("pppp"),
					Url:      convutils.Ref("http://uuu.com"),
				}
				dockerRegistry, _, _ := createDockerRegistry(&randoms.ClusterId, dockerRegistryReq, router, headerForAdmin(randoms))

				dockerRegistryPrivateReq := buildRandomTenantDockerRegistry()
				dockerRegistryPrivateReq.Meta.Private = convutils.Ref(true)
				dockerRegistryPrivateReq.Spec = api.DockerRegistryCreationSpec{
					User:     convutils.Ref("uuuu"),
					Password: convutils.Ref("pppp"),
					Url:      convutils.Ref("http://uuu.com"),
				}
				privateDocker, _, _ := createDockerRegistry(&randoms.ClusterId, dockerRegistryPrivateReq, router, headerForAdmin(randoms))

				genericSecretReq := buildRandomClusterGenericSecret(randoms)
				genericSecretReq.Spec = api.GenericSecretCreationSpec{
					KeyValuePairs: &[]api.KeyValuePair{
						{
							Key:   convutils.Ref("aaaa"),
							Value: convutils.Ref("bbbb"),
						},
					},
				}
				genericSecret, _, _ := createGenericSecret(&randoms.ClusterId, genericSecretReq, router, headerForAdmin(randoms))

				privateGenericSecretReq := buildRandomTenantGenericSecret()
				privateGenericSecretReq.Meta.Private = convutils.Ref(true)
				privateGenericSecretReq.Spec = api.GenericSecretCreationSpec{
					KeyValuePairs: &[]api.KeyValuePair{
						{
							Key:   convutils.Ref("private"),
							Value: convutils.Ref("cred"),
						},
					},
				}
				privateGenericSecret, _, _ := createGenericSecret(&randoms.ClusterId, privateGenericSecretReq, router, headerForAdmin(randoms))

				// not to be synced
				createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

				ngcApiKeyReq := buildRandomClusterNgcApiKey(randoms)
				ngcApiKey, _, _ := createNgcApiKey(ngcApiKeyReq, router, headerForAdmin(randoms))

				//
				//   clean all the resources from the tracker mock
				//
				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

				//
				//  right after creation, everything should be in creating state
				//
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta):       {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(genericSecret.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(genericSecret.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}
				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//  check that correct updates are reported to the agent
				//
				after := int64(0)

				expectedUpdates := api.CredentialsUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  1,
						HighestUpdate: 9,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 9, // we also have the extra pvc
					},
					Assets: []api.CredentialsUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](password.Meta),
							Spec: api.CredentialsUpdateSpec{
								Password: &passwordReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 1,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](accessKeyDept.Meta),
							Spec: api.CredentialsUpdateSpec{
								AccessKey: &accessKeyDeptReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 2,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](accessKeyProj.Meta),
							Spec: api.CredentialsUpdateSpec{
								AccessKey: &accessKeyProjReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 3,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](dockerRegistry.Meta),
							Spec: api.CredentialsUpdateSpec{
								DockerRegistry: &dockerRegistryReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 4,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](privateDocker.Meta),
							Spec: api.CredentialsUpdateSpec{
								DockerRegistry: &dockerRegistryPrivateReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 5,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](genericSecret.Meta),
							Spec: api.CredentialsUpdateSpec{
								GenericSecret: &genericSecretReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 6,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](privateGenericSecret.Meta),
							Spec: api.CredentialsUpdateSpec{
								GenericSecret: &privateGenericSecretReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 7,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](ngcApiKey.Meta),
							Spec: api.CredentialsUpdateSpec{
								GenericSecret: &api.GenericSecretCreationSpec{
									KeyValuePairs: &[]api.KeyValuePair{
										{Key: convutils.Ref(dbmodel.NGC_API_KEY), Value: &ngcApiKeyReq.Spec.NgcApiKey},
									},
								},
								DockerRegistry: &api.DockerRegistryCreationSpec{
									User:     convutils.Ref(dbmodel.NGC_REGISTRY_USER),
									Url:      convutils.Ref(dbmodel.NGC_REGISTRY_URL),
									Password: &ngcApiKeyReq.Spec.NgcApiKey,
								},
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 9,
						},
					},
				}

				updates, listErr := listCredentialsUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdates)

				after = max(after, updates.SystemInfo.HighestUpdate)

				//
				//  status is still creating
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta):       {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(genericSecret.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//   agent report about success/failure of the creation in the cluster
				//
				report := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[0].Meta.Id,
							UpdateCount: expectedUpdates.Assets[0].UpdateCount,
						},
						{
							Id:          expectedUpdates.Assets[1].Meta.Id,
							UpdateCount: expectedUpdates.Assets[1].UpdateCount,
							Message:     convutils.Ref("The credential couldn't be synchronized in the cluster"),
						},
						{
							Id:          expectedUpdates.Assets[2].Meta.Id,
							UpdateCount: expectedUpdates.Assets[2].UpdateCount,
						},
						{
							Id:          expectedUpdates.Assets[3].Meta.Id,
							UpdateCount: expectedUpdates.Assets[3].UpdateCount,
							Message:     convutils.Ref("The credential couldn't be synchronized in the cluster"),
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				after = max(after, updates.SystemInfo.HighestUpdate)

				//
				//  status needs to be updated
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta):       {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(genericSecret.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//  agent report that the generic secret is ok now
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[5].Meta.Id,
							UpdateCount: expectedUpdates.Assets[5].UpdateCount,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  k8s objects tracker received information about the docker registry and the password, that they have an issue
				//
				genericSecretIssues := []k8sObjectsApi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_CLUSTER),
						ScopeId:   randoms.ClusterId,
						Issue:     k8sObjectsApi.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					},
				}

				passwordIssues := []k8sObjectsApi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_PROJECT),
						ScopeId:   randoms.ProjectIdStr,
						Issue:     k8sObjectsApi.CLUSTEROBJECTISSUETYPE_NOT_CREATED,
					},
				}

				k8sObjectsMock := k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(randoms.ClusterId, &passwordReq, &passwordIssues).
					WithGenericSecret(randoms.ClusterId, &genericSecretReq, &genericSecretIssues)

				k8sObjectsTracker.SetMock(k8sObjectsMock)

				//
				//  status needs to be updated with the k8s objects tracker findings
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundCredentials,
						Issues:  toAssetIssues(passwordIssues),
						Url:     salesforceUrl,
					},
					metaToString(genericSecret.Meta): {
						Message: issuesFoundCredentials,
						Status:  api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Issues:  toAssetIssues(genericSecretIssues),
						Url:     salesforceUrl,
					},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//   delete the problematic generic secret
				//
				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGenericSecret, genericSecret.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				genericSecret.Meta.UpdateCount = 10 // following the deletion

				//
				//  status needs to be updated to deleting
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundCredentials,
						Issues:  toAssetIssues(passwordIssues),
						Url:     salesforceUrl,
					},
					metaToString(genericSecret.Meta): {
						Status: api.ASSETSYNCSTATUSENUM_DELETING,
					},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//   the agent pulled the deletion - no change in the statuses
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{},
				}

				expectedUpdates = api.CredentialsUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  11,
						HighestUpdate: 11,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 11,
					},
					Assets: []api.CredentialsUpdate{
						{
							Meta:        srcToTarget[api.AssetMeta, api.AssetMetaForSync](genericSecret.Meta),
							IsDeleted:   pointer.Bool(true),
							UpdateCount: 11,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(1))

				after = max(after, updates.ResponseInfo.HighestUpdate)

				testutils.Expect(updates).ToEqual(expectedUpdates)

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//  the agent reported success in deleting the generic secret
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          genericSecret.Meta.Id,
							UpdateCount: 11,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  the generic secret is gone
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(password.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundCredentials,
						Issues:  toAssetIssues(passwordIssues),
						Url:     salesforceUrl,
					},
					metaToString(accessKeyDept.Meta):  {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(accessKeyProj.Meta):  {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(dockerRegistry.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The credential couldn't be synchronized to the cluster"},
					metaToString(ngcApiKey.Meta):      {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)
			})
		})

		Context("sync handles sync of pvc assets", func() {
			It("handles all pvcs properly", func() {
				// to be synced
				pvcReq1 := testing.BuildRandomProjectPVCAsset(randoms)
				pvcReq1.Meta.Name = "pvc1"
				pvc1, _, _ := createPVCAsset(pvcReq1, router, headerForAdmin(randoms))

				pvcReq2 := testing.BuildRandomDepartmentPVCAsset(randoms)
				pvcReq2.Meta.Name = "pvc2"
				pvc2, _, _ := createPVCAsset(pvcReq2, router, headerForAdmin(randoms))

				pvcReq3 := testing.BuildRandomDepartmentPVCAsset(randoms)
				pvcReq3.Meta.Name = "pvc3"
				pvc3, _, _ := createPVCAsset(pvcReq3, router, headerForAdmin(randoms))

				// not to be synced
				createAccessKey(&randoms.ClusterId, buildRandomDepartmentAccessKey(randoms), router, headerForAdmin(randoms))

				//
				//  right after creation, everything should be in creating state
				//
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(pvc1.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				expectedCompliance := map[string]api.ComplianceInfo{
					metaToString(pvc1.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc2.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc3.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//  check that correct updates are reported to the agent
				//
				after := int64(0)

				expectedUpdates := api.DatasourceUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  1,
						HighestUpdate: 3,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 4, // we also have the extra credentials
					},
					Assets: []api.DatasourceUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](pvc1.Meta),
							Spec: api.DatasourceUpdateSpec{
								Pvc: &pvc1.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 1,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](pvc2.Meta),
							Spec: api.DatasourceUpdateSpec{
								Pvc: &pvc2.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 2,
						},
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](pvc3.Meta),
							Spec: api.DatasourceUpdateSpec{
								Pvc: &pvc3.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 3,
						},
					},
				}

				updates, listErr := listDatasourcesUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdates)

				after = max(after, updates.ResponseInfo.HighestUpdate)

				//
				//  status is still creating
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvc1.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				// agent report about success/failure of the creation in the cluster of two pvcs but there
				// is async issue with the 3rd pvc
				//
				report := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[0].Meta.Id,
							UpdateCount: expectedUpdates.Assets[0].UpdateCount,
						},
						{
							Id:          expectedUpdates.Assets[1].Meta.Id,
							UpdateCount: expectedUpdates.Assets[1].UpdateCount,
						},
						{
							Id:          expectedUpdates.Assets[2].Meta.Id,
							UpdateCount: expectedUpdates.Assets[2].UpdateCount,
							Message:     convutils.Ref("The data source couldn't be synchronized to the cluster"),
						},
					},
				}

				updates, listErr = listDatasourcesUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				after = max(after, updates.ResponseInfo.HighestUpdate)

				//
				//  status needs to be updated
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvc1.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The data source couldn't be synchronized to the cluster"},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvc1.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc2.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc3.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: syncFailedMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//  agent report that the first pvc is ok
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[0].Meta.Id,
							UpdateCount: expectedUpdates.Assets[0].UpdateCount,
						},
					},
				}

				updates, listErr = listDatasourcesUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  k8s objects tracker received information about the first pvc is problematic
				//
				pvc1Issues := []k8sObjectsApi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_CLUSTER),
						ScopeId:   randoms.ClusterId,
						Issue:     k8sObjectsApi.CLUSTEROBJECTISSUETYPE_NOT_CREATED,
					},
				}

				k8sObjectsMock := k8sObjectsTracker.NewSimpleMockClient().
					WithPvc(randoms.ClusterId, &pvcReq1, &pvc1Issues)

				k8sObjectsTracker.SetMock(k8sObjectsMock)

				//
				//  status needs to be updated with the k8s objects tracker findings
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvc1.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundDatasources,
						Issues:  toAssetIssues(pvc1Issues),
						Url:     salesforceUrl,
					},
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The data source couldn't be synchronized to the cluster"},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvc1.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: issuesFoundMessage}}},
					metaToString(pvc2.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc3.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: syncFailedMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//   delete the problematic pvc
				//
				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvc1.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				//
				//  status needs to be updated to deletion
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvc1.Meta): {
						Status: api.ASSETSYNCSTATUSENUM_DELETING,
					},
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The data source couldn't be synchronized to the cluster"},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvc1.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: beingDeletedMessage}}},
					metaToString(pvc2.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc3.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: syncFailedMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//   the agent pulled the deletion - no change in the statuses
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{},
				}

				expectedUpdates = api.DatasourceUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  6,
						HighestUpdate: 6,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 6, // we also have the extra pvc
					},
					Assets: []api.DatasourceUpdate{
						{
							Meta:        srcToTarget[api.AssetMeta, api.AssetMetaForSync](pvc1.Meta),
							IsDeleted:   pointer.Bool(true),
							UpdateCount: 6,
							Spec: api.DatasourceUpdateSpec{
								Pvc: &api.PvcAssetSpec{
									ClaimName: pvc1.Spec.ClaimName,
								},
							},
						},
					},
				}

				updates, listErr = listDatasourcesUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(1))

				after = max(after, updates.ResponseInfo.HighestUpdate)

				testutils.Expect(updates).ToEqual(expectedUpdates)

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//  the agent reported success in deleting the generic secret
				//
				report = api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          pvc1.Meta.Id,
							UpdateCount: 6,
						},
					},
				}

				updates, listErr = listDatasourcesUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  the pvc is gone
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvc2.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvc3.Meta): {Status: api.ASSETSYNCSTATUSENUM_FAILED, Message: "The data source couldn't be synchronized to the cluster"},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvc2.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvc3.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: syncFailedMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)
			})
		})

		Context("sync skips autoDelete workload-volume pvc assets", func() {
			It("does not sync pvc assets marked autoDelete=true", func() {
				// persistent pvc (autoDelete=false) - owned by the asset sync agent, must be synced
				persistentReq := testing.BuildRandomProjectPVCAsset(randoms)
				persistentReq.Meta.Name = "persistent-pvc"
				persistent, _, _ := createPVCAsset(persistentReq, router, headerForAdmin(randoms))

				// workload-volume pvc (autoDelete=true) - created and owned by the workload runner,
				// must NOT be synced by the agent
				workloadVolumeReq := testing.BuildRandomProjectPVCAsset(randoms)
				workloadVolumeReq.Meta.Name = "workload-volume-pvc"
				workloadVolumeReq.Meta.AutoDelete = convutils.Ref(true)
				createPVCAsset(workloadVolumeReq, router, headerForAdmin(randoms))

				//
				//  only the persistent pvc should be reported to the agent
				//
				updates, listErr := listDatasourcesUpdatesForTest(randoms, router, int64(0), nil)
				Expect(listErr).To(BeNil())

				Expect(updates.Assets).To(HaveLen(1))
				Expect(updates.Assets[0].Meta.Id).To(Equal(persistent.Meta.Id))
			})
		})

		Context("old clusters do not support agent sync", func() {
			It("assets will be hard deleted", func() {
				// to be synced
				passwordNewReq := buildRandomProjectPassword(randoms)
				passwordNew, _, _ := createPassword(&randoms.ClusterId, passwordNewReq, router, headerForAdmin(randoms))

				passwordOldReq := buildRandomProjectPassword(anotherOldCluster)
				passwordOld, _, _ := createPassword(&randoms.ClusterId, passwordOldReq, router, headerForAdmin(anotherOldCluster))

				//
				//  right after creation, no status yet, but there is a variation b/w agent and non-agent
				//
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(passwordNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(passwordOld.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING, Message: noClusterUpdateMessage},
				}
				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//  check that correct updates are reported to the agent
				//
				after := int64(0)

				expectedUpdates := api.CredentialsUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  1,
						HighestUpdate: 1,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 2, // we also have the extra pvc
					},
					Assets: []api.CredentialsUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](passwordNew.Meta),
							Spec: api.CredentialsUpdateSpec{
								Password: &passwordNewReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 1,
						},
					},
				}

				updates, listErr := listCredentialsUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdates)

				after = expectedUpdates.SystemInfo.HighestUpdate

				//
				//  agent created the resource
				//
				report := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[0].Meta.Id,
							UpdateCount: expectedUpdates.Assets[0].UpdateCount,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  k8s objects tracker received information about the two assets
				//
				passwordIssues := []k8sObjectsApi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_PROJECT),
						ScopeId:   randoms.ProjectIdStr,
						Issue:     k8sObjectsApi.CLUSTEROBJECTISSUETYPE_NOT_CREATED,
					},
				}

				k8sObjectsMock := k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(anotherOldCluster.ClusterId, &passwordOldReq, &passwordIssues).
					WithPassword(randoms.ClusterId, &passwordNewReq, &passwordIssues)

				k8sObjectsTracker.SetMock(k8sObjectsMock)

				//
				//  status needs to be updated with the k8s objects tracker findings
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(passwordNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundCredentials,
						Issues:  toAssetIssues(passwordIssues),
						Url:     salesforceUrl,
					},
					metaToString(passwordOld.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundCredentials,
						Issues:  toAssetIssues(passwordIssues),
						Url:     salesforceUrl,
					},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//   delete both assets
				//
				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPassword, passwordOld.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(anotherOldCluster)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				deleteHttpReq, err = http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPassword, passwordNew.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp = httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				//
				//  status needs to be updated to deleting on the agent, but on old clusters it is hard delete
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(passwordNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_DELETING},
				}
				validateCredentialsStatusForTest(randoms, router, expected)
			})
		})

		Context("old clusters do not support agent sync", func() {
			It("assets will be hard deleted", func() {
				// to be synced
				pvcNewReq := testing.BuildRandomProjectPVCAsset(randoms)
				pvcNewReq.Meta.Name = "pvc-new"
				pvcNew, _, _ := createPVCAsset(pvcNewReq, router, headerForAdmin(randoms))

				pvcOldReq := testing.BuildRandomProjectPVCAsset(anotherOldCluster)
				pvcOldReq.Meta.Name = "pvc-old"
				pvcOld, _, _ := createPVCAsset(pvcOldReq, router, headerForAdmin(anotherOldCluster))

				//
				//  right after creation, no status yet, but there is a variation b/w agent and non-agent
				//
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(pvcNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(pvcOld.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING, Message: noClusterUpdateMessage},
				}

				expectedCompliance := map[string]api.ComplianceInfo{
					metaToString(pvcOld.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
					metaToString(pvcNew.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: notReadyMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//  check that correct updates are reported to the agent
				//
				after := int64(0)

				expectedUpdates := api.DatasourceUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  1,
						HighestUpdate: 1,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 2, // we also have the extra pvc
					},
					Assets: []api.DatasourceUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](pvcNew.Meta),
							Spec: api.DatasourceUpdateSpec{
								Pvc: &pvcNewReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 1,
						},
					},
				}

				updates, listErr := listDatasourcesUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdates)

				after = expectedUpdates.SystemInfo.HighestUpdate

				//
				//  agent created the resource
				//
				report := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          expectedUpdates.Assets[0].Meta.Id,
							UpdateCount: expectedUpdates.Assets[0].UpdateCount,
						},
					},
				}

				updates, listErr = listDatasourcesUpdatesForTest(randoms, router, after, &report)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				//
				//  k8s objects tracker received information about the two assets
				//
				pvcIssues := []k8sObjectsApi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_PROJECT),
						ScopeId:   randoms.ProjectIdStr,
						Issue:     k8sObjectsApi.CLUSTEROBJECTISSUETYPE_NOT_CREATED,
					},
				}

				k8sObjectsMock := k8sObjectsTracker.NewSimpleMockClient().
					WithPvc(anotherOldCluster.ClusterId, &pvcOldReq, &pvcIssues).
					WithPvc(randoms.ClusterId, &pvcNewReq, &pvcIssues)

				k8sObjectsTracker.SetMock(k8sObjectsMock)

				//
				//  status needs to be updated with the k8s objects tracker findings
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvcNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundDatasources,
						Issues:  toAssetIssues(pvcIssues),
						Url:     salesforceUrl,
					},
					metaToString(pvcOld.Meta): {Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
						Message: issuesFoundDatasources,
						Issues:  toAssetIssues(pvcIssues),
						Url:     salesforceUrl,
					},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvcOld.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: issuesFoundMessage}}},
					metaToString(pvcNew.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: issuesFoundMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)

				//
				//   delete both assets
				//
				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvcOld.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(anotherOldCluster)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				deleteHttpReq, err = http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvcNew.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp = httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				//
				//  status needs to be updated to deleting on the agent, but on old clusters it is hard delete
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(pvcNew.Meta): {Status: api.ASSETSYNCSTATUSENUM_DELETING},
				}

				expectedCompliance = map[string]api.ComplianceInfo{
					metaToString(pvcNew.Meta): {Compliance: false, Reason: []api.ComplianceInfoReason{{Details: beingDeletedMessage}}},
				}

				validateDatasourcesStatusForTest(randoms, router, expected, expectedCompliance)
			})
		})

		Context("two clusters in parallel", func() {
			It("should work properly", func() {
				// to be synced
				drClusterAReq := buildRandomProjectDockerRegistry(randoms)
				drClusterAReq.Meta.Name = "dr-cluster-a"
				drClusterA, _, _ := createDockerRegistry(&randoms.ClusterId, drClusterAReq, router, headerForAdmin(randoms))

				drClusterBReq := buildRandomProjectDockerRegistry(anotherNewCluster)
				drClusterBReq.Meta.Name = "dr-cluster-b"
				drClusterB, _, _ := createDockerRegistry(&randoms.ClusterId, drClusterBReq, router, headerForAdmin(anotherNewCluster))

				//
				//  right after creation, no status yet, but there is a variation b/w agent and non-agent
				//
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				//
				//  check that correct update is reported to each agent
				//
				after := int64(0)

				expectedUpdatesA := api.CredentialsUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  1,
						HighestUpdate: 1,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 2, // we also have the extra pvc
					},
					Assets: []api.CredentialsUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](drClusterA.Meta),
							Spec: api.CredentialsUpdateSpec{
								DockerRegistry: &drClusterAReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 1,
						},
					},
				}

				updates, listErr := listCredentialsUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdatesA)

				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				expectedUpdatesB := api.CredentialsUpdatesResp{
					RequestInfo: api.RequestInfo{
						AfterUpdate: after,
					},
					ResponseInfo: api.ResponseInfo{
						LowestUpdate:  2,
						HighestUpdate: 2,
					},
					SystemInfo: api.SystemInfo{
						HighestUpdate: 2, // we also have the extra pvc
					},
					Assets: []api.CredentialsUpdate{
						{
							Meta: srcToTarget[api.AssetMeta, api.AssetMetaForSync](drClusterB.Meta),
							Spec: api.CredentialsUpdateSpec{
								DockerRegistry: &drClusterBReq.Spec,
							},
							IsDeleted:   pointer.Bool(false),
							UpdateCount: 2,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(anotherNewCluster, router, after, nil)
				Expect(listErr).To(BeNil())

				testutils.Expect(updates).ToEqual(expectedUpdatesB)

				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				after = expectedUpdatesA.SystemInfo.HighestUpdate

				//
				// k8s objects tracker update of the assets resources created in the cluster
				//
				k8sObjectsMock := (k8sObjectsTracker.NewSimpleMockClient().
					WithDockerRegistry(randoms.ClusterId, &drClusterAReq, nil).
					WithDockerRegistry(randoms.ClusterId, &drClusterBReq, nil))
				k8sObjectsTracker.SetMock(k8sObjectsMock)

				//
				// each cluster sends its report
				//
				reportA := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          drClusterA.Meta.Id,
							UpdateCount: 1,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(randoms, router, after, &reportA)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				after = max(after, updates.ResponseInfo.HighestUpdate)

				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, Message: noIssuesCredMessage},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				reportB := api.ClusterSyncReport{
					Updates: []api.ClusterSyncReportUpdate{
						{
							Id:          drClusterB.Meta.Id,
							UpdateCount: 2,
						},
					},
				}

				updates, listErr = listCredentialsUpdatesForTest(anotherNewCluster, router, after, &reportB)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(0))

				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, Message: noIssuesCredMessage},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, Message: noIssuesCredMessage},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				after = max(after, updates.ResponseInfo.HighestUpdate)

				//
				//   delete both assets
				//
				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlDockerRegistry, drClusterA.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				//
				//  status needs to be updated to deleting on the agent, but on old clusters it is hard delete
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_DELETING},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, Message: noIssuesCredMessage},
				}

				validateCredentialsStatusForTest(randoms, router, expected)

				deleteHttpReq, err = http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlDockerRegistry, drClusterB.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(anotherNewCluster)

				deleteHttpResp = httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

				//
				//  status needs to be updated to deleting on the agent, but on old clusters it is hard delete
				//
				expected = map[string]api.AssetClusterStatusInfo{
					metaToString(drClusterA.Meta): {Status: api.ASSETSYNCSTATUSENUM_DELETING},
					metaToString(drClusterB.Meta): {Status: api.ASSETSYNCSTATUSENUM_DELETING},
				}

				validateCredentialsStatusForTest(randoms, router, expected)
			})
		})

		XContext("private credentials sync", func() {
			It("only syncs private credentials to clusters that support them", func() {
				// Create a private credential
				privateGenericSecretReq := buildRandomTenantGenericSecret()
				privateGenericSecretReq.Meta.Private = convutils.Ref(true)
				privateGenericSecretReq.Spec = api.GenericSecretCreationSpec{
					KeyValuePairs: &[]api.KeyValuePair{
						{
							Key:   convutils.Ref("private"),
							Value: convutils.Ref("cred"),
						},
					},
				}
				createGenericSecret(&randoms.ClusterId, privateGenericSecretReq, router, headerForAdmin(randoms))

				// Create a non-private credential
				genericSecretReq := buildRandomClusterGenericSecret(randoms)
				genericSecretReq.Spec = api.GenericSecretCreationSpec{
					KeyValuePairs: &[]api.KeyValuePair{
						{
							Key:   convutils.Ref("aaaa"),
							Value: convutils.Ref("bbbb"),
						},
					},
				}
				genericSecret, _, _ := createGenericSecret(&randoms.ClusterId, genericSecretReq, router, headerForAdmin(randoms))

				// Check initial status
				expected := map[string]api.AssetClusterStatusInfo{
					metaToString(genericSecret.Meta): {Status: api.ASSETSYNCSTATUSENUM_CREATING},
				}
				validateCredentialsStatusForTest(randoms, router, expected)

				// Check updates for new cluster (should include both private and non-private)
				after := int64(0)
				updates, listErr := listCredentialsUpdatesForTest(randoms, router, after, nil)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(2)) // Both credentials should be synced

				// Check updates for old cluster (should only include non-private)
				updates, listErr = listCredentialsUpdatesForTest(anotherOldCluster, router, after, nil)
				Expect(listErr).To(BeNil())
				Expect(len(updates.Assets)).To(Equal(1)) // Only non-private credential should be synced
				Expect(updates.Assets[0].Meta.Id).To(Equal(genericSecret.Meta.Id))
			})
		})

	})
})

func listCredentialsUpdatesForTest(randoms *testutils.RandomValues, router chi.Router, after int64, report *api.ClusterSyncReport) (*api.CredentialsUpdatesResp, *api.Error) {
	url := fmt.Sprintf("%s?clusterId=%s&after=%d", urlGetCredentialsUpdates, randoms.ClusterId, after)

	if report == nil {
		report = &api.ClusterSyncReport{}
	}

	syncHttpReq, err := http.NewRequest("POST", url, makePayload(report))
	Expect(err).To(BeNil())
	syncHttpReq.Header = *headerForAdmin(randoms)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, syncHttpReq)

	if resp.Code != http.StatusOK {
		var updatesError api.Error
		jsonErr := json.Unmarshal(resp.Body.Bytes(), &updatesError)
		Expect(jsonErr).To(BeNil())

		return nil, &updatesError
	}

	var updatesResp api.CredentialsUpdatesResp
	jsonErr := json.Unmarshal(resp.Body.Bytes(), &updatesResp)
	Expect(jsonErr).To(BeNil())

	return &updatesResp, nil
}

func listDatasourcesUpdatesForTest(randoms *testutils.RandomValues, router chi.Router, after int64, report *api.ClusterSyncReport) (*api.DatasourceUpdatesResp, *api.Error) {
	url := fmt.Sprintf("%s?clusterId=%s&after=%d", urlGetDatasourceUpdates, randoms.ClusterId, after)

	if report == nil {
		report = &api.ClusterSyncReport{}
	}

	syncHttpReq, err := http.NewRequest("POST", url, makePayload(report))
	Expect(err).To(BeNil())
	syncHttpReq.Header = *headerForAdmin(randoms)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, syncHttpReq)

	if resp.Code != http.StatusOK {
		var updatesError api.Error
		jsonErr := json.Unmarshal(resp.Body.Bytes(), &updatesError)
		Expect(jsonErr).To(BeNil())

		return nil, &updatesError
	}

	var updatesResp api.DatasourceUpdatesResp
	jsonErr := json.Unmarshal(resp.Body.Bytes(), &updatesResp)
	Expect(jsonErr).To(BeNil())

	return &updatesResp, nil
}

func validateCredentialsStatusForTest(randoms *testutils.RandomValues, router chi.Router, expected map[string]api.AssetClusterStatusInfo) {
	syncHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?statusInfo=true", urlCredentials), nil)
	Expect(err).To(BeNil())
	syncHttpReq.Header = *headerForAdmin(randoms)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, syncHttpReq)
	Expect(resp.Code).To(Equal(http.StatusOK))

	var listResp api.CredentialsListResponse
	jsonErr := json.Unmarshal(resp.Body.Bytes(), &listResp)
	Expect(jsonErr).To(BeNil())

	actual := make(map[string]api.AssetClusterStatusInfo)

	for _, entry := range listResp.Entries {
		if entry.Status != nil {
			actual[metaToString(entry.Meta)] = *entry.Status
		}
	}

	//testutils.Expect(actual).ToEqual(expected)
	Expect(actual).To(HaveLen(len(expected)))
	for k, v := range expected {
		Expect(actual).To(HaveKeyWithValue(k, v))
	}
}

func validateDatasourcesStatusForTest(randoms *testutils.RandomValues, router chi.Router,
	expected map[string]api.AssetClusterStatusInfo, expectedCompliance map[string]api.ComplianceInfo) {
	syncHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?statusInfo=true&complyToWorkloadType=%s&complyToProject=%d",
		urlDatasource, api.WORKLOADTYPE_WORKSPACE, randoms.ProjectId), nil)
	Expect(err).To(BeNil())
	syncHttpReq.Header = *headerForAdmin(randoms)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, syncHttpReq)
	Expect(resp.Code).To(Equal(http.StatusOK))

	var listResp api.DatasourceListResponse
	jsonErr := json.Unmarshal(resp.Body.Bytes(), &listResp)
	Expect(jsonErr).To(BeNil())

	actual := make(map[string]api.AssetClusterStatusInfo)
	actualCompliance := make(map[string]api.ComplianceInfo)

	for _, entry := range listResp.Entries {
		if entry.Status != nil {
			actual[metaToString(entry.Meta)] = *entry.Status
		}
		if entry.Compliance != nil {
			actualCompliance[metaToString(entry.Meta)] = *entry.Compliance
		}
	}

	testutils.Expect(actual).ToEqual(expected)
	testutils.Expect(actualCompliance).ToEqual(expectedCompliance)
}

/////////////////////////////////////////////
// copy helpers for the test
/////////////////////////////////////////////

type copierSource interface {
	api.AssetMeta
}

type copierTarget interface {
	api.AssetMetaForSync
}

func srcToTarget[S copierSource, T copierTarget](source S) T {
	var result T
	err := copier.Copy(&result, &source)
	Expect(err).To(BeNil())
	return result
}

func metaToString(meta api.AssetMeta) string {
	scopeInfo := ScopeInfoOfMeta(&meta)
	return fmt.Sprintf("%s:%s %s:%s", meta.Kind, meta.Name, meta.Scope, scopeInfo.ScopeId)
}

func toAssetIssues(issues []k8sObjectsApi.ClusterObjectIssue) []api.AssetClusterStatusIssue {
	result := make([]api.AssetClusterStatusIssue, 0, len(issues))

	for _, issue := range issues {
		result = append(result, api.AssetClusterStatusIssue{
			ScopeType: api.ScopeType(issue.ScopeType),
			ScopeId:   issue.ScopeId,
			Issue:     api.AssetClusterStatusIssueEnum(issue.Issue),
		})
	}

	return result
}
