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

const (
	publicUrl  = "public-url"
	privateUrl = "private-url"
)

func newTwoRepoConfig() *config.Installation {
	return &config.Installation{
		HelmRegistryUrl:                 privateUrl,
		HelmRegistryUsername:            "$oauthtoken",
		HelmRegistryPasswordPlaceholder: "<NGC_KEY>",
		HelmRegistryUrlPublic:           publicUrl,
		Version:                         "3.1.0",
		MinVersion:                      "2.25",
	}
}

var publicVersions = clusterinstallation.VersionsList{{"3.0.2"}, {"3.1.0"}}
var privateVersions = clusterinstallation.VersionsList{{"2.25.3"}, {"2.26.1"}}

var _ = Describe("GetHelmRepoForVersion", func() {
	var helmActions *mocks.HelmActions
	var service clusterinstallation.ClusterInstallService

	BeforeEach(func() {
		helmActions = mocks.NewHelmActions(GinkgoT())
		service = clusterinstallation.NewClusterInstallService(helmActions)
	})

	givenIndexes := func(public, private clusterinstallation.VersionsList, publicErr, privateErr error) {
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(public, publicErr).Maybe()
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(private, privateErr).Maybe()
	}

	It("uses the public repo, without credentials, for a version it lists", func() {
		givenIndexes(publicVersions, privateVersions, nil, nil)
		repo := service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), "3.1.0")
		Expect(repo).To(Equal(clusterinstallation.HelmRepo{URL: publicUrl}))
	})

	It("uses the private repo, with credentials, for a version only it lists", func() {
		givenIndexes(publicVersions, privateVersions, nil, nil)
		repo := service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), "2.26.1")
		Expect(repo).To(Equal(clusterinstallation.HelmRepo{URL: privateUrl, Username: "$oauthtoken", PasswordPlaceholder: "<NGC_KEY>"}))
	})

	It("prefers the public repo when both list the version", func() {
		givenIndexes(publicVersions, clusterinstallation.VersionsList{{"3.1.0"}}, nil, nil)
		Expect(service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), "3.1.0").URL).To(Equal(publicUrl))
	})

	It("trusts configured credentials when the private index cannot be read", func() {
		givenIndexes(publicVersions, nil, nil, errors.New("401 unauthorized"))
		Expect(service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), "2.26.1").URL).To(Equal(privateUrl))
	})

	It("falls back to the public repo when the version is unlisted and no credentials are configured", func() {
		givenIndexes(publicVersions, nil, nil, errors.New("401 unauthorized"))
		cfg := newTwoRepoConfig()
		cfg.HelmRegistryUsername = ""
		cfg.HelmRegistryPasswordPlaceholder = ""
		repo := service.GetHelmRepoForVersion(context.Background(), cfg, "2.26.1")
		Expect(repo).To(Equal(clusterinstallation.HelmRepo{URL: publicUrl}))
	})

	It("uses a private repo that lists the version even without credentials (e.g. an open mirror)", func() {
		givenIndexes(publicVersions, privateVersions, nil, nil)
		cfg := newTwoRepoConfig()
		cfg.HelmRegistryUsername = ""
		cfg.HelmRegistryPasswordPlaceholder = ""
		Expect(service.GetHelmRepoForVersion(context.Background(), cfg, "2.26.1").URL).To(Equal(privateUrl))
	})

	It("uses the private repo when the public one is unreachable and it lists the version", func() {
		givenIndexes(nil, privateVersions, errors.New("dial tcp: i/o timeout"), nil)
		Expect(service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), "2.25.3").URL).To(Equal(privateUrl))
	})

	DescribeTable("uses the public repo for non-concrete versions without reading any index",
		func(version string) {
			repo := service.GetHelmRepoForVersion(context.Background(), newTwoRepoConfig(), version)
			Expect(repo.URL).To(Equal(publicUrl))
			Expect(repo.Username).To(BeEmpty())
			helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, mock.Anything)
		},
		Entry("range expression", ">0.0.0-0"),
		Entry("development", "development"),
	)

	It("uses only the private repo when no public repo is configured", func() {
		cfg := newTwoRepoConfig()
		cfg.HelmRegistryUrlPublic = ""
		repo := service.GetHelmRepoForVersion(context.Background(), cfg, "3.1.0")
		Expect(repo.URL).To(Equal(privateUrl))
		Expect(repo.Username).To(Equal("$oauthtoken"))
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, mock.Anything)
	})

	It("ignores the public repo when the private URL is blanked, as the air-gap values do", func() {
		cfg := newTwoRepoConfig()
		cfg.HelmRegistryUrl = ""
		Expect(service.GetHelmRepoForVersion(context.Background(), cfg, "3.1.0").URL).To(BeEmpty())
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, mock.Anything)
	})
})

var _ = Describe("ClusterVersionList with two repositories", func() {
	var helmActions *mocks.HelmActions
	var service clusterinstallation.ClusterInstallService

	BeforeEach(func() {
		helmActions = mocks.NewHelmActions(GinkgoT())
		service = clusterinstallation.NewClusterInstallService(helmActions)
	})

	It("combines the latest versions published in both repositories", func() {
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(publicVersions, nil)
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(privateVersions, nil)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newTwoRepoConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"2.25": "2.25.3", "2.26": "2.26.1", "3.0": "3.0.2", "3.1": "3.1.0"}))
	})

	It("keeps the other repository's versions when one repository fails", func() {
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(publicVersions, nil)
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(nil, errors.New("401 unauthorized"))

		versions, err := service.GetClusterSupportedVersions(context.Background(), newTwoRepoConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.0": "3.0.2", "3.1": "3.1.0"}))
	})

	It("keeps the private repository's versions when the public one is unreachable", func() {
		helmActions.On("GetVersionListFromHelm", mock.Anything, publicUrl).Return(nil, errors.New("dial tcp: i/o timeout"))
		helmActions.On("GetVersionListFromHelm", mock.Anything, privateUrl).Return(privateVersions, nil)

		versions, err := service.GetClusterSupportedVersions(context.Background(), newTwoRepoConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"2.25": "2.25.3", "2.26": "2.26.1"}))
	})

	It("falls back to the current version when every repository fails", func() {
		helmActions.On("GetVersionListFromHelm", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))

		versions, err := service.GetClusterSupportedVersions(context.Background(), newTwoRepoConfig())

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.1": "3.1.0"}))
	})

	It("does not contact any repository when the private URL is blanked", func() {
		cfg := newTwoRepoConfig()
		cfg.HelmRegistryUrl = ""

		versions, err := service.GetClusterSupportedVersions(context.Background(), cfg)

		Expect(err).To(BeNil())
		Expect(versions).To(Equal(map[string]string{"3.1": "3.1.0"}))
		helmActions.AssertNotCalled(GinkgoT(), "GetVersionListFromHelm", mock.Anything, mock.Anything)
	})
})
