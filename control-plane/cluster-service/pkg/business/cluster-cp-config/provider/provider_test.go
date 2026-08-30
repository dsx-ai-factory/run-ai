package provider_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/testutils"
	pmocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/business/cluster-cp-config/provider"
	bcmocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/business/cluster-cp-config/provider"
	errtypes "github.com/run-ai/runai/control-plane/cluster-service/pkg/common/errors"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	idmclienttypes "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	tmclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	tmmocks "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

var once sync.Once

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	// Register config types needed for tests
	_ = runai_config.Register[*config.Service]()
	_ = runai_config.Register[*config.Installation]()
	_ = runai_config.Register[*config.Environment]()
	_ = runai_config.Register[*config.ImagePullSecret]()
	_ = runai_config.Register[*authorization_client.Config]()

	testutils.RegisterPostgresSuite(config.Schema, "file://../../../database/migrations")

	RunSpecs(t, "ClusterControlPlaneConfigProvider Suite")
}

type Mocks struct {
	clusterOAuthClientGetter    *pmocks.ClusterOAuthClientGetter
	clusterAccessControlCreator *pmocks.ClusterAccessControlCreator
	tenantSettingsClient        *bcmocks.TenantSettingsClient
	tmInternalApi               *tmmocks.InternalApi
	tmTenantsApi                *tmmocks.TenantsApi
}

type Input struct {
	tenantID    int32
	clusterUUID string
}

