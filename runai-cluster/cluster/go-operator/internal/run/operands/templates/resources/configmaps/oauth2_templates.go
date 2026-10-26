/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	_ "embed"
	"fmt"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed templates/error.html
var oauth2ErrorPageHTML string

func NewOAuth2Templates(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *OAuth2Templates {
	return &OAuth2Templates{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type OAuth2Templates struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (ot *OAuth2Templates) OperandName() string {
	return ot.operandName
}

func (ot *OAuth2Templates) Object(ctx context.Context) (client.Object, error) {
	configMap := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: oauth2proxy.ConfigMapName, Namespace: ot.runaiConfig.Namespace}
	if err := ot.k8sReader.Get(ctx, key, configMap); err != nil && !kerrors.IsNotFound(err) {
		return nil, fmt.Errorf("get OAuth2 templates configmap %s: %w", key, err)
	}

	ownerKind := runv1.GroupVersion.WithKind(runv1.RunaiConfigKind).GroupKind()
	if owner := metav1.GetControllerOf(configMap); owner != nil &&
		(schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind).GroupKind() != ownerKind ||
			owner.Name != ot.runaiConfig.Name || owner.UID != ot.runaiConfig.UID) {
		return nil, fmt.Errorf("configmap %s is controlled by %s %s, refusing to overwrite", key, owner.Kind, owner.Name)
	}
	// Remote resources use owner annotations instead of owner references.
	for annotation, expected := range map[string]string{
		runv1.NamespacedNameAnnotation: client.ObjectKeyFromObject(ot.runaiConfig).String(),
		runv1.TypeAnnotation:           ownerKind.String(),
	} {
		if actual, exists := configMap.Annotations[annotation]; exists && actual != expected {
			return nil, fmt.Errorf("configmap %s has conflicting %s ownership annotation", key, annotation)
		}
	}

	configMap.Name = key.Name
	configMap.Namespace = key.Namespace
	if configMap.Labels == nil {
		configMap.Labels = map[string]string{}
	}
	configMap.Labels[commonconstants.ClusterWideResourceLabel] = "true"
	configMap.Data = map[string]string{oauth2proxy.TemplateKey: oauth2ErrorPageHTML}
	configMap.BinaryData = nil
	return configMap, nil
}

func (ot *OAuth2Templates) Status(ctx context.Context) error {
	configMap := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: oauth2proxy.ConfigMapName, Namespace: ot.runaiConfig.Namespace}
	if err := ot.k8sReader.Get(ctx, key, configMap); err != nil {
		return fmt.Errorf("get OAuth2 templates configmap %s: %w", key, err)
	}
	if configMap.Data[oauth2proxy.TemplateKey] != oauth2ErrorPageHTML {
		return fmt.Errorf("configmap %s does not contain the current OAuth2 error page", key)
	}
	if configMap.Labels[commonconstants.ClusterWideResourceLabel] != "true" {
		return fmt.Errorf("configmap %s is not labelled for cluster-wide replication", key)
	}
	return nil
}

func (ot *OAuth2Templates) Enabled() bool {
	return true
}
