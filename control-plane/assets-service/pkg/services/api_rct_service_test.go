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
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
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

	// buildClusterWithKubeVersion returns a DisplayedCluster at the requested
	// cluster version with Status.Platform present. If kubeVer is non-nil,
	// Platform.KubeVersion is populated; if nil, Platform is present but
	// KubeVersion is left unset (exercises the "cluster did not report a
	// kubernetes version" 503 branch that BuildDisplayedCluster's default
	// kube 1.34.0 always satisfies).
	buildClusterWithKubeVersion := func(clusterVer string, kubeVer *string) openapi.DisplayedCluster {
		dc := openapi.DisplayedCluster{Uuid: randoms.ClusterId, TenantId: int32(randoms.TenantId)}
		dc.Version.Set(&clusterVer)
		config := openapi.NewClusterReportedConfig()
		sub := false
		config.SubdomainEnabled.Set(&sub)
		status := openapi.NewClusterDisplayedStatus()
		status.SetConfig(*config)
		platform := openapi.NewClusterReportedStatusPlatform()
		if kubeVer != nil {
			platform.KubeVersion.Set(kubeVer)
		}
		status.SetPlatform(*platform)
		dc.Status.Set(status)
		return dc
	}

	// postRctToClusterWithKubeVersion installs a cluster mock at the given
	// cluster version + kube version (nil kubeVer = Platform present but
	// KubeVersion unset) and issues a valid RCT creation POST.
	postRctToClusterWithKubeVersion := func(clusterVer string, kubeVer *string) *httptest.ResponseRecorder {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, buildClusterWithKubeVersion(clusterVer, kubeVer)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		return postRctWithObjectSpec(validRctObjectSpec())
	}

	// postRctWithObjectMeta issues a cluster-scoped RCT creation POST with
	// the given ObjectMeta and a valid ObjectSpec. Used by the ObjectMeta
	// validation tests.
	postRctWithObjectMeta := func(meta *api.K8sObjectMeta) *httptest.ResponseRecorder {
		req := api.RctCreationRequest{
			Meta: api.AssetCreationRequest{
				Name:      testutils.RandString(5),
				Scope:     api.SCOPE_CLUSTER,
				ClusterId: convutils.Ref(randoms.ClusterId),
			},
			Spec: api.RctSpec{
				ObjectMeta: meta,
				ObjectSpec: validRctObjectSpec(),
			},
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

	Context("kubernetes version check (MinDraK8sVersion = 1.34.0)", func() {
		It("rejects a cluster reporting kube below the DRA floor (1.33.0)", func() {
			resp := postRctToClusterWithKubeVersion("3.1.0", convutils.Ref("1.33.0"))
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			// JSON response bodies HTML-escape `>` as `\u003e`; the backtick
			// raw string matches the escaped form verbatim.
			Expect(resp.Body.String()).To(ContainSubstring(`require kubernetes \u003e= 1.34.0`))
		})

		It("returns 503 when the cluster reports Platform but no KubeVersion", func() {
			resp := postRctToClusterWithKubeVersion("3.1.0", nil)
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable))
			Expect(resp.Body.String()).To(ContainSubstring("did not report a kubernetes version"))
		})

		It("returns 503 when the cluster reports an unparseable kube version", func() {
			resp := postRctToClusterWithKubeVersion("3.1.0", convutils.Ref("not-a-version"))
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable))
			Expect(resp.Body.String()).To(ContainSubstring("unrecognized kubernetes version"))
		})
	})

	Context("cross-kind id guard (Update/Delete against a non-RCT asset id)", func() {
		// Uses HostPath as the "wrong kind" — any non-RCT project-scoped asset
		// would work. The test proves that dao.GetAssetMetaById (which is
		// kind-agnostic) doesn't let a caller drive foreign-kind data through
		// the RCT service by targeting the RCT URL with a non-RCT id.
		It("returns 404 on PUT when the id belongs to a different asset kind", func() {
			hostPath, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

			updateReq := api.RctUpdateRequest{
				Meta: api.AssetUpdateRequest{Name: testutils.RandString(5)},
				Spec: api.RctSpec{ObjectSpec: validRctObjectSpec()},
			}
			payload, err := json.Marshal(updateReq)
			Expect(err).To(BeNil())
			httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlRct, hostPath.Meta.Id), bytes.NewReader(payload))
			Expect(err).To(BeNil())
			httpReq.Header = *headerForAdmin(randoms)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)

			Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("resource claim template with id"))
			Expect(recorder.Body.String()).To(ContainSubstring("not found"))
		})

		It("returns 404 on DELETE when the id belongs to a different asset kind", func() {
			hostPath, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

			httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlRct, hostPath.Meta.Id), nil)
			Expect(err).To(BeNil())
			httpReq.Header = *headerForAdmin(randoms)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)

			Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("resource claim template with id"))
			Expect(recorder.Body.String()).To(ContainSubstring("not found"))
		})
	})

	Context("ObjectMeta validation (K8s label/annotation syntax)", func() {
		It("accepts a valid label + annotation payload", func() {
			meta := &api.K8sObjectMeta{
				Labels:      &map[string]string{"app": "my-app", "example.com/tier": "gpu"},
				Annotations: &map[string]string{"example.com/owner": "team-a"},
			}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})

		It("accepts an empty label value (K8s spec allows it)", func() {
			meta := &api.K8sObjectMeta{Labels: &map[string]string{"app": ""}}
			Expect(postRctWithObjectMeta(meta).Code).To(Equal(http.StatusAccepted))
		})

		It("rejects an invalid label key character", func() {
			meta := &api.K8sObjectMeta{Labels: &map[string]string{"my@label": "v"}}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("spec.objectMeta.labels"))
			Expect(resp.Body.String()).To(ContainSubstring("valid Kubernetes qualified name"))
		})

		It("rejects a label key name longer than 63 characters", func() {
			meta := &api.K8sObjectMeta{Labels: &map[string]string{strings.Repeat("a", 64): "v"}}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("spec.objectMeta.labels"))
		})

		It("rejects a label key prefix longer than 253 characters", func() {
			meta := &api.K8sObjectMeta{Labels: &map[string]string{strings.Repeat("a", 254) + "/name": "v"}}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("prefix component"))
		})

		It("rejects an invalid label value character", func() {
			meta := &api.K8sObjectMeta{Labels: &map[string]string{"app": "bad@value"}}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("valid Kubernetes label value"))
		})

		It("rejects an invalid annotation key character", func() {
			meta := &api.K8sObjectMeta{Annotations: &map[string]string{"my@ann": "v"}}
			resp := postRctWithObjectMeta(meta)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.String()).To(ContainSubstring("spec.objectMeta.annotations"))
		})

		It("enforces the same label rules on the update path", func() {
			// Create a valid RCT first, then PUT with a bad label.
			created, _, _ := createRctAsset(buildRandomClusterRctAsset(randoms), router, headerForAdmin(randoms))

			badUpdate := api.RctUpdateRequest{
				Meta: api.AssetUpdateRequest{Name: testutils.RandString(5)},
				Spec: api.RctSpec{
					ObjectMeta: &api.K8sObjectMeta{Labels: &map[string]string{"bad@key": "v"}},
					ObjectSpec: validRctObjectSpec(),
				},
			}
			payload, err := json.Marshal(badUpdate)
			Expect(err).To(BeNil())
			httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlRct, created.Meta.Id), bytes.NewReader(payload))
			Expect(err).To(BeNil())
			httpReq.Header = *headerForAdmin(randoms)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("spec.objectMeta.labels"))
		})
	})
})
