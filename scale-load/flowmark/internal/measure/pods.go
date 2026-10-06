/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"regexp"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var PodKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}

var podIndexSuffix = regexp.MustCompile(`-\d+-\d+$`)

func WatchWorkloadPods(ctx context.Context, kube k8sclient.WithWatch, namespace string) (*Watcher, error) {
	return Watch(ctx, kube, PodKind, namespace, podWorkloadName)
}

func podWorkloadName(pod *unstructured.Unstructured) string {
	if owners := pod.GetOwnerReferences(); len(owners) > 0 && owners[0].Name != "" {
		return owners[0].Name
	}
	return podIndexSuffix.ReplaceAllString(pod.GetName(), "")
}
