package init_containers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	sdkcommon "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func exposedUrlsMapFromInfo(info *composer.ExposedUrlInfo) map[string]flds.ExposedUrlItemAndRules {
	out := make(map[string]flds.ExposedUrlItemAndRules, len(info.Items))
	for nick, item := range info.Items {
		if item == nil {
			out[nick] = flds.ExposedUrlItemAndRules{Value: nil}
			continue
		}
		out[nick] = flds.ExposedUrlItemAndRules{
			Value: &flds.ExposedUrlItemValue{
				ContainerPort:   item.ContainerPort,
				AuthorizedUsers: item.AuthorizedUsers,
			},
		}
	}
	return out
}

var _ = Describe("BuildUsernameLoaderInitContainer", func() {
	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
	})

	It("should return nil when there are no authorized users on any url", func() {
		info := &composer.ExposedUrlInfo{
			Config: composer.DefaultClusterStorageConfig(),
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080))},
			},
		}
		container, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(container).To(BeNil())
	})

	It("should return nil when external auth is disabled even if users are present", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(false)

		info := &composer.ExposedUrlInfo{
			Config: config,
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
			},
		}
		container, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(container).To(BeNil())
	})

	It("should write a base64-decoding command, image and resources for a single url with users", func() {
		info := &composer.ExposedUrlInfo{
			Config: composer.DefaultClusterStorageConfig(),
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com,bob@example.com")},
			},
		}

		c, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(c).NotTo(BeNil())
		Expect(c.Name).To(Equal("runai-auth-user-loader"))
		Expect(c.Image).To(Equal(sdkcommon.WorkloadRunaiOauth2ProxyUsernameLoaderImageRef))
		Expect(c.Command).To(HaveLen(3))
		Expect(c.Command[0]).To(Equal("sh"))
		Expect(c.Command[1]).To(Equal("-c"))
		Expect(c.Command[2]).To(Equal("echo -e \"YWxpY2VAZXhhbXBsZS5jb20KYm9iQGV4YW1wbGUuY29t\" | base64 -d > /etc/runai/auth/users/dXJsLTA="))
		Expect(c.VolumeMounts).To(HaveLen(1))
		Expect(c.VolumeMounts[0].Name).To(Equal("runai-auth"))
		Expect(c.VolumeMounts[0].MountPath).To(Equal("/etc/runai/auth/users"))
		Expect(c.Resources.Requests.Cpu().Equal(resource.MustParse("50m"))).To(BeTrue())
		Expect(c.Resources.Limits.Memory().Equal(resource.MustParse("100Mi"))).To(BeTrue())
		Expect(c.SecurityContext).NotTo(BeNil())
		Expect(c.SecurityContext.AllowPrivilegeEscalation).To(Equal(convutils.Ref(false)))
		Expect(c.SecurityContext.RunAsNonRoot).To(Equal(convutils.Ref(true)))
		Expect(c.SecurityContext.RunAsUser).To(Equal(convutils.Ref(int64(65534))))
	})

	It("should join multiple urls in sorted nick order with the serving-port entry", func() {
		info := &composer.ExposedUrlInfo{
			Config:                     composer.DefaultClusterStorageConfig(),
			ServingPortAuthorizedUsers: convutils.Ref("carol@example.com"),
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-1": {Nick: "url-1", AuthorizedUsers: convutils.Ref("bob@example.com")},
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
			},
		}

		c, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(c).NotTo(BeNil())
		Expect(c.Command[2]).To(ContainSubstring("/etc/runai/auth/users/dXJsLTA="))
		Expect(c.Command[2]).To(ContainSubstring(" && "))
	})

	It("should let the SCC assign an allowed UID on OpenShift", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(true)
		config.SetIsOpenshift(true)

		info := &composer.ExposedUrlInfo{
			Config: config,
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
			},
		}

		c, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(c).NotTo(BeNil())
		Expect(c.SecurityContext.RunAsNonRoot).To(Equal(convutils.Ref(true)))
		Expect(c.SecurityContext.RunAsUser).To(BeNil())
	})

	It("applies configured username-loader resources", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(true)
		expected := corev1.ResourceRequirements{Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("75m"),
		}}
		config.SetUsernameLoaderResources(expected)
		info := &composer.ExposedUrlInfo{
			Config: config,
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
			},
		}

		container, err := BuildUsernameLoaderInitContainer(info)

		Expect(err).To(BeNil())
		Expect(container.Resources).To(Equal(expected))
	})

	It("distinguishes unreported defaults from explicitly empty resources", func() {
		item := map[string]*composer.ExposedUrlInfoItem{
			"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
		}
		unsetConfig := composer.StaticClusterStorageConfig{}
		unsetConfig.SetExternalAuthUrlEnabled(true)
		unset, unsetErr := BuildUsernameLoaderInitContainer(&composer.ExposedUrlInfo{Config: unsetConfig, Items: item})
		explicitConfig := composer.StaticClusterStorageConfig{}
		explicitConfig.SetExternalAuthUrlEnabled(true)
		explicitConfig.SetUsernameLoaderResources(corev1.ResourceRequirements{})
		explicit, explicitErr := BuildUsernameLoaderInitContainer(&composer.ExposedUrlInfo{Config: explicitConfig, Items: item})

		Expect(unsetErr).To(BeNil())
		Expect(explicitErr).To(BeNil())
		Expect(unset.Resources.Requests.Cpu().Equal(resource.MustParse("50m"))).To(BeTrue())
		Expect(unset.Resources.Limits.Memory().Equal(resource.MustParse("100Mi"))).To(BeTrue())
		Expect(explicit.Resources).To(Equal(corev1.ResourceRequirements{}))
	})

	It("reports invalid quantities with the username-loader field name", func() {
		config := clustersApi.ClusterReportedConfig{}
		config.SetExternalAuthUrlEnabled(true)
		config.SetUsernameLoaderResources(composerReportedResources(nil, map[string]string{"memory": "invalid"}))
		info := &composer.ExposedUrlInfo{
			Config: composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(config)),
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")},
			},
		}

		_, err := BuildUsernameLoaderInitContainer(info)

		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("usernameLoaderResources"))
		Expect(err.Message).To(ContainSubstring(`limits.memory has invalid quantity "invalid"`))
	})
})

