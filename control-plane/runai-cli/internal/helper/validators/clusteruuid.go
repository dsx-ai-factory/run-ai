/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
)

const reconfigureMessage = "to fix this, run `runai cluster` to reconfigure it."

// ValidateClusterUUID validates that the cluster UUID is aligned with the cluster name.
func ValidateClusterUUID(txn *transaction.Txn, clusterUUID, clusterName, cpURL, token string) error {
	logger := config.GetInstance().GetLogger()

	// Sanity check: this code path should never be reached.
	if clusterUUID == "" {
		logger.Debug("clusterUUID is not set, skipping validation")
		return nil
	}

	clusterObj, err := cluster.GetClusterByName(txn, cpURL, token, clusterName)
	if err != nil {
		logger.Debug("failed to get the cluster by name.", zap.String("clusterName", clusterName),
			zap.Error(err))
		return fmt.Errorf("cluster '%s' does not exist or is misconfigured. %s", clusterName,
			reconfigureMessage)
	}

	// Sanity check: this code path should never be reached.
	if clusterObj == nil {
		return fmt.Errorf("cluster parsing error. %s", reconfigureMessage)
	}

	if clusterObj.Uuid != clusterUUID {
		logger.Debug("configured cluster UUID does not match the cluster UUID",
			zap.String("configuredClusterUUID", clusterUUID), zap.String("clusterUUID", clusterObj.Uuid),
			zap.String("clusterName", clusterName))
		return fmt.Errorf("configured cluster '%s' has a UUID mismatch. %s", clusterName, reconfigureMessage)
	}

	logger.Debug("cluster UUID match cluster name", zap.String("clusterUUID", clusterUUID),
		zap.String("clusterName", clusterName))
	return nil
}
