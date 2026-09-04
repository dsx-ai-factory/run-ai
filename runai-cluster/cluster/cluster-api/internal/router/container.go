/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	openapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
)

type Container struct {
	Router                       chi.Router
	WorkloadMiddleware           func(http.Handler) http.Handler
	ProjectMiddleware            func(http.Handler) http.Handler
	WorkloadAttributesMiddleware func(http.Handler) http.Handler
	AuthzMiddleware              func(http.Handler) http.Handler
	AuthMiddleware               func(http.Handler) http.Handler
	MetricsMiddleware            func(http.Handler) http.Handler
	K8sObjectMiddleware          func(http.Handler) http.Handler
	K8sObjectService             openapi.K8sObjectAPIServicer
	MapHandler                   map[string]UrlHandler
}

func NewContainer(r chi.Router, workloadMiddleware, projectMiddleware,
	workloadAttributesMiddleware, authzMiddleware,
	authMiddleware, metricsMiddleware func(http.Handler) http.Handler, MapHandler map[string]UrlHandler,
) Container {
	return Container{
		Router:                       r,
		WorkloadMiddleware:           workloadMiddleware,
		ProjectMiddleware:            projectMiddleware,
		WorkloadAttributesMiddleware: workloadAttributesMiddleware,
		AuthzMiddleware:              authzMiddleware,
		AuthMiddleware:               authMiddleware,
		MetricsMiddleware:            metricsMiddleware,
		MapHandler:                   MapHandler,
	}
}

type ContainerBuilder struct {
	container Container
}

func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{}
}

func (b *ContainerBuilder) WithRouter(r chi.Router) *ContainerBuilder {
	b.container.Router = r
	return b
}

func (b *ContainerBuilder) WithWorkloadMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.WorkloadMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithProjectMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.ProjectMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithWorkloadAttributesMiddleware(
	middleware func(http.Handler) http.Handler,
) *ContainerBuilder {
	b.container.WorkloadAttributesMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithAuthzMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.AuthzMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithAuthMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.AuthMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithMetricsMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.MetricsMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithHandler(handlers map[string]UrlHandler) *ContainerBuilder {
	b.container.MapHandler = handlers
	return b
}

func (b *ContainerBuilder) WithK8sObjectMiddleware(middleware func(http.Handler) http.Handler) *ContainerBuilder {
	b.container.K8sObjectMiddleware = middleware
	return b
}

func (b *ContainerBuilder) WithK8sObjectService(service openapi.K8sObjectAPIServicer) *ContainerBuilder {
	b.container.K8sObjectService = service
	return b
}

func (b *ContainerBuilder) Build() Container {
	return b.container
}
