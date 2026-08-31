/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"

	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	runaicommon "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

// convertImage keeps user-provided scheduler-stack image overrides functional.
// Unset images remain nil so KAI applies its upstream defaults.
func convertImage(image *runaicommon.Image) *kaicommon.Image {
	if image == nil {
		return nil
	}

	repository := image.Registry
	if repository == nil && (image.Repository != nil || image.Env != nil) {
		legacyRepository := ptr.Deref(image.Repository, runaicommon.DefaultRepositeory)
		legacyEnv := ptr.Deref(image.Env, runaicommon.DefaultEnv)
		repository = ptr.To(fmt.Sprintf("%s%s", legacyRepository, legacyEnv))
	}

	converted := &kaicommon.Image{
		Repository: repository,
		Name:       image.Name,
		Tag:        image.Tag,
	}
	if image.PullPolicy != "" {
		converted.PullPolicy = ptr.To(corev1.PullPolicy(image.PullPolicy))
	}
	return converted
}
