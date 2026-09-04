/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

type ParametrizedQuery struct {
	Query  string
	Params []any
}

func NewParametrizedQuery() *ParametrizedQuery {
	return &ParametrizedQuery{
		Params: make([]any, 0),
	}
}

func (nvq *ParametrizedQuery) Add(query string, params ...any) *ParametrizedQuery {
	nvq.Query += " " + query
	nvq.Params = append(nvq.Params, params...)
	return nvq
}
