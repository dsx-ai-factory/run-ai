/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform_identifier

import (
	"context"
	"fmt"
	"strings"

	internalv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kubernetesFlavorAKSNodeLabelPrefix    = "kubernetes.azure.com"
	kubernetesFlavorOKEOCINodeLabelPrefix = "oci.oraclecloud.com"
)

type crdName string

const (
	crdNameOpenShift crdName = "clusterversions.config.openshift.io"
	crdNameRKE       crdName = "clusterauthtokens.cluster.cattle.io"
	crdNameGKE       crdName = "allowlistedworkloads.auto.gke.io"
	crdNameEKS       crdName = "securitygrouppolicies.vpcresources.k8s.aws"
)

var (
	flavorToUniqueCRDMap = map[internalv1.KubernetesFlavor]crdName{
		internalv1.KubernetesFlavorOpenShift: crdNameOpenShift,
		internalv1.KubernetesFlavorRKE:       crdNameRKE,
		internalv1.KubernetesFlavorGKE:       crdNameGKE,
		internalv1.KubernetesFlavorEKS:       crdNameEKS,
	}
)

func Identify(ctx context.Context, k8s client.Client) (*internalv1.Platform, error) {
	flavor, err := identifyKubernetesFlavor(ctx, k8s)
	if err != nil {
		return nil, fmt.Errorf("failed to identify kubernetes flavor: %w", err)
	}

	version, err := identifyVersion(ctx, k8s, flavor)
	if err != nil {
		return nil, fmt.Errorf("failed to identify kubernetes version: %w", err)
	}

	return &internalv1.Platform{
		KubernetesFlavor: flavor,
		Version:          version,
	}, nil
}

func identifyKubernetesFlavor(ctx context.Context, k8s client.Client) (internalv1.KubernetesFlavor, error) {
	for flavor, uniqueCRDName := range flavorToUniqueCRDMap {
		crd := apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(uniqueCRDName),
			},
		}

		err := k8s.Get(ctx, client.ObjectKeyFromObject(&crd), &crd)
		if err == nil {
			return flavor, nil
		}
		if err != nil && !errors.IsNotFound(err) {
			return internalv1.KubernetesFlavorUnknown, err
		}
	}

	// AKS has no special API resource that I could find, so I will be using node labels
	// instead in order to categorize the cluster as AKS
	nodeList := &v1.NodeList{}
	err := k8s.List(ctx, nodeList)
	if err != nil {
		return internalv1.KubernetesFlavorUnknown, err
	}

	for _, node := range nodeList.Items {
		labels := node.GetLabels()

		for k := range labels {
			if strings.HasPrefix(k, kubernetesFlavorAKSNodeLabelPrefix) {
				return internalv1.KubernetesFlavorAKS, nil
			}
		}
	}

	isOKE, err := identifyOKEFlavour(ctx, k8s)
	if err != nil {
		return internalv1.KubernetesFlavorUnknown, err
	}
	if isOKE {
		return internalv1.KubernetesFlavorOKE, nil
	}

	// If none of the above flavors were detected, default to vanilla
	return internalv1.KubernetesFlavorVanilla, nil
}

// Supports OpenShift only
func identifyVersion(ctx context.Context, k8s client.Client, kubernetesFlavor internalv1.KubernetesFlavor) (string, error) {
	switch kubernetesFlavor {
	case internalv1.KubernetesFlavorOpenShift:
		return getVersionOCP(ctx, k8s)
	default:
		return "", nil
	}
}

func identifyOKEFlavour(ctx context.Context, k8s client.Client) (bool, error) {
	nodeList := &v1.NodeList{}
	err := k8s.List(ctx, nodeList)
	if err != nil {
		return false, err
	}
	for _, node := range nodeList.Items {
		labels := node.GetLabels()
		for k := range labels {
			if strings.HasPrefix(k, kubernetesFlavorOKEOCINodeLabelPrefix) {
				return true, nil
			}
		}
	}
	return false, nil

}
