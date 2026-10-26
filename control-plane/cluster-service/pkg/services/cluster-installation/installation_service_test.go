/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation_test

import (
	"context"
	"errors"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	clusterinstallation "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("ClusterVersionList", func() {
	Context("getVersionList", func() {
		It("should return version list happy flow", func() {
			helmActions := mocks.NewHelmActions(GinkgoT())
			versionsList := clusterinstallation.VersionsList{{"2.18.0"}, {"2.18.22"}, {"2.15.33"}, {"2.13.44"}, {"2.13.55"}, {"2.19.22"}, {"2.20.22"}}
			helmActions.On("GetVersionListFromHelm", mock.Anything, mock.Anything).Return(versionsList, nil)
			service := clusterinstallation.NewClusterInstallService(helmActions)
			installationConfig := &config.Installation{
				HelmRegistryUrl: "test-url",
				Version:         "2.19.22",
				MinVersion:      "2.13",
				MaxVersion:      "",
				ExcludeVersions: []string{"2.15"},
			}
			versions, err := service.GetClusterSupportedVersions(context.Background(), installationConfig)

			// Assert
			Expect(err).To(BeNil())
			Expect(len(versions)).To(Equal(3))
		})
		It("should return current version empty url - airgapped", func() {
			helmActions := mocks.NewHelmActions(GinkgoT())
			service := clusterinstallation.NewClusterInstallService(helmActions)
			installationConfig := &config.Installation{
				HelmRegistryUrl: "",
				Version:         "2.19.22",
				MinVersion:      "2.13",
				MaxVersion:      "",
				ExcludeVersions: []string{"2.15"},
			}
			versions, err := service.GetClusterSupportedVersions(context.Background(), installationConfig)

			// Assert
			Expect(err).To(BeNil())
			Expect(len(versions)).To(Equal(1))
			Expect(versions["2.19"]).To(Equal("2.19.22"))
		})
		It("should return current version GetVersionListFromHelm returns an error", func() {
			helmActions := mocks.NewHelmActions(GinkgoT())
			helmActions.On("GetVersionListFromHelm", mock.Anything, mock.Anything).Return(nil, errors.New("error getting versions"))
			service := clusterinstallation.NewClusterInstallService(helmActions)
			installationConfig := &config.Installation{
				HelmRegistryUrl: "url",
				Version:         "2.19.22",
				MinVersion:      "2.13",
				MaxVersion:      "",
				ExcludeVersions: []string{"2.15"},
			}
			versions, err := service.GetClusterSupportedVersions(context.Background(), installationConfig)

			// Assert
			Expect(err).To(BeNil())
			Expect(len(versions)).To(Equal(1))
			Expect(versions["2.19"]).To(Equal("2.19.22"))
		})
	})

	Context("generateInstallationStr", func() {
		It("should generate a public NGC installation string without credentials", func() {
			// Arrange
			data := &clusterinstallation.InstallationData{
				RepoURL:          "https://helm.ngc.nvidia.com/nvidia/runai-oss",
				Host:             "test-url",
				BootstrapToken:   "test-secret",
				Uuid:             "test-uuid",
				RemoteClusterUrl: "test-cluster-url",
				Version:          "test-version",
			}

			expectedInstallationStr := `helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="test-version" --create-namespace`

			helmActions := mocks.NewHelmActions(GinkgoT())
			versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
			actualInstallationStr := versionHandler.GetInstallationString(data)

			// Assert
			Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
		})

		It("should use the runai repository name regardless of the repository URL", func() {
			// Arrange
			data := &clusterinstallation.InstallationData{
				RepoURL:          "test-storage-staging-url",
				Host:             "test-url",
				BootstrapToken:   "test-secret",
				Uuid:             "test-uuid",
				RemoteClusterUrl: "test-cluster-url",
				Version:          ">0.0.0-0",
			}

			expectedInstallationStr := `helm repo add runai test-storage-staging-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version=">0.0.0-0" --create-namespace`
			helmActions := mocks.NewHelmActions(GinkgoT())
			versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
			actualInstallationStr := versionHandler.GetInstallationString(data)

			// Assert
			Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
		})
	})

	Context("helmRegistryUsername", func() {
		It("should include username and password placeholder when both are provided", func() {
			data := &clusterinstallation.InstallationData{
				RepoURL:              "test-storage-url",
				HelmRegistryUsername: "$oauthtoken",
				PasswordPlaceholder:  "<Registry password>",
				Host:                 "test-url",
				BootstrapToken:       "test-secret",
				Uuid:                 "test-uuid",
				RemoteClusterUrl:     "test-cluster-url",
				Version:              "2.24.0",
			}

			expectedInstallationStr := `helm repo add runai test-storage-url \
--username='$oauthtoken' --password='<Registry password>' --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

			helmActions := mocks.NewHelmActions(GinkgoT())
			versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
			actualInstallationStr := versionHandler.GetInstallationString(data)

			Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
		})

		It("should use custom password placeholder when provided", func() {
			data := &clusterinstallation.InstallationData{
				RepoURL:              "test-storage-url",
				HelmRegistryUsername: "$oauthtoken",
				PasswordPlaceholder:  "<NGC_API_KEY>",
				Host:                 "test-url",
				BootstrapToken:       "test-secret",
				Uuid:                 "test-uuid",
				RemoteClusterUrl:     "test-cluster-url",
				Version:              "2.24.0",
			}

			expectedInstallationStr := `helm repo add runai test-storage-url \
--username='$oauthtoken' --password='<NGC_API_KEY>' --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

			helmActions := mocks.NewHelmActions(GinkgoT())
			versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
			actualInstallationStr := versionHandler.GetInstallationString(data)

			Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
		})

		It("should not include username and password when username is empty", func() {
			data := &clusterinstallation.InstallationData{
				RepoURL:              "test-storage-url",
				HelmRegistryUsername: "",
				Host:                 "test-url",
				BootstrapToken:       "test-secret",
				Uuid:                 "test-uuid",
				RemoteClusterUrl:     "test-cluster-url",
				Version:              "2.24.0",
			}

			expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

			helmActions := mocks.NewHelmActions(GinkgoT())
			versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
			actualInstallationStr := versionHandler.GetInstallationString(data)

			Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
		})
	})

	Context("ingressClassFlag", func() {
		Context("new cluster installation", func() {
			It("should include haproxy ingress class for version 2.24.0", func() {
				// Arrange
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.24.0",
					LastLiveness:     nil,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should include haproxy ingress class for version greater than 2.24.0", func() {
				// Arrange
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.25.1",
					LastLiveness:     nil,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.25.1" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should not include ingress class flag for version less than 2.24.0", func() {
				// Arrange
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.23.0",
					LastLiveness:     nil,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.23.0" --create-namespace`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should not include ingress class flag for version 2.15.0 (less than 2.24.0)", func() {
				// Arrange
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.15.0",
					LastLiveness:     nil,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.15.0" --create-namespace`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})
		})

		Context("fipsModeFlag", func() {
			It("should include fips mode flag when fips mode is off", func() {
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.24.0",
					FipsMode:         "off",
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy \
--set global.fipsMode=off`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
				actualInstallationStr := versionHandler.GetInstallationString(data)

				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should include fips mode flag when fips mode is on", func() {
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.24.0",
					FipsMode:         "on",
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy \
--set global.fipsMode=on`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
				actualInstallationStr := versionHandler.GetInstallationString(data)

				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should include fips mode flag when fips mode is only", func() {
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.23.0",
					FipsMode:         "only",
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.23.0" --create-namespace \
--set global.fipsMode=only`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)
				actualInstallationStr := versionHandler.GetInstallationString(data)

				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})
		})

		Context("cluster upgrade", func() {
			It("should include reported ingress class when cluster reports specific ingress class", func() {
				// Arrange
				lastLiveness := time.Now()
				nginxIngressClass := "nginx"
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.24.0",
					LastLiveness:     &lastLiveness,
					IngressClass:     &nginxIngressClass,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=nginx`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should include haproxy when cluster reports haproxy ingress class", func() {
				// Arrange
				lastLiveness := time.Now()
				haproxyIngressClass := "haproxy"
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.23.0",
					LastLiveness:     &lastLiveness,
					IngressClass:     &haproxyIngressClass,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.23.0" --create-namespace \
--set clusterConfig.global.ingress.ingressClass=haproxy`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should not include ingress class flag when cluster does not report ingress class", func() {
				// Arrange
				lastLiveness := time.Now()
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.24.0",
					LastLiveness:     &lastLiveness,
					IngressClass:     nil,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.24.0" --create-namespace`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})

			It("should not include ingress class flag when cluster reports empty ingress class", func() {
				// Arrange
				lastLiveness := time.Now()
				emptyIngressClass := ""
				data := &clusterinstallation.InstallationData{
					RepoURL:          "test-storage-url",
					Host:             "test-url",
					BootstrapToken:   "test-secret",
					Uuid:             "test-uuid",
					RemoteClusterUrl: "test-cluster-url",
					Version:          "2.25.0",
					LastLiveness:     &lastLiveness,
					IngressClass:     &emptyIngressClass,
				}

				expectedInstallationStr := `helm repo add runai test-storage-url --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
--set controlPlane.url=test-url \
--set controlPlane.bootstrapToken=test-secret \
--set cluster.uid=test-uuid \
--set cluster.url=test-cluster-url --version="2.25.0" --create-namespace`

				helmActions := mocks.NewHelmActions(GinkgoT())
				versionHandler := clusterinstallation.NewClusterInstallService(helmActions)

				// Act
				actualInstallationStr := versionHandler.GetInstallationString(data)

				// Assert
				Expect(actualInstallationStr).To(Equal(expectedInstallationStr))
			})
		})
	})
})
