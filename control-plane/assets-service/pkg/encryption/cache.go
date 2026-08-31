/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"sync"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	corev1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var once sync.Once

func initTenantKeysCache() {
	once.Do(func() {
		err := cache.RegisterStore[corev1.Secret](options.Ttl(runai_config.Get[*config.TenantKeysCache]().Expiration()))
		if err != nil {
			logging.Warnf("Failed to register cache store for type corev1/secret: %v", err)
		}
	})
}
