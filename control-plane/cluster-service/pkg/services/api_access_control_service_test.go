package services_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authztypes "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authzmocks "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	access_control "github.com/run-ai/runai/control-plane/cluster-service/pkg/business/cluster-infra/access-control"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/business/cluster-infra/access-control/authn"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/business/cluster-infra/access-control/authz"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	idmtypes "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	idmmocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var _ = Describe("ApiAccessControlService", func() {
	const testSuite = "access_control"

	type Mocks struct {
		idmClient           *idmmocks.IIdm
		authorizationClient *authzmocks.IAuthorizationClient
	}

	var (
		mocks *Mocks

		ctx                     context.Context
		cluster                 *models.Cluster
		accessControlApiService *services.AccessControlApiService
	)

	BeforeEach(func() {
		mocks = &Mocks{
			idmClient:           &idmmocks.IIdm{},
			authorizationClient: &authzmocks.IAuthorizationClient{},
		}

		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "1001")
		var apiErr *api.Error
		cluster, apiErr = dao.CreateCluster(ctx, &models.Cluster{
			Name:     clusterNameGenerator.Generate(testSuite),
			Domain:   "domain",
			Version:  "2.1",
			TenantId: 1001,
		})
		Expect(apiErr).Should(BeNil())

		accessControlManager := access_control.NewClusterAccessControlManager(authn.NewAuthnManager(mocks.idmClient), authz.NewAuthzManager(mocks.authorizationClient))
		accessControlApiService = services.NewAccessControlApiService(accessControlManager)
	})

	Describe("SetClusterAccessControl", func() {
		Context("Happy flow", func() {
			BeforeEach(func() {
				mocks.idmClient.On("CreateInternalClient", ctx, mock.Anything).Return(&idmtypes.CreateInternalClient201Response{
					Id: ptr.To("client-id"),
				}, nil, nil)

				mocks.authorizationClient.On("CreateInternalAccessRule", ctx).Return(authztypes.ApiCreateInternalAccessRuleRequest{}, nil)
				mocks.authorizationClient.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, nil, nil)
			})

			It("Should set cluster access control correctly", func() {
				resp, err := accessControlApiService.SetClusterAccessControl(ctx, cluster.ID.String())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusOK))
				Expect(resp.Body).To(BeNil())
			})
		})

		Context("Sad flow", func() {
			When("Failed to create client", func() {
				BeforeEach(func() {
					mocks.idmClient.On("CreateInternalClient", ctx, mock.Anything).Return(nil, nil, fmt.Errorf("error"))
				})

				It("Should return error", func() {
					resp, err := accessControlApiService.SetClusterAccessControl(ctx, cluster.ID.String())
					Expect(err).ShouldNot(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusInternalServerError))
					Expect(resp.Body).ToNot(BeNil())
				})
			})
		})
	})
})
