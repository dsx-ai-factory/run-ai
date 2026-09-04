/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package projects

import (
	"sync"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

var once sync.Once

func initProjectMapperCache() {
	once.Do(func() {
		err := cache.RegisterStore[BiDirectionalMapping](
			options.Ttl(runai_config.Get[*config.ProjectsCache]().Expiration()),
			options.MaxRecordSizeKB(runai_config.Get[*config.ProjectsCache]().MaxRecordSizeKB),
		)
		if err != nil {
			logging.Warnf("Failed to register cache store for type bi directional mapping: %v", err)
		}
	})
}
