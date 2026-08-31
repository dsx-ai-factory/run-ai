/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

import (
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
)

type Appender[T any] struct {
	DuplicatorMapper
	list          []T
	lastCreatedAt time.Time
}

func NewAppender[T any](limit int) *Appender[T] {
	return &Appender[T]{
		DuplicatorMapper: NewIdsMapper(limit),
		list:             make([]T, 0), // Initialize as empty slice, not nil
	}
}

// AddPage add the workload to the page appender with duplicate ID validation
func (a *Appender[T]) AddPage(entries []T) error {
	var (
		ea  EntryAdapter
		err error
	)

	for _, entry := range entries {
		ea, err = GetEntryAdapter(entry)
		if err != nil {
			return err
		}
		if a.ReachedLimit() {
			// reached limit, set last enty creation time
			return a.setLastCreatedAt(ea)
		}
		a.add(ea.GetId(), entry)
	}

	// reached limit, set last enty creation time
	return nil
}

func (a *Appender[T]) GetList() []T {
	return a.list
}

func (a *Appender[T]) add(id string, entry T) bool {
	if !a.ShouldAdd(id) {
		return false
	}

	a.list = append(a.list, entry)
	return true
}

func (a *Appender[T]) GetNextToken() (*structs.NextToken, error) {
	if a.lastCreatedAt.IsZero() {
		return nil, nil
	}
	return structs.NextTokenFromTimestamp(a.lastCreatedAt)
}

func (a *Appender[T]) setLastCreatedAt(ea EntryAdapter) error {
	if ea == nil {
		return nil
	}
	if ea.GetCreatedAt().IsZero() {
		return ErrorNextTokenIsZero
	}
	a.lastCreatedAt = ea.GetCreatedAt()
	return nil
}
