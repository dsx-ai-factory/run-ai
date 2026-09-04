/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InformersApi interface {
	RemoveInformer(ctx context.Context, gvk schema.GroupVersionKind) error
	GetObjectInformers() map[schema.GroupVersionKind]cachedclient.CachedInformer
}

type CachedAdditionalGvksApi interface {
	AddAdditionalGroupKind(gk schema.GroupKind)
	RemoveAdditionalGroupKind(gk schema.GroupKind)
}

func getGvksMapPerGroupKind[T any](gvks map[schema.GroupVersionKind]T) map[schema.GroupKind][]schema.GroupVersionKind {
	gkToGvks := map[schema.GroupKind][]schema.GroupVersionKind{}
	for gvk := range gvks {
		gk := gvk.GroupKind()

		if _, ok := gkToGvks[gk]; !ok {
			gkToGvks[gk] = []schema.GroupVersionKind{}
		}

		gkToGvks[gk] = append(gkToGvks[gk], gvk)
	}
	return gkToGvks
}
