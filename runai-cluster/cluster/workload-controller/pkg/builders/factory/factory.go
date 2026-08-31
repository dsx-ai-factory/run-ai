/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RunaiJobBuilder   = "runaiJobBuilder"
	InferenceBuilder  = "inferenceBuilder"
	MpiOldBuilder     = "mpiJobBuilder" // MPI builder based on Job workload -> should be removed
	MpiBuilder        = "mpiBuilder"    // MPI builder based on Distributed workload
	PyTorchBuilder    = "pyTorchBuilder"
	TFJobBuilder      = "TFJobBuilder"
	JAXJobBuilder     = "JAXJobBuilder"
	XGBoostJobBuilder = "XGBoostJobBuilder"
)

// BuildersRegistry maps builder nick --> to the constructor that creates it
// a builder's constructor receive two pointers
//
//	a fromObject -> e.g. runaiJobbuilder
//	a toResources -> e.g. RunaiJob
var BuildersRegistry = map[string]func(fromObject interface{}, toResources interface{},
	k8sClient client.Client, options BuilderOptions,
	podLabels map[string]string, podAnnotations map[string]string) Builder{}
