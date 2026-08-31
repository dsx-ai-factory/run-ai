/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	ResourcePods    = "pods"
	SubResourceExec = "exec"
)

type PodExecute struct {
	client kubernetes.Interface
}

func NewPodExecute(client kubernetes.Interface) *PodExecute {
	return &PodExecute{client: client}
}

func (p *PodExecute) CreateRequest(namespace string, options *stream.ExecOptions) (*rest.Request, error) {
	execOptions, err := options.ToK8sPodExecOptions()
	if err != nil {
		return nil, err
	}

	req := p.client.CoreV1().RESTClient().Post().
		Resource(ResourcePods).
		Namespace(namespace).
		Name(options.Pod).
		SubResource(SubResourceExec)
	req.VersionedParams(
		execOptions,
		scheme.ParameterCodec,
	)

	return req, nil
}
