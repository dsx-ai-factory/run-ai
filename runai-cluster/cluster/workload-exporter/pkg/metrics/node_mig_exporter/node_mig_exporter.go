/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node_mig_exporter

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	GpuProductLabel  = "nvidia.com/gpu.product"
	MigStrategyLabel = "nvidia.com/mig.strategy"
	MigEnabledLabel  = "node-role.kubernetes.io/runai-mig-enabled"
	TrueLabelValue   = "true"

	MigStrategyNone   = "none"
	MigStrategySingle = "single"
	MigStrategyMixed  = "mixed"

	DefaultMigStrategy = MigStrategyNone

	// MIG device families
	A100 = "A100"
	H100 = "H100"
)

// MigCapableLabel is set to "true" by GPU Feature Discovery on nodes whose GPUs support MIG.
const MigCapableLabel = "nvidia.com/mig.capable"

type NodeMigExporter struct {
	node *v1.Node
}

func NewMigStrategyExporter(node *v1.Node) *NodeMigExporter {
	return &NodeMigExporter{node: node}
}

func (ne *NodeMigExporter) IsMigNode() bool {
	return ne.getMigStrategy() != MigStrategyNone
}

// getMigStrategy returns the node MIG strategy, or none when the node GPUs do not support MIG.
func (ne *NodeMigExporter) getMigStrategy() string {
	if !ne.isMigCapableLabeled() && !ne.supportMIG() {
		return DefaultMigStrategy
	}

	migStrategyByLabel := ne.getMigStrategyFromLabel()
	switch migStrategyByLabel {
	case MigStrategyMixed:
		if ne.isMigEnabledLabeled() {
			return MigStrategyMixed
		}
	case MigStrategySingle:
		if ne.isMigEnabledLabeled() {
			return MigStrategySingle
		}
	}
	return DefaultMigStrategy
}

func (ne *NodeMigExporter) getMigStrategyFromLabel() string {
	migStrategy, found := ne.node.Labels[MigStrategyLabel]
	if !found || len(migStrategy) == 0 {
		return MigStrategyNone
	}
	if migStrategy != MigStrategyMixed && migStrategy != MigStrategySingle {
		return MigStrategyNone
	}
	return migStrategy
}

func (ne *NodeMigExporter) isMigEnabledLabeled() bool {
	isMigEnabled, found := ne.node.Labels[MigEnabledLabel]
	return found && isMigEnabled == TrueLabelValue
}

// isMigCapableLabeled reports whether GPU Feature Discovery labeled the node with nvidia.com/mig.capable=true.
// When the label is absent or not "true", the caller falls back to the product family list in supportMIG.
func (ne *NodeMigExporter) isMigCapableLabeled() bool {
	return ne.node.Labels[MigCapableLabel] == TrueLabelValue
}

func (ne *NodeMigExporter) supportMIG() bool {
	gpuProduct, found := ne.node.Labels[GpuProductLabel]
	if !found {
		return false
	}
	for _, deviceFamily := range GetMigDeviceFamilies() {
		if strings.Contains(gpuProduct, deviceFamily) {
			return true
		}
	}
	return false
}

// GetMigDeviceFamilies returns the list of GPU device families that support MIG.
func GetMigDeviceFamilies() []string {
	return []string{A100, H100}
}
