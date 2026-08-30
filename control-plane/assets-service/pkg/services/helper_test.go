package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

// mockTenantCtx matches auth.TenantIdOf: CfTenantId must be a string
// (auth.MockAuthContext stores tenantId as int32 and would yield tenant 0).
func mockTenantCtx(randoms *testutils.RandomValues) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfSubjectID, "test")
	return context.WithValue(ctx, runaictx.CfTenantId, randoms.TenantIdStr)
}

var _ = Describe("services helper tests", func() {
	Context("kind environment - type Distributed with DistFramework", func() {
		It("should pass the validate of asset create request meta", func() {
			mpi := "MPI"
			validationError := ValidateAssetCreateRequestMeta(context.Background(), &api.AssetCreationRequest{
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Distributed:   convutils.Ref(true),
					DistFramework: &mpi,
				},
			}, api.ASSETKIND_ENVIRONMENT)
			Expect(validationError).To(BeNil())
		})

		It("kind environment - should reject the validate of asset create request meta, DistFramework is missing", func() {
			validationError := ValidateAssetCreateRequestMeta(context.Background(), &api.AssetCreationRequest{
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Distributed:   convutils.Ref(true),
					DistFramework: nil,
				},
			}, api.ASSETKIND_ENVIRONMENT)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("you must specify a DistFramework value when setting the type to Distributed"))
		})

		It("kind environment - should reject the validate of asset create request meta, Distributed is missing", func() {
			mpi := "MPI"
			validationError := ValidateAssetCreateRequestMeta(context.Background(), &api.AssetCreationRequest{
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Training:      convutils.Ref(false),
					DistFramework: &mpi,
				},
			}, api.ASSETKIND_ENVIRONMENT)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("you need to provide a Distributed value when using the DistFramework"))
		})

		It("kind compute - should reject the validate of asset create request meta, Distributed is missing", func() {
			mpi := "MPI"
			validationError := ValidateAssetCreateRequestMeta(context.Background(), &api.AssetCreationRequest{
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Training:      convutils.Ref(false),
					DistFramework: &mpi,
				},
			}, api.ASSETKIND_COMPUTE)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("workloadSupportedTypes is not supported for kind compute"))
		})
	})

	Context("ValidateClusterIdBelongsToTenant", func() {
		var randoms *testutils.RandomValues

		BeforeEach(func() {
			randoms = testutils.NewRandomValues()
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.22.0", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		It("returns nil when clusterId is nil", func() {
			Expect(ValidateClusterIdBelongsToTenant(mockTenantCtx(randoms), nil)).To(BeNil())
		})

		It("returns nil when clusterId is an empty string", func() {
			Expect(ValidateClusterIdBelongsToTenant(mockTenantCtx(randoms), convutils.Ref(""))).To(BeNil())
		})

		It("returns nil when cluster belongs to the caller tenant", func() {
			Expect(ValidateClusterIdBelongsToTenant(mockTenantCtx(randoms), convutils.Ref(randoms.ClusterId))).To(BeNil())
		})

		It("returns 403 when cluster does not exist", func() {
			unknownClusterId := "00000000-0000-0000-0000-000000000000"
			err := ValidateClusterIdBelongsToTenant(mockTenantCtx(randoms), &unknownClusterId)
			Expect(err).NotTo(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusForbidden))
			Expect(err.Message).To(ContainSubstring("not permitted to perform this operation"))
		})

		It("returns 403 when cluster belongs to a different tenant", func() {
			otherTenant := randoms.CloneWithDifferentTenant()
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(otherTenant.ClusterId, testing.BuildDisplayedCluster(otherTenant, "2.22.0", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

			err := ValidateClusterIdBelongsToTenant(mockTenantCtx(randoms), convutils.Ref(otherTenant.ClusterId))
			Expect(err).NotTo(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusForbidden))
			Expect(err.Message).To(ContainSubstring("not permitted to perform this operation"))
		})

		It("blocks asset creation at scope=cluster when clusterId belongs to a different tenant", func() {
			otherTenant := randoms.CloneWithDifferentTenant()
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.22.0", false)).
				WithCluster(otherTenant.ClusterId, testing.BuildDisplayedCluster(otherTenant, "2.22.0", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

			err := ValidateAssetCreateRequestMeta(mockTenantCtx(randoms), &api.AssetCreationRequest{
				Name:      "x",
				Scope:     api.SCOPE_CLUSTER,
				ClusterId: convutils.Ref(otherTenant.ClusterId),
			}, api.ASSETKIND_HOST_PATH)
			Expect(err).NotTo(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusForbidden))
		})
	})
})

func validateAssetSpec(randoms *testutils.RandomValues, router chi.Router, url string, id string, expectedSpec interface{}) {
	desc := fmt.Sprintf("validateAssetSpec %s/%s", url, id)

	//
	//   get the asset see that the spec is identical to the spec
	//
	getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", url, id), nil)
	Expect(err).To(BeNil(), desc)
	getHttpReq.Header = *headerForAdmin(randoms)

	getHttpResp := httptest.NewRecorder()
	router.ServeHTTP(getHttpResp, getHttpReq)
	Expect(getHttpResp.Code).To(Equal(http.StatusOK), desc)

	actualAsMap := make(map[string]interface{})
	err = json.Unmarshal(getHttpResp.Body.Bytes(), &actualAsMap)
	Expect(err).To(BeNil(), desc)

	expectedAsMap, err := gojsonq.LoadJson(expectedSpec)
	Expect(err).To(BeNil(), desc)

	actualAsYaml, err := yaml.Marshal(actualAsMap["spec"])
	Expect(err).To(BeNil(), desc)

	expectedAsYaml, err := yaml.Marshal(expectedAsMap)
	Expect(err).To(BeNil(), desc)

	Expect(string(actualAsYaml)).To(Equal(string(expectedAsYaml)))
}

func updateAssetSpec(randoms *testutils.RandomValues, router chi.Router, url string, id string, updatedAsset interface{}) {
	payload, err := json.Marshal(updatedAsset)
	Expect(err).To(BeNil())

	updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", url, id), strings.NewReader(string(payload)))
	Expect(err).To(BeNil())
	updateHttpReq.Header = *headerForAdmin(randoms)

	updateHttpResp := httptest.NewRecorder()
	router.ServeHTTP(updateHttpResp, updateHttpReq)

	Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted), updateHttpResp.Body.String())
}
