/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sync"
)

const (
	Jax        extensions.Type = runv1.RunaiPublicJax
	Pytorch    extensions.Type = runv1.RunaiPublicPyTorch
	Tensorflow extensions.Type = runv1.RunaiPublicTF
	MPI        extensions.Type = runv1.RunaiPublicMPI
	XGBoost    extensions.Type = runv1.RunaiPublicXGBoost
)

const (
	Suspension extensions.CapabilityType = "suspension"
)

var SupportedExtensions = extensions.Group{
	Jax:        extensions.Extension{},
	Pytorch:    extensions.Extension{},
	Tensorflow: extensions.Extension{},
	MPI:        extensions.Extension{},
	XGBoost:    extensions.Extension{},
}
var ExtensionToJobKindMap = map[extensions.Type]string{
	MPI:        constants.KindMpiJob,
	Pytorch:    constants.KindPyTorchJob,
	Tensorflow: constants.KindTFJob,
	XGBoost:    constants.KindXGBoostJob,
	Jax:        constants.KindJAXJob,
}

var JobKindToExtensionMap = func() map[string]extensions.Type {
	var reverseMapOnce sync.Once
	var reverseMap map[string]extensions.Type

	reverseMapOnce.Do(func() {
		reverseMap = make(map[string]extensions.Type)
		for key, value := range ExtensionToJobKindMap {
			reverseMap[value] = key
		}
	})
	return reverseMap
}()
