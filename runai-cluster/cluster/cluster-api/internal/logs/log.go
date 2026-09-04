/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"context"
	"io"

	"k8s.io/client-go/kubernetes"
)

type PodStreamer struct {
	client kubernetes.Interface
}

func NewPodStreamer(client kubernetes.Interface) *PodStreamer {
	return &PodStreamer{client: client}
}

func (p *PodStreamer) Stream(ctx context.Context, namespace string, options *Options) (io.ReadCloser, error) {
	logOptions, err := options.ToK8sLogsOptions()
	if err != nil {
		return nil, err
	}

	return p.client.CoreV1().Pods(namespace).GetLogs(options.Pod, logOptions).Stream(ctx)
}
