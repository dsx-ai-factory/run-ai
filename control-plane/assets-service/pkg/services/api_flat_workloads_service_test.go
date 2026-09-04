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
	"os"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/ghodss/yaml"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/stretchr/testify/assert"
)

const (
	assetsPath = "../assets"

	spec              = "spec"
	passwordAssetId   = "passwordAssetId"
	accessKeyAssetId  = "accessKeyAssetId"
	credentialAssetId = "credentialAssetId"

	environmentVariablesKey = "environmentVariables"
)

var urlByType = map[api.AssetKind]string{
	// datasources
	api.ASSETKIND_HOST_PATH:     urlHostPath,
	api.ASSETKIND_NFS:           urlNFS,
	api.ASSETKIND_PVC:           urlPVC,
	api.ASSETKIND_GIT:           urlGit,
	api.ASSETKIND_S3:            urlS3,
	api.ASSETKIND_CONFIG_MAP:    urlConfigMap,
	api.ASSETKIND_SECRET_VOLUME: urlSecretVolume,
	// credentials
	api.ASSETKIND_PASSWORD:       urlPassword,
	api.ASSETKIND_ACCESS_KEY:     urlAccessKey,
	api.ASSETKIND_GENERIC_SECRET: urlGenericSecret,
}

type TestDetails struct {
	workloadType string

	specFname     string
	expectedError string
	expectedFname string

	environmentWorkloadTypes *api.WorkloadSupportedTypes

	runtimeData RuntimeTestData // to be loaded by the test itself
}

type RuntimeTestData struct {
	router   chi.Router
	randoms  *testutils.RandomValues
	basePath string

	// we need the WS mock in order to get the info about the workload that
	// was sent to workload service CRUD
	wsFactory workloadsServiceClient.SimpleMockClientFactory

	// keep an eye on the credentials we already created for this test so we
	// won't attempt to create it twice
	credentials map[string]string
}

var (
	envForWorkspace  = api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
	envForTraining   = api.WorkloadSupportedTypes{Training: convutils.Ref(true)}
	envForInference  = api.WorkloadSupportedTypes{Inference: convutils.Ref(true)}
	envForTensorFlow = api.WorkloadSupportedTypes{Distributed: convutils.Ref(true), DistFramework: convutils.Ref("TF")}
	envForMpi        = api.WorkloadSupportedTypes{Distributed: convutils.Ref(true), DistFramework: convutils.Ref("MPI")}
)

func createExistingResources(runtimeData RuntimeTestData) {
	pvcCreReq := api.PvcCreationRequest{
		Meta: api.AssetCreationRequest{
			Scope:     api.SCOPE_PROJECT,
			ProjectId: convutils.Ref(int32(runtimeData.randoms.ProjectId)),
		},
		Spec: api.PvcAssetSpec{
			ExistingPvc: convutils.Ref(true),
			ClaimName:   convutils.Ref("existing-claim"),
		},
	}

	configMapReq := api.ConfigMapCreationRequest{
		Meta: api.AssetCreationRequest{
			Scope:     api.SCOPE_PROJECT,
			ProjectId: convutils.Ref(int32(runtimeData.randoms.ProjectId)),
		},
		Spec: api.ConfigMapAssetSpec{
			ConfigMap: convutils.Ref("my-config-map"),
		},
	}

	passwordCreReq := api.PasswordCreationRequest{
		Meta: api.AssetCreationRequest{
			Scope:     api.SCOPE_PROJECT,
			ProjectId: convutils.Ref(int32(runtimeData.randoms.ProjectId)),
		},
		Spec: api.PasswordCreationSpec{
			ExistingSecretName: convutils.Ref("my-existing-password-secret"),
		},
	}

	accessKeyCreReq := api.AccessKeyCreationRequest{
		Meta: api.AssetCreationRequest{
			Scope:     api.SCOPE_PROJECT,
			ProjectId: convutils.Ref(int32(runtimeData.randoms.ProjectId)),
		},
		Spec: api.AccessKeyCreationSpec{
			ExistingSecretName: convutils.Ref("my-existing-access-key-secret"),
		},
	}

	genericSecretCreReq := api.GenericSecretCreationRequest{
		Meta: api.AssetCreationRequest{
			Scope:     api.SCOPE_PROJECT,
			ProjectId: convutils.Ref(int32(runtimeData.randoms.ProjectId)),
		},
		Spec: api.GenericSecretCreationSpec{
			ExistingSecretName: convutils.Ref("my-existing-generic-secret"),
		},
	}

	k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
		WithPvc(runtimeData.randoms.ClusterId, &pvcCreReq, nil).
		WithConfigMap(runtimeData.randoms.ClusterId, &configMapReq, nil).
		WithPassword(runtimeData.randoms.ClusterId, &passwordCreReq, nil).
		WithAccessKey(runtimeData.randoms.ClusterId, &accessKeyCreReq, nil).
		WithGenericSecret(runtimeData.randoms.ClusterId, &genericSecretCreReq, nil))
}

