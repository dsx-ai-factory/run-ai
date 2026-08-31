/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pevent

import (
	"context"
	"reflect"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// don't remove fields so we can be backward compatible, add new fields and handle the deprecation of the old ones
type Event[T any] struct {
	CtxMetaData map[string]interface{}
	Payload     *T
	Event       *T // deprecated
}

func New[T any]() *Event[T] {
	return &Event[T]{
		CtxMetaData: make(map[string]interface{}),
		Payload:     new(T),
	}
}

func (msg *Event[T]) GetTopicName() string {
	return reflect.TypeOf(msg.Payload).Elem().Name()
}

func (msg *Event[T]) WithCtx(ctx context.Context) *Event[T] {
	msg.CtxMetaData = runaictx.ToMap(ctx)
	return msg
}

func (msg *Event[T]) WithPayload(payload *T) *Event[T] {
	msg.Payload = payload
	return msg
}

func (msg *Event[T]) ExtractCtx() context.Context {
	return runaictx.FromMap(context.TODO(), msg.CtxMetaData)
}

func (e *Event[T]) GetPayload() *T {
	if isZeroValue(e.Payload) {
		return e.Event
	}
	return e.Payload
}

func isZeroValue[T any](val *T) bool {
	if val == nil {
		return true
	}
	zero := reflect.Zero(reflect.TypeOf(*val)).Interface()
	return reflect.DeepEqual(*val, zero)
}
