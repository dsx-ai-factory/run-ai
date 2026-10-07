/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
)

func GetEncryptionKey() ([]byte, error) {
	key, err := config.GetAdsEncryptionKey()
	if err != nil {
		return nil, err
	}
	return key, nil
}