var ComposerTests = []TableEntry{
	Entry("gpu portion + specific params", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "rich_specific_params",
		expectedFname:            "rich_specific_params_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("gpu memory + no specific params", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "no_specific_params",
		expectedFname:            "no_specific_params_train",
		environmentWorkloadTypes: &envForTraining,
	}),

	Entry("distributed with master", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "distributed_with_master",
		expectedFname:            "distributed_with_master",
		environmentWorkloadTypes: &envForTensorFlow,
	}),

	Entry("distributed with no master (TF)", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "distributed_no_master",
		expectedFname:            "distributed_no_master",
		environmentWorkloadTypes: &envForTensorFlow,
	}),

	Entry("distributed same as but also assets", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "distributed_same_with_assets",
		expectedError:            "assets should not be provided when sameAsWorker is requested",
		environmentWorkloadTypes: &envForMpi,
	}),

	Entry("distributed master missing", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "distributed_master_missing",
		expectedError:            "assets should be provided for the master replica",
		environmentWorkloadTypes: &envForMpi,
	}),

	Entry("distributed master same as worker", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:                "distributed_master_same",
		expectedFname:            "distributed_master_same",
		environmentWorkloadTypes: &envForTensorFlow,
	}),

	Entry("inference basic test", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:                "inference_basic",
		expectedFname:            "inference_basic",
		environmentWorkloadTypes: &envForInference,
	}),

	/* To open JIRA for this validation, the API does not enforce it
	   Entry("inference basic test <-> workspace only asset", TestDetails{
	   	workloadType:             string(workloadsApi.WORKLOADTYPE_INFERENCE),
	   	specFname:                "inference_basic",
	   	expectedError:            "bla bla bla",
	   	environmentWorkloadTypes: &envForTraining,
	   }),
	*/

	Entry("jupyter and tensor", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "jupyter_tensor",
		expectedFname:            "jupyter_tensor_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("mixed tools", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "mixed_tools",
		expectedFname:            "mixed_tools_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("private datasources", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "private_datasources",
		expectedFname:            "private_datasources_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("private git with ssh", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "private_git_with_ssh",
		expectedFname:            "private_git_with_ssh_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("environment variable with credentials in the environment asset + specific run params", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "env_var_total",
		expectedFname:            "env_var_total_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("environment variable with credentials in the environment asset", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "env_var_without_specific",
		expectedFname:            "env_var_without_specific_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),

	Entry("environment variable with configmap in the environment asset", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "env_var_configmap_no_specific",
		expectedFname:            "env_var_configmap_no_specific_ws",
		environmentWorkloadTypes: &envForWorkspace,
	}),
	Entry("workspace with compute asset with multiple devices", TestDetails{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "multiple_devices",
		expectedFname:            "multiple_devices",
		environmentWorkloadTypes: &envForWorkspace,
	}),
}

type ActualResults struct {
	FlatWorkloads *workloadsServiceClient.SimpleMockClient `json:"flatWorkloads,omitempty"`
}

