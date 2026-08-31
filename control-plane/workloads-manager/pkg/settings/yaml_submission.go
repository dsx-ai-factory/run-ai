/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type YamlSubmission struct {
	Enabled bool `json:"enabled"`
}

func NewYamlSubmission() manager.Settinger {
	return &YamlSubmission{}
}

func (l *YamlSubmission) Category() string {
	return WorkloadsCategory
}

func (l *YamlSubmission) Description() string {
	return "Enable to submit a range of workload types using a standard Kubernetes YAML."
}

func (l *YamlSubmission) DisplayName() string {
	return "Submit supported workload types via YAML"
}

func (l *YamlSubmission) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_YAML_SUBMISSION)
}

func (l *YamlSubmission) SetDefault() {
	l.Enabled = true
}

func (l *YamlSubmission) Stage() manager.Stage {
	return manager.GA
}

func (l *YamlSubmission) Validate(value interface{}) bool {
	return true
}
