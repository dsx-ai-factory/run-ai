/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/go-chi/chi/v5"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	clusterDownScopes = datastructs.NewSet(clusterScope, departmentScope, projectScope)
	projectScopeOnly  = datastructs.NewSet(projectScope)
	tenantScopeOnly   = datastructs.NewSet(tenantScope)
)

// AssetTestFactory defines an interface that each asset being tested should implement
type AssetTestFactory interface {
	// createSystemBasedAsset should implement creating an asset with system scope
	createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte)

	// createTenantBasedAsset should implement creating an asset with tenant scope
	createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte)

	// createClusterBasedAsset should implement creating an asset with cluster scope
	createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte)

	// createProjectBasedAsset should implement creating an asset with project scope
	createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte)

	// createDepartmentBasedAsset should implement creating an asset with department scope
	createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte)

	// unmarshalList receives the result of list response and returns an array of metadata
	// for all the assets included in the list
	unmarshalList(bytes []byte) []api.AssetMeta

	// marshalCreationReq and marshalUpdateReq returns creation/update request for this asset as a string
	marshalCreationReq(request api.AssetCreationRequest) string
	marshalUpdateReq(createdSpec []byte) string

	// generic function to run a list request and validate that it contains the provided set of asset specs
	// it is actually a generic test implemented by the GenericTest functions below. the function can also
	// test the usage info, i.e. the set of workloads which depend on all the assets included in the response
	// (we do not test each asset separately, it checks all the assets included in the response together).
	validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], assetSpec ...[]byte)

	// add the asset id to a workload creation request. if the asset is o a kind that does not take part
	// in workspaces (for example: s3 secret which cannot be included in a workspace directly) the function
	// should leave the output workload without changing it and return false.
	addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool
}

type AssetTest struct {
	kind                      api.AssetKind
	factory                   AssetTestFactory
	url                       string
	assetDeletionResponseCode int
	assetUpdateResponseCode   int
	testUsageInfo             bool // indication if the asset supports usage info
	testComplianceInfo        bool // indication if the asset supports compliance info
	testStatusInfo            bool // indication if the asset supports status info
	customAuthorization       bool // assets with custom defined API authorization
	supportsPresets           bool // supports preset assets
	supportsWorkloadType      bool // supports preset assets
	supportsAutoDeletion      bool
	notSupportedPre217        bool // we need it for some tests which checks behavior in old clusters
	notSupportedPre220        bool // we need it for some tests which checks behavior in old clusters

	renameNotAllowedDuringSync bool

	cannotAppearInWorkload     bool // asset cannot be part of workload submission directly
	skipClusterDeleteEventTest bool // indicate to skip cluster scope on delete event test

	supportsCleanup bool

	forbiddenScopes *datastructs.Set[api.Scope]
	allowedScopes   *datastructs.Set[api.Scope]
}

func (a *AssetTest) canRunOnScope(scope api.Scope) bool {
	if a.allowedScopes != nil {
		return a.allowedScopes.Has(scope)
	}

	if a.forbiddenScopes != nil {
		return !a.forbiddenScopes.Has(scope)
	}

	return true
}

func (a *AssetTest) canRunOnlyOnScope(scope api.Scope) bool {
	if a.allowedScopes != nil {
		if len(*a.allowedScopes) == 1 && a.allowedScopes.Has(scope) {
			return true
		}
	}
	return false
}

