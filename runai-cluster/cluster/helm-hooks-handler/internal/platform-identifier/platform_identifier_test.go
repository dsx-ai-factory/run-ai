/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform_identifier

import (
	"context"
	"testing"

	ocpv1 "github.com/openshift/api/config/v1"
	internalv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx = context.Background()
)

func TestVanillaIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorVanilla, platform.KubernetesFlavor)
	tu.assert.Equal("", platform.Version)
}

func TestOpenShiftIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	err := tu.k8s.Create(context.Background(),
		customResourceDefinition(string(crdNameOpenShift)))
	tu.assert.ErrorIs(err, nil)

	err = createOcpClusterVersion(tu.k8s, "4.7.0")
	tu.assert.ErrorIs(err, nil)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorOpenShift, platform.KubernetesFlavor)
	tu.assert.Equal("4.7.0", platform.Version)
}

func TestRKEIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	err := tu.k8s.Create(context.Background(),
		customResourceDefinition(string(crdNameRKE)))
	tu.assert.ErrorIs(err, nil)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorRKE, platform.KubernetesFlavor)
	tu.assert.Equal("", platform.Version)
}

func TestGKEIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	err := tu.k8s.Create(context.Background(),
		customResourceDefinition(string(crdNameGKE)))
	tu.assert.ErrorIs(err, nil)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorGKE, platform.KubernetesFlavor)
	tu.assert.Equal("", platform.Version)
}

func TestAKSIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	err := tu.k8s.Create(context.Background(),
		node("aks-node", map[string]string{
			kubernetesFlavorAKSNodeLabelPrefix: "random-aks-label-value",
		}))
	tu.assert.ErrorIs(err, nil)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorAKS, platform.KubernetesFlavor)
	tu.assert.Equal("", platform.Version)
}

func TestEKSIdentification(t *testing.T) {
	tu := newTestingUtils(t)

	err := tu.k8s.Create(context.Background(),
		customResourceDefinition(string(crdNameEKS)))
	tu.assert.ErrorIs(err, nil)

	platform, err := Identify(ctx, tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.Equal(internalv1.KubernetesFlavorEKS, platform.KubernetesFlavor)
	tu.assert.Equal("", platform.Version)
}

func createOcpClusterVersion(k8s client.Client, version string) error {
	clusterVersion := ocpv1.ClusterVersion{
		ObjectMeta: v1.ObjectMeta{
			Name: "version",
		},
		Status: ocpv1.ClusterVersionStatus{
			Desired: ocpv1.Release{
				Version: version,
			},
		},
	}

	return k8s.Create(context.Background(), &clusterVersion)
}

func TestOKENodeLabelsIdentification(t *testing.T) {
	tu := newTestingUtils(t)
	err := tu.k8s.Create(context.Background(),
		node("oke-node", map[string]string{
			kubernetesFlavorOKEOCINodeLabelPrefix + "/fault-domain": "FAULT-DOMAIN-3",
		}))
	tu.assert.ErrorIs(err, nil)

	isOKE, err := identifyOKEFlavour(context.Background(), tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.True(isOKE)
}

func TestOKENoIdentification(t *testing.T) {
	tu := newTestingUtils(t)
	err := tu.k8s.Create(context.Background(),
		node("regular-node", map[string]string{
			"kubernetes.io/hostname": "test",
		}))
	tu.assert.ErrorIs(err, nil)

	isOKE, err := identifyOKEFlavour(context.Background(), tu.k8s)
	tu.assert.ErrorIs(err, nil)
	tu.assert.False(isOKE)
}
