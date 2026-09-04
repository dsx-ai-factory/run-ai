/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// gvkComponentId returns the canonical component id for a mapping dependency:
// "group/kind/version".
func gvkComponentId(gvk api.Gvk) string {
	return fmt.Sprintf("%s/%s/%s", convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version)
}