var _ = Describe("Workload composer validations", func() {
	router := CreateRouter()
	randoms := testutils.NewRandomValues()

	wsFactory := workloadsServiceClient.SimpleMockClientFactory{
		SimpleMockClient: &workloadsServiceClient.SimpleMockClient{},
	}

	BeforeEach(func() {
		dbTesting.CleanTables()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)

		workloadsServiceClient.SetMock(&wsFactory)

		wsFactory.Clear()

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	Context("input VS expected tests", func() {
		DescribeTable("should create workloads properly", func(testDetails TestDetails) {
			testDetails.runtimeData.router = router
			testDetails.runtimeData.randoms = randoms
			testDetails.runtimeData.basePath = "."
			testDetails.runtimeData.wsFactory = wsFactory
			testDetails.runtimeData.credentials = make(map[string]string)
			testDetails.RunWorkloadsTest()
		}, ComposerTests)
	})
})

func (td *TestDetails) RunWorkloadsTest() {
	testBrief := fmt.Sprintf("Spec: %s, Type: %s", td.specFname, td.workloadType)

	createExistingResources(td.runtimeData)

	specYaml, err := testutils.LoadInputFile(td.specFname, td.runtimeData.basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", td.specFname))

	url, payload := td.determineUrlAndPayload(specYaml, testBrief)

	createHttpReq, err := http.NewRequest("POST", url, strings.NewReader(string(gojsonq.ToBytes(payload))))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(td.runtimeData.randoms)

	createHttpResp := httptest.NewRecorder()
	td.runtimeData.router.ServeHTTP(createHttpResp, createHttpReq)
	if td.expectedError != "" {
		Expect(createHttpResp.Code).NotTo(Equal(http.StatusOK))
		Expect(createHttpResp.Body.String()).To(ContainSubstring(td.expectedError))
		return
	}

	Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

	td.runtimeData.wsFactory.SanitizeWorkloads()

	actualResults := ActualResults{
		FlatWorkloads: td.runtimeData.wsFactory.SimpleMockClient, // those are the flat records that were created
	}

	CompareFiles(testBrief, &actualResults, td.expectedFname, td.runtimeData.basePath)
}

func (td *TestDetails) updateWorkloadMetadata(meta *api.WorkloadCreationMeta) {
	meta.Name = td.expectedFname
	if meta.Name == "" { // this is only for errors that do not have expected file
		meta.Name = td.specFname
	}

	meta.ClusterId = td.runtimeData.randoms.ClusterId
	meta.ProjectId = td.runtimeData.randoms.ProjectIdStr
}

func (td *TestDetails) determineUrlAndPayload(specYaml []byte, testBrief string) (string, map[string]interface{}) {
	var payload map[string]interface{}
	var url string

	switch td.workloadType {
	case string(workloadsApi.WORKLOADTYPE_WORKSPACE):
		if specYaml != nil {
			var request api.WorkspaceCreationRequestV2
			err := yaml.Unmarshal(specYaml, &request)
			Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", td.specFname), testBrief)

			td.updateWorkloadMetadata(&request.Meta)
			td.createAssets(&request.Assets)
			if request.SpecificRunParams != nil {
				td.createCredentialsFromEnvar(request.SpecificRunParams.EnvironmentVariables)
			}

			payload, err = gojsonq.LoadJson(request)
			Expect(err).To(BeNil())
		}

		url = urlWorkspaceV2

	case string(workloadsApi.WORKLOADTYPE_TRAINING):
		if specYaml != nil {
			var request api.TrainingCreationRequestV2
			err := yaml.Unmarshal(specYaml, &request)
			Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", td.specFname), testBrief)

			td.updateWorkloadMetadata(&request.Meta)
			td.createAssets(&request.Assets)
			if request.SpecificRunParams != nil {
				td.createCredentialsFromEnvar(request.SpecificRunParams.EnvironmentVariables)
			}

			if request.Distributed != nil && request.Distributed.Master != nil {
				td.createMasterAssets(request.Distributed.Master.Assets)
			}

			payload, err = gojsonq.LoadJson(request)
			Expect(err).To(BeNil())
		}

		url = urlTrainingV2

	case string(workloadsApi.WORKLOADTYPE_INFERENCE):
		if specYaml != nil {
			var request api.InferenceCreationRequestV2
			err := yaml.Unmarshal(specYaml, &request)
			Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", td.specFname), testBrief)

			td.updateWorkloadMetadata(&request.Meta)
			td.createAssets(&request.Assets)
			if request.SpecificRunParams != nil {
				td.createCredentialsFromEnvar(request.SpecificRunParams.EnvironmentVariables)
			}

			payload, err = gojsonq.LoadJson(request)
			Expect(err).To(BeNil())
		}

		url = urlInferenceV2
	}

	return url, payload
}

func (td *TestDetails) createAssets(assets *api.AssetsIds) {
	td.createEnvironmentAsset(&assets.Environment)
	td.createComputeAsset(assets.Compute)
	for idx := range assets.Datasources {
		td.createDatasourceAsset(&assets.Datasources[idx])
	}
}

func (td *TestDetails) createCredentialsFromEnvar(environmentVariables *[]api.EnvironmentVariableOfAsset) {
	if environmentVariables == nil {
		return
	}
	for _, envVar := range *environmentVariables {
		if envVar.Credential == nil {
			continue
		}
		envVar.Credential.AssetId = convutils.Ref(td.createCredentialsAsset(api.ASSETKIND_GENERIC_SECRET, *envVar.Credential.AssetId))
	}
}

func (td *TestDetails) createMasterAssets(assets *api.OptionalAssetIds) {
	if assets == nil {
		return
	}
	td.createEnvironmentAsset(&assets.Environment)
	td.createComputeAsset(assets.Compute)
	for idx := range assets.Datasources {
		td.createDatasourceAsset(&assets.Datasources[idx])
	}
}

func (td *TestDetails) createEnvironmentAsset(asset *string) {
	if asset == nil || *asset == "" {
		return
	}

	var request map[string]interface{}

	path := fmt.Sprintf("%s/environment.%s", assetsPath, *asset)
	specYaml, err := testutils.LoadInputFile(path, td.runtimeData.basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", path))

	err = yaml.Unmarshal(specYaml, &request)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", path))

	td.updateAssetMetadata(request, *asset, td.environmentWorkloadTypes)

	*asset = td.createAssetForTest(urlEnvironment, request)
}

func (td *TestDetails) createComputeAsset(asset *string) {
	if asset == nil || *asset == "" {
		return
	}

	var request map[string]interface{}

	path := fmt.Sprintf("%s/compute.%s", assetsPath, *asset)
	specYaml, err := testutils.LoadInputFile(path, td.runtimeData.basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", path))

	err = yaml.Unmarshal(specYaml, &request)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", path))

	td.updateAssetMetadata(request, *asset, nil)

	*asset = td.createAssetForTest(urlCompute, request)
}

func (td *TestDetails) createDatasourceAsset(datasource *api.AssetIdAndKind) {
	if datasource == nil {
		return
	}

	var request map[string]interface{}

	path := fmt.Sprintf("%s/%s.%s", assetsPath, datasource.Kind, datasource.Id)
	specYaml, err := testutils.LoadInputFile(path, td.runtimeData.basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", path))

	err = yaml.Unmarshal(specYaml, &request)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s", path))

	td.updateAssetMetadata(request, datasource.Id, nil)

	//
	//  create dependent credentials
	//
	spec := convutils.MapMe(request[spec])
	if spec[passwordAssetId] != nil {
		splits := strings.Split(spec[passwordAssetId].(string), ".")
		if len(splits) != 2 {
			return
		}
		realId := td.createCredentialsAsset(api.AssetKind(splits[0]), splits[1])
		spec[passwordAssetId] = realId
	}
	if spec[accessKeyAssetId] != nil {
		realId := td.createCredentialsAsset(api.ASSETKIND_ACCESS_KEY, spec[accessKeyAssetId].(string))
		spec[accessKeyAssetId] = realId
	}
	if spec[credentialAssetId] != nil {
		realId := td.createCredentialsAsset(api.ASSETKIND_GENERIC_SECRET, spec[credentialAssetId].(string))
		spec[credentialAssetId] = realId
	}

	url := urlByType[datasource.Kind]
	Expect(url).NotTo(Equal(""), fmt.Sprintf("test is not supporting datasource kind %s", datasource.Kind))

	datasource.Id = td.createAssetForTest(url, request)
}

func (td *TestDetails) createCredentialsAsset(kind api.AssetKind, mockId string) string {
	var request map[string]interface{}

	path := fmt.Sprintf("%s/%s.%s", assetsPath, kind, mockId)
	if td.runtimeData.credentials[path] != "" {
		return td.runtimeData.credentials[path]
	}

	specYaml, err := testutils.LoadInputFile(path, td.runtimeData.basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", path))

	err = yaml.Unmarshal(specYaml, &request)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", path))

	url := urlByType[kind]
	Expect(url).NotTo(Equal(""), fmt.Sprintf("test is not supporting credentials kind %s", kind))

	td.updateAssetMetadata(request, mockId, nil)

	realId := td.createAssetForTest(url, request)
	td.runtimeData.credentials[path] = realId // keep it for another asset that uses the same credentials

	return realId
}

func (td *TestDetails) createAssetForTest(url string, payload map[string]interface{}) string {
	createHttpReq, err := http.NewRequest("POST", url, strings.NewReader(string(gojsonq.ToBytes(payload))))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(td.runtimeData.randoms)

	createHttpResp := httptest.NewRecorder()
	td.runtimeData.router.ServeHTTP(createHttpResp, createHttpReq)

	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted), createHttpResp.Body.String())

	var resp api.EnvironmentAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp.Meta.Id
}

func (td *TestDetails) updateAssetMetadata(request map[string]interface{}, name string, supportedTypes *api.WorkloadSupportedTypes) {
	assetCreationRequest := api.AssetCreationRequest{
		Name:                   name,
		Scope:                  api.SCOPE_PROJECT,
		ClusterId:              &td.runtimeData.randoms.ClusterId,
		ProjectId:              convutils.Ref(int32(td.runtimeData.randoms.ProjectId)),
		WorkloadSupportedTypes: supportedTypes,
	}

	request["meta"] = assetCreationRequest

	spec := convutils.MapMe(request[spec])
	if spec[environmentVariablesKey] != nil {
		envVars := spec[environmentVariablesKey].([]interface{})
		for _, envVar := range envVars {
			envVarMap := convutils.MapMe(envVar)
			if envVarMap["credential"] != nil {
				cred := convutils.MapMe(envVarMap["credential"])
				cred["assetId"] = td.createCredentialsAsset(api.ASSETKIND_GENERIC_SECRET, cred["assetId"].(string))
			}
		}
	}
}

func CompareFiles(testBrief string, actualResults interface{}, expectedFname string, basePath string) {
	var actualYaml []byte
	actualYaml, marshalError := yaml.Marshal(actualResults)
	Expect(marshalError).To(BeNil(), testBrief)

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testBrief)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
		osError := os.WriteFile(tmpPath, actualYaml, 0o755)
		if osError != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osError.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}
