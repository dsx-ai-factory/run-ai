/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package control_plane_details_downloader

import (
	"context"
	"testing"

	cpclusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	rwlc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs"
	runaiv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

var (
	cpd = InstallationConfigurationData{
		Internal: runaiv1_internal.Internal{
			Global: &runaiv1_internal.Global{
				Platform: &runaiv1_internal.Platform{
					KubernetesFlavor: runaiv1_internal.KubernetesFlavorOpenShift,
				},
				ControlPlane: &runaiv1_internal.ControlPlane{
					Config: &runaiv1_internal.ClusterControlPlaneConfig{
						ClusterControlPlaneConfig: cpclusterclient.ClusterControlPlaneConfig{
							ClusterName: ptr.To("cluster"),
							ResearcherService: &cpclusterclient.ClusterControlPlaneConfigResearcherService{
								LogoutLink: ptr.To("logout"),
							},
						},
					},
					URL:         "url",
					ClusterUUID: "cluster-uuid",
				},
			},
		},
	}
)

func TestControlPlaneDetailsDownloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControlPlaneDetailsDownloader Suite")
}

func TestCreateControlPlaneDetailsSecret(t *testing.T) {
	k8s := fake.NewClientBuilder().Build()

	err := CreateOrReplaceInstallationConfigurationSecret(k8s, &cpd)
	assert.Equal(t, err, nil)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      installationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
	}

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)
	assert.Equal(t, err, nil)

	cpd2 := &InstallationConfigurationData{}
	err = yaml.Unmarshal(secret.Data[installationConfigurationKey], cpd2)
	assert.Equal(t, err, nil)

	assert.Equal(t, cpd2.Global.Platform.KubernetesFlavor, cpd.Global.Platform.KubernetesFlavor)
	assert.Equal(t, *cpd2.Global.ControlPlane.Config.ClusterName, *cpd.Global.ControlPlane.Config.ClusterName)
	assert.Equal(t, *cpd2.Global.ControlPlane.Config.ResearcherService.LogoutLink,
		*cpd.Global.ControlPlane.Config.ResearcherService.LogoutLink)

	assert.Equal(t, cpd2.Global.ControlPlane.ClusterUUID, cpd.Global.ControlPlane.ClusterUUID)
	assert.Equal(t, cpd2.Global.ControlPlane.URL, cpd.Global.ControlPlane.URL)
}

func TestReplaceControlPlaneDetailsSecret(t *testing.T) {
	k8s := fake.NewClientBuilder().WithObjects(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      installationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
	}).Build()

	err := CreateOrReplaceInstallationConfigurationSecret(k8s, &cpd)
	assert.Equal(t, err, nil)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      installationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
	}

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)
	assert.Equal(t, err, nil)

	cpd2 := &InstallationConfigurationData{}
	err = yaml.Unmarshal(secret.Data[installationConfigurationKey], cpd2)
	assert.Equal(t, err, nil)

	assert.Equal(t, cpd2.Global.Platform.KubernetesFlavor, cpd.Global.Platform.KubernetesFlavor)
	assert.Equal(t, *cpd2.Global.ControlPlane.Config.ClusterName, *cpd.Global.ControlPlane.Config.ClusterName)
	assert.Equal(t, *cpd2.Global.ControlPlane.Config.ResearcherService.LogoutLink,
		*cpd.Global.ControlPlane.Config.ResearcherService.LogoutLink)
}

func TestGetClientIDOrDefault(t *testing.T) {
	clientID := getClientIDOrDefault("")
	assert.Equal(t, defaultControlPlaneClientID, clientID)

	nonDefaultClientID := "non-default"
	clientID = getClientIDOrDefault(nonDefaultClientID)
	assert.Equal(t, nonDefaultClientID, clientID)
}

