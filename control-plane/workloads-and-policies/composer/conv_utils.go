/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"strings"
	"unsafe"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

func NonEmptyStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

type EnumType interface {
	workloadsApi.GpuComputeMode |
		workloadsManagerApi.GpuComputeMode |
		workloadsApi.ImagePullPolicy |
		workloadsApi.SeccompProfileType |
		workloadsApi.UidGidSource |
		workloadsApi.DistributedFramework |
		workloadsApi.CleanPodPolicy |
		workloadsApi.MpiLauncherCreationPolicy |
		workloadsApi.Capability |
		workloadsApi.GitAuthenticationMethod |
		workloadsApi.TolerationOperator |
		workloadsApi.ProbeHandlerScheme |
		workloadsApi.DistributedInferenceStartupPolicy |
		workloadsManagerApi.ImagePullPolicy
}

func EnumToString[T EnumType](s *T) *string {
	if s == nil {
		return nil
	}
	return convutils.Ref(string(*s))
}

// CastType casts a pointer to struct or simple type - to a pointer to the same type taken
// from another service. The main use is for WorkloadsManager service calling Workloads service,
// so we need to cast *WorkloadsManagerApi.NullableString to *WorkloadsApi.NullableString
func CastType[Src any, Dst any](src *Src) *Dst {
	// Safety Check: Ensure the element sizes match.
	var s Src
	var d Dst
	if unsafe.Sizeof(s) != unsafe.Sizeof(d) {
		panic(fmt.Sprintf("size mismatch: cannot cast %T (%d bytes) to %T (%d bytes)",
			s, unsafe.Sizeof(s), d, unsafe.Sizeof(d)))
	}

	return (*Dst)(unsafe.Pointer(src))
}

// CastSlice is the same as CastType, but used for casting arrays rather than type pointers
func CastSlice[Src any, Dst any](src []Src) []Dst {
	// Safety Check: Ensure the element sizes match.
	// If they differ, array indexing (stride) will be completely wrong.
	var s Src
	var d Dst
	if unsafe.Sizeof(s) != unsafe.Sizeof(d) {
		panic(fmt.Sprintf("size mismatch: cannot cast %T (%d bytes) to %T (%d bytes)",
			s, unsafe.Sizeof(s), d, unsafe.Sizeof(d)))
	}

	return *(*[]Dst)(unsafe.Pointer(&src))
}

func CastStruct[Src any, Dst any](src *Src) *Dst {
	// Safety Check: Ensure the element sizes match.
	// If they differ, array indexing (stride) will be completely wrong.
	var s Src
	var d Dst
	if unsafe.Sizeof(s) != unsafe.Sizeof(d) {
		panic(fmt.Sprintf("size mismatch: cannot cast %T (%d bytes) to %T (%d bytes)",
			s, unsafe.Sizeof(s), d, unsafe.Sizeof(d)))
	}

	return (*Dst)(unsafe.Pointer(src))
}
