/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GvkWorkloadHandler struct {
	gvk          schema.GroupVersionKind
	cachedClient client.Client

	kartaTopOwnerHandler *KartaTopOwnerHandler
}

func (h *GvkWorkloadHandler) ShouldUpdateNetworkStatus() bool {
	return h.gvk.Group == "leaderworkerset.x-k8s.io" && h.gvk.Kind == "LeaderWorkerSet"
}

func NewGvkWorkloadHandler(cachedClient client.Client, draParser *dra.DRAParser, gvk schema.GroupVersionKind) *GvkWorkloadHandler {
	kartaTopOwnerHandler := NewKartaTopOwnerHandler(cachedClient, draParser)

	return &GvkWorkloadHandler{
		gvk:                  gvk,
		cachedClient:         cachedClient,
		kartaTopOwnerHandler: kartaTopOwnerHandler,
	}
}

func (h *GvkWorkloadHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	gvkObject := &unstructured.Unstructured{}
	gvkObject.SetGroupVersionKind(h.gvk)

	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return gvkObject })
}

func (h *GvkWorkloadHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	kartaOfCrd, err := h.getKartaOfGvk(ctx, object.GetObjectKind().GroupVersionKind())
	if err != nil {
		log.Error().Msgf("Failed to get Karta of GVK %v: %v", h.gvk, err)
		return err
	}

	log.Info().Msgf("Found Karta <%s> for GVK <%v>, parsing object <%s/%s>", kartaOfCrd.Name, h.gvk, object.GetNamespace(), object.GetName())

	return h.kartaTopOwnerHandler.Parse(ctx, object, kartaOfCrd, externalWorkload)
}

func (h *GvkWorkloadHandler) getKartaOfGvk(ctx context.Context, gvk schema.GroupVersionKind) (*kartav1alpha1.Karta, error) {
	kartaOfCrd, err := utils.GetKartaOfGvk(ctx, h.cachedClient, gvk)
	if err != nil {
		return nil, err
	}

	if kartaOfCrd.DeletionTimestamp != nil {
		return nil, fmt.Errorf("karta of gvk is being deleted")
	}
	return kartaOfCrd, nil
}
