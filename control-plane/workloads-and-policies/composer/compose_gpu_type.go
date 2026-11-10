// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package composer

import (
	"net/http"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"k8s.io/apimachinery/pkg/util/validation"
)

const NvidiaGPUProductLabel = "nvidia.com/gpu.product"

func (fc *FieldsComposer) ComposeGPUType(src *string, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}
		if *src == "" || len(validation.IsValidLabelValue(*src)) != 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "gpuType must be an exact, nonempty nvidia.com/gpu.product label value")
		}
		return fc.CopyString(src, target, setter, cookie)()
	}
}
