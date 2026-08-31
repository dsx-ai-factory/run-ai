/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cachedclient

import (
	"context"

	"github.com/rs/zerolog/log"

	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CachedClient interface {
	// Client - Get and List would be cached, other operations would be done on the writer client
	client.Client

	StartCache(ctx context.Context)
	WaitForCachesToSync(ctx context.Context, stopCh <-chan struct{})
}

type {{cookiecutter.__app_name_camel_case_upper}}CachedClient struct {
	client.Client
	cachedClient        cache.Cache
	scheme              *apiruntime.Scheme
}

func NewCachedClient(ctx context.Context, writerClient client.Client,
	kubeConfig *rest.Config, scheme *apiruntime.Scheme) CachedClient {
	cachedClient := createCache(ctx, kubeConfig, scheme)

	return &{{cookiecutter.__app_name_camel_case_upper}}CachedClient{
		Client:              writerClient,
		cachedClient:        cachedClient,
		scheme:              scheme,
	}
}

// Get is a cached operation
func (c *{{cookiecutter.__app_name_camel_case_upper}}CachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object,
	opts ...client.GetOption) error {
	return c.cachedClient.Get(ctx, key, obj, opts...)
}

// List is a cached operation
func (c *{{cookiecutter.__app_name_camel_case_upper}}CachedClient) List(ctx context.Context, list client.ObjectList,
	opts ...client.ListOption) error {
	return c.cachedClient.List(ctx, list, opts...)
}

func (c *{{cookiecutter.__app_name_camel_case_upper}}CachedClient) StartCache(ctx context.Context) {
	go func() {
		if err := c.cachedClient.Start(ctx); err != nil {
			log.Fatal().Msgf("Failed to start cached client, error: %s", err.Error())
		}
	}()
}

func (c *{{cookiecutter.__app_name_camel_case_upper}}CachedClient) WaitForCachesToSync(ctx context.Context, stopCh <-chan struct{}) {
	if !c.cachedClient.WaitForCacheSync(ctx) {
		log.Fatal().Msgf("failed waiting for cache to sync")
	}
}

func createCache(ctx context.Context, kubeConfig *rest.Config, scheme *apiruntime.Scheme) cache.Cache {
	cacheOptions := cache.Options{
		Scheme: scheme,
	}
	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		log.Fatal().Msgf("Failed to create cached client, error: %s", err.Error())
	}

	return cachedClient
}
