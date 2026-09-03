/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package options

import (
	"net/url"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
)

const (
	queryParamPod = "pod"
)

type Options struct {
	Project           string
	WorkloadName      string
	WorkloadType      api.WorkloadType
	WorkloadFramework api.Framework
	Address           string
	Pod               string
	PodRunningTimeout time.Duration
	Ports             []string
}

func (o *Options) ToQueryParams() (*url.Values, error) {
	vars := url.Values{}
	if o.Pod != "" {
		vars.Set(queryParamPod, o.Pod)
	}
	return &vars, nil
}
