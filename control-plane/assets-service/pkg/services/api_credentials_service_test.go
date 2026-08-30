package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"

	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

const (
	urlCredentials = "/api/v1/asset/credentials"
)

var _ = Describe("credentials specific tests", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	randoms = testutils.NewRandomValues()
	differentProject := randoms.CloneWithDifferentProject()

	var accessKeyA api.AccessKey
	var passwordA api.Password

	BeforeEach(func() {
		dbTesting.CleanTables()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject))
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		accessKeyA, _, _ = createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
		passwordA, _, _ = createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
	})

	Context("list credentials with various parameters", func() {
		It("should give the correct set", func() {
			//
			//    list all credentials
			//
			listHttpReq, err := http.NewRequest("GET", urlCredentials, nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var listResp api.CredentialsListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
			Expect(err).To(BeNil())

			validateCredentials("list all", listResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_ACCESS_KEY: {Expected: 1},
				api.ASSETKIND_PASSWORD:   {Expected: 1},
			})

			//
			//    list different project
			//
			listHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?%s", urlCredentials, differentProject.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp = httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
			Expect(err).To(BeNil())
			validateCredentials("project + tenants", listResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_ACCESS_KEY: {Expected: 0},
				api.ASSETKIND_PASSWORD:   {Expected: 0},
			})
		})
	})

	Context("credentials - usage info", func() {
		It("should give the correct usage", func() {
			accessKeyB, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
			passwordB, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
			accessKeyC, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
			passwordC, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
			gs1, _, _ := createGenericSecret(&randoms.ClusterId, buildRandomProjectGenericSecret(randoms), router, headerForAdmin(randoms))
			gs2, _, _ := createGenericSecret(&randoms.ClusterId, buildRandomProjectGenericSecret(randoms), router, headerForAdmin(randoms))

			createS3Asset(buildRandomProjectS3(randoms, nil), router, headerForAdmin(randoms))
			s3XUsingA, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyA.Meta.Id), router, headerForAdmin(randoms))
			s3YUsingA, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyA.Meta.Id), router, headerForAdmin(randoms))
			s3UsingB, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyB.Meta.Id), router, headerForAdmin(randoms))

			createGitAsset(buildRandomProjectGit(randoms, nil), router, headerForAdmin(randoms))
			gitXUsingA, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordA.Meta.Id), router, headerForAdmin(randoms))
			gitYUsingA, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordA.Meta.Id), router, headerForAdmin(randoms))
			gitUsingB, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordB.Meta.Id), router, headerForAdmin(randoms))
			gitUsingGs1, _, _ := createGitAsset(buildRandomProjectGit(randoms, &gs1.Meta.Id), router, headerForAdmin(randoms))

			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			environmentCreationReq := buildRandomTenantEnvironmentAsset(randoms)
			addEnvVarToEnvironmentCreationRequest(&environmentCreationReq, api.EnvironmentVariableOfAsset{
				Name: convutils.Ref(testutils.RandString(5)),
				Credential: &api.EnvironmentVariableCredential{
					AssetId: &accessKeyB.Meta.Id,
					Key:     convutils.Ref(testutils.RandString(5)),
				},
			})
			envUsingAccessKeyB, _, _ := createEnvironmentAsset(environmentCreationReq, router, headerForAdmin(randoms))

			//
			//    expected data
			//
			expectedUsageOfAsset := map[string]datastructs.Set[string]{
				accessKeyA.Meta.Id: {
					s3XUsingA.Meta.Id: false,
					s3YUsingA.Meta.Id: false,
				},
				accessKeyB.Meta.Id: {
					s3UsingB.Meta.Id:           false,
					envUsingAccessKeyB.Meta.Id: false,
				},
				accessKeyC.Meta.Id: {},
				passwordA.Meta.Id: {
					gitXUsingA.Meta.Id: false,
					gitYUsingA.Meta.Id: false,
				},
				passwordB.Meta.Id: {
					gitUsingB.Meta.Id: false,
				},
				passwordC.Meta.Id: {},
				gs1.Meta.Id: {
					gitUsingGs1.Meta.Id: false,
				},
				gs2.Meta.Id: {},
			}

			//
			//    list all credentials
			//
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", urlCredentials), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var listResp api.CredentialsListResponse
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &listResp)
			Expect(err).To(BeNil())
			Expect(len(listResp.Entries)).To(Equal(len(expectedUsageOfAsset)))

			for _, entry := range listResp.Entries {
				validateUsageInfoOfCredentials("list all", entry.Meta, entry.UsedBy, expectedUsageOfAsset[entry.Meta.Id])
			}

			//
			//    list only access keys
			//
			getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", urlAccessKey), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp = httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var listAccessKeyResp api.AccessKeyListResponse
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &listAccessKeyResp)
			Expect(err).To(BeNil())
			Expect(len(listAccessKeyResp.Entries)).To(Equal(3))

			for _, entry := range listAccessKeyResp.Entries {
				validateUsageInfoOfCredentials("list access keys", entry.Meta, entry.UsedBy, expectedUsageOfAsset[entry.Meta.Id])
			}

			//
			//    list only passwords
			//
			getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", urlPassword), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp = httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var listPasswordResp api.PasswordListResponse
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &listPasswordResp)
			Expect(err).To(BeNil())
			Expect(len(listPasswordResp.Entries)).To(Equal(3))

			for _, entry := range listPasswordResp.Entries {
				validateUsageInfoOfCredentials("list all", entry.Meta, entry.UsedBy, expectedUsageOfAsset[entry.Meta.Id])
			}

			//
			//    get access key
			//
			getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", urlAccessKey, accessKeyA.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp = httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var getAccessKeyResp api.AccessKey
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &getAccessKeyResp)
			Expect(err).To(BeNil())

			validateUsageInfoOfCredentials("get access key", accessKeyA.Meta, getAccessKeyResp.UsedBy, expectedUsageOfAsset[accessKeyA.Meta.Id])

			//
			//    get password
			//
			getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", urlPassword, passwordA.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp = httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var getPasswordResp api.Password
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &getPasswordResp)
			Expect(err).To(BeNil())

			validateUsageInfoOfCredentials("get password", passwordA.Meta, getPasswordResp.UsedBy, expectedUsageOfAsset[passwordA.Meta.Id])

			//
			// get generic secret
			//
			getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", urlGenericSecret, gs1.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp = httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var getGenericSecretResp api.GenericSecret
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &getGenericSecretResp)
			Expect(err).To(BeNil())
			validateUsageInfoOfCredentials("get generic secret", gs1.Meta, getGenericSecretResp.UsedBy, expectedUsageOfAsset[gs1.Meta.Id])
		})
	})

	Context("test includeDescendants and filter by type params", func() {
		var anotherCluster *testutils.RandomValues
		var anotherProj *testutils.RandomValues

		BeforeEach(func() {
			dbTesting.CleanTables()

			anotherProj = randoms.CloneWithDifferentProject()
			anotherCluster = randoms.CloneWithDifferentCluster()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherProj, anotherCluster)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherProj, anotherCluster))
			authorization.SetMock(authMock)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18", false)).
				WithCluster(anotherCluster.ClusterId, testing.BuildDisplayedCluster(anotherCluster, "2.18", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		It("should return correct credential assets with includeDescendants flag", func() {
			// Create credentials assets at different scope levels
			tenantPassword, _, _ := createPassword(&randoms.ClusterId, buildRandomTenantPassword(), router, headerForAdmin(randoms))
			clusterAccessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomClusterAccessKey(randoms), router, headerForAdmin(randoms))
			anotherClusterAccessKey, _, _ := createAccessKey(&anotherCluster.ClusterId, buildRandomClusterAccessKey(anotherCluster), router, headerForAdmin(anotherCluster))
			departmentGenericSecret, _, _ := createGenericSecret(&randoms.ClusterId, buildRandomDepartmentGenericSecret(randoms), router, headerForAdmin(randoms))
			projectNgcApiKey, _, _ := createNgcApiKey(buildRandomProjectNgcApiKey(randoms), router, headerForAdmin(randoms))
			anotherProjectNgcApiKey, _, _ := createNgcApiKey(buildRandomProjectNgcApiKey(anotherProj), router, headerForAdmin(anotherProj))

			// Test with no params - expect to get all credentials assets
			resp := fetchCredentialsAssets(router, randoms, urlCredentials)
			expectCredentialsAssets(resp.Entries, []string{tenantPassword.Meta.Id, clusterAccessKey.Meta.Id, anotherClusterAccessKey.Meta.Id, departmentGenericSecret.Meta.Id, projectNgcApiKey.Meta.Id, anotherProjectNgcApiKey.Meta.Id})

			// Test with type param - expect to get only credentials
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?type=%s", urlCredentials, api.ASSETKIND_NGC_API_KEY))
			expectCredentialsAssets(resp.Entries, []string{projectNgcApiKey.Meta.Id, anotherProjectNgcApiKey.Meta.Id})

			// Test with cluster id - expect to get the cluster asset and all assets that are above (1 tenant)
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s", urlCredentials, randoms.ClusterId))
			expectCredentialsAssets(resp.Entries, []string{clusterAccessKey.Meta.Id, tenantPassword.Meta.Id})

			// Test with cluster id and includeDescendants - expect to get the cluster asset, all assets that are above (1 tenant) and all that are below (1 department, 2 projects)
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&includeDescendants=true", urlCredentials, randoms.ClusterId))
			expectCredentialsAssets(resp.Entries, []string{clusterAccessKey.Meta.Id, tenantPassword.Meta.Id, departmentGenericSecret.Meta.Id, projectNgcApiKey.Meta.Id, anotherProjectNgcApiKey.Meta.Id})

			// Test with cluster id and scope - expect to get just the cluster asset
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&scope=cluster", urlCredentials, randoms.ClusterId))
			expectCredentialsAssets(resp.Entries, []string{clusterAccessKey.Meta.Id})

			// Test with department id - expect to get the department asset and all assets that are above (1 cluster, 1 tenant)
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s", urlCredentials, randoms.DepartmentId))
			expectCredentialsAssets(resp.Entries, []string{departmentGenericSecret.Meta.Id, clusterAccessKey.Meta.Id, tenantPassword.Meta.Id})

			// Test with department id and scope - expect to get just the department asset
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&scope=department", urlCredentials, randoms.DepartmentId))
			expectCredentialsAssets(resp.Entries, []string{departmentGenericSecret.Meta.Id})

			// Test with department id and includeDescendants - expect to get the department asset, all assets that are above (1 cluster, 1 tenant) and no descendants (department doesn't support includeDescendants yet)
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&includeDescendants=true", urlCredentials, randoms.DepartmentId))
			expectCredentialsAssets(resp.Entries, []string{clusterAccessKey.Meta.Id, tenantPassword.Meta.Id, departmentGenericSecret.Meta.Id})

			// Test with project id - expect to get the project asset and all assets that are above (1 department, 1 cluster, 1 tenant)
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?projectId=%v", urlCredentials, randoms.ProjectId))
			expectCredentialsAssets(resp.Entries, []string{projectNgcApiKey.Meta.Id, departmentGenericSecret.Meta.Id, clusterAccessKey.Meta.Id, tenantPassword.Meta.Id})

			// Test with project id and scope - expect to get just the project asset
			resp = fetchCredentialsAssets(router, randoms, fmt.Sprintf("%s?projectId=%v&scope=project", urlCredentials, randoms.ProjectId))
			expectCredentialsAssets(resp.Entries, []string{projectNgcApiKey.Meta.Id})
		})
	})

	Context("create accessKey using existing secret", func() {
		It("plain text fields must not be provided", func() {
			accessKey := buildRandomProjectAccessKey(randoms)

			//  AccessKeyId provided -> invalid

			accessKey.Spec.ExistingSecretName = convutils.Ref("secret")
			accessKey.Spec.AccessKeyId = convutils.Ref("hello")
			accessKey.Spec.SecretAccessKey = nil

			createHttpReq, err := http.NewRequest("POST", urlAccessKey, makePayload(accessKey))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("accessKeyId should not be provided when creating credentials from existing secret"))

			//  SecretAccessKry provided -> invalid

			accessKey.Spec.ExistingSecretName = convutils.Ref("secret")
			accessKey.Spec.AccessKeyId = nil
			accessKey.Spec.SecretAccessKey = convutils.Ref("hello")

			createHttpReq, err = http.NewRequest("POST", urlAccessKey, makePayload(accessKey))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("secretAccessKey should not be provided when creating credentials from existing secret"))
		})
	})

	Context("create password using existing secret", func() {
		It("plain text fields must not be provided", func() {
			password := buildRandomProjectPassword(randoms)

			//  user provided -> invalid

			password.Spec.ExistingSecretName = convutils.Ref("secret")
			password.Spec.User = convutils.Ref("hello")
			password.Spec.Password = nil

			createHttpReq, err := http.NewRequest("POST", urlPassword, makePayload(password))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("user name should not be provided when creating credentials from existing secret"))

			//  password provided -> invalid

			password.Spec.ExistingSecretName = convutils.Ref("secret")
			password.Spec.User = nil
			password.Spec.Password = convutils.Ref("hello")

			createHttpReq, err = http.NewRequest("POST", urlPassword, makePayload(password))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("password should not be provided when creating credentials from existing secret"))
		})
	})

	Context("create generic-secret using existing secret", func() {
		It("plain text fields must not be provided", func() {
			genericSecret := buildRandomProjectGenericSecret(randoms)

			genericSecret.Spec.ExistingSecretName = convutils.Ref("")
			createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(genericSecret))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))

			genericSecret.Spec.ExistingSecretName = convutils.Ref("secret")

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithGenericSecret(randoms.ClusterId, &genericSecret, nil))

			createHttpReq, err = http.NewRequest("POST", urlGenericSecret, makePayload(genericSecret))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("create accessKey - provided password", func() {
		It("credentials must be provided", func() {
			accessKey := buildRandomProjectAccessKey(randoms)

			//  missing AccessKeyId - invalid

			accessKey.Spec.ExistingSecretName = nil
			accessKey.Spec.AccessKeyId = nil
			accessKey.Spec.SecretAccessKey = convutils.Ref("hello")

			createHttpReq, err := http.NewRequest("POST", urlAccessKey, makePayload(accessKey))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("mandatory accessKeyId is not provided"))

			//  missing secretAccessKey - invalid

			accessKey.Spec.ExistingSecretName = nil
			accessKey.Spec.AccessKeyId = convutils.Ref("secret")
			accessKey.Spec.SecretAccessKey = nil

			createHttpReq, err = http.NewRequest("POST", urlAccessKey, makePayload(accessKey))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("mandatory secretAccessKey is not provided"))
		})
	})

	Context("create password from provided input", func() {
		It("plain text fields must be provided", func() {
			password := buildRandomProjectPassword(randoms)

			//  missing user - invalid

			password.Spec.ExistingSecretName = nil
			password.Spec.User = nil
			password.Spec.Password = convutils.Ref("hello")

			createHttpReq, err := http.NewRequest("POST", urlPassword, makePayload(password))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("mandatory user name is not provided"))

			//  missing password -> invalid

			password.Spec.ExistingSecretName = nil
			password.Spec.User = convutils.Ref("hello")
			password.Spec.Password = nil

			createHttpReq, err = http.NewRequest("POST", urlPassword, makePayload(password))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("mandatory password is not provided"))
		})
	})

	Context("credentials - cluster info", func() {
		It("should give name of the asset in the cluster", func() {
			// Creating pairs for each of the credentials types (1 with existing 1 without)

			// generic-secret
			genericWithExistingReq := buildRandomProjectGenericSecret(randoms)
			genericWithExistingReq.Spec.ExistingSecretName = convutils.Ref("secret-existing-generic")

			genericWithoutExistingReq := buildRandomProjectGenericSecret(randoms)
			genericWithoutExistingReq.Spec.ExistingSecretName = nil
			genericWithoutExistingReq.Spec.KeyValuePairs = &[]api.KeyValuePair{
				{
					Key:   convutils.Ref("key1"),
					Value: convutils.Ref("value1"),
				},
			}

			genericWithExisting, _, _ := createGenericSecret(&randoms.ClusterId, genericWithExistingReq, router, headerForAdmin(randoms))
			genericWithoutExisting, _, _ := createGenericSecret(&randoms.ClusterId, genericWithoutExistingReq, router, headerForAdmin(randoms))

			// access-key
			accessKeyWithoutExistingReq := buildRandomProjectAccessKey(randoms)

			accessKeyWithExistingReq := buildRandomProjectAccessKey(randoms)
			accessKeyWithExistingReq.Spec.ExistingSecretName = convutils.Ref("secret-existing-ac")
			accessKeyWithExistingReq.Spec.AccessKeyId = nil
			accessKeyWithExistingReq.Spec.SecretAccessKey = nil

			accessKeyWithoutExisting, _, _ := createAccessKey(&randoms.ClusterId, accessKeyWithoutExistingReq, router, headerForAdmin(randoms))
			accessKeyWithExisting, _, _ := createAccessKey(&randoms.ClusterId, accessKeyWithExistingReq, router, headerForAdmin(randoms))

			// password
			passwordWithoutExistingReq := buildRandomProjectPassword(randoms)

			passwordWithExistingReq := buildRandomProjectPassword(randoms)
			passwordWithExistingReq.Spec.ExistingSecretName = convutils.Ref("secret-existing-pass")
			passwordWithExistingReq.Spec.User = nil
			passwordWithExistingReq.Spec.Password = nil

			passwordWithExisting, _, _ := createPassword(&randoms.ClusterId, passwordWithExistingReq, router, headerForAdmin(randoms))
			passwordWithoutExisting, _, _ := createPassword(&randoms.ClusterId, passwordWithoutExistingReq, router, headerForAdmin(randoms))

			// docker-registry
			dockerRegistryWithoutExistingReq := buildRandomProjectDockerRegistry(randoms)

			dockerRegistryWithExistingReq := buildRandomProjectDockerRegistry(randoms)
			dockerRegistryWithExistingReq.Spec.ExistingSecretName = convutils.Ref("secret-existing-docker")
			dockerRegistryWithExistingReq.Spec.Password = nil
			dockerRegistryWithExistingReq.Spec.Url = nil
			dockerRegistryWithExistingReq.Spec.User = nil

			dockerRegistryWithoutExisting, _, _ := createDockerRegistry(&randoms.ClusterId, dockerRegistryWithoutExistingReq, router, headerForAdmin(randoms))
			dockerRegistryWithExisting, _, _ := createDockerRegistry(&randoms.ClusterId, dockerRegistryWithExistingReq, router, headerForAdmin(randoms))

			// expected data

			expectedNames := map[string]*string{
				genericWithExisting.Meta.Id:           genericWithExisting.Spec.ExistingSecretName,
				passwordWithExisting.Meta.Id:          passwordWithExisting.Spec.ExistingSecretName,
				accessKeyWithExisting.Meta.Id:         accessKeyWithExisting.Spec.ExistingSecretName,
				dockerRegistryWithExisting.Meta.Id:    dockerRegistryWithExisting.Spec.ExistingSecretName,
				genericWithoutExisting.Meta.Id:        convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[genericWithoutExisting.Meta.Kind], genericWithoutExisting.Meta.Name))),
				accessKeyWithoutExisting.Meta.Id:      convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[accessKeyWithoutExisting.Meta.Kind], accessKeyWithoutExisting.Meta.Name))),
				passwordWithoutExisting.Meta.Id:       convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[passwordWithoutExisting.Meta.Kind], passwordWithoutExisting.Meta.Name))),
				dockerRegistryWithoutExisting.Meta.Id: convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[dockerRegistryWithoutExisting.Meta.Kind], dockerRegistryWithoutExisting.Meta.Name))),
			}

			//
			//    list all credentials
			//
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", urlCredentials), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var listResp api.CredentialsListResponse
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &listResp)
			Expect(err).To(BeNil())

			for _, entry := range listResp.Entries {
				name := expectedNames[entry.Meta.Id]

				if name != nil {
					Expect(entry.ClusterInfo.Resources).To(HaveLen(1))
					Expect(entry.ClusterInfo.Resources[0].Name).To(Equal(name))
				}
			}
		})
	})

	/* At this point we are not supporting spec update
	   Context("update accessKey using existing secret", func() {
	       It("plain text fields must not be provided", func() {
	           accessKey := buildRandomProjectAccessKey(randoms)

	           //  AccessKeyId provided -> invalid

	           accessKeyUpdReq := api.AccessKeyUpdateRequest{
	               Meta: api.AssetUpdateRequest{
	                   Name: accessKey.Meta.Name,
	               },
	               Spec: &api.AccessKeyUpdateSpec{
	                   ExistingSecretName: convutils.Ref("secret"),
	                   AccessKeyId:        convutils.Ref("hello"),
	                   SecretAccessKey:    nil,
	               },
	           }

	           updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAccessKey, accessKeyA.Meta.Id), makePayload(accessKeyUpdReq))
	           updateHttpReq.Header = *headerForAdmin(randoms)
	           Expect(err).To(BeNil())

	           updateHttpResp := httptest.NewRecorder()
	           router.ServeHTTP(updateHttpResp, updateHttpReq)
	           Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	           Expect(updateHttpResp.Body.String()).To(ContainSubstring("accessKeyId should not be provided when creating credentials from existing secret"))

	           //  SecretAccessKry provided -> invalid

	           accessKeyUpdReq.Spec.ExistingSecretName = convutils.Ref("secret")
	           accessKeyUpdReq.Spec.AccessKeyId = nil
	           accessKeyUpdReq.Spec.SecretAccessKey = convutils.Ref("hello")

	           updateHttpReq, err = http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAccessKey, accessKeyA.Meta.Id), makePayload(accessKeyUpdReq))
	           updateHttpReq.Header = *headerForAdmin(randoms)
	           Expect(err).To(BeNil())

	           updateHttpResp = httptest.NewRecorder()
	           router.ServeHTTP(updateHttpResp, updateHttpReq)
	           Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	           Expect(updateHttpResp.Body.String()).To(ContainSubstring("secretAccessKey should not be provided when creating credentials from existing secret"))
	       })
	   })
	*/

	/* At this point we are not supporting spec update
	   Context("update password using existing secret", func() {
	       It("plain text fields must not be provided", func() {
	           password := buildRandomProjectPassword(randoms)

	           //  user provided -> invalid

	           passwordUpdReq := api.PasswordUpdateRequest{
	               Meta: api.AssetUpdateRequest{
	                   Name: password.Meta.Name,
	               },
	               Spec: api.PasswordUpdateSpec{
	                   ExistingSecretName: convutils.Ref("secret"),
	                   User:               convutils.Ref("hello"),
	                   Password:           nil,
	               },
	           }

	           updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlPassword, passwordA.Meta.Id), makePayload(passwordUpdReq))
	           updateHttpReq.Header = *headerForAdmin(randoms)
	           Expect(err).To(BeNil())

	           updateHttpResp := httptest.NewRecorder()
	           router.ServeHTTP(updateHttpResp, updateHttpReq)
	           Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	           Expect(updateHttpResp.Body.String()).To(ContainSubstring("user name should not be provided when creating credentials from existing secret"))

	           //  password provided -> invalid

	           passwordUpdReq.Spec.ExistingSecretName = convutils.Ref("secret")
	           passwordUpdReq.Spec.User = nil
	           passwordUpdReq.Spec.Password = convutils.Ref("hello")

	           updateHttpReq, err = http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlPassword, passwordA.Meta.Id), makePayload(passwordUpdReq))
	           updateHttpReq.Header = *headerForAdmin(randoms)
	           Expect(err).To(BeNil())

	           updateHttpResp = httptest.NewRecorder()
	           router.ServeHTTP(updateHttpResp, updateHttpReq)
	           Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	           Expect(updateHttpResp.Body.String()).To(ContainSubstring("password should not be provided when creating credentials from existing secret"))
	       })
	   })
	*/

	/* At this point we are not supporting spec update
	       Context("update accessKey - provided credentials", func() {
	   		It("credentials must be provided", func() {
	   			accessKey := buildRandomProjectAccessKey(randoms)

	   			//  missing AccessKeyId - invalid

	   			accessKeyUpdReq := api.AccessKeyUpdateRequest{
	   				Meta: api.AssetUpdateRequest{
	   					Name: accessKey.Meta.Name,
	   				},
	   				Spec: api.AccessKeyUpdateSpec{
	   					ExistingSecretName: nil,
	   					AccessKeyId:        nil,
	   					SecretAccessKey:    convutils.Ref("hello"),
	   				},
	   			}

	   			updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAccessKey, accessKeyA.Meta.Id), makePayload(accessKeyUpdReq))
	   			updateHttpReq.Header = *headerForAdmin(randoms)
	   			Expect(err).To(BeNil())

	   			updateHttpResp := httptest.NewRecorder()
	   			router.ServeHTTP(updateHttpResp, updateHttpReq)
	   			Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	   			Expect(updateHttpResp.Body.String()).To(ContainSubstring("mandatory accessKeyId is not provided"))

	   			//  missing secretAccessKey - invalid

	   			accessKeyUpdReq.Spec.ExistingSecretName = nil
	   			accessKeyUpdReq.Spec.AccessKeyId = convutils.Ref("secret")
	   			accessKeyUpdReq.Spec.SecretAccessKey = nil

	   			updateHttpReq, err = http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAccessKey, accessKeyA.Meta.Id), makePayload(accessKeyUpdReq))
	   			updateHttpReq.Header = *headerForAdmin(randoms)
	   			Expect(err).To(BeNil())

	   			updateHttpResp = httptest.NewRecorder()
	   			router.ServeHTTP(updateHttpResp, updateHttpReq)
	   			Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	   			Expect(updateHttpResp.Body.String()).To(ContainSubstring("mandatory secretAccessKey is not provided"))
	   		})
	   	})
	*/

	/* At this point we are not supporting spec update
	       Context("update password from provided input", func() {
	   		It("plain text fields must be provided", func() {
	   			password := buildRandomProjectPassword(randoms)

	   			//  missing user - invalid

	   			passwordUpdReq := api.PasswordUpdateRequest{
	   				Meta: api.AssetUpdateRequest{
	   					Name: password.Meta.Name,
	   				},
	   				Spec: api.PasswordCreationSpec{
	   					ExistingSecretName: nil,
	   					User:               nil,
	   					Password:           convutils.Ref("hello"),
	   				},
	   			}

	   			updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlPassword, passwordA.Meta.Id), makePayload(passwordUpdReq))
	   			updateHttpReq.Header = *headerForAdmin(randoms)
	   			Expect(err).To(BeNil())

	   			updateHttpResp := httptest.NewRecorder()
	   			router.ServeHTTP(updateHttpResp, updateHttpReq)
	   			Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	   			Expect(updateHttpResp.Body.String()).To(ContainSubstring("mandatory user name is not provided"))

	   			//  missing password -> invalid

	   			passwordUpdReq.Spec.ExistingSecretName = nil
	   			passwordUpdReq.Spec.User = convutils.Ref("hello")
	   			passwordUpdReq.Spec.Password = nil

	   			updateHttpReq, err = http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlPassword, passwordA.Meta.Id), makePayload(passwordUpdReq))
	   			updateHttpReq.Header = *headerForAdmin(randoms)
	   			Expect(err).To(BeNil())

	   			updateHttpResp = httptest.NewRecorder()
	   			router.ServeHTTP(updateHttpResp, updateHttpReq)
	   			Expect(updateHttpResp.Code).To(Equal(http.StatusBadRequest))
	   			Expect(updateHttpResp.Body.String()).To(ContainSubstring("mandatory password is not provided"))
	   		})
	   	})
	*/
})

