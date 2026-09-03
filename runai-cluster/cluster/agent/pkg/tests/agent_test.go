/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/adapters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	"github.com/google/uuid"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	assets_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	openapi2 "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	org_unit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	dpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"

	infpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"

	ipv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"

	tpv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"

	policy_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"

	workloads_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"

	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	EventuallyTimeout = 30 * time.Second
)

var _ = Describe("Agent tests", func() {
	projectsOrgUnitPath := "/api/v1/org-unit/projects/sync"
	departmentsOrgUnitPath := "/api/v1/org-unit/departments/sync"
	nodepoolsPath := "/api/v1/node-pools"
	accessrulesPath := "/api/v1/authorization/access-rules/sync"
	interactivePolicyPath := "/api/v2/policy/workspaces/sync"
	trainingPolicyPath := "/api/v2/policy/trainings/sync"
	interactiveWorkloadPath := "/api/v1/workloads/workspaces/updates"
	distributedPolicyPath := "/api/v2/policy/distributed/sync"
	inferencePolicyPath := "/api/v2/policy/inferences/sync"
	trainingWorkloadPath := "/api/v1/workloads/trainings/updates"
	distributedWorkloadPath := "/api/v1/workloads/distributed/updates"
	inferenceWorkloadPath := "/api/v1/workloads/inferences/updates"
	datasourceAssetsPath := "/api/v1/asset/datasource/updates"
	workloadMgrPath := "/api/v2/workloads/updates"
	credentialsAssetsPath := "/api/v1/asset/credentials/updates"
	datavolumesPath := "/api/v1/datavolumes/sync"
	usersPath := fmt.Sprintf("/v1/k8s/clusters/%s/users", TestClusterId)
	groupsPath := fmt.Sprintf("/v1/k8s/clusters/%s/groups", TestClusterId)
	workloadsTypesPath := "/api/v1/workload-types"
	workloadsPrioritiesPath := "/api/v1/workload-priorities-sync"
	workloadsCategoriesPath := "/api/v1/workload-categories"
	networkTopologiesPath := "/api/v1/network-topologies"

	var agent *pkg.RunaiAgent
	var cancel context.CancelFunc
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var k8sClient client.WithWatch

	var existingProjectToBeReconciled *v2.Project
	var projectToBeDeleted *v2.Project
	var projectToBeDeletedAndCreated *v2.Project

	var existingDepartmentToBeReconciled *schedv2alpha2.Department
	var departmentToBeDeleted *schedv2alpha2.Department

	var existingNodePoolToBeReconciled *kainodepool.NodePool
	var nodePoolToBeDeleted *kainodepool.NodePool

	var existingDatavolumeToBeReconciled *v1alpha1.DataVolume
	var datavolumeToBeDeleted *v1alpha1.DataVolume

	var existingAccessRuleToBeReconciled *v1alpha1.AccessRule
	var accessRuleToBeDeleted *v1alpha1.AccessRule

	var existingInteractivePolicyToBeReconciled *ipv1.InteractivePolicy
	var existingInteractivePolicyToDeleted *ipv1.InteractivePolicy

	var existingTrainingPolicyToBeReconciled *tpv1.TrainingPolicy
	var existingTrainingPolicyToBeDeleted *tpv1.TrainingPolicy

	var existingDistributedPolicyToBeReconciled *dpv1.DistributedPolicy
	var existingDistributedPolicyToBeDeleted *dpv1.DistributedPolicy

	var existingInferencePolicyToBeReconciled *infpv1.InferencePolicy
	var existingInferencePolicyToBeDeleted *infpv1.InferencePolicy

	var existingNetworkTopologyToBeReconciled *kaiv1alpha1.Topology
	var networkTopologyToBeDeleted *kaiv1alpha1.Topology

	var existingClusterNetworkTopology1 *kaiv1alpha1.Topology
	var existingClusterNetworkTopology2 *kaiv1alpha1.Topology

	var existingClusterKarta1 *kartav1alpha1.Karta
	var existingClusterKarta2 *kartav1alpha1.Karta
	var existingClusterKarta3 *kartav1alpha1.Karta
	var existingKartaToBeDeleted *kartav1alpha1.Karta

	var stopper context.Context

	BeforeEach(func() {
		GinkgoWriter.Println("Creating server for tests...")
		server = CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for tests %s \n", server.URL())

		server.AllowUnhandledRequests = true
	})

	Context("runai backend is available", func() {
		BeforeEach(func() {
			server.RouteToHandler("GET", projectsOrgUnitPath, func(writer http.ResponseWriter, request *http.Request) {
				q := request.URL.Query()
				offset := getIntFromQuery(q, "offset")
				resObj := &org_unit_openapi.GetSyncProjects200Response{Projects: []org_unit_openapi.ProjectSync{}}
				switch offset {
				case 0:
					resObj = createProjectsOrgUnitGetResponse(Project1Name, Project2Name)
				case 2:
					resObj = createProjectsOrgUnitGetResponse(Project3Name, Project4Name)
				case 4:
					resObj = createProjectsOrgUnitGetDelResponse()
				}
				ghttp.RespondWithJSONEncoded(http.StatusOK, &resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET org-unit projects mock server calls...")

			server.RouteToHandler("GET", departmentsOrgUnitPath, func(writer http.ResponseWriter, request *http.Request) {
				q := request.URL.Query()
				offset := getIntFromQuery(q, "offset")
				resObj := &org_unit_openapi.GetSyncDepartments200Response{Departments: []org_unit_openapi.DepartmentSync{}}
				switch offset {
				case 0:
					resObj = createDepartmentsOrgUnitGetResponse(Department1Name, Department2Name)
				case 2:
					resObj = createDepartmentsOrgUnitGetResponse(Department3Name, Department4Name)
				case 4:
					resObj = createDepartmentsOrgUnitGetResponse(BadDeletionDepartment)
				}
				ghttp.RespondWithJSONEncoded(http.StatusOK, &resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET org-unit departments mock server calls...")

			server.RouteToHandler("GET", nodepoolsPath, func(writer http.ResponseWriter, request *http.Request) {
				nodepools := createNodePoolsGetResponse(NodePoolDName, NodePool1Name, NodePool2Name, NodePool3Name, NodePoolDeletingName)

				resObj := &openapi2.GetNodepools200Response{Nodepools: nodepools}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET nodepools mock server calls...")

			server.RouteToHandler("GET", datavolumesPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createDatavolumesGetResponse(DataVolume1Name, DataVolume2Name, DataVolume3Name, DataVolume4Name, DataVolumeDeletingName)
				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET datavolumes mock server calls...")

			server.RouteToHandler("POST", interactivePolicyPath, func(writer http.ResponseWriter, request *http.Request) {
				interactivePolicies := []policy_openapi.PolicySyncResponsePoliciesInner{
					BackendPolicy1,
					BackendPolicy5,
					BackendPolicy2, BackendPolicy3, BackendPolicy4,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &policy_openapi.PolicySyncResponse{
					RequestInfo: &policy_openapi.PolicySyncResponseRequestInfo{
						After: intTo64Pointer(0),
					},
					ResponseInfo: &policy_openapi.PolicySyncResponseResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Policies: []policy_openapi.PolicySyncResponsePoliciesInner{},
				}

				for _, interactivePolicy := range interactivePolicies {
					if int64(after) >= *interactivePolicy.UpdateCounter {
						continue
					}
					if *interactivePolicy.UpdateCounter <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = interactivePolicy.UpdateCounter
					}
					if *interactivePolicy.UpdateCounter >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = interactivePolicy.UpdateCounter
					}
					resObj.Policies = append(resObj.Policies, interactivePolicy)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST interactive policy mock server calls...")

			server.RouteToHandler("POST", trainingPolicyPath, func(writer http.ResponseWriter, request *http.Request) {
				trainingPolicies := []policy_openapi.PolicySyncResponsePoliciesInner{
					BackendPolicy1,
					BackendPolicy5,
					BackendPolicy2, BackendPolicy3, BackendPolicy4,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &policy_openapi.PolicySyncResponse{
					RequestInfo: &policy_openapi.PolicySyncResponseRequestInfo{
						After: intTo64Pointer(0),
					},
					ResponseInfo: &policy_openapi.PolicySyncResponseResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Policies: []policy_openapi.PolicySyncResponsePoliciesInner{},
				}

				for _, trainingPolicy := range trainingPolicies {
					if int64(after) >= *trainingPolicy.UpdateCounter {
						continue
					}
					if *trainingPolicy.UpdateCounter <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = trainingPolicy.UpdateCounter
					}
					if *trainingPolicy.UpdateCounter >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = trainingPolicy.UpdateCounter
					}
					resObj.Policies = append(resObj.Policies, trainingPolicy)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST training policy mock server calls...")

			server.RouteToHandler("POST", distributedPolicyPath, func(writer http.ResponseWriter, request *http.Request) {
				distributedPolicies := []policy_openapi.PolicySyncResponsePoliciesInner{
					BackendPolicy6,
					BackendPolicy7,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &policy_openapi.PolicySyncResponse{
					RequestInfo: &policy_openapi.PolicySyncResponseRequestInfo{
						After: intTo64Pointer(0),
					},
					ResponseInfo: &policy_openapi.PolicySyncResponseResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Policies: []policy_openapi.PolicySyncResponsePoliciesInner{},
				}

				for _, distributedPolicy := range distributedPolicies {
					if int64(after) >= *distributedPolicy.UpdateCounter {
						continue
					}
					if *distributedPolicy.UpdateCounter <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = distributedPolicy.UpdateCounter
					}
					if *distributedPolicy.UpdateCounter >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = distributedPolicy.UpdateCounter
					}
					resObj.Policies = append(resObj.Policies, distributedPolicy)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST distributed policy mock server calls...")

			server.RouteToHandler("POST", inferencePolicyPath, func(writer http.ResponseWriter, request *http.Request) {
				inferencePolicies := []policy_openapi.PolicySyncResponsePoliciesInner{
					BackendPolicy9,
					BackendPolicy10,
					BackendPolicy2, BackendPolicy3, BackendPolicy4,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &policy_openapi.PolicySyncResponse{
					RequestInfo: &policy_openapi.PolicySyncResponseRequestInfo{
						After: intTo64Pointer(0),
					},
					ResponseInfo: &policy_openapi.PolicySyncResponseResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Policies: []policy_openapi.PolicySyncResponsePoliciesInner{},
				}

				for _, inferencePolicy := range inferencePolicies {
					if int64(after) >= *inferencePolicy.UpdateCounter {
						continue
					}
					if *inferencePolicy.UpdateCounter <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = inferencePolicy.UpdateCounter
					}
					if *inferencePolicy.UpdateCounter >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = inferencePolicy.UpdateCounter
					}
					resObj.Policies = append(resObj.Policies, inferencePolicy)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST inference policy mock server calls...")

			server.RouteToHandler("POST", interactiveWorkloadPath, func(writer http.ResponseWriter, request *http.Request) {
				interactiveWorkloads := []workloads_openapi.WorkloadsSyncInfo{
					BackendNonDistributedWorkload1,
					BackendNonDistributedWorkload2,
					BackendNonDistributedWorkload3,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "updatesAfter")
				resObj := &workloads_openapi.WorkloadsSyncResponse{
					RequestInfo: &workloads_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &workloads_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Workloads: []workloads_openapi.WorkloadsSyncInfo{},
				}

				for _, interactiveWorkload := range interactiveWorkloads {
					if int64(after) >= *interactiveWorkload.UpdateCount {
						continue
					}
					if *interactiveWorkload.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = interactiveWorkload.UpdateCount
					}
					if *interactiveWorkload.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = interactiveWorkload.UpdateCount
					}
					resObj.Workloads = append(resObj.Workloads, interactiveWorkload)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST interactive workloads mock server calls...")

			server.RouteToHandler("POST", trainingWorkloadPath, func(writer http.ResponseWriter, request *http.Request) {
				trainingWorkloads := []workloads_openapi.WorkloadsSyncInfo{
					BackendNonDistributedWorkload1,
					BackendNonDistributedWorkload2,
					BackendNonDistributedWorkload3,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "updatesAfter")
				resObj := &workloads_openapi.WorkloadsSyncResponse{
					RequestInfo: &workloads_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &workloads_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Workloads: []workloads_openapi.WorkloadsSyncInfo{},
				}

				for _, trainingWorkload := range trainingWorkloads {
					if int64(after) >= *trainingWorkload.UpdateCount {
						continue
					}
					if *trainingWorkload.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = trainingWorkload.UpdateCount
					}
					if *trainingWorkload.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = trainingWorkload.UpdateCount
					}
					resObj.Workloads = append(resObj.Workloads, trainingWorkload)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST training workloads mock server calls...")

			server.RouteToHandler("POST", distributedWorkloadPath, func(writer http.ResponseWriter, request *http.Request) {
				distributedWorkloads := []workloads_openapi.WorkloadsSyncInfo{
					BackendDistributedWorkload1,
					BackendDistributedWorkload2,
					BackendDistributedWorkload3,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "updatesAfter")
				resObj := &workloads_openapi.WorkloadsSyncResponse{
					RequestInfo: &workloads_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &workloads_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Workloads: []workloads_openapi.WorkloadsSyncInfo{},
				}

				for _, distributedWorkload := range distributedWorkloads {
					if int64(after) >= *distributedWorkload.UpdateCount {
						continue
					}
					if *distributedWorkload.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = distributedWorkload.UpdateCount
					}
					if *distributedWorkload.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = distributedWorkload.UpdateCount
					}
					resObj.Workloads = append(resObj.Workloads, distributedWorkload)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST distributed workloads mock server calls...")

			server.RouteToHandler("POST", inferenceWorkloadPath, func(writer http.ResponseWriter, request *http.Request) {
				inferenceWorkloads := []workloads_openapi.WorkloadsSyncInfo{
					BackendInferenceWorkload1,
					BackendInferenceWorkload2,
					BackendInferenceWorkload3,
				}
				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "updatesAfter")
				resObj := &workloads_openapi.WorkloadsSyncResponse{
					RequestInfo: &workloads_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &workloads_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Workloads: []workloads_openapi.WorkloadsSyncInfo{},
				}

				for _, inferenceWorkload := range inferenceWorkloads {
					if int64(after) >= *inferenceWorkload.UpdateCount {
						continue
					}
					if *inferenceWorkload.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = inferenceWorkload.UpdateCount
					}
					if *inferenceWorkload.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = inferenceWorkload.UpdateCount
					}
					resObj.Workloads = append(resObj.Workloads, inferenceWorkload)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST inference workloads mock server calls...")

			server.RouteToHandler("POST", datasourceAssetsPath, func(writer http.ResponseWriter, request *http.Request) {
				datasourceUpdates := allDatasourceUpdates

				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &assets_openapi.DatasourceUpdatesResp{
					RequestInfo: &assets_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &assets_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Assets: []assets_openapi.DatasourceUpdate{},
				}

				for _, datasourceUpdate := range datasourceUpdates {
					if datasourceUpdate.UpdateCount <= int64(after) {
						continue
					}

					if datasourceUpdate.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = convutils.Ref(datasourceUpdate.UpdateCount)
					}
					if datasourceUpdate.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = convutils.Ref(datasourceUpdate.UpdateCount)
					}
					resObj.Assets = append(resObj.Assets, datasourceUpdate)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST datasources mock server calls...")

			server.RouteToHandler("POST", credentialsAssetsPath, func(writer http.ResponseWriter, request *http.Request) {
				credentialsUpdates := allCredentialsUpdates

				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "after")
				resObj := &assets_openapi.CredentialsUpdatesResp{
					RequestInfo: &assets_openapi.RequestInfo{
						AfterUpdate: int64(0),
					},
					ResponseInfo: &assets_openapi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Assets: []assets_openapi.CredentialsUpdate{},
				}

				for _, credentialsUpdate := range credentialsUpdates {
					if credentialsUpdate.UpdateCount <= int64(after) {
						continue
					}

					if credentialsUpdate.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = convutils.Ref(credentialsUpdate.UpdateCount)
					}
					if credentialsUpdate.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = convutils.Ref(credentialsUpdate.UpdateCount)
					}
					resObj.Assets = append(resObj.Assets, credentialsUpdate)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST datasources mock server calls...")

			server.RouteToHandler("GET", accessrulesPath, func(writer http.ResponseWriter, request *http.Request) {
				accessRules := []string{
					AccessRule1Name,                                   // exists - should be updated
					AccessRule2Name, AccessRule3Name, AccessRule4Name, // don't exist - should be created
					AccessRule5ToBeDeleted,      // exists - should be deleted
					AccessRule6AlreadyDeleted,   // doesn't exist - should try to delete and succeed cause errors.IsNotFound
					AccessRule7FutureUpdateTime, // will be returned only once we update it's UpdateAt in the test
				}

				// Parse the query parameters from the request.
				q := request.URL.Query()
				lastUpdated := getTimeFromQuery(q, "lastUpdated")
				limit := getIntFromQuery(q, "limit")
				offset := getIntFromQuery(q, "offset")

				accessRulesToReturn := filterOutLastUpdatedIncludeDeleted(accessRules, lastUpdated)
				accessRulesToReturn = filterOutOffsetLimit(accessRulesToReturn, offset, limit)

				resObj := createAccessRulesGetResponse(accessRulesToReturn...)
				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET accessrules mock server calls...")

			server.RouteToHandler("GET", usersPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createUsersGetResponse(User1Name, User2Name, User3Name)
				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET users mock server calls...")

			server.RouteToHandler("GET", groupsPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createGroupsGetResponse(Group1Name, Group2Name, Group3Name)
				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET users mock server calls...")

			server.RouteToHandler("GET", workloadsTypesPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createListWorkloadsTypesResponse()

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			server.RouteToHandler("GET", workloadsPrioritiesPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createListWorkloadsPrioritiesResponse()

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET workloads priorities mock server calls...")

			server.RouteToHandler("GET", workloadsCategoriesPath, func(writer http.ResponseWriter, request *http.Request) {
				resObj := createListWorkloadsCategories()

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET workloads categories  mock server calls...")

			server.RouteToHandler("GET", networkTopologiesPath, func(writer http.ResponseWriter, request *http.Request) {
				networkTopologies := []openapi2.NetworkTopology{
					BackendNetworkTopology1,
					BackendNetworkTopology2,
					BackendNetworkTopology3,
					BackendNetworkTopology4,
					BackendNetworkTopologyCluster1,
					BackendNetworkTopologyCluster2,
				}
				resObj := &openapi2.GetNetworkTopologies200Response{NetworkTopologies: networkTopologies}
				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created GET network topologies mock server calls...")

			server.RouteToHandler("POST", workloadMgrPath, func(writer http.ResponseWriter, request *http.Request) {
				workloadsV2Updates := allWorkloadsV2Updates

				// parse the query parameters from the request.
				q := request.URL.Query()
				after := getIntFromQuery(q, "updatesAfter")
				resObj := &workloadsMgrApi.WorkloadsV2Updates{
					RequestInfo: &workloadsMgrApi.RequestInfo{
						AfterUpdate: int64(after),
					},
					ResponseInfo: &workloadsMgrApi.ResponseInfo{
						LowestUpdate:  intTo64Pointer(0),
						HighestUpdate: intTo64Pointer(0),
					},
					Workloads: []workloadsMgrApi.WorkloadV2Update{},
				}

				for _, workload := range workloadsV2Updates {
					if workload.UpdateCount <= int64(after) {
						continue
					}

					if workload.UpdateCount <= *resObj.ResponseInfo.LowestUpdate {
						resObj.ResponseInfo.LowestUpdate = convutils.Ref(workload.UpdateCount)
					}
					if workload.UpdateCount >= *resObj.ResponseInfo.HighestUpdate {
						resObj.ResponseInfo.HighestUpdate = convutils.Ref(workload.UpdateCount)
					}
					resObj.Workloads = append(resObj.Workloads, workload)
				}

				ghttp.RespondWithJSONEncoded(http.StatusOK, resObj)(writer, request)
			})
			GinkgoWriter.Printf("Created POST workload V2 mock server calls...")

			stopper, cancel = context.WithCancel(context.TODO())
		})

		JustBeforeEach(func() {
			existingProjectToBeReconciled = getBasicProjectObject(Project1Name, strconv.Itoa(Proj1Id), Department1Name, strconv.Itoa(Dep1Id), NamespaceForSpec1)
			projectToBeDeleted = getBasicProjectObject(Project4Name, strconv.Itoa(4), Department1Name, strconv.Itoa(Dep1Id), "")
			projectToBeDeletedAndCreated = getBasicProjectObject(Project3Name, strconv.Itoa(22), Department1Name, strconv.Itoa(Dep1Id), "")
			// user and group that are not part of any other project..
			projectToBeDeleted.Spec.Permissions = v2.ProjectPermissions{
				Users:  []string{User3Name},
				Groups: []string{Group3Name},
			}
			projectToBeDeletedAndCreated.Spec.Permissions = v2.ProjectPermissions{
				Users:  []string{User3Name},
				Groups: []string{Group3Name},
			}

			existingDepartmentToBeReconciled = getBasicDepartmentObject(Department1Name, strconv.Itoa(Dep1Id))
			departmentToBeDeleted = getBasicDepartmentObject(Department4Name, strconv.Itoa(Dep4Id))

			existingNodePoolToBeReconciled = getBasicNodePoolObject(NodePool1Name)
			nodePoolToBeDeleted = getBasicNodePoolObject(NodePoolDeletingName)

			existingDatavolumeToBeReconciled = getBasicDatavolumeObject(DataVolume1Name)
			datavolumeToBeDeleted = getBasicDatavolumeObject(DataVolumeDeletingName)

			existingAccessRuleToBeReconciled = getBasicAccessRuleObject(AccessRule1Name, strconv.Itoa(int(BackendAccessRule2.Id)))
			accessRuleToBeDeleted = getBasicAccessRuleObject(AccessRule5ToBeDeleted, strconv.Itoa(int(BackendAccessRule5ToBeDeleted.Id)))

			existingInteractivePolicyToBeReconciled = getBasicInteractivePolicyObject(policy1.name, policy1.namespace, policy1.preUpdatedCounter, policy1.scope, policy1.scopeId)
			existingInteractivePolicyToDeleted = getBasicInteractivePolicyObject(policy5.name, policy5.namespace, policy5.preUpdatedCounter, policy5.scope, policy5.scopeId)

			existingTrainingPolicyToBeReconciled = getBasicTrainingPolicyObject(policy1.name, policy1.namespace, policy1.preUpdatedCounter, policy1.scope, strconv.Itoa(policy1.scopeId), policy1.spec.(tpv1.TrainingSpec))
			existingTrainingPolicyToBeDeleted = getBasicTrainingPolicyObject(policy5.name, policy5.namespace, policy5.preUpdatedCounter, policy5.scope, strconv.Itoa(policy5.scopeId), policy5.spec.(tpv1.TrainingSpec))

			existingDistributedPolicyToBeReconciled = getBasicDistributedPolicyObject(policy8.name, policy8.namespace, policy8.preUpdatedCounter, policy8.scope, strconv.Itoa(policy8.scopeId),
				policy8.spec.(dpv1.DistributedSpec), policy8.masterSpec.(*dpv1.DistributedMasterSpec))
			existingDistributedPolicyToBeDeleted = getBasicDistributedPolicyObject(policy6.name, policy6.namespace, policy6.preUpdatedCounter, policy6.scope, strconv.Itoa(policy6.scopeId),
				policy6.spec.(dpv1.DistributedSpec), policy6.masterSpec.(*dpv1.DistributedMasterSpec))

			existingInferencePolicyToBeReconciled = getBasicInferencePolicyObject(policy9.name, policy9.namespace, policy9.preUpdatedCounter, policy9.scope, strconv.Itoa(policy9.scopeId), policy9.spec.(infpv1.InferenceSpec))
			existingInferencePolicyToBeDeleted = getBasicInferencePolicyObject(policy10.name, policy10.namespace, policy10.preUpdatedCounter, policy10.scope, strconv.Itoa(policy10.scopeId), policy10.spec.(infpv1.InferenceSpec))

			existingNetworkTopologyToBeReconciled = getBasicNetworkTopologyObject(NetworkTopology1Name)
			networkTopologyToBeDeleted = getBasicNetworkTopologyObject(NetworkTopologyDeletingName)

			// Create cluster network topologies that should not be deleted
			existingClusterNetworkTopology1 = getBasicClusterNetworkTopologyObject(NetworkTopologyCluster1Name)
			existingClusterNetworkTopology2 = getBasicClusterNetworkTopologyObject(NetworkTopologyCluster2Name)

			existingPriorityClass := getPriorityClass(priorityClassUpdatedName, priorityOriginalValue, priorityClassUpdatedId, clusterCommonConstants.PriorityClassRunaiSource)
			existingNonRunaiPriorityClass := getPriorityClass(priorityClassNonRunaiName, priorityOriginalValue, "", "")

			existingClusterKarta1 = getBasicKartaObject(KartaAppsDeploymentName, KartaGroupApps, KartaVersionV1, KartaKindDeployment)
			existingClusterKarta2 = getBasicKartaObject(KartaRayClusterName, KartaGroupRayIo, KartaVersionV1Alpha2, KartaKindRayCluster)
			existingClusterKarta3 = getBasicKartaObject(KartaRayServiceName, KartaGroupRayIo, KartaVersionV1Alpha2, KartaKindRayService)
			existingClusterKarta3.Spec.StructureDefinition.RootComponent.Name = "some-other-name-for-differentiate"
			existingKartaToBeDeleted = getBasicKartaObject(KartaBatchJobName, KartaGroupBatch, KartaVersionV1, KartaKindJob)

			GinkgoWriter.Println("Starting agent")
			backendClient = backendclient.NewBackendClient(server.URL(), &backendclient.Options{UseDebugLogLevel: false})
			// backendClient.OrgUnitClient = ()
			k8sClient = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(
				existingProjectToBeReconciled, projectToBeDeleted, projectToBeDeletedAndCreated,
				existingDepartmentToBeReconciled, departmentToBeDeleted,
				existingNodePoolToBeReconciled, nodePoolToBeDeleted,
				existingDatavolumeToBeReconciled, datavolumeToBeDeleted,
				existingAccessRuleToBeReconciled, accessRuleToBeDeleted,
				existingInteractivePolicyToBeReconciled, existingInteractivePolicyToDeleted,
				existingTrainingPolicyToBeReconciled, existingTrainingPolicyToBeDeleted,
				existingDistributedPolicyToBeReconciled, existingDistributedPolicyToBeDeleted,
				existingInferencePolicyToBeReconciled, existingInferencePolicyToBeDeleted,
				existingNetworkTopologyToBeReconciled, networkTopologyToBeDeleted,
				existingClusterNetworkTopology1, existingClusterNetworkTopology2,
				existingProjPvcA, existingDeptPvcA, existingProjPvcB, existingDeptPvcB,
				existingProjPasswordA, existingDeptAccessKey, existingProjDockerRegistry, existingDeptGenericSecret,
				existingPriorityClass, existingNonRunaiPriorityClass,
				existingV2WorkloadA, existingV2WorkloadB,
				existingClusterKarta1, existingClusterKarta3, existingKartaToBeDeleted,
			).WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					err := client.Create(ctx, obj, opts...)
					if err != nil {
						return err
					}
					obj.SetCreationTimestamp(metav1.Now())
					obj.SetUID(types.UID(uuid.New().String()))
					return nil
				},
				// The newer controller-runtime fake client strips TypeMeta (apiVersion/kind)
				// from objects on Get. Some production code (e.g. AssetsObjectInspector.DeepCopySpec)
				// switches on the fetched object's Kind, so restore the GVK from the scheme here.
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if err := c.Get(ctx, key, obj, opts...); err != nil {
						return err
					}
					if obj.GetObjectKind().GroupVersionKind().Empty() {
						if gvks, _, err := scheme.ObjectKinds(obj); err == nil && len(gvks) > 0 {
							obj.GetObjectKind().SetGroupVersionKind(gvks[0])
						}
					}
					return nil
				},
			}).Build()

			seedKaiOrgUnits(k8sClient)

			agent = pkg.NewRunaiAgent(backendClient, k8sClient, k8sClient, config.Config().AgentSyncIntervalMillis)
		})

		JustAfterEach(func() {
			cancel()
		})

		Describe("projects and departments", func() {
			It("Projects from backend created + unnecessary project deleted + updated projects reconciled", func() {
				// reset the shared mock-control var (the legacy projects spec leaves it false)
				Project2InDepartment2 = true

				// the existing project1 does not match the backend project
				Expect(getProjectFromClient(Project1Name, k8sClient)).ToNot(BeNil())
				validateExistingKaiProjectDoesNotMatchBackendSpec(Project1Name, k8sClient)

				Expect(getProjectFromClient(Project2Name, k8sClient)).To(BeNil())
				Expect(getProjectFromClient(projectToBeDeleted.Name, k8sClient)).ToNot(BeNil())
				deletedAndCreatedProject := getProjectFromClient(projectToBeDeletedAndCreated.Name, k8sClient)
				Expect(deletedAndCreatedProject).ToNot(BeNil())
				Expect(deletedAndCreatedProject.Labels[v2.ProjectIdLabel]).To(Equal(OrgUnitProject3del.Id))

				go agent.Start(stopper)

				validateKaiProjectExistsEventually(Project1Name, k8sClient)
				validateKaiProjectExistsEventually(Project2Name, k8sClient)
				validateKaiProjectExistsEventually(Project3Name, k8sClient)
				validateKaiProjectDeletedEventually(projectToBeDeleted.Name, k8sClient)

				// after the reconcile; the existing project1 should match the backend project
				validateExistingKaiProjectMatchesBackendSpec(Project1Name, k8sClient)
				validateExistingKaiProjectMatchesBackendSpec(Project2Name, k8sClient)
				validateExistingKaiProjectMatchesBackendSpec(Project3Name, k8sClient)

				// update a random field in project 2 (parent department)
				proj2 := getProjectFromClient(Project2Name, k8sClient)
				proj2.Spec.Parent = proj2.Spec.Parent + "some-suffix"
				Expect(k8sClient.Update(context.Background(), proj2)).To(Succeed())

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					proj := getProjectFromClient(Project2Name, k8sClient)
					g.Expect(proj.Spec.Parent).To(Equal(common.GetStringVal(BackendProjectsMap[Project2Name].DepartmentName)))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiProjectMatchesBackendSpec(Project2Name, k8sClient)

				// update the projectId label in project 2
				proj2 = getProjectFromClient(Project2Name, k8sClient)
				proj2.Labels[v2.ProjectIdLabel] = "wronggggg-id"
				Expect(k8sClient.Update(context.Background(), proj2)).To(Succeed())

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					proj := getProjectFromClient(Project2Name, k8sClient)
					g.Expect(proj.Labels[v2.ProjectIdLabel]).To(Equal(strconv.Itoa(Proj2Id)))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiProjectMatchesBackendSpec(Project2Name, k8sClient)

				// delete the projectId label in project 2
				proj2 = getProjectFromClient(Project2Name, k8sClient)
				delete(proj2.Labels, v2.ProjectIdLabel)
				Expect(k8sClient.Update(context.Background(), proj2)).To(Succeed())

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					proj := getProjectFromClient(Project2Name, k8sClient)
					g.Expect(proj.Labels[v2.ProjectIdLabel]).To(Equal(strconv.Itoa(Proj2Id)))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiProjectMatchesBackendSpec(Project2Name, k8sClient)

				// update the departmentId label in project 2
				proj2 = getProjectFromClient(Project2Name, k8sClient)
				proj2.Labels[schedv2alpha2.DepartmentIdLabel] = "wronggggg-id"
				Expect(k8sClient.Update(context.Background(), proj2)).To(Succeed())

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					proj := getProjectFromClient(Project2Name, k8sClient)
					g.Expect(proj.Labels[schedv2alpha2.DepartmentIdLabel]).To(Equal(strconv.Itoa(Dep2Id)))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiProjectMatchesBackendSpec(Project2Name, k8sClient)

				// update var so returned project2 will be of department1 instead of department2
				Project2InDepartment2 = false
				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					proj := getProjectFromClient(Project2Name, k8sClient)
					g.Expect(proj.Labels[schedv2alpha2.DepartmentIdLabel]).ToNot(Equal(strconv.Itoa(Dep2Id)))
					g.Expect(proj.Labels[schedv2alpha2.DepartmentIdLabel]).To(Equal(strconv.Itoa(Dep1Id)))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiProjectDoesNotMatchBackendSpec(Project2Name, k8sClient)
			})

			It("Departments from backend created + unnecessary department deleted + updated departments reconciled", func() {
				// the existing department1 does not match the backend department
				Expect(getDepartmentFromClient(Department1Name, k8sClient)).ToNot(BeNil())
				validateExistingKaiDepartmentDoesNotMatchBackendSpec(Department1Name, k8sClient)

				Expect(getDepartmentFromClient(Department2Name, k8sClient)).To(BeNil())
				Expect(getDepartmentFromClient(Department3Name, k8sClient)).To(BeNil())
				Expect(getDepartmentFromClient(departmentToBeDeleted.Name, k8sClient)).ToNot(BeNil())

				go agent.Start(stopper)

				validateKaiDepartmentExistsEventually(Department1Name, k8sClient)
				validateKaiDepartmentExistsEventually(Department2Name, k8sClient)
				validateKaiDepartmentExistsEventually(Department3Name, k8sClient)
				validateKaiDepartmentDeletedEventually(departmentToBeDeleted.Name, k8sClient)

				// after the reconcile; the existing department1 should match the backend department
				validateExistingKaiDepartmentMatchesBackendSpec(Department1Name, k8sClient)
				validateExistingKaiDepartmentMatchesBackendSpec(Department2Name, k8sClient)
				validateExistingKaiDepartmentMatchesBackendSpec(Department3Name, k8sClient)

				// update a random field in department 2 (clear its queues)
				dep2 := getDepartmentFromClient(Department2Name, k8sClient)
				queuesToCompareTo := dep2.Spec.Queues
				dep2.Spec.Queues = []kainodepool.QueueConfig{}
				Expect(k8sClient.Update(context.Background(), dep2)).To(Succeed())
				Expect(getDepartmentFromClient(Department2Name, k8sClient).Spec.Queues).ToNot(Equal(queuesToCompareTo))

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					dep := getDepartmentFromClient(Department2Name, k8sClient)
					g.Expect(dep.Spec.Queues).To(Equal(queuesToCompareTo))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiDepartmentMatchesBackendSpec(Department2Name, k8sClient)

				// delete the departmentId label in department 2
				dep2 = getDepartmentFromClient(Department2Name, k8sClient)
				dep2IdToCompareTo := dep2.Labels[schedv2alpha2.DepartmentIdLabel]
				delete(dep2.Labels, schedv2alpha2.DepartmentIdLabel)
				Expect(k8sClient.Update(context.Background(), dep2)).To(Succeed())
				Expect(getDepartmentFromClient(Department2Name, k8sClient).Labels[schedv2alpha2.DepartmentIdLabel]).ToNot(Equal(dep2IdToCompareTo))

				// expect it to be reconciled eventually...
				Eventually(func(g Gomega) {
					dep := getDepartmentFromClient(Department2Name, k8sClient)
					g.Expect(dep.Labels[schedv2alpha2.DepartmentIdLabel]).To(Equal(dep2IdToCompareTo))
				}, DefaultTimeout).Should(Succeed())
				validateExistingKaiDepartmentMatchesBackendSpec(Department2Name, k8sClient)
			})
		})

		It("syncs node pools to the kai.resources NodePool CRD", func() {
			NodePoolsFeatureTurnedOff = false

			go agent.Start(stopper)

			Eventually(func(g Gomega) {
				kaiNodePool1 := getNodePoolFromClient(NodePool1Name, k8sClient)
				g.Expect(kaiNodePool1).NotTo(BeNil())
				g.Expect(kaiNodePool1.Spec.LabelKey).To(Equal(OpenApiNodepool1.LabelKey))
				g.Expect(kaiNodePool1.Spec.LabelValue).To(Equal(OpenApiNodepool1.LabelValue))
				g.Expect(kaiNodePool1.Annotations[kainodepool.AnnotationGPUNetworkAccelerationDetection]).To(Equal("Use"))
				g.Expect(kaiNodePool1.Annotations[kainodepool.AnnotationGPUNetworkAccelerationLabelKey]).To(Equal("GPU-DETECTION-LABEL"))
			}, DefaultTimeout).Should(Succeed())
		})

		It("NodePools from backend created + unnecessary nodepool deleted + updated nodepool reconciled", func() {
			// the existing kai nodepool1 does not match the backend nodepool
			Expect(getNodePoolFromClient(NodePool1Name, k8sClient)).ToNot(BeNil())
			validateExistingNodePoolDoesNotMatchBackendSpec(NodePool1Name, k8sClient)

			Expect(getNodePoolFromClient(NodePool2Name, k8sClient)).To(BeNil())
			Expect(getNodePoolFromClient(NodePool3Name, k8sClient)).To(BeNil())
			Expect(getNodePoolFromClient(nodePoolToBeDeleted.Name, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			validateNodePoolExistsEventually(NodePoolDName, k8sClient)
			validateNodePoolExistsEventually(NodePool1Name, k8sClient)
			validateNodePoolExistsEventually(NodePool2Name, k8sClient)
			validateNodePoolExistsEventually(NodePool3Name, k8sClient)
			validateNodePoolDeletedEventually(nodePoolToBeDeleted.Name, k8sClient)

			// after the reconcile; the existing kai nodepool1 should match the backend nodepool
			validateExistingNodePoolMatchesBackendSpec(NodePoolDName, k8sClient)
			validateExistingNodePoolMatchesBackendSpec(NodePool1Name, k8sClient)
			validateExistingNodePoolMatchesBackendSpec(NodePool2Name, k8sClient)
			validateExistingNodePoolMatchesBackendSpec(NodePool3Name, k8sClient)

			// update a random field in nodepool 2
			np2 := getNodePoolFromClient(NodePool2Name, k8sClient)
			labelKeyToCompareTo := np2.Spec.LabelKey
			np2.Spec.LabelKey = np2.Spec.LabelKey + "some-suffix"
			Expect(k8sClient.Update(context.Background(), np2)).To(Succeed())
			Expect(getNodePoolFromClient(NodePool2Name, k8sClient).Spec.LabelKey).ToNot(Equal(labelKeyToCompareTo))

			// expect it to be reconciled eventually...
			Eventually(func(g Gomega) {
				np := getNodePoolFromClient(NodePool2Name, k8sClient)
				g.Expect(np.Spec.LabelKey).To(Equal(labelKeyToCompareTo))
			}, DefaultTimeout).Should(Succeed())
			validateExistingNodePoolMatchesBackendSpec(NodePool2Name, k8sClient)

			// now "turn off" node pool feature
			NodePoolsFeatureTurnedOff = true

			// expect all other node pools (except the default) to be deleted
			validateNodePoolDeletedEventually(NodePool1Name, k8sClient)
			validateNodePoolDeletedEventually(NodePool2Name, k8sClient)
			validateNodePoolDeletedEventually(NodePool3Name, k8sClient)
			validateExistingNodePoolMatchesBackendSpec(NodePoolDName, k8sClient)
		})

		It("DataVolumes from backend created + unnecessary datavolume deleted + updated datavolume reconciled", func() {
			// the existing datavolume1 does not match the backend datavolume
			Expect(getDatavolumeFromClient(DataVolume1Name, k8sClient)).ToNot(BeNil())
			validateExistingDatavolumeDoesNotMatchBackendSpec(DataVolume1Name, k8sClient)

			Expect(getDatavolumeFromClient(DataVolume2Name, k8sClient)).To(BeNil())
			Expect(getDatavolumeFromClient(DataVolume3Name, k8sClient)).To(BeNil())
			Expect(getDatavolumeFromClient(DataVolume4Name, k8sClient)).To(BeNil())
			Expect(getDatavolumeFromClient(DataVolumeDeletingName, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			validateDatavolumeExistsEventually(DataVolume1Name, k8sClient)
			validateDatavolumeExistsEventually(DataVolume2Name, k8sClient)
			validateDatavolumeExistsEventually(DataVolume3Name, k8sClient)
			validateDatavolumeExistsEventually(DataVolume4Name, k8sClient)
			validateDatavolumeDeletedEventually(DataVolumeDeletingName, k8sClient)

			// after the reconcile; the existing datavolume1 should match the backend datavolume
			validateExistingDatavolumeMatchesBackendSpec(DataVolume1Name, k8sClient)
			validateExistingDatavolumeMatchesBackendSpec(DataVolume2Name, k8sClient)
			validateExistingDatavolumeMatchesBackendSpec(DataVolume3Name, k8sClient)
			validateExistingDatavolumeMatchesBackendSpec(DataVolume4Name, k8sClient)
		})

		It("AccessRule from backend created + backend-deleted accessrule deleted from cluster", func() {
			// the existing accessrule1 does not match the backend accessrule
			Expect(getAccessRuleFromClient(AccessRule1Name, k8sClient)).ToNot(BeNil())
			validateExistingAccessRuleDoesNotMatchBackendSpec(AccessRule1Name, k8sClient)

			Expect(getAccessRuleFromClient(AccessRule2Name, k8sClient)).To(BeNil())
			Expect(getAccessRuleFromClient(AccessRule3Name, k8sClient)).To(BeNil())
			Expect(getAccessRuleFromClient(AccessRule4Name, k8sClient)).To(BeNil())
			Expect(getAccessRuleFromClient(accessRuleToBeDeleted.Name, k8sClient)).ToNot(BeNil())
			Expect(getAccessRuleFromClient(AccessRule6AlreadyDeleted, k8sClient)).To(BeNil())
			Expect(getAccessRuleFromClient(AccessRule7FutureUpdateTime, k8sClient)).To(BeNil())

			go agent.Start(stopper)

			validateAccessRuleExistsEventually(AccessRule1Name, k8sClient)
			validateAccessRuleExistsEventually(AccessRule2Name, k8sClient)
			validateAccessRuleExistsEventually(AccessRule3Name, k8sClient)
			validateAccessRuleExistsEventually(AccessRule4Name, k8sClient)
			validateAccessRuleDeletedEventually(accessRuleToBeDeleted.Name, k8sClient)
			Expect(getAccessRuleFromClient(AccessRule6AlreadyDeleted, k8sClient)).To(BeNil())
			Expect(getAccessRuleFromClient(AccessRule7FutureUpdateTime, k8sClient)).To(BeNil())

			// after the reconcile; the existing accessrule1 should match the backend accessrule
			validateExistingAccessRuleMatchesBackendSpec(AccessRule1Name, k8sClient)
			validateExistingAccessRuleMatchesBackendSpec(AccessRule2Name, k8sClient)
			validateExistingAccessRuleMatchesBackendSpec(AccessRule3Name, k8sClient)
			validateExistingAccessRuleMatchesBackendSpec(AccessRule4Name, k8sClient)

			// now; edit the "UpdateAt" for the AccessRule7FutureUpdateTime - so it will return in the next agent loop
			BackendAccessRule7FutureUpdateTime.UpdatedAt = time.Now()
			validateAccessRuleExistsEventually(AccessRule7FutureUpdateTime, k8sClient)
			validateExistingAccessRuleMatchesBackendSpec(AccessRule7FutureUpdateTime, k8sClient)

			// in the real cluster - we will have a webhook preventing us from updating accessrules;
			// but here in the test - we can update an accessrule and see it NOT being reconciled -
			// as it shouldn't return from the backend anymore

			// update a random field in accessrule 2
			ar2 := getAccessRuleFromClient(AccessRule2Name, k8sClient)
			roleToCompareTo := ar2.Spec.RoleName
			ar2.Spec.RoleName = ar2.Spec.RoleName + "some-suffix"
			Expect(k8sClient.Update(context.Background(), ar2)).To(Succeed())
			Expect(getAccessRuleFromClient(AccessRule2Name, k8sClient).Spec.RoleName).ToNot(Equal(roleToCompareTo))

			Consistently(func(g Gomega) {
				ar := getAccessRuleFromClient(AccessRule2Name, k8sClient)
				g.Expect(ar.Spec.RoleName).To(Equal(roleToCompareTo))
			}, ConsistentlyTimeout).ShouldNot(Succeed())
		})

		It("Interactive policy from backend created + updated + deleted", func() {
			// policy5 - should exist before the agent starts
			Expect(getInteractivePolicyFromClient(policy5.name, policy5.namespace, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			// policy1 - should be updated with the backend policy the scope labels and updated counter
			validateIPCounterValueExistsEventually(policy1.name, policy1.namespace, k8sClient, strconv.Itoa(policy1.afterUpdateCounter))

			// policy5 - should be deleted (use Eventually: deletion may land in a later sync cycle)
			Eventually(func() *ipv1.InteractivePolicy {
				return getInteractivePolicyFromClient(policy5.name, policy5.namespace, k8sClient)
			}, DefaultTimeout).Should(BeNil())
		})

		It("Training policy from backend created + updated + deleted", func() {
			// policy5 - should exist before the agent starts
			Expect(getTrainingPolicyFromClient(policy5.name, policy5.namespace, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			// policy1 - should be updated with the backend policy spec (image) and the scope labels and updated counter
			validateTPCounterValueExistsEventually(policy1.name, policy1.namespace, k8sClient, strconv.Itoa(policy1.afterUpdateCounter))

			// policy5 - should be deleted (use Eventually: deletion may land in a later sync cycle)
			Eventually(func() *tpv1.TrainingPolicy {
				return getTrainingPolicyFromClient(policy5.name, policy5.namespace, k8sClient)
			}, DefaultTimeout).Should(BeNil())
		})

		It("pvc assets from backend", func() {
			go agent.Start(stopper)

			for _, pvcAndOptionalUpdate := range postSyncUpdatePvcs {
				validatePvcExistsEventually(pvcAndOptionalUpdate, k8sClient)
				validatePvcLabelsEventually(pvcAndOptionalUpdate, k8sClient)
				validatePvcAnnotationsEventually(pvcAndOptionalUpdate, k8sClient)
			}

			for _, pvc := range postSyncDeletedPvcs {
				validatePvcDoesNotExistsEventually(pvc, k8sClient)
			}

			// we had updates 3-6
			validateAgentConfigWithDatasourceCounter(k8sClient, "6")
		})

		It("credentials assets from backend", func() {
			go agent.Start(stopper)

			for _, secretAndOptionalUpdate := range postSyncUpdateSecrets {
				validateSecretExistsEventually(secretAndOptionalUpdate, k8sClient)
			}

			for _, secret := range postSyncDeletedSecrets {
				validateSecretDoesNotExistsEventually(secret, k8sClient)
			}

			// we had updates 9-15
			validateAgentConfigWithCredentialsCounter(k8sClient, "16")
		})

		It("rejects interactive workloads from the legacy backend", func() {
			go agent.Start(stopper)

			validateAgentConfigWithWorkloadCounters(k8sClient, configmaps.InteractiveWorkloadsSyncCounter, fmt.Sprintf("%d", *BackendNonDistributedWorkload3.UpdateCount))
			Expect(getInteractiveWorkloadFromClient(ClusterInteractiveWorkload1.Name, ClusterInteractiveWorkload1.Namespace, k8sClient)).To(BeNil())
			Expect(getInteractiveWorkloadFromClient(ClusterInteractiveWorkload2.Name, ClusterInteractiveWorkload2.Namespace, k8sClient)).To(BeNil())
			Expect(getInteractiveWorkloadFromClient(ClusterInteractiveWorkload3.Name, ClusterInteractiveWorkload3.Namespace, k8sClient)).To(BeNil())
		})

		It("Training policy from backend created + updated + deleted", func() {
			// policy6 - should exist before the agent starts
			Expect(getDistributedPolicyFromClient(policy6.name, policy6.namespace, k8sClient)).ToNot(BeNil())
			Expect(getDistributedPolicyFromClient(policy6.name, policy6.namespace, k8sClient).Annotations[clusterCommonConstants.AnnotationsBackendPolicyCounter]).To(Equal(strconv.Itoa(policy6.preUpdatedCounter)))
			Expect(getDistributedPolicyFromClient(policy6.name, policy6.namespace, k8sClient).Labels[clusterCommonConstants.PolicyLabel]).To(Equal("true"))

			go agent.Start(stopper)

			// policy6 - should be deleted
			validateDPDeletedEventually(policy6.name, policy6.namespace, k8sClient)
		})

		It("rejects training workloads from the legacy backend", func() {
			go agent.Start(stopper)

			validateAgentConfigWithWorkloadCounters(k8sClient, configmaps.TrainingWorkloadsSyncCounter, fmt.Sprintf("%d", *BackendNonDistributedWorkload3.UpdateCount))
			Expect(getTrainingWorkloadFromClient(ClusterTrainingWorkload1.Name, ClusterTrainingWorkload1.Namespace, k8sClient)).To(BeNil())
			Expect(getTrainingWorkloadFromClient(ClusterTrainingWorkload2.Name, ClusterTrainingWorkload2.Namespace, k8sClient)).To(BeNil())
			Expect(getTrainingWorkloadFromClient(ClusterTrainingWorkload3.Name, ClusterTrainingWorkload3.Namespace, k8sClient)).To(BeNil())
		})

		It("rejects distributed workloads from the legacy backend", func() {
			go agent.Start(stopper)

			validateAgentConfigWithWorkloadCounters(k8sClient, configmaps.DistributedWorkloadsSyncCounter, fmt.Sprintf("%d", *BackendDistributedWorkload3.UpdateCount))
			Expect(getDistributedWorkloadFromClient(ClusterDistributedWorkload1.Name, ClusterDistributedWorkload1.Namespace, k8sClient)).To(BeNil())
			Expect(getDistributedWorkloadFromClient(ClusterDistributedWorkload2.Name, ClusterDistributedWorkload2.Namespace, k8sClient)).To(BeNil())
			Expect(getDistributedWorkloadFromClient(ClusterDistributedWorkload3.Name, ClusterDistributedWorkload3.Namespace, k8sClient)).To(BeNil())
		})

		It("Inference policy from backend created + updated + deleted", func() {
			// policy10 - should exist before the agent starts
			Expect(getInferencePolicyFromClient(policy10.name, policy10.namespace, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			// policy9 - should be updated with the backend policy spec (image) and the scope labels and updated counter
			validateINFPCounterValueExistsEventually(policy9.name, policy9.namespace, k8sClient, strconv.Itoa(policy9.afterUpdateCounter))

			// policy10 - should be deleted
			validateINFPDeletedEventually(policy10.name, policy10.namespace, k8sClient)
		})

		It("rejects inference workloads from the legacy backend", func() {
			go agent.Start(stopper)

			validateAgentConfigWithWorkloadCounters(k8sClient, configmaps.InferenceWorkloadsSyncCounter, fmt.Sprintf("%d", *BackendInferenceWorkload3.UpdateCount))
			Expect(getInferenceWorkloadFromClient(ClusterInferenceWorkload1.Name, ClusterInferenceWorkload1.Namespace, k8sClient)).To(BeNil())
			Expect(getInferenceWorkloadFromClient(ClusterInferenceWorkload2.Name, ClusterInferenceWorkload2.Namespace, k8sClient)).To(BeNil())
			Expect(getInferenceWorkloadFromClient(ClusterInferenceWorkload3.Name, ClusterInferenceWorkload3.Namespace, k8sClient)).To(BeNil())
		})

		It("Workloads v2 from backend", func() {
			go agent.Start(stopper)

			validateAgentConfigWithWorkloadV2Counters(k8sClient, fmt.Sprintf("%d", lastWorkloadV2Update.UpdateCount))

			for _, workloadV2 := range workloadsV2WithUpdates {
				gvk := workloadV2.Workload.GetObjectKind().GroupVersionKind()

				if workloadV2.Update == nil {
					// resources that were not updated are expected to remain in the cluster
					validateWorkloadV2ExistsEventually(workloadV2.Workload.GetName(), workloadV2.Workload.GetNamespace(), gvk, k8sClient)
				} else {
					if workloadV2.Update.DesiredPhase == workloadsMgrApi.DESIREDPHASE_DELETED {
						validateWorkloadV2DoesNotExistEventually(workloadV2.Workload.GetName(), workloadV2.Workload.GetNamespace(), gvk, k8sClient)
					} else {
						validateWorkloadV2ExistsEventually(workloadV2.Workload.GetName(), workloadV2.Workload.GetNamespace(), gvk, k8sClient)
						validateWorkloadV2CounterExistsEventually(workloadV2.Workload.GetName(), workloadV2.Workload.GetNamespace(), gvk, workloadV2.Update.UpdateCount, k8sClient)
					}
				}
			}
		})

		It("test chunkSize = 0 (backend response size < 4)", func() {
			// create agent config map with init counter 14 to make chunkSize = 0 (backend response size < 4)
			_, err := configmaps.CreateAgentConfigIfNotExists(k8sClient, k8sClient)
			Expect(err).To(BeNil())
			err = configmaps.UpdateTrainingPoliciesSyncCounters(k8sClient, k8sClient, 14)
			Expect(err).To(BeNil())

			go agent.Start(stopper)

			// policy1 - should not be updated (policy1 counter was 11 and we start from 14 so it should not be updated)
			Expect(getTrainingPolicyFromClient(policy1.name, policy1.namespace, k8sClient).Annotations[clusterCommonConstants.AnnotationsBackendPolicyCounter]).To(Equal(strconv.Itoa(policy1.preUpdatedCounter)))

			// policy5 - should be deleted (policy5 counter was 15 and we start from 14 so it should be deleted)
			validateTPDeletedEventually(policy5.name, policy5.namespace, k8sClient)
		})

		It("test agent config is created ", func() {
			// make sure agent config does not exist before the agent starts
			Expect(getAgentConfigFromClient(k8sClient)).To(BeNil())

			go agent.Start(stopper)

			// agent config should be created
			validateAgentConfigExistEventually(k8sClient)
		})

		It("Workloads config from backend", func() {
			go agent.Start(stopper)
			validateWorkloadsConfigMap(k8sClient)
		})

		It("Workloads priorities", func() {
			go agent.Start(stopper)
			validateWorkloadsPriorities(priorityClassAddedName, 10, priorityClassAddedId, clusterCommonConstants.PriorityClassRunaiSource, k8sClient)
			validateWorkloadsPriorities(priorityClassUpdatedName, 20, priorityClassUpdatedId, clusterCommonConstants.PriorityClassRunaiSource, k8sClient)
			validateWorkloadsPriorities(priorityClassNonRunaiName, priorityOriginalValue, "", "", k8sClient)
			validateWorkloadsPrioritiesNotExisting(priorityClassDeletedName, k8sClient)
		})

		It("Network topologies from backend created + unnecessary network topology deleted", func() {
			// the existing network topology1 should already exist and match backend spec
			Expect(getNetworkTopologyFromClient(NetworkTopology1Name, k8sClient)).ToNot(BeNil())

			Expect(getNetworkTopologyFromClient(NetworkTopology2Name, k8sClient)).To(BeNil())
			Expect(getNetworkTopologyFromClient(NetworkTopology3Name, k8sClient)).To(BeNil())
			Expect(getNetworkTopologyFromClient(NetworkTopology4Name, k8sClient)).To(BeNil())
			Expect(getNetworkTopologyFromClient(NetworkTopologyDeletingName, k8sClient)).ToNot(BeNil())

			go agent.Start(stopper)

			validateNetworkTopologyExistsEventually(NetworkTopology1Name, k8sClient)
			validateNetworkTopologyExistsEventually(NetworkTopology2Name, k8sClient)
			validateNetworkTopologyExistsEventually(NetworkTopology3Name, k8sClient)
			validateNetworkTopologyExistsEventually(NetworkTopology4Name, k8sClient)
			validateNetworkTopologyExistsEventually(NetworkTopologyCluster1Name, k8sClient)
			validateNetworkTopologyExistsEventually(NetworkTopologyCluster2Name, k8sClient)
			validateNetworkTopologyDeletedEventually(NetworkTopologyDeletingName, k8sClient)

			// validate that all network topologies exist and match backend specifications
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopology1Name, adapters.NetworkTopologyRunAiSource, k8sClient)
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopology2Name, adapters.NetworkTopologyRunAiSource, k8sClient)
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopology3Name, adapters.NetworkTopologyRunAiSource, k8sClient)
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopology4Name, adapters.NetworkTopologyRunAiSource, k8sClient)

			// validate that cluster network topologies exist and match backend specifications
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopologyCluster1Name, ClusterSource, k8sClient)
			validateExistingNetworkTopologyMatchesBackendSpec(NetworkTopologyCluster2Name, ClusterSource, k8sClient)
		})

		It("Kartas from workload types are created and deleted as expected (no updates)", func() {
			expected1 := buildExpectedKartaWithLabels(existingClusterKarta1, backendWorkloadTypeConfigs[0].Id, backendWorkloadTypeConfigs[0].Kartas[0].UpdatedAt)
			expected2 := buildExpectedKartaWithLabels(existingClusterKarta2, backendWorkloadTypeConfigs[1].Id, backendWorkloadTypeConfigs[1].Kartas[0].UpdatedAt)
			// the existing has the name set, but the expected should not have it
			expectedClusterKarta3 := existingClusterKarta3.DeepCopy()
			expectedClusterKarta3.Spec.StructureDefinition.RootComponent.Name = ""
			expected3 := buildExpectedKartaWithLabels(expectedClusterKarta3, backendWorkloadTypeConfigs[2].Id, backendWorkloadTypeConfigs[2].Kartas[0].UpdatedAt)

			// Validate initial state
			Expect(getKartaFromClient(existingClusterKarta2.Name, k8sClient)).To(BeNil())
			Expect(getKartaFromClient(existingKartaToBeDeleted.Name, k8sClient)).ToNot(BeNil())
			// labels are not updated yet - should not be up to date
			preExisting1 := getKartaFromClient(existingClusterKarta1.Name, k8sClient)
			Expect(preExisting1).ToNot(BeNil())
			Expect(validateKartaIsUpToDate(expected1, preExisting1)).To(BeFalse())
			// spec is not updated yet - should not be up to date
			preExisting3 := getKartaFromClient(existingClusterKarta3.Name, k8sClient)
			Expect(preExisting3).ToNot(BeNil())
			Expect(validateKartaIsUpToDate(expected3, preExisting3)).To(BeFalse())

			go agent.Start(stopper)

			validateKartaExistsEventually(existingClusterKarta2.Name, k8sClient)
			validateKartaExistsEventually(existingClusterKarta1.Name, k8sClient)
			validateKartaDeletedEventually(existingKartaToBeDeleted.Name, k8sClient)
			// 3 - should be re-created since the spec is different
			validateKartaExistsEventually(existingClusterKarta3.Name, k8sClient)

			// validate that Kartas in the cluster exist and match expected - with correct labels
			current1 := getKartaFromClient(existingClusterKarta1.Name, k8sClient)
			Expect(validateKartaIsUpToDate(expected1, current1)).To(BeTrue())
			Expect(current1.CreationTimestamp).To(Equal(preExisting1.CreationTimestamp), "should be the same Karta resource - only updated, not deleted and re-created")

			Expect(validateKartaIsUpToDate(expected2, getKartaFromClient(existingClusterKarta2.Name, k8sClient))).To(BeTrue())

			current3 := getKartaFromClient(existingClusterKarta3.Name, k8sClient)
			Expect(validateKartaIsUpToDate(expected3, current3)).To(BeTrue())
			Expect(current3.CreationTimestamp).ToNot(Equal(preExisting3.CreationTimestamp), "should NOT be the same Karta resource - deleted and re-created")
		})
	})
})
