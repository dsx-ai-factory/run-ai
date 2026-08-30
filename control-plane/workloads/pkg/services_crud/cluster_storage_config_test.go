package services_crud

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
)

func displayedClusterWithStorageImages(clusterId, gitImage, s3Image string) *openapi.DisplayedCluster {
	config := openapi.NewClusterReportedConfig()
	config.SetGitSyncImage(gitImage)
	config.SetS3FileSystemImage(s3Image)

	status := openapi.NewClusterDisplayedStatus()
	status.SetConfig(*config)

	cluster := &openapi.DisplayedCluster{Uuid: clusterId}
	cluster.Status.Set(status)
	return cluster
}

var _ = Describe("buildClusterStorageConfig", func() {
	const clusterId = "cluster-storage-config-test"
	ctx := context.Background()

	It("uses the cluster-reported git-sync and S3 images", func() {
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
			WithCluster(clusterId, displayedClusterWithStorageImages(clusterId, "reg/git-sync:reported", "reg/goofys:reported")))

		cfg, err := buildClusterStorageConfig(ctx, clusterId)

		Expect(err).To(BeNil())
		Expect(cfg.GitSyncImage()).To(Equal("reg/git-sync:reported"))
		Expect(cfg.S3FsImage()).To(Equal("reg/goofys:reported"))
	})

	It("fails the submission when the cluster cannot be read", func() {
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock())

		cfg, err := buildClusterStorageConfig(ctx, "unknown-cluster")

		Expect(cfg).To(BeNil())
		Expect(err).NotTo(BeNil())
	})

	It("uses the cluster-reported oidc issuer when set", func() {
		issuer := "https://cp.example.com/auth/realms/acme"
		cluster := displayedClusterWithStorageImages(clusterId, "reg/git:1", "reg/s3:1")
		config := cluster.Status.Get().GetConfig()
		config.SetRunaiOauth2ProxyOidcIssuerUrl(issuer)
		cluster.Status.Get().SetConfig(config)
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithCluster(clusterId, cluster))

		cfg, err := buildClusterStorageConfig(ctx, clusterId)

		Expect(err).To(BeNil())
		Expect(cfg.OAuth2ProxyOidcIssuerUrl()).To(Equal(issuer))
	})

	It("keeps the default oidc issuer when the cluster does not report one", func() {
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
			WithCluster(clusterId, displayedClusterWithStorageImages(clusterId, "reg/git:1", "reg/s3:1")))

		cfg, err := buildClusterStorageConfig(ctx, clusterId)

		Expect(err).To(BeNil())
		Expect(cfg.OAuth2ProxyOidcIssuerUrl()).To(Equal(composer.DefaultClusterStorageConfig().OAuth2ProxyOidcIssuerUrl()))
	})

	It("falls back to default custom CA wiring and keeps reported images", func() {
		cluster := displayedClusterWithStorageImages(clusterId, "reg/git:1", "reg/s3:1")
		gitCA := openapi.NewClusterReportedCustomCA()
		gitCA.SetEnabled(true)
		gitCA.SetSecretName("git-ca")
		gitCA.SetSecretKey("git-ca.pem")
		config := cluster.Status.Get().GetConfig()
		config.SetCustomCAGit(*gitCA)
		cluster.Status.Get().SetConfig(config)
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithCluster(clusterId, cluster))

		cfg, err := buildClusterStorageConfig(ctx, clusterId)

		Expect(err).To(BeNil())
		Expect(cfg.GitSyncImage()).To(Equal("reg/git:1"))
		Expect(cfg.IsCustomCAGitEnabled()).To(BeTrue())
		Expect(cfg.CustomCAGitSecret()).To(Equal("git-ca"))
	})
})