// NewAssetKindTable creates a ginkgo table which runs over all supported assets
func NewAssetKindTable(router chi.Router) []ginkgo.TableEntry {
	return []ginkgo.TableEntry{
		// list environments missing workloadSupportedTypes in the meta
		ginkgo.Entry(string(api.ASSETKIND_ENVIRONMENT), AssetTest{
			kind: api.ASSETKIND_ENVIRONMENT, factory: NewEnvironmentTestFactory(router), url: urlEnvironment,
			testUsageInfo:        true,
			testComplianceInfo:   true,
			supportsPresets:      true,
			supportsWorkloadType: true,
			supportsCleanup:      true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_COMPUTE), AssetTest{
			kind: api.ASSETKIND_COMPUTE, factory: NewComputeTestFactory(router), url: urlCompute,
			testUsageInfo:      true,
			testComplianceInfo: true,
			supportsPresets:    true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_PASSWORD), AssetTest{
			kind: api.ASSETKIND_PASSWORD, factory: NewPasswordTestFactory(router),
			url:            urlPassword,
			testStatusInfo: true,

			cannotAppearInWorkload:     true,
			renameNotAllowedDuringSync: true,
			supportsCleanup:            true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_ACCESS_KEY), AssetTest{
			kind: api.ASSETKIND_ACCESS_KEY, factory: NewAccessKeysTestFactory(router),
			url:            urlAccessKey,
			testStatusInfo: true,

			cannotAppearInWorkload:     true,
			renameNotAllowedDuringSync: true,
			supportsCleanup:            true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_DOCKER_REGISTRY), AssetTest{
			kind: api.ASSETKIND_DOCKER_REGISTRY, factory: NewDockerRegistryTestFactory(router),
			url: urlDockerRegistry,

			cannotAppearInWorkload:     true,
			renameNotAllowedDuringSync: true,
			supportsCleanup:            true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_GENERIC_SECRET), AssetTest{
			kind: api.ASSETKIND_GENERIC_SECRET, factory: NewGenericSecretsTestFactory(router),
			url:           urlGenericSecret,
			allowedScopes: &projectScopeOnly,

			cannotAppearInWorkload:     true,
			renameNotAllowedDuringSync: true,
			supportsCleanup:            true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_S3), AssetTest{
			kind: api.ASSETKIND_S3, factory: NewS3TestFactory(router), url: urlS3,
			testUsageInfo:      true,
			testComplianceInfo: true,
			supportsPresets:    true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_GIT), AssetTest{
			kind: api.ASSETKIND_GIT, factory: NewGitTestFactory(router), url: urlGit,
			testUsageInfo:      true,
			testComplianceInfo: true,
			supportsPresets:    true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_NFS), AssetTest{
			kind: api.ASSETKIND_NFS, factory: NewNFSTestFactory(router), url: urlNFS,
			testUsageInfo:      true,
			testComplianceInfo: true,
			supportsPresets:    true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_HOST_PATH), AssetTest{
			kind: api.ASSETKIND_HOST_PATH, factory: NewHostPathTestFactory(router), url: urlHostPath,
			testUsageInfo:      true,
			testComplianceInfo: true,
			supportsPresets:    true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_PVC), AssetTest{
			kind: api.ASSETKIND_PVC, factory: NewPVCTestFactory(router), url: urlPVC,
			testUsageInfo:        true,
			testComplianceInfo:   true,
			testStatusInfo:       true,
			supportsAutoDeletion: true,
			supportsCleanup:      true,
			allowedScopes:        &clusterDownScopes,
		}),

		ginkgo.Entry(string(api.ASSETKIND_CONFIG_MAP), AssetTest{
			kind: api.ASSETKIND_CONFIG_MAP, factory: NewConfigMapTestFactory(router), url: urlConfigMap,
			testUsageInfo:      true,
			testComplianceInfo: true,
			testStatusInfo:     true,
			allowedScopes:      &clusterDownScopes,
			notSupportedPre220: true,
			supportsCleanup:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_SECRET_VOLUME), AssetTest{
			kind: api.ASSETKIND_SECRET_VOLUME, factory: NewSecretVolumeTestFactory(router), url: urlSecretVolume,
			testUsageInfo:              true,
			testComplianceInfo:         true,
			allowedScopes:              &projectScopeOnly,
			notSupportedPre217:         true,
			skipClusterDeleteEventTest: true, // Skipping test cluster scope because it requires some refactor in the test i/f
			supportsCleanup:            true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_WORKLOAD_TEMPLATE), AssetTest{
			kind: api.ASSETKIND_WORKLOAD_TEMPLATE, factory: NewTemplateTestFactory(router),
			url: urlTemplate,
		}),

		ginkgo.Entry(string(api.ASSETKIND_REGISTRY), AssetTest{
			kind: api.ASSETKIND_REGISTRY, factory: NewRegistryTestFactory(router),
			url:                       urlRegistry,
			assetDeletionResponseCode: 204,
			assetUpdateResponseCode:   200,
			allowedScopes:             &tenantScopeOnly,
			customAuthorization:       true,
			cannotAppearInWorkload:    true,
		}),

		ginkgo.Entry(string(api.ASSETKIND_NGC_API_KEY), AssetTest{
			kind: api.ASSETKIND_NGC_API_KEY, factory: NewNgcApiKeyTestFactory(router),
			url:                        urlNgcApiKey,
			cannotAppearInWorkload:     true,
			renameNotAllowedDuringSync: true,
			supportsCleanup:            true,
			testStatusInfo:             false,
			allowedScopes:              &clusterDownScopes,
		}),
	}
}

// AssetSpec constraint should contain all the possible asset spec in order to allow generic
// tests to be supported for all assets. See below for further information
type AssetSpec interface {
	api.ComputeAssetSpec |
		api.EnvironmentAssetSpec |
		api.AccessKeySpec |
		api.PasswordSpec |
		api.DockerRegistrySpec |
		api.GitAssetSpec |
		api.S3AssetSpec |
		api.NfsAssetSpec |
		api.PvcAssetSpec |
		api.HostPathSpec |
		api.WorkloadTemplate |
		api.RegistrySpec |
		api.ConfigMapAssetSpec |
		api.GenericSecretSpec |
		api.SecretAssetSpec |
		api.NgcApiKeySpec
}

