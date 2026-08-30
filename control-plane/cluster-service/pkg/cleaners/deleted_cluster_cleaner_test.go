package cleaners_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	clusterdeletemocks "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/testutils"
	clustercleaner "github.com/run-ai/runai/control-plane/cluster-service/pkg/cleaners"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	authorizationapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorizationUtils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/stretchr/testify/mock"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
)

const (
	tenantId = 1
)

type ClusterCleanerMocks struct {
	clusterDeleter *clusterdeletemocks.ClusterDeleter
}

func initService() (*clustercleaner.ClusterCleaner, *ClusterCleanerMocks) {
	mocks := &ClusterCleanerMocks{
		clusterDeleter: &clusterdeletemocks.ClusterDeleter{},
	}
	return clustercleaner.NewClusterCleaner(mocks.clusterDeleter, 30*time.Minute, 3*time.Second), mocks
}

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	authorization_client.Init(nil)

	// Register config types needed for tests
	_ = runai_config.Register[*config.Service]()

	_ = runai_config.Register[*authorization_client.Config]()

	testutils.RegisterPostgresSuite(config.Schema, "file://../database/migrations")

	RunSpecs(t, "Cluster cleaner Suite")
}

var _ = Describe("CleanClusters", func() {
	var ctx context.Context
	BeforeEach(func() {
		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})
	})
	Describe("cluster cleaner", func() {
		It("One cluster should be deleted", func() {
			clusterCleaner, mocks := initService()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Once()
			cluster1Name := "clusterDeleted1"
			cluster2Name := "clusterDeleted2"

			cluster1Model := &models.Cluster{
				Name:     cluster1Name,
				Domain:   "cluster-11",
				Version:  "",
				TenantId: tenantId,
			}

			cluster2Model := &models.Cluster{
				Name:     cluster2Name,
				Domain:   "cluster-22",
				Version:  "",
				TenantId: tenantId,
			}
			cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
			cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

			requestDeletion := time.Now().Add(-5 * time.Hour)

			dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster1.ID.String(), tenantId, requestDeletion)
			dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster2.ID.String(), tenantId, time.Now())
			clusterCleaner.Init(ctx)
			mocks.clusterDeleter.AssertExpectations(GinkgoT())
			dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			dao.DeleteCluster(ctx, cluster2.ID.String(), tenantId)
		})
		It("No clusters should be deleted", func() {
			clusterCleaner, mocks := initService()
			mocks.clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Once()
			cluster1Name := "clusterDeleted1"
			cluster2Name := "clusterDeleted2"

			cluster1Model := &models.Cluster{
				Name:     cluster1Name,
				Domain:   "cluster-11",
				Version:  "",
				TenantId: tenantId,
			}

			cluster2Model := &models.Cluster{
				Name:     cluster2Name,
				Domain:   "cluster-22",
				Version:  "",
				TenantId: tenantId,
			}
			cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
			cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

			dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster1.ID.String(), tenantId, time.Now())
			dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster2.ID.String(), tenantId, time.Now())
			clusterCleaner.Init(ctx)
			mocks.clusterDeleter.AssertNotCalled(GinkgoT(), "ForceDeleteCluster", mock.Anything, mock.Anything)
			dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			dao.DeleteCluster(ctx, cluster2.ID.String(), tenantId)
		})
	})
})
