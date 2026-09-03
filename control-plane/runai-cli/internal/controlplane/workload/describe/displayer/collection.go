/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

type CollectionBuilder[T any] struct {
	items []T
}

func NewCollectionBuilder[T any]() *CollectionBuilder[T] {
	return &CollectionBuilder[T]{items: []T{}}
}

func (b *CollectionBuilder[T]) Add(item T) *CollectionBuilder[T] {
	b.items = append(b.items, item)
	return b
}

func (b *CollectionBuilder[T]) Set(items []T) *CollectionBuilder[T] {
	b.items = items
	return b
}

func (b *CollectionBuilder[T]) Build() []T {
	return b.items
}

type PodsBuilder = CollectionBuilder[Pod]
type EventsBuilder = CollectionBuilder[Event]
type NetworksBuilder = CollectionBuilder[Network]
type IssuesBuilder = CollectionBuilder[Issue]

func NewPodsBuilder() *PodsBuilder {
	return NewCollectionBuilder[Pod]()
}

func NewEventsBuilder() *EventsBuilder {
	return NewCollectionBuilder[Event]()
}

func NewNetworksBuilder() *NetworksBuilder {
	return NewCollectionBuilder[Network]()
}

func NewIssuesBuilder() *IssuesBuilder {
	return NewCollectionBuilder[Issue]()
}