func TestAppendControlPlaneImagePullSecretIfNeeded(t *testing.T) {
	tests := []struct {
		name       string
		existing   []string
		dockerJSON string
		expected   []string
	}{
		{
			name:       "saas disabled image pull secret keeps tenant-provided secret only",
			existing:   []string{"tenant-registry-secret"},
			dockerJSON: "",
			expected:   []string{"tenant-registry-secret"},
		},
		{
			name:       "saas disabled image pull secret with no tenant secrets stays empty",
			existing:   nil,
			dockerJSON: " ",
			expected:   nil,
		},
		{
			name:       "cp token with tenant-provided secret appends runai-reg-creds",
			existing:   []string{"tenant-registry-secret"},
			dockerJSON: "e30=",
			expected:   []string{"tenant-registry-secret", hooks.RunAIControlPlaneImagePullSecretName},
		},
		{
			name:       "cp token with no tenant secrets adds runai-reg-creds",
			existing:   nil,
			dockerJSON: "e30=",
			expected:   []string{hooks.RunAIControlPlaneImagePullSecretName},
		},
		{
			name:       "runai-reg-creds is not duplicated",
			existing:   []string{hooks.RunAIControlPlaneImagePullSecretName},
			dockerJSON: "e30=",
			expected:   []string{hooks.RunAIControlPlaneImagePullSecretName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendControlPlaneImagePullSecretIfNeeded(tt.existing, tt.dockerJSON)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterOutApplicativeRunaiRemoteWrite_FilterOutApplicativeRunai(t *testing.T) {
	input := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://aaa.example.com/api/v1/receive"),
			Label: ptr.To(string(rwlc.ClusterMetricsRemoteWriteLabel)),
		},
		{
			Url:   ptr.To("https://app.example.com/api/v1/receive"),
			Label: ptr.To(rwlc.ApplicationMetricsRemoteWriteLabel),
		},
	}
	expected := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://aaa.example.com/api/v1/receive"),
			Label: ptr.To(rwlc.ClusterMetricsRemoteWriteLabel),
		},
	}

	result := filterOutApplicativeRunaiRemoteWrite(input)
	assert.Equal(t, expected, result)
}

func TestFilterOutApplicativeRunaiRemoteWrite_NoApplicativeRunaiToFilter(t *testing.T) {
	input := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://aaa.example.com/api/v1/receive"),
			Label: ptr.To(rwlc.ClusterMetricsRemoteWriteLabel),
		},
		{
			Url:   ptr.To("https://bbb.example.com/api/v1/receive"),
			Label: ptr.To("other"),
		},
	}
	expected := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://aaa.example.com/api/v1/receive"),
			Label: ptr.To(rwlc.ClusterMetricsRemoteWriteLabel),
		},
		{
			Url:   ptr.To("https://bbb.example.com/api/v1/receive"),
			Label: ptr.To("other"),
		},
	}

	result := filterOutApplicativeRunaiRemoteWrite(input)
	assert.Equal(t, expected, result)
}

func TestFilterOutApplicativeRunaiRemoteWrite_EmptyInput(t *testing.T) {
	input := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{}
	expected := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{}

	result := filterOutApplicativeRunaiRemoteWrite(input)
	assert.Equal(t, expected, result)
}

func TestFilterOutApplicativeRunaiRemoteWrite_OnlyApplicativeRunai(t *testing.T) {
	input := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://app.example.com/api/v1/receive"),
			Label: ptr.To(rwlc.ApplicationMetricsRemoteWriteLabel),
		},
	}
	expected := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{}

	result := filterOutApplicativeRunaiRemoteWrite(input)
	assert.Equal(t, expected, result)
}

func TestFilterOutApplicativeRunaiRemoteWrite_NilLabelShouldNotBeFiltered(t *testing.T) {
	input := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://app.example.com/api/v1/receive"),
			Label: nil,
		},
	}
	expected := []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
		{
			Url:   ptr.To("https://app.example.com/api/v1/receive"),
			Label: nil,
		},
	}

	result := filterOutApplicativeRunaiRemoteWrite(input)
	assert.Equal(t, expected, result)
}
