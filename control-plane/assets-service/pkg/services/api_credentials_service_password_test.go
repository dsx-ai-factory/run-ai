package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

const (
	urlPassword = "/api/v1/asset/credentials/password"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type PasswordTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewPasswordTestFactory(router chi.Router) *PasswordTestFactory {
	return &PasswordTestFactory{router: router}
}

func (c *PasswordTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.PasswordListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *PasswordTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.PasswordCreationRequest{
		Meta: creationMeta,
		Spec: api.PasswordCreationSpec{
			User:     convutils.Ref(testutils.RandString(8)),
			Password: convutils.Ref(testutils.RandString(8)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *PasswordTestFactory) marshalUpdateReq([]byte) string {
	request := api.PasswordUpdateRequest{
		Meta: api.CredentialUpdateRequest{
			// Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		           Spec: &api.PasswordUpdateSpec{
		   			User:     convutils.Ref(testutils.RandString(8)),
		   			Password: convutils.Ref(testutils.RandString(8)),
		   		},
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *PasswordTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	password, payload, checkSum := createPassword(&randoms.ClusterId, buildRandomSystemPassword(), c.router, header)
	return password.Meta, payload, checkSum
}

func (c *PasswordTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	password, payload, checkSum := createPassword(&randoms.ClusterId, buildRandomTenantPassword(), c.router, header)
	return password.Meta, payload, checkSum
}

func (c *PasswordTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	password, payload, checkSum := createPassword(&randoms.ClusterId, buildRandomClusterPassword(randoms), c.router, header)
	return password.Meta, payload, checkSum
}

func (c *PasswordTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	password, payload, checkSum := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), c.router, header)
	return password.Meta, payload, checkSum
}

func (c *PasswordTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	password, payload, checkSum := createPassword(&randoms.ClusterId, buildRandomDepartmentPassword(randoms), c.router, header)
	return password.Meta, payload, checkSum
}

func (c *PasswordTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.PasswordSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *PasswordTestFactory) addToWorkloadCreationReq(*api.WorkloadCreationRequest, *api.AssetMeta) bool {
	return false // password asset cannot be part of workload creation request
}

// ///////////////////////////////////////////////////////////////////////////////////////////
//
//	Asset Builders
//
// ///////////////////////////////////////////////////////////////////////////////////////////
func buildRandomPasswordHelper() api.PasswordCreationRequest {
	return api.PasswordCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.PasswordCreationSpec{
			User:     convutils.Ref(testutils.RandString(8)),
			Password: convutils.Ref(testutils.RandString(8)),
		},
	}
}

func buildRandomProjectPassword(randoms *testutils.RandomValues) api.PasswordCreationRequest {
	ans := buildRandomPasswordHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentPassword(randoms *testutils.RandomValues) api.PasswordCreationRequest {
	ans := buildRandomPasswordHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantPassword() api.PasswordCreationRequest {
	ans := buildRandomPasswordHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterPassword(randoms *testutils.RandomValues) api.PasswordCreationRequest {
	ans := buildRandomPasswordHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemPassword() api.PasswordCreationRequest {
	ans := buildRandomPasswordHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createPassword(clusterId *string, creationRequest api.PasswordCreationRequest, router chi.Router, header *http.Header) (api.Password, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlPassword, makePayload(creationRequest))
	Expect(err).To(BeNil())

	if clusterId != nil && creationRequest.Spec.ExistingSecretName != nil {
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithPassword(*clusterId, &creationRequest, nil))
	}

	if header != nil {
		createHttpReq.Header = *header
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.Password
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestCredentialsPassword", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test password routes", func() {
		BeforeEach(func() {
			dbTesting.CleanTables()

			randoms = testutils.NewRandomValues()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms))
			authorization.SetMock(authMock)

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		Context("Test Delete password with existing data source which used it", func() {
			It("deletion should return precondition failed used by git", func() {
				password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
				createGitAsset(buildRandomProjectGit(randoms, convutils.Ref(password.Meta.Id)), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPassword, password.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
			It("deletion should return precondition failed used by Secret Volume", func() {
				password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
				createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, &password.Meta.Id), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPassword, password.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})

		// Context("Test rename of password", func() {
		// 	It("deletion should return precondition failed", func() {
		// 		password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))
		// 		createS3Asset(buildRandomProjectS3(randoms, &password.Meta.Id), router, headerForAdmin(randoms))

		// 		//
		// 		//   simulate sync of the asset)
		// 		//
		// 		var metaForUpdate dbmodel.AssetMetaRec
		// 		copier.Copy(&metaForUpdate.AssetMeta, &password.Meta)
		// 		metaForUpdate.AssetMeta.DeletedAt.Valid = false

		// 		syncErr := dao.UpdateSyncStatusSendToAgent(context.Background(), *password.Meta.TenantId, *password.Meta.ClusterId, []dbmodel.AssetsToSync{
		// 			{
		// 				AssetMetaRec:        metaForUpdate,
		// 				AssetsGenericSecret: &dbmodel.GenericSecretRecord{},
		// 			},
		// 		})
		// 		Expect(syncErr).To(BeNil())

		// 		dao.HandleAgentSyncUpdates(context.Background(), *password.Meta.TenantId, *password.Meta.ClusterId, []api.ClusterSyncReportUpdate{
		// 			{
		// 				Id:          password.Meta.Id,
		// 				UpdateCount: password.Meta.UpdateCount,
		// 			},
		// 		})

		// 		renamePayload := api.PasswordUpdateRequest{
		// 			Meta: api.AssetUpdateRequest{
		// 				Name: "new-name",
		// 			},
		// 		}
		// 		updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlPassword, password.Meta.Id), makePayload(renamePayload))
		// 		Expect(err).To(BeNil())
		// 		updateHttpReq.Header = *headerForAdmin(randoms)

		// 		updateHttpResp := httptest.NewRecorder()
		// 		router.ServeHTTP(updateHttpResp, updateHttpReq)
		// 		Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))
		// 		Expect(updateHttpResp.Body.String()).To(ContainSubstring("new-name"))
		// 	})
		// })

		Context("ListGet credentials asset", func() {
			It("should mask the secret", func() {
				password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))

				//
				//   List all
				//
				listHttpReq, err := http.NewRequest("GET", urlPassword, nil)
				Expect(err).To(BeNil())
				listHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var listResp api.PasswordListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
				Expect(err).To(BeNil())
				Expect(len(listResp.Entries)).To(Equal(1))
				Expect(*listResp.Entries[0].Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(*listResp.Entries[0].Spec.Password).To(Equal(encryption.SecretMask))

				//
				//   get by id
				//
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlPassword, password.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp api.Password
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(*getResp.Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(*getResp.Spec.Password).To(Equal(encryption.SecretMask))
			})
		})

		Context("create global (tenant) scope password", func() {
			It("should succeed, tenant scope is allowed", func() {
				createPayload := buildRandomTenantPassword()
				createHttpReq, err := http.NewRequest("POST", urlPassword, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
		})
		Context("Test Delete password used by an environment asset", func() {
			It("deletion should return precondition failed", func() {
				password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(randoms), router, headerForAdmin(randoms))

				environmentCreationReq := buildRandomTenantEnvironmentAsset(randoms)
				addEnvVarToEnvironmentCreationRequest(&environmentCreationReq, api.EnvironmentVariableOfAsset{
					Name: convutils.Ref(testutils.RandString(5)),
					Credential: &api.EnvironmentVariableCredential{
						AssetId: &password.Meta.Id,
						Key:     convutils.Ref(testutils.RandString(5)),
					},
				})
				createEnvironmentAsset(environmentCreationReq, router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPassword, password.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})
	})
})