var _ = Describe("ClusterControlPlaneConfigProvider", func() {
	var (
		mocks *Mocks
		p     *provider.ClusterControlPlaneConfigProvider

		tenantID    int32 = 1
		clusterUUID uuid.UUID
	)

	BeforeEach(func() {
		once.Do(func() {
			authConfig := runai_config.Get[*config.Auth]()
			authConfig.KeycloakExternalEndpoint = "keycloak"
			runai_config.Set(authConfig)

			cluster := createCluster()
			clusterUUID = cluster.ID
		})

		mocks = initNaiveMocks()
		installation := runai_config.Get[*config.Installation]()
		installation.InstallationType = "selfhosted"
		runai_config.Set(installation)

		envConfig := runai_config.Get[*config.Environment]()
		envConfig.SetDefaults()
		runai_config.Set(envConfig)

		imagePullSecret := runai_config.Get[*config.ImagePullSecret]()
		imagePullSecret.Enabled = true
		imagePullSecret.StaticImagePullSecret = "imagepullsecret"
		runai_config.Set(imagePullSecret)

		p = provider.NewClusterControlPlaneConfigProvider(
			mocks.clusterOAuthClientGetter,
			mocks.clusterAccessControlCreator,
			mocks.tenantSettingsClient,
			&tmclient.APIClient{
				InternalAPI: mocks.tmInternalApi,
				TenantsAPI:  mocks.tmTenantsApi,
			},
		)
	})

	cases := map[string]struct {
		setup     func(*Mocks)
		input     func() Input
		assertion func(config *api.ClusterControlPlaneConfig, err error)
	}{
		"success - full": {
			setup: func(mocks *Mocks) {
				appMetricStore := runai_config.Get[*config.ApplicationMetricStore]()
				envConfig := runai_config.Get[*config.Environment]()

				appMetricStore.Url = "app_url"
				appMetricStore.Username = "app_username"
				appMetricStore.Password = "app_password"
				appMetricStore.Label = "applicative_runai"
				runai_config.Set(appMetricStore)

				envConfig.EnvIdentifier = "staging"
				envConfig.ReleaseName = "release.name1"
				envConfig.FipsMode = "off"
				runai_config.Set(envConfig)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(config *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(config).To(Equal(&api.ClusterControlPlaneConfig{
					FipsMode: "off",
					Agent: api.ClusterControlPlaneConfigAgent{
						Client: api.ClusterControlPlaneConfigAgentClient{
							Id:     "id",
							Name:   "clientId",
							Secret: "secret",
						},
					},
					Workloads: api.ClusterControlPlaneConfigWorkloads{
						Client: api.ClusterControlPlaneConfigWorkloadsClient{
							Id:     "id",
							Name:   "clientId",
							Secret: "secret",
						},
					},
					MetricsRemoteWrite: api.ClusterControlPlaneConfigMetricsRemoteWrite{
						Url:   "url",
						Label: "runai",
						OrgID: "mimirId",
						Credentials: api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
							Username: "username",
							Password: "password",
						},
					},
					MetricsRemoteWrites: []api.ClusterControlPlaneConfigMetricsRemoteWrite{
						{
							Url:   "url",
							Label: "runai",
							OrgID: "mimirId",
							Credentials: api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
								Username: "username",
								Password: "password",
							},
						},
						{
							Url:   "app_url",
							Label: "applicative_runai",
							Credentials: api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
								Username: "app_username",
								Password: "app_password",
							},
						},
					},
					ResearcherService: api.ClusterControlPlaneConfigResearcherService{
						JwksURL:    fmt.Sprintf("https://%s/auth/realms/%s/protocol/openid-connect/certs", "keycloak", "name"),
						LogoutLink: "",
						OauthProxy: api.ClusterControlPlaneConfigResearcherServiceOauthProxy{
							ClientId:      "dummy",
							ClientSecret:  "dummy",
							OidcIssuerUrl: fmt.Sprintf("https://%s/auth/realms/%s", "keycloak", "name"),
						},
					},
					ClusterName:      "name",
					Environment:      "staging",
					ReleaseName:      "release.name1",
					TenantId:         1,
					InstallationType: "selfhosted",
					DockerConfigJson: base64.StdEncoding.EncodeToString([]byte("imagepullsecret")),
				}))
			},
		},
		"success - fips mode on": {
			setup: func(mocks *Mocks) {
				envConfig := runai_config.Get[*config.Environment]()
				envConfig.FipsMode = "on"
				runai_config.Set(envConfig)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(cfg *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(cfg).To(Not(BeNil()))
				Expect(cfg.FipsMode).To(Equal("on"))
			},
		},
		"success - fips mode only": {
			setup: func(mocks *Mocks) {
				envConfig := runai_config.Get[*config.Environment]()
				envConfig.FipsMode = "only"
				runai_config.Set(envConfig)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(cfg *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(cfg).To(Not(BeNil()))
				Expect(cfg.FipsMode).To(Equal("only"))
			},
		},
		"success - installation type launchpad": {
			setup: func(mocks *Mocks) {
				installation := runai_config.Get[*config.Installation]()
				installation.InstallationType = "launchpad"
				runai_config.Set(installation)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(cfg *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(cfg).To(Not(BeNil()))
				Expect(cfg.InstallationType).To(Equal("launchpad"))
			},
		},
		"success - installation type airgap": {
			setup: func(mocks *Mocks) {
				installation := runai_config.Get[*config.Installation]()
				installation.InstallationType = "airgap"
				runai_config.Set(installation)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(cfg *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(cfg).To(Not(BeNil()))
				Expect(cfg.InstallationType).To(Equal("airgap"))
			},
		},
		"success - installation type saas": {
			setup: func(mocks *Mocks) {
				installation := runai_config.Get[*config.Installation]()
				installation.InstallationType = "saas"
				runai_config.Set(installation)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(cfg *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(cfg).To(Not(BeNil()))
				Expect(cfg.InstallationType).To(Equal("saas"))
			},
		},
		"success - metric store without auth": {
			setup: func(mocks *Mocks) {
				mocks.tenantSettingsClient.On("GetTenantMetricStore").Unset().On("GetTenantMetricStore", mock.Anything, mock.Anything).
					Return(&metrics_api.MetricStore{
						MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
							Write: metrics_api.MetricStoreEndpoint{
								Url: "url",
							},
						}),
					}, nil)
			},
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(config *api.ClusterControlPlaneConfig, err error) {
				Expect(err).To(BeNil())
				Expect(config).To(Not(BeNil()))
				Expect(config.MetricsRemoteWrites[0]).To(Equal(api.ClusterControlPlaneConfigMetricsRemoteWrite{
					Label: "runai",
					Url:   "url",
					OrgID: "mimirId",
					Credentials: api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
						Username: "",
						Password: "",
					},
				}))
			},
		},
		"failure - non-existing cluster": {
			input: func() Input {
				return Input{
					tenantID:    tenantID,
					clusterUUID: uuid.New().String(), // non-existing cluster
				}
			},
			assertion: func(config *api.ClusterControlPlaneConfig, err error) {
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errtypes.ErrNotFound)).To(BeTrue())
			},
		},
		"failure - cluster belongs to a different tenant": {
			input: func() Input {
				otherTenantCluster, apiErr := dao.CreateCluster(context.Background(), &models.Cluster{
					Name:     "other-tenant-cluster",
					TenantId: tenantID + 1,
				})
				Expect(apiErr).To(BeNil())

				return Input{
					tenantID:    tenantID,
					clusterUUID: otherTenantCluster.ID.String(),
				}
			},
			assertion: func(config *api.ClusterControlPlaneConfig, err error) {
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errtypes.ErrNotFound)).To(BeTrue())
			},
		},
		"failure - non-existing tenant": {
			setup: func(mocks *Mocks) {
				mocks.tmTenantsApi.On("GetTenantByIdInternalExecute").Unset().On("GetTenantByIdInternalExecute", mock.Anything).
					Return(nil, &http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))
			},
			input: func() Input {
				return Input{
					tenantID:    999, // non-existing tenant
					clusterUUID: clusterUUID.String(),
				}
			},
			assertion: func(config *api.ClusterControlPlaneConfig, err error) {
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errtypes.ErrNotFound)).To(BeTrue())
			},
		},
	}

	for name, c := range cases {
		c := c
		It(fmt.Sprintf("should act correctly in case of %s", name), func() {
			if c.setup != nil {
				c.setup(mocks)
			}

			input := c.input()
			config, err := p.GetClusterControlPlaneConfig(context.Background(), input.tenantID, input.clusterUUID)

			c.assertion(config, err)
		})
	}
})

