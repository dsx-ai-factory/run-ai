/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"

	"go.uber.org/zap"
)

func isDistributedFrameworkSuspendable(extensions extensions.Group, framework string) (bool, error) {
	if extensions == nil {
		return false, fmt.Errorf("cluster dependencies are nil")
	}

	frameworkExtension, ok := distributed.JobKindToExtensionMap[framework]
	if !ok {
		return false, fmt.Errorf("unable to find the extension for the requested framework %s", framework)
	}
	status, ok := extensions[frameworkExtension]
	if !ok {
		return false, fmt.Errorf("missing extension framework status for %s", framework)
	}
	if !status.IsEnabled() {
		zap.S().Warnf("Framework %s is not enabled treat as non suspeandable", framework)
		return false, nil
	}
	return status.IsCapable(distributed.Suspension), nil
}
