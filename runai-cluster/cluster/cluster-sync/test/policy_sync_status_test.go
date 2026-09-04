/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	policyswagger "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-policy-service-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	distributedpolicy "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	interactivepolicy "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	resourceUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"net/http"
	"time"
)

var _ = Describe("PolicyHandler", func() {
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var syncPathWorkspace = "/api/v2/policy/workspaces/sync/status"
	var syncPathDistributed = "/api/v2/policy/distributed/sync/status"
	var putRestRequestsCh chan policyswagger.SyncStatusUpdate
	var policyHandler *handlers.PolicyHandler
	var policyName = "policyName"
	var updateCounter = int64(12)

	BeforeEach(func() {
		GinkgoWriter.Println("Creating server for test...")
		server = CreateBasicRunaiBackend()
		server.AllowUnhandledRequests = true
		DeferCleanup(func() {
			server.Close()
		})

		putRestRequestsCh = make(chan policyswagger.SyncStatusUpdate, 1)
		DeferCleanup(func() { close(putRestRequestsCh) })
	})

	Context("Verifying policy sync status is accurate, by its CRD", func() {
		BeforeEach(func() {
			server.RouteToHandler("PUT", syncPathWorkspace, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in workspace policySyncStatus PUT Mock")
				var body policyswagger.SyncStatusUpdate
				decoder := json.NewDecoder(request.Body)
				err := decoder.Decode(&body)
				Expect(err).To(BeNil())
				putRestRequestsCh <- body
			})
			server.RouteToHandler("PUT", syncPathDistributed, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in distributed policySyncStatus PUT Mock")
				var body policyswagger.SyncStatusUpdate
				decoder := json.NewDecoder(request.Body)
				err := decoder.Decode(&body)
				Expect(err).To(BeNil())
				putRestRequestsCh <- body
			})
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme())
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			policyHandler = handlers.NewPolicyHandler(backendClient, fakeCachedClient)
		})

		It("Test valid policy with correct sync status without labels", func() {
			validPolicyWithSyncStatus := interactivepolicy.InteractivePolicy{
				ObjectMeta: resourceUtils.NewRunaiObject("minMaxOk"),
				Spec: interactivepolicy.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: intFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
				Status: common.PolicyStatus{
					SyncStatus: common.PolicySyncStatus{
						LastSyncTimestamp:         metav1.NewTime(time.Now()),
						LastSyncRevision:          12,
						LastAttemptedSyncRevision: 12,
					},
				},
			}

			err := policyHandler.HandleAddUpdate(&validPolicyWithSyncStatus, constants.AddOperationString)
			if err != nil {
				Expect(err.Error()).To(Equal("unable to identify relevant scope"))
			}
		})

		It("Test valid policy with correct sync status with tenant label", func() {
			validPolicyWithSyncStatus := interactivepolicy.InteractivePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: clusterCommonConstants.NamespaceRunAi,
					Labels: map[string]string{
						clusterCommonConstants.TenantScopeLabel: "true",
					},
				},
				Spec: interactivepolicy.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: intFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
				Status: common.PolicyStatus{
					SyncStatus: common.PolicySyncStatus{
						LastSyncTimestamp:         metav1.NewTime(time.Now()),
						LastSyncRevision:          updateCounter,
						LastAttemptedSyncRevision: updateCounter,
					},
				},
			}

			err := policyHandler.HandleAddUpdate(&validPolicyWithSyncStatus, constants.AddOperationString)
			Expect(err).To(BeNil())

			requestBodyMessage := ReadFromChannel(putRestRequestsCh, 1)
			updateRequest := requestBodyMessage[0]
			Expect(*updateRequest.UpdateCounter).To(Equal(updateCounter))
			Expect(*updateRequest.Action).To(Equal(constants.ApplyAction))
			Expect(*updateRequest.ScopeType).To(Equal(policyswagger.SCOPETYPE_TENANT))
		})

		It("Test valid policy with errored sync status & delete action", func() {
			syncErrorMsg := "Error updating interactive policy"
			scopeId := "1"
			validPolicyWithErroredSyncStatus := interactivepolicy.InteractivePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: clusterCommonConstants.NamespaceRunAi,
					Labels: map[string]string{
						clusterCommonConstants.ProjectScopeLabel: scopeId,
					},
				},
				Spec: interactivepolicy.InteractiveSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: intFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
				Status: common.PolicyStatus{
					SyncStatus: common.PolicySyncStatus{
						LastSyncTimestamp:         metav1.NewTime(time.Now()),
						LastSyncRevision:          updateCounter,
						LastAttemptedSyncRevision: updateCounter + 1,
						SyncErrorMsg:              syncErrorMsg,
					},
				},
			}
			err := policyHandler.HandleAddUpdate(&validPolicyWithErroredSyncStatus, constants.DeleteOperationString)
			Expect(err).To(BeNil())

			requestBodyMessage := ReadFromChannel(putRestRequestsCh, 1)
			updateRequest := requestBodyMessage[0]
			Expect(*updateRequest.UpdateCounter).To(Equal(updateCounter))
			Expect(*updateRequest.Action).To(Equal(constants.DeleteOperationString))
			Expect(*updateRequest.ScopeType).To(Equal(policyswagger.SCOPETYPE_PROJECT))
			Expect(*updateRequest.ScopeId.Get()).To(Equal(scopeId))
			Expect(*updateRequest.ErrorMessage.Get()).To(Equal(syncErrorMsg))

		})

		It("Test valid policy with correct sync status with tenant label - distributed", func() {
			validPolicyWithSyncStatus := distributedpolicy.DistributedPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: clusterCommonConstants.NamespaceRunAi,
					Labels: map[string]string{
						clusterCommonConstants.TenantScopeLabel: "true",
					},
				},
				Spec: distributedpolicy.DistributedSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: intFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
				Status: common.PolicyStatus{
					SyncStatus: common.PolicySyncStatus{
						LastSyncTimestamp:         metav1.NewTime(time.Now()),
						LastSyncRevision:          updateCounter,
						LastAttemptedSyncRevision: updateCounter,
					},
				},
			}

			err := policyHandler.HandleAddUpdate(&validPolicyWithSyncStatus, constants.AddOperationString)
			Expect(err).To(BeNil())

			requestBodyMessage := ReadFromChannel(putRestRequestsCh, 1)
			updateRequest := requestBodyMessage[0]
			Expect(*updateRequest.UpdateCounter).To(Equal(updateCounter))
			Expect(*updateRequest.Action).To(Equal(constants.ApplyAction))
			Expect(*updateRequest.ScopeType).To(Equal(policyswagger.SCOPETYPE_TENANT))
		})

		It("Test valid policy with errored sync status & delete action - distributed", func() {
			syncErrorMsg := "Error updating distributed policy"
			scopeId := "1"
			validPolicyWithErroredSyncStatus := distributedpolicy.DistributedPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: clusterCommonConstants.NamespaceRunAi,
					Labels: map[string]string{
						clusterCommonConstants.ProjectScopeLabel: scopeId,
					},
				},
				Spec: distributedpolicy.DistributedSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							RunAsUid: intFieldMinMax(ptr.To(4), ptr.To(6)),
						},
					},
				},
				Status: common.PolicyStatus{
					SyncStatus: common.PolicySyncStatus{
						LastSyncTimestamp:         metav1.NewTime(time.Now()),
						LastSyncRevision:          updateCounter,
						LastAttemptedSyncRevision: updateCounter + 1,
						SyncErrorMsg:              syncErrorMsg,
					},
				},
			}
			err := policyHandler.HandleAddUpdate(&validPolicyWithErroredSyncStatus, constants.DeleteOperationString)
			Expect(err).To(BeNil())

			requestBodyMessage := ReadFromChannel(putRestRequestsCh, 1)
			updateRequest := requestBodyMessage[0]
			Expect(*updateRequest.UpdateCounter).To(Equal(updateCounter))
			Expect(*updateRequest.Action).To(Equal(constants.DeleteOperationString))
			Expect(*updateRequest.ScopeType).To(Equal(policyswagger.SCOPETYPE_PROJECT))
			Expect(*updateRequest.ScopeId.Get()).To(Equal(scopeId))
			Expect(*updateRequest.ErrorMessage.Get()).To(Equal(syncErrorMsg))
		})
	})
})

func intFieldMinMax(min *int, max *int) *fields.IntField {
	return &fields.IntField{
		Rules: &fields.IntRules{
			Min: min,
			Max: max,
		},
	}
}