func initNaiveMocks() *Mocks {
	mocks := &Mocks{
		clusterOAuthClientGetter:    &pmocks.ClusterOAuthClientGetter{},
		clusterAccessControlCreator: &pmocks.ClusterAccessControlCreator{},
		tenantSettingsClient:        &bcmocks.TenantSettingsClient{},
		tmInternalApi:               tmmocks.NewInternalApi(GinkgoT()),
		tmTenantsApi:                tmmocks.NewTenantsApi(GinkgoT()),
	}

	mocks.clusterOAuthClientGetter.On("GetClient", mock.Anything, mock.Anything, mock.Anything).
		Return(idmclienttypes.Client{
			Id:       "id",
			ClientId: "clientId",
			Secret:   "secret",
		}, nil).Maybe()

	mocks.clusterAccessControlCreator.On("CreateAccessControl", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()

	mocks.tenantSettingsClient.On("GetTenantMetricStore", mock.Anything, mock.Anything).
		Return(&metrics_api.MetricStore{
			MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
				Write: metrics_api.MetricStoreEndpoint{
					Url: "url",
					Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
						Type: metrics_api.METRICSTOREENDPOINTAUTHTYPE_BASIC.Ptr(),
						Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
							Username: "username",
							Password: "password",
						}),
					}),
				},
			}),
		}, nil).Maybe()

	mocks.tmTenantsApi.On("GetTenantByIdInternal", mock.Anything, mock.Anything).Return(tmclient.ApiGetTenantByIdInternalRequest{}).Maybe()
	mocks.tmTenantsApi.On("GetTenantByIdInternalExecute", mock.Anything).
		Return(&tmclient.Tenant{
			MimirId: ptr.To("mimirId"),
			Name:    "name",
		}, nil, nil).Maybe()

	return mocks
}

func createCluster() *models.Cluster {
	cluster, apiErr := dao.CreateCluster(context.Background(), &models.Cluster{
		Name:     "name",
		TenantId: 1,
	})
	Expect(apiErr).To(BeNil())
	Expect(cluster).NotTo(BeNil())

	return cluster
}
