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
	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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

// rctDeviceClassForTests is the single DeviceClass name every test spec built
// via validRctObjectSpec references. Shared with the k8s-objects-tracker mock
// setup in each BeforeEach (services_crud_test.go, services_auth_test.go, this
// file) so RCT creation passes the new validateRctDeviceClasses gate without
// each site duplicating the string literal.
const rctDeviceClassForTests = "gpu.nvidia.com"

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
						"deviceClassName": rctDeviceClassForTests,
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

		// Seed the DeviceClass referenced by validRctObjectSpec. Every test
		// in this suite POSTs through postRctWithObjectSpec / the factory,
		// which both consult the tracker as part of validateRctDeviceClasses.
		// Tests that specifically exercise the validator (missing names,
		// mixed Exactly/FirstAvailable, tracker outage) override this mock
		// locally so this seed is a safe non-interfering default.
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
			WithDeviceClass(randoms.ClusterId, rctDeviceClassForTests))
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

	// ObjectSpec validation is unit-tested directly in
	// k8s_object_validation_test.go against the same rctObjectSpecSchema.

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

	// specWithDeviceClasses builds an ObjectSpec with (a) one Exactly request
	// per name in exactlyNames and (b) one FirstAvailable request whose
	// subrequests cover every name in firstAvailableNames. Keeping both
	// branches populated in a single spec is the point: validateRctDeviceClasses
	// has to walk both and the test needs to prove it.
	specWithDeviceClasses := func(exactlyNames []string, firstAvailableNames []string) map[string]any {
		requests := make([]any, 0, len(exactlyNames)+1)
		for i, name := range exactlyNames {
			requests = append(requests, map[string]any{
				"name": fmt.Sprintf("ex%d", i),
				"exactly": map[string]any{
					"deviceClassName": name,
					"count":           1,
					"allocationMode":  "ExactCount",
				},
			})
		}
		if len(firstAvailableNames) > 0 {
			subs := make([]any, 0, len(firstAvailableNames))
			for i, name := range firstAvailableNames {
				subs = append(subs, map[string]any{
					"name":            fmt.Sprintf("fa%d", i),
					"deviceClassName": name,
					"count":           1,
					"allocationMode":  "ExactCount",
				})
			}
			requests = append(requests, map[string]any{
				"name":           "fa-req",
				"firstAvailable": subs,
			})
		}
		return map[string]any{"devices": map[string]any{"requests": requests}}
	}

	Context("DeviceClass existence validation (validateRctDeviceClasses)", func() {
		// These tests re-install the tracker mock locally so they control
		// exactly which DeviceClass names are known on the cluster. The
		// BeforeEach seed is overwritten by each SetMock below.

		It("accepts an Exactly request whose deviceClassName is known on the cluster", func() {
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithDeviceClass(randoms.ClusterId, "gpu.nvidia.com"))

			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"gpu.nvidia.com"}, nil))
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})

		It("rejects an Exactly request whose deviceClassName is unknown on the cluster", func() {
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithDeviceClass(randoms.ClusterId, "gpu.nvidia.com"))

			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"compute-domain.nvidia.com"}, nil))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("device class(es) not found"))
			Expect(resp.Body.String()).To(ContainSubstring("compute-domain.nvidia.com"))
			Expect(resp.Body.String()).To(ContainSubstring(randoms.ClusterId))
			// The seeded class should not appear in the "missing" list.
			Expect(resp.Body.String()).NotTo(ContainSubstring("gpu.nvidia.com"))
		})

		It("accepts a spec with both Exactly and FirstAvailable when every referenced class is known", func() {
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithDeviceClass(randoms.ClusterId, "gpu.nvidia.com").
				WithDeviceClass(randoms.ClusterId, "mig.nvidia.com").
				WithDeviceClass(randoms.ClusterId, "vfio.gpu.nvidia.com"))

			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"gpu.nvidia.com"},
				[]string{"mig.nvidia.com", "vfio.gpu.nvidia.com"}))
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})

		It("rejects only the unknown names across both branches (sorted) and preserves the known ones", func() {
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithDeviceClass(randoms.ClusterId, "gpu.nvidia.com").
				WithDeviceClass(randoms.ClusterId, "mig.nvidia.com"))

			// Exactly uses one known (gpu) and one unknown (zzz-unknown.nvidia.com).
			// FirstAvailable uses one known (mig) and one unknown (aaa-missing.nvidia.com).
			// Deduplication plus sorted output means the message lists
			// "aaa-missing.nvidia.com, zzz-unknown.nvidia.com" and nothing else.
			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"gpu.nvidia.com", "zzz-unknown.nvidia.com"},
				[]string{"mig.nvidia.com", "aaa-missing.nvidia.com"}))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring(
				"aaa-missing.nvidia.com, zzz-unknown.nvidia.com"))
			// Known names must not be reported as missing.
			Expect(resp.Body.String()).NotTo(ContainSubstring(`"gpu.nvidia.com,`))
			Expect(resp.Body.String()).NotTo(ContainSubstring(`"mig.nvidia.com,`))
		})

		It("deduplicates: the same deviceClassName in both Exactly and FirstAvailable is reported once", func() {
			// Tracker knows nothing — both references to the same name
			// should produce exactly one entry in the error list, not two.
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"gpu.nvidia.com"},
				[]string{"gpu.nvidia.com"}))
			Expect(resp.Code).To(Equal(http.StatusBadRequest), resp.Body.String())
			// Count occurrences of the name in the body. Exact-match: one hit.
			body := resp.Body.String()
			occurrences := strings.Count(body, "gpu.nvidia.com")
			Expect(occurrences).To(Equal(1), "expected the duplicate name to appear exactly once in: "+body)
		})

		It("fails open on tracker outage: request succeeds even when the tracker returns an error", func() {
			// The validator is a usability aid, not an authoritative gate —
			// the scheduler enforces at pod-bind time. Transient tracker
			// outages must not block legitimate RCT edits.
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithError(http.StatusServiceUnavailable, "tracker is down"))

			resp := postRctWithObjectSpec(specWithDeviceClasses(
				[]string{"would-be-rejected.nvidia.com"}, nil))
			Expect(resp.Code).To(Equal(http.StatusAccepted), resp.Body.String())
		})
	})

	Context("ObjectMeta / ObjectSpec validation wiring", func() {
		// The validators themselves are unit-tested in
		// k8s_object_validation_test.go. This single integration test
		// proves that both CreateRct and UpdateRctById actually call them.
		It("create + update reject a bad label via validateK8sObjectMeta", func() {
			// CreateRct with a valid meta must succeed (proves the create
			// path runs validateK8sObjectMeta and it accepts good input).
			created, _, _ := createRctAsset(buildRandomClusterRctAsset(randoms), router, headerForAdmin(randoms))

			// UpdateRctById with a bad label must fail (proves the update
			// path runs validateK8sObjectMeta and it rejects bad input).
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

	Context("workloads using the resource claim template (GetRctWorkloads)", func() {
		var wsFactory workloadsServiceClient.SimpleMockClientFactory

		BeforeEach(func() {
			wsFactory = workloadsServiceClient.NewSimpleMockClientFactory()
			workloadsServiceClient.SetMock(wsFactory)
		})

		getRctWorkloads := func(id string, header *http.Header) *httptest.ResponseRecorder {
			httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/workloads", urlRct, id), nil)
			Expect(err).To(BeNil())
			httpReq.Header = *header
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)
			return recorder
		}

		DescribeTable("forwards the template name and the asset's scope to the workloads service",
			func(build func(*testutils.RandomValues) api.RctCreationRequest, scopeType workloadsApi.ScopeType, scopeId func() string) {
				created, _, _ := createRctAsset(build(randoms), router, headerForAdmin(randoms))

				resp := getRctWorkloads(created.Meta.Id, headerForAdmin(randoms))
				Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
				Expect(wsFactory.ResourceClaimTemplateWorkloadsRequests).To(Equal([]workloadsApi.InternalResourceClaimTemplateWorkloadsRequest{
					{Name: created.Meta.Name, ScopeType: scopeType, ScopeId: scopeId()},
				}))
			},
			Entry("cluster scope", buildRandomClusterRctAsset, workloadsApi.SCOPETYPE_CLUSTER, func() string { return randoms.ClusterId }),
			Entry("department scope", buildRandomDepartmentRctAsset, workloadsApi.SCOPETYPE_DEPARTMENT, func() string { return randoms.DepartmentId }),
			Entry("project scope", buildRandomProjectRctAsset, workloadsApi.SCOPETYPE_PROJECT, func() string { return randoms.ProjectIdStr }),
		)

		It("returns the workloads, pods and bound claims reported by the workloads service", func() {
			created, _, _ := createRctAsset(buildRandomClusterRctAsset(randoms), router, headerForAdmin(randoms))
			wsFactory.ResourceClaimTemplateWorkloads = &workloadsApi.ResourceClaimTemplateWorkloadsResponse{
				Next: convutils.Ref(int32(1)),
				Workloads: []workloadsApi.ResourceClaimTemplateWorkload{{
					WorkloadId: "workload-1",
					Name:       "train",
					Type:       "Training",
					ProjectId:  randoms.ProjectIdStr,
					Pods: []workloadsApi.ResourceClaimTemplateWorkloadPod{
						{PodId: "pod-1", PodName: "train-0", ResourceClaims: []workloadsApi.PodResourceClaimBinding{
							{Name: "gpu", ClaimName: *workloadsApi.NewNullableString(convutils.Ref("train-0-gpu-x7k2p"))},
						}},
						{PodId: "pod-2", PodName: "train-1", ResourceClaims: []workloadsApi.PodResourceClaimBinding{
							{Name: "gpu"},
						}},
					},
				}},
			}

			resp := getRctWorkloads(created.Meta.Id, headerForAdmin(randoms))
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var body api.ResourceClaimTemplateWorkloadsResponse
			Expect(json.Unmarshal(resp.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(Equal(api.ResourceClaimTemplateWorkloadsResponse{
				Next: 1,
				Workloads: []api.ResourceClaimTemplateWorkload{{
					WorkloadId: "workload-1",
					Name:       "train",
					Type:       "Training",
					ProjectId:  randoms.ProjectIdStr,
					Pods: []api.ResourceClaimTemplateWorkloadPod{
						{PodId: "pod-1", PodName: "train-0", ResourceClaims: []api.PodResourceClaimBinding{
							{Name: "gpu", ClaimName: convutils.Ref("train-0-gpu-x7k2p")},
						}},
						{PodId: "pod-2", PodName: "train-1", ResourceClaims: []api.PodResourceClaimBinding{
							{Name: "gpu"},
						}},
					},
				}},
			}))
		})

		It("returns 404 when the id belongs to a different asset kind", func() {
			hostPath, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

			resp := getRctWorkloads(hostPath.Meta.Id, headerForAdmin(randoms))
			Expect(resp.Code).To(Equal(http.StatusNotFound), resp.Body.String())
			Expect(wsFactory.ResourceClaimTemplateWorkloadsRequests).To(BeEmpty())
		})

		It("returns 404 for a deleted resource claim template", func() {
			created, _, _ := createRctAsset(buildRandomClusterRctAsset(randoms), router, headerForAdmin(randoms))
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlRct, created.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(randoms)
			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusAccepted), deleteResp.Body.String())

			resp := getRctWorkloads(created.Meta.Id, headerForAdmin(randoms))
			Expect(resp.Code).To(Equal(http.StatusNotFound), resp.Body.String())
			Expect(wsFactory.ResourceClaimTemplateWorkloadsRequests).To(BeEmpty())
		})

		It("returns 403 when the caller may not read the template's project", func() {
			forbidden := randoms.
				CloneWithDifferentProject().Forbidden().
				CloneWithDifferentDepartment().Forbidden().
				CloneWithDifferentCluster().Forbidden()
			authorization.SetMock(testing.NewRandomsBasedAuthMockClient(randoms, forbidden))
			org_unit.SetMock(org_unit.NewSimpleMock(randoms, forbidden))
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
				WithCluster(forbidden.ClusterId, testing.BuildDisplayedCluster(forbidden, testing.UnitTestDefaultClusterVersion, false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithDeviceClass(forbidden.ClusterId, rctDeviceClassForTests))

			created, _, _ := createRctAsset(buildRandomProjectRctAsset(forbidden), router, headerForAdmin(randoms))

			resp := getRctWorkloads(created.Meta.Id, headerForNonAdmin(randoms))
			Expect(resp.Code).To(Equal(http.StatusForbidden), resp.Body.String())
			Expect(wsFactory.ResourceClaimTemplateWorkloadsRequests).To(BeEmpty())
		})

		It("passes through the workloads service error", func() {
			created, _, _ := createRctAsset(buildRandomClusterRctAsset(randoms), router, headerForAdmin(randoms))
			wsFactory.ResourceClaimTemplateWorkloadsErr = workloadsApi.NewError(http.StatusServiceUnavailable, "workloads service is down")

			resp := getRctWorkloads(created.Meta.Id, headerForAdmin(randoms))
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable), resp.Body.String())
			Expect(resp.Body.String()).To(ContainSubstring("workloads service is down"))
		})
	})
})
