/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"bytes"
	"context"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var TestFixedDate = metav1.NewTime(time.Date(2022, 2, 22, 17, 0, 0, 0, time.UTC))

type ResourceBuilderOptions struct {
	Owner client.Object
}

type ExternalBuilderOptions struct {
	Owner          client.Object
	DisableAuth    *bool
	AuthPortOffset int
}

type ComputeBuilderOptions struct {
	Owner           client.Object
	ExitingPodsSpec map[constants.ReplicaType]*v1.PodSpec
}

type PreviewBuilder interface {
	GetPreview(ctx context.Context, workload *job.JobWorkload) ([]common.Resource, error)
}

type ExternalBuilder interface {
	Build(ctx context.Context, workload *job.JobWorkload, options *ExternalBuilderOptions) ([]client.Object, error)
	PreviewBuilder
}

type ResourceBuilder interface {

	//
	// Build - building the resource schemas
	//
	Build(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]client.Object, error)

	PreviewBuilder
}

type ComputeBuilder interface {
	Build(ctx context.Context, workload *job.JobWorkload, options *ComputeBuilderOptions) (*ComputeResources, error)

	PreviewBuilder
}

// used for testing to dump the resource to yaml
type ResourceYaml interface {
	ToYaml(objects []client.Object, w *bytes.Buffer) error
}