func validateCredentials(desc string, entries []api.CredentialsListResponseEntry, counts map[api.AssetKind]ExpectedVsActualInt) {
	// Check actual assets in the response
	for _, entry := range entries {
		if countsOfAsset, ok := counts[entry.Meta.Kind]; ok {
			countsOfAsset.Actual += 1
			counts[entry.Meta.Kind] = countsOfAsset
		} else {
			counts[entry.Meta.Kind] = ExpectedVsActualInt{Actual: 1}
		}
	}

	// Do the comparison
	for kind, counts := range counts {
		Expect(counts.Actual).To(Equal(counts.Expected), fmt.Sprintf("%s: for kind %s", desc, kind))
	}
}

func validateUsageInfoOfCredentials(_ string, assetMeta api.AssetMeta, usedBy *api.AssetUsageInfo, expected datastructs.Set[string]) {
	actualUsage := datastructs.Set[string]{}
	if usedBy == nil {
		Expect(actualUsage).To(Equal(expected), fmt.Sprintf("expecting %s asset %s not to be used", assetMeta.Kind, assetMeta.Id))
		return
	}

	Expect(usedBy.Workspaces).To(BeNil())
	Expect(usedBy.Trainings).To(BeNil())
	Expect(usedBy.Templates).To(BeNil())
	Expect(usedBy.Assets).NotTo(BeNil())
	Expect(usedBy.Assets.Environment).To(BeNil())
	Expect(usedBy.Assets.Compute).To(BeNil())

	for _, usedByAsset := range usedBy.Assets.Datasources {
		actualUsage.Add(usedByAsset.Id)
	}
	for _, usedByAsset := range usedBy.Assets.Environments {
		actualUsage.Add(usedByAsset.Id)
	}

	Expect(actualUsage).To(Equal(expected), fmt.Sprintf("actual VS expected of %s asset id %s", assetMeta.Kind, assetMeta.Id))
}

func fetchCredentialsAssets(router chi.Router, randoms *testutils.RandomValues, url string) api.CredentialsListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	listHttpReq.Header = *headerForAdmin(randoms)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK))

	var resp api.CredentialsListResponse
	respJsonErr := json.Unmarshal(listHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp
}

func expectCredentialsAssets(entries []api.CredentialsListResponseEntry, expectedIds []string) {
	Expect(len(entries)).To(Equal(len(expectedIds)), "Expected %d credentials assets but got %d", len(expectedIds), len(entries))

	actualIds := make([]string, 0, len(entries))
	for _, entry := range entries {
		actualIds = append(actualIds, entry.Meta.Id)
	}

	for _, expectedId := range expectedIds {
		Expect(actualIds).To(ContainElement(expectedId))
	}
}
