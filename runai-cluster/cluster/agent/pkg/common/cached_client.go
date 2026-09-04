/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"github.com/rs/zerolog/log"

	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

func CreateCache(kubeConfig *rest.Config, scheme *apiruntime.Scheme) cache.Cache {
	cacheOptions := cache.Options{
		Scheme: scheme,
	}

	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		log.Fatal().Msgf("Failed to create cached client, error: %s", err.Error())
	}

	return cachedClient
}

func StartCache(ctx context.Context, cachedClient cache.Cache) {
	go func() {
		if err := cachedClient.Start(ctx); err != nil {
			log.Fatal().Msgf("Failed to start cache client, error: %s", err.Error())
		}
	}()

	if !cachedClient.WaitForCacheSync(ctx) {
		log.Fatal().Msgf("Failed waiting for cache to sync")
	}
}
