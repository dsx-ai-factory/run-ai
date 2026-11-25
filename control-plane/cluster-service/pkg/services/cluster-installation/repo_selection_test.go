/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation_test

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	clusterinstallation "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("SelectHelmRepoForVersion", func() {
	const privateUrl = "https://helm.ngc.nvidia.com/nvidia/runai"
	const publicUrl = "https://helm.ngc.nvidia.com/nvidia/runai-oss"
	installationConfig := &config.Installation{
		HelmRegistryUrl:                 privateUrl,
		HelmRegistryUsername:            "$oauthtoken",
		HelmRegistryPasswordPlaceholder: "<NGC_KEY>",
		HelmRegistryUrlPublic:           publicUrl,
	}

	DescribeTable("picks the repository by cluster version",
		func(version, expectedUrl string, expectCreds bool) {
			repo := clusterinstallation.SelectHelmRepoForVersion(installationConfig, version)
			Expect(repo.URL).To(Equal(expectedUrl))
			if expectCreds {
				Expect(repo.Username).To(Equal("$oauthtoken"))
				Expect(repo.PasswordPlaceholder).To(Equal("<NGC_KEY>"))
			} else {
				Expect(repo.Username).To(BeEmpty())
				Expect(repo.PasswordPlaceholder).To(BeEmpty())
			}
		},
		Entry("3.0 uses the public repo", "3.0.0", publicUrl, false),
		Entry("3.1 uses the public repo", "3.1.4", publicUrl, false),
		Entry("3.0 prerelease sorts below 3.0.0 and uses the private repo", "3.0.0-rc.1", privateUrl, true),
		Entry("2.26 uses the private repo", "2.26.1", privateUrl, true),
		Entry("2.25 uses the private repo", "2.25.0", privateUrl, true),
		Entry("range expression means latest and uses the public repo", ">0.0.0-0", publicUrl, false),
		Entry("development means latest and uses the public repo", "development", publicUrl, false),
	)

	It("uses the public repo for old versions when no private credentials are configured", func() {
		cfg := *installationConfig
		cfg.HelmRegistryUsername = ""
		cfg.HelmRegistryPasswordPlaceholder = ""
		repo := clusterinstallation.SelectHelmRepoForVersion(&cfg, "2.26.1")
		Expect(repo.URL).To(Equal(publicUrl))
		Expect(repo.Username).To(BeEmpty())
		Expect(repo.PasswordPlaceholder).To(BeEmpty())
	})

	It("keeps a custom legacy registry for old versions even without credentials", func() {
		cfg := *installationConfig
		cfg.HelmRegistryUrl = "https://charts.example.com"
		cfg.HelmRegistryUsername = ""
		repo := clusterinstallation.SelectHelmRepoForVersion(&cfg, "2.26.1")
		Expect(repo.URL).To(Equal("https://charts.example.com"))
	})

	It("ignores the public repo for air-gapped installs", func() {
		cfg := *installationConfig
		cfg.InstallationType = "airgap"
		Expect(clusterinstallation.SelectHelmRepoForVersion(&cfg, "3.1.0").URL).To(Equal(privateUrl))
	})

	It("ignores the public repo when the legacy URL is blanked, as the air-gap values do", func() {
		cfg := *installationConfig
		cfg.HelmRegistryUrl = ""
		Expect(clusterinstallation.SelectHelmRepoForVersion(&cfg, "3.1.0").URL).To(BeEmpty())
	})

	It("uses the legacy repo for every version when no public repo is configured", func() {
		cfg := *installationConfig
		cfg.HelmRegistryUrlPublic = ""
		repo := clusterinstallation.SelectHelmRepoForVersion(&cfg, "3.2.0")
		Expect(repo.URL).To(Equal(privateUrl))
		Expect(repo.Username).To(Equal("$oauthtoken"))
	})
})

var _ = Describe("ClusterVersionList with two repositories", func() {
	const publicUrl, privateUrl = "public-url", "private-url"
	newConfig := func() *config.Installation {
		return &config.Installation{
			HelmRegistryUrl:       privateUrl,
			HelmRegistryUrlPublic: publicUrl,
			Version:               "3.1.0",
			MinVersion:            "2.25",
		}
	}

	It("merges versions published in both repositories", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(clusterinstallation.VersionsList{{"3.0.2"}, {"3.1.0"}}, nil)
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(clusterinstallation.VersionsList{{"2.25.3"}, {"2.26.1"}}, nil)
		service := clusterinstallation.NewClusterInstallService(helmActions)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"2.25": "2.25.3", "2.26": "2.26.1", "3.0": "3.0.2", "3.1": "3.1.0"}))
	})

	It("keeps the other repository's versions when one repository fails", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(clusterinstallation.VersionsList{{"3.0.2"}, {"3.1.0"}}, nil)
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(nil, errors.New("401 unauthorized"))
		service := clusterinstallation.NewClusterInstallService(helmActions)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.0": "3.0.2", "3.1": "3.1.0"}))
	})

	It("falls back to the current version when every repository fails", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
		service := clusterinstallation.NewClusterInstallService(helmActions)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.1": "3.1.0"}))
	})

	It("skips the private NGC repo when no credentials are configured", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, "https://helm.ngc.nvidia.com/nvidia/runai-oss").
			Return(clusterinstallation.VersionsList{{"3.0.2"}, {"3.1.0"}}, nil)
		service := clusterinstallation.NewClusterInstallService(helmActions)
		cfg := newConfig()
		cfg.HelmRegistryUrl = "https://helm.ngc.nvidia.com/nvidia/runai"
		cfg.HelmRegistryUrlPublic = "https://helm.ngc.nvidia.com/nvidia/runai-oss"

		versions, err := service.GetClusterSupportedVersions(context.Background(), cfg)

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.0": "3.0.2", "3.1": "3.1.0"}))
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, "https://helm.ngc.nvidia.com/nvidia/runai")
	})

	It("keeps the private repo's versions when the public repo is unreachable", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(nil, errors.New("dial tcp: i/o timeout"))
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(clusterinstallation.VersionsList{{"2.25.3"}}, nil)
		service := clusterinstallation.NewClusterInstallService(helmActions)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"2.25": "2.25.3"}))
	})

	It("never contacts the public repository for air-gapped installs", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(clusterinstallation.VersionsList{{"3.1.0"}}, nil)
		service := clusterinstallation.NewClusterInstallService(helmActions)
		cfg := newConfig()
		cfg.InstallationType = "airgap"

		versions, err := service.GetClusterSupportedVersions(context.Background(), cfg)

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.1": "3.1.0"}))
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, publicUrl)
	})

	It("does not contact any repository when the legacy URL is blanked", func() {
		helmActions := mocks.NewHelmActions(GinkgoT())
		service := clusterinstallation.NewClusterInstallService(helmActions)
		cfg := newConfig()
		cfg.HelmRegistryUrl = ""

		versions, err := service.GetClusterSupportedVersions(context.Background(), cfg)

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.1": "3.1.0"}))
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, mock.Anything)
	})
})
