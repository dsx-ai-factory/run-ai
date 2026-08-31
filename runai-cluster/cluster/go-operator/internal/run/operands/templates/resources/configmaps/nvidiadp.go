/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	"errors"
	"fmt"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewNvidiaDevicePlugin(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *NvidiaDevicePlugin {
	return &NvidiaDevicePlugin{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type NvidiaDevicePlugin struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (nvidiadp *NvidiaDevicePlugin) OperandName() string {
	return nvidiadp.operandName
}

func (nvidiadp *NvidiaDevicePlugin) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = "nvidia-device-plugin-config"
		namespace = nvidiadp.runaiConfig.Namespace
	)

	configMap := &corev1.ConfigMap{}

	// Get the existing configMap if it exists to consume any cluster-set values
	err := nvidiadp.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, err
	}

	configMap.Name = name
	configMap.Namespace = namespace

	return configMap, nil
}

func (nvidiadp *NvidiaDevicePlugin) Status(ctx context.Context) error {
	configMap, err := nvidiadp.Object(ctx)
	if err != nil {
		return err
	}

	if configMap.GetUID() == "" {
		return errors.New(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName()))
	}

	return nil
}

func (nvidiadp *NvidiaDevicePlugin) Enabled() bool {
	return true
}
