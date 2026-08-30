package init_containers

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("BuildOAuth2Bundle", func() {
	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
	})

	It("returns an empty bundle when info is nil", func() {
		bundle, err := BuildOAuth2Bundle(nil)
		Expect(err).To(BeNil())
		Expect(bundle).NotTo(BeNil())
		Expect(bundle.Proxies).To(BeEmpty())
		Expect(bundle.UsernameLoader).To(BeNil())
		Expect(bundle.Volumes).To(BeEmpty())
	})

	It("returns an empty bundle when info has no items", func() {
		bundle, err := BuildOAuth2Bundle(&composer.ExposedUrlInfo{
			Config:       oauth2DefaultConfigWithImage(),
			WorkloadName: "my-workload",
			Items:        map[string]*composer.ExposedUrlInfoItem{},
		})
		Expect(err).To(BeNil())
		Expect(bundle).NotTo(BeNil())
		Expect(bundle.Proxies).To(BeEmpty())
		Expect(bundle.UsernameLoader).To(BeNil())
		Expect(bundle.Volumes).To(BeEmpty())
	})

	It("returns an empty bundle when external auth is disabled and no item overrides it", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(false)
		bundle, err := BuildOAuth2Bundle(&composer.ExposedUrlInfo{
			Config:       config,
			WorkloadName: "my-workload",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app", OAuthPort: convutils.Ref(int32(4180))},
			},
		})
		Expect(err).To(BeNil())
		Expect(bundle.Proxies).To(BeEmpty())
		Expect(bundle.UsernameLoader).To(BeNil())
		Expect(bundle.Volumes).To(BeEmpty())
	})

	It("assembles the proxy, volumes, and (when an allowlist is present) the username loader", func() {
		bundle, err := BuildOAuth2Bundle(&composer.ExposedUrlInfo{
			Config:       oauth2DefaultConfigWithImage(),
			WorkloadName: "my-workload",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {
					Nick:            "url-0",
					ContainerPort:   convutils.Ref(int32(8080)),
					Url:             "https://app.example.com/app",
					OAuthPort:       convutils.Ref(int32(4180)),
					AuthorizedUsers: convutils.Ref("alice@example.com"),
				},
			},
		})
		Expect(err).To(BeNil())
		Expect(bundle.Proxies).To(HaveLen(1))
		Expect(bundle.Proxies[0].Name).To(Equal("runai-oauth2-proxy-0"))
		Expect(bundle.UsernameLoader).NotTo(BeNil())
		Expect(bundle.UsernameLoader.Name).To(Equal("runai-auth-user-loader"))
		Expect(bundle.Volumes).NotTo(BeEmpty())
	})

	It("omits the username loader when no item carries an authorizedUsers list", func() {
		bundle, err := BuildOAuth2Bundle(&composer.ExposedUrlInfo{
			Config:       oauth2DefaultConfigWithImage(),
			WorkloadName: "my-workload",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app", OAuthPort: convutils.Ref(int32(4180))},
			},
		})
		Expect(err).To(BeNil())
		Expect(bundle.Proxies).To(HaveLen(1))
		Expect(bundle.UsernameLoader).To(BeNil())
		Expect(bundle.Volumes).NotTo(BeEmpty())
	})

	It("propagates errors from the underlying proxy builder", func() {
		bundle, err := BuildOAuth2Bundle(&composer.ExposedUrlInfo{
			Config:       composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(clustersApi.ClusterReportedConfig{})),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app"},
			},
		})
		Expect(bundle).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusServiceUnavailable)))
	})
})