var _ = Describe("BuildUsernameLoaderInitContainer parity with the cluster builder", func() {
	var info *composer.ExposedUrlInfo

	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
		info = &composer.ExposedUrlInfo{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.ExposedUrlInfoItem{},
		}
	})

	expectParity := func() {
		GinkgoHelper()
		clusterOut := builders.GetUsernameLoaderInitContainer(exposedUrlsMapFromInfo(info), info.ServingPortAuthorizedUsers)
		cpOut, err := BuildUsernameLoaderInitContainer(info)
		Expect(err).To(BeNil())
		Expect(cpOut).To(Equal(clusterOut))
	}

	It("mirrors the cluster when no users are present", func() {
		info.Items["url-0"] = &composer.ExposedUrlInfoItem{Nick: "url-0", ContainerPort: convutils.Ref(int32(8080))}
		expectParity()
	})

	It("mirrors the cluster for a single url with authorized users", func() {
		info.Items["url-0"] = &composer.ExposedUrlInfoItem{Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com,bob@example.com")}
		expectParity()
	})

	It("mirrors the cluster across multiple urls in sorted order", func() {
		info.Items["url-1"] = &composer.ExposedUrlInfoItem{Nick: "url-1", AuthorizedUsers: convutils.Ref("bob@example.com")}
		info.Items["url-0"] = &composer.ExposedUrlInfoItem{Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")}
		expectParity()
	})

	It("mirrors the cluster when a serving-port user set is present", func() {
		info.ServingPortAuthorizedUsers = convutils.Ref("carol@example.com")
		info.Items["url-0"] = &composer.ExposedUrlInfoItem{Nick: "url-0", AuthorizedUsers: convutils.Ref("alice@example.com")}
		expectParity()
	})

	It("mirrors the cluster when a nil item sorts before a real one", func() {
		info.Items["url-0"] = nil
		info.Items["url-1"] = &composer.ExposedUrlInfoItem{Nick: "url-1", AuthorizedUsers: convutils.Ref("alice@example.com")}
		expectParity()
	})
})
