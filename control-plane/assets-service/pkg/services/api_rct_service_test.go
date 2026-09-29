/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const urlRct = "/api/v1/k8s-objects/resource-claim-templates"

// validRctObjectSpec returns a minimal but valid resource.k8s.io/v1
// ResourceClaimSpec payload used by every RCT test factory create call.
// Keep this minimal on purpose: the framework only exercises CRUD/RBAC/
// tenancy plumbing — spec-shape validation lives in validateRctObjectSpec
// and is tested separately.
func validRctObjectSpec() map[string]any {
	return map[string]any{
		"devices": map[string]any{
			"requests": []any{
				map[string]any{
					"name": "gpu",
					"exactly": map[string]any{
						"deviceClassName": "gpu.nvidia.com",
						"count":           1,
						"allocationMode":  "ExactCount",
					},
				},
			},
		},
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type RctTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewRctTestFactory(router chi.Router) *RctTestFactory {
	return &RctTestFactory{router: router}
}

func (c *RctTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.RctListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *RctTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.RctCreationRequest{
		Meta: creationMeta,
		Spec: api.RctSpec{
			ObjectSpec: validRctObjectSpec(),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *RctTestFactory) marshalUpdateReq([]byte) string {
	request := api.RctUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: api.RctSpec{
			ObjectSpec: validRctObjectSpec(),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

// createSystemBasedAsset — RCT does not support system scope. Return empty;
// the framework's canRunOnScope + allowedScopes gate keeps the caller from
// invoking this arm.
func (c *RctTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, _ *http.Header) (api.AssetMeta, []byte, []byte) {
	return api.AssetMeta{}, nil, nil
}

func (c *RctTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createRctAsset(buildRandomTenantRctAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *RctTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createRctAsset(buildRandomClusterRctAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *RctTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createRctAsset(buildRandomProjectRctAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *RctTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createRctAsset(buildRandomDepartmentRctAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *RctTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.RctSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

// addToWorkloadCreationReq — RCT is not yet wired into workload creation.
// Returning false makes the framework skip workload-integration sub-tests
// (same pattern the credential kinds use).
func (c *RctTestFactory) addToWorkloadCreationReq(_ *api.WorkloadCreationRequest, _ *api.AssetMeta) bool {
	return false
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomRctHelper() api.RctCreationRequest {
	return api.RctCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.RctSpec{
			ObjectSpec: validRctObjectSpec(),
		},
	}
}

func buildRandomTenantRctAsset() api.RctCreationRequest {
	ans := buildRandomRctHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterRctAsset(randoms *testutils.RandomValues) api.RctCreationRequest {
	ans := buildRandomRctHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomProjectRctAsset(randoms *testutils.RandomValues) api.RctCreationRequest {
	ans := buildRandomRctHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentRctAsset(randoms *testutils.RandomValues) api.RctCreationRequest {
	ans := buildRandomRctHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func createRctAsset(creation api.RctCreationRequest, router chi.Router, header *http.Header) (api.Rct, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlRct, makePayload(creation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.Rct
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

// RCT-specific tests. Cross-cutting CRUD/RBAC/tenancy/list tests run via the
// DescribeTable in NewAssetKindTable. This suite covers the RCT-only surfaces
// that the generic framework doesn't touch:
//
//   - Two-layer ObjectSpec validation (santhosh-tekuri schema + typed decode
//     with DisallowUnknownFields), exercising each layer separately.
//   - Kubernetes cluster capability gate (MinVersionResourceClaimTemplates =
//     "3.1.0"), which only fires when kind == RCT.
var _ = Describe("Rct API tests", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()
		randoms = testutils.NewRandomValues()

		authorization.SetMock(testing.NewRandomsBasedAuthMockClient(randoms))
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})
	})

	// postRctWithObjectSpec builds and sends a minimal cluster-scoped RCT
	// creation request with the supplied ObjectSpec. Cluster scope is chosen
	// deliberately: it avoids the org_unit ancestor resolution that project /
	// department scope needs, so any failure comes from the surface actually
	// under test (ObjectSpec validation or cluster capability), not setup.
	postRctWithObjectSpec := func(spec map[string]any) *httptest.ResponseRecorder {
		req := api.RctCreationRequest{
			Meta: api.AssetCreationRequest{
				Name:      testutils.RandString(5),
				Scope:     api.SCOPE_CLUSTER,
				ClusterId: convutils.Ref(randoms.ClusterId),
			},
			Spec: api.RctSpec{ObjectSpec: spec},
		}
		payload, err := json.Marshal(req)
		Expect(err).To(BeNil())
		httpReq, err := http.NewRequest("POST", urlRct, bytes.NewReader(payload))
		Expect(err).To(BeNil())
		httpReq.Header = *headerForAdmin(randoms)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httpReq)
		return recorder
	}

	Context("ObjectSpec validation", func() {
		It("accepts a minimal valid ResourceClaimSpec", func() {
			resp := postRctWithObjectSpec(validRctObjectSpec())
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})

		It("rejects an empty objectSpec (short-circuit before schema)", func() {
			// Can't build this via the Go struct path — omitempty on
			// `ObjectSpec map[string]interface{}` strips empty maps at marshal
			// time, so the request arrives with a zero-valued Spec and the
			// generated AssertRctCreationRequestRequired wrapper rejects it as
			// "spec required" before validateRctObjectSpec runs. Construct
			// the JSON explicitly so ObjectSpec is present-but-empty.
			body := fmt.Sprintf(
				`{"meta":{"name":%q,"scope":"cluster","clusterId":%q},"spec":{"objectSpec":{}}}`,
				testutils.RandString(5), randoms.ClusterId,
			)
			httpReq, err := http.NewRequest("POST", urlRct, strings.NewReader(body))
			Expect(err).To(BeNil())
			httpReq.Header = *headerForAdmin(randoms)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("spec.objectSpec is required"))
		})

		It("rejects a top-level typo like 'devics' via typed-decode layer", func() {
			// K8s OpenAPI schemas are permissive (no additionalProperties: false),
			// so layer 1 (santhosh-tekuri) accepts unknown fields. Layer 2
			// (json.DisallowUnknownFields against resourcev1.ResourceClaimSpec)
			// is what catches this — the regression that motivated adding the
			// second layer in the first place.
			resp := postRctWithObjectSpec(map[string]any{
				"devics":  map[string]any{},
				"devices": map[string]any{"requests": []any{}},
			})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			// The response body is JSON-encoded, so the inner double-quotes
			// around the field name arrive as `\"...\"`. Backticks let us
			// match that verbatim without Go-string escaping the backslashes.
			Expect(resp.Body.String()).To(ContainSubstring(`unknown field \"devics\"`))
		})

		It("rejects an unknown nested field inside devices.requests[] via typed-decode layer", func() {
			resp := postRctWithObjectSpec(map[string]any{
				"devices": map[string]any{
					"requests": []any{
						map[string]any{
							"name":                "gpu",
							"exactly":             map[string]any{"deviceClassName": "gpu.nvidia.com", "count": 1},
							"unexpectedNestedKey": "oops",
						},
					},
				},
			})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			// See note above on why the backticks + escaped quotes.
			Expect(resp.Body.String()).To(ContainSubstring(`unknown field \"unexpectedNestedKey\"`))
		})

		It("rejects a shape mismatch (devices.requests as string) via schema layer", func() {
			// Layer 1 fires before layer 2 for structural type violations
			// because typed decode would fail with a Go json unmarshal error
			// that's harder to read; the schema error names the offending
			// path directly.
			resp := postRctWithObjectSpec(map[string]any{
				"devices": map[string]any{
					"requests": "not-an-array",
				},
			})
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("resource.k8s.io/v1 ResourceClaimSpec"))
		})
	})

	// postRctToClusterVersion re-installs the cluster mock at the given
	// cluster version and issues a valid RCT creation POST. Everything else
	// (ObjectSpec, scope, auth) is held constant so any non-202 comes from
	// the capability check.
	postRctToClusterVersion := func(clusterVersion string) *httptest.ResponseRecorder {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, clusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		return postRctWithObjectSpec(validRctObjectSpec())
	}

	Context("cluster capability check (MinVersionResourceClaimTemplates = 3.1.0)", func() {
		It("accepts a cluster at exactly the minimum supported version", func() {
			resp := postRctToClusterVersion("3.1.0")
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})

		It("rejects a cluster one patch below the minimum supported version", func() {
			resp := postRctToClusterVersion("3.0.9")
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("not supported for this cluster version"))
		})

		It("rejects a cluster far below (2.15.0, the OldClustersPreK8sObjectsTracker mock)", func() {
			// Matches OldClustersPreK8sObjectsTracker — the same version the
			// shared cannotDeleteAssetBeingUsed helper uses and why we set
			// minClusterVersion: "3.1.0" on the RCT entry.
			resp := postRctToClusterVersion("2.15.0")
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("not supported for this cluster version"))
		})
	})
})
