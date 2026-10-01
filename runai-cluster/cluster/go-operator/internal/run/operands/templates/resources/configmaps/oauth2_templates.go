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