// ListEntry is a generic list response structure for a given asset
type ListEntry[AS AssetSpec] struct {
	Meta   api.AssetMeta       `json:"meta,omitempty"`
	Spec   AS                  `json:"spec,omitempty"`
	UsedBy *api.AssetUsageInfo `json:"usedBy,omitempty"`
}
type ListResponse[AS AssetSpec] struct {
	Entries []ListEntry[AS] `json:"entries"`
}

type GenericListEntry struct {
	Meta   api.AssetMeta               `json:"meta,omitempty"`
	Spec   map[string]interface{}      `json:"spec,omitempty"`
	Status *api.AssetClusterStatusInfo `json:"status,omitempty"`
}

type GenericListResponse struct {
	Entries []GenericListEntry `json:"entries"`
}

// GenericTest assist in implementing generic tests which can run cross all assets
type GenericTest[AS AssetSpec] struct{}

// validateListContent runs a given list command and validates that it contains the provided set of assets.
// it is not invoked directly but each asset calls it as part of its implementation of validateListContent
// of the AssetTestFactory interface (see above)
// it can also compare usedBy information of all the assets included in the response VS expected usage info.
func (gt GenericTest[AssetSpec]) validateListContent(router chi.Router, listRequest *http.Request, expectedUsedBy datastructs.Set[string], expectedAssetSpecs [][]byte) {
	// Normalize expectedAssetSpecs by filtering all empty asset specs
	var normalizeExpectedAssetSpecs [][]byte
	for _, specs := range expectedAssetSpecs {
		if specs != nil {
			normalizeExpectedAssetSpecs = append(normalizeExpectedAssetSpecs, specs)
		}
	}
	expectedAssetSpecs = normalizeExpectedAssetSpecs

	listAllHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listAllHttpResp, listRequest)
	Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

	var listResponse ListResponse[AssetSpec]
	err := json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
	Expect(err).To(BeNil())

	Expect(len(listResponse.Entries)).To(Equal(len(expectedAssetSpecs)))

	actual := map[string]ListEntry[AssetSpec]{}
	actualUsedBy := datastructs.Set[string]{}

	for _, entry := range listResponse.Entries {
		if entry.UsedBy != nil && entry.UsedBy.Workspaces != nil {
			for _, ws := range entry.UsedBy.Workspaces {
				actualUsedBy.Add(ws.Id)
			}
		}
		if entry.UsedBy != nil && entry.UsedBy.Trainings != nil {
			for _, training := range entry.UsedBy.Trainings {
				actualUsedBy.Add(training.Id)
			}
		}
		if entry.UsedBy != nil && entry.UsedBy.Templates != nil {
			for _, template := range entry.UsedBy.Templates {
				actualUsedBy.Add(template.Id)
			}
		}
		entry.UsedBy = nil           // clean the usedBy information as this is checked separately
		entry.Meta.ProjectName = nil // ignore the projectName in the comparison because the expected typically does not have it
		actual[entry.Meta.Id] = entry
	}

	expected := map[string]ListEntry[AssetSpec]{}
	for _, createdSpec := range expectedAssetSpecs {
		var entry ListEntry[AssetSpec]
		err := json.Unmarshal(createdSpec, &entry)
		Expect(err).To(BeNil())
		expected[entry.Meta.Id] = entry
	}

	Expect(actual).To(Equal(expected))
	if expectedUsedBy != nil {
		Expect(actualUsedBy).To(Equal(expectedUsedBy))
	}
}

type MetaAndUsageEntries struct {
	Entries []MetaAndUsage `json:"entries,omitempty"`
}

type MetaAndUsage struct {
	Meta       api.AssetMeta       `json:"meta,omitempty"`
	UsedBy     *api.AssetUsageInfo `json:"usedBy,omitempty"`
	UsageTimes *api.UsageTimesInfo `json:"usageTimes,omitempty"`
}

// run a get request for specific asset and validate its usage info VS expected usage info
func validateUsageInfoOfAsset(router chi.Router, getHttpReq *http.Request, expectedUsage datastructs.Set[string]) {
	getHttpResp := httptest.NewRecorder()
	router.ServeHTTP(getHttpResp, getHttpReq)
	Expect(getHttpResp.Code).To(Equal(http.StatusOK))

	var getResponse MetaAndUsage
	err := json.Unmarshal(getHttpResp.Body.Bytes(), &getResponse)
	Expect(err).To(BeNil())

	actualUsedBy := datastructs.Set[string]{}

	if getResponse.UsedBy != nil {
		if getResponse.UsedBy.Workspaces != nil {
			for _, ws := range getResponse.UsedBy.Workspaces {
				actualUsedBy.Add(ws.Id)
			}
		}
		if getResponse.UsedBy.Trainings != nil {
			for _, training := range getResponse.UsedBy.Trainings {
				actualUsedBy.Add(training.Id)
			}
		}
		if getResponse.UsedBy.Templates != nil {
			for _, temp := range getResponse.UsedBy.Templates {
				actualUsedBy.Add(temp.Id)
			}
		}
	}

	Expect(actualUsedBy).To(Equal(expectedUsage))
}
