/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

import (
	kfv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	TrainingOperatorTagsConfigMapName = "training-operator-tags-config"
	TrainingOperatorMinSuspendVersion = "1.8.1"
	MpiOperatorMinSuspendVersion      = "0.4"
	TrainingOperatorNamespace         = "kubeflow"
	TrainingOperatorDeploymentName    = "training-operator"
	MpiOperatorNamespace              = "mpi-operator"
	MpiOperatorDeploymentName         = "mpi-operator"
	DefaultLWSNamespace               = "lws-system"
	LWSDeploymentName                 = "lws-controller-manager"
	EnableSchemeFlag                  = "--enable-scheme"
)

var TrainingOperatorFrameworksToKindMap = map[string]string{
	runv1.RunaiPublicTF:      kfv1.TFJobKind,
	runv1.RunaiPublicPyTorch: kfv1.PyTorchJobKind,
	runv1.RunaiPublicXGBoost: kfv1.XGBoostJobKind,
	runv1.RunaiPublicJax:     kfv1.JAXJobKind,
}
var TrainingOperatorFrameworksToSingularMap = map[string]string{
	runv1.RunaiPublicTF:      kfv1.TFJobSingular,
	runv1.RunaiPublicPyTorch: kfv1.PyTorchJobSingular,
	runv1.RunaiPublicXGBoost: kfv1.XGBoostJobSingular,
	runv1.RunaiPublicJax:     kfv1.JAXJobSingular,
}
