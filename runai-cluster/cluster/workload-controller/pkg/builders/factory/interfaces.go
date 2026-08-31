/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"bytes"
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BuilderOptions struct {
	DryRun bool
	Owner  client.Object
}

// Builder - we have multiple builders for runaijobs, mpijobs, etc. each builder should
// implement the builder interface
type Builder interface {

	//
	// Build - building the resources schemas
	//
	Build(ctx context.Context) error
	//
	// GetJob - return the job object
	//
	GetJob() client.Object

	//
	// Create - creating the resources in k8s
	//
	Create(ctx context.Context) error

	//
	// ToYaml produce a yaml containing all the resources - this is for debug purposes mainly
	//
	ToYaml(ctx context.Context, result *bytes.Buffer) error
}

type CommonResourceBuilderFactory interface {
	CreateCommonResourceBuilder(workloadType constants.WorkloadType, jobType string)
}

type JobBuilderFactory interface {
	CreateJobBuilder(workloadType constants.WorkloadType)
}
