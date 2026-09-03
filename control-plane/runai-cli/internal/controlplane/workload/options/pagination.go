/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package options

import "k8s.io/utils/ptr"

const (
	DefaultLimit = 100
)

type Paginationer interface {
	GetLimit() int32
	GetOffset() int32
}

type Pagination struct {
	Limit  *int32
	Offset int32
}

func (p *Pagination) GetLimit() int32 {
	return ptr.Deref(p.Limit, DefaultLimit)
}

func (p *Pagination) GetOffset() int32 {
	return p.Offset
}
