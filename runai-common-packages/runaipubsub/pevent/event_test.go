/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pevent

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

const (
	tenantIdValue           = "tenant-id"
	subjectValue            = "subject"
	groupsValue1            = "group1"
	groupsValue2            = "group2"
	supplementalGroupsValue = "supplementalGroups"
)

type TestEvent struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestMagCreation(t *testing.T) {
	testEventActual := &TestEvent{Name: "test", Value: 1}
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdValue)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectValue)
	ctx = context.WithValue(ctx, runaictx.CfGroups, []string{groupsValue1, groupsValue2})
	ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, supplementalGroupsValue)
	msg := New[TestEvent]().WithPayload(testEventActual).WithCtx(ctx)
	assert.Equal(t, msg.Payload, testEventActual)
	assert.Equal(t, msg.CtxMetaData["tenant_id"], tenantIdValue)
	assert.Equal(t, msg.CtxMetaData["subject_id"], subjectValue)
	assert.Equal(t, msg.CtxMetaData["groups"], []string{groupsValue1, groupsValue2})
	assert.Equal(t, msg.CtxMetaData["supplemental_groups"], supplementalGroupsValue)
}

func TestCtxCreation(t *testing.T) {
	testEventActual := &TestEvent{Name: "test", Value: 1}
	msg := New[TestEvent]().WithPayload(testEventActual)
	msg.CtxMetaData["tenant_id"] = tenantIdValue
	msg.CtxMetaData["subject_id"] = subjectValue
	msg.CtxMetaData["groups"] = []string{groupsValue1, groupsValue2}
	msg.CtxMetaData["supplemental_groups"] = supplementalGroupsValue
	ctx := msg.ExtractCtx()
	assert.Equal(t, ctx.Value(runaictx.CfTenantId), tenantIdValue)
	assert.Equal(t, ctx.Value(runaictx.CfSubjectID), subjectValue)
	assert.Equal(t, ctx.Value(runaictx.CfGroups), []string{groupsValue1, groupsValue2})
	assert.Equal(t, ctx.Value(runaictx.CfSupplementalGroups), supplementalGroupsValue)
}

func TestGetPayload_ReturnsPayload(t *testing.T) {
	e := &Event[TestEvent]{
		Payload: &TestEvent{Name: "fallback", Value: 1},
		Event:   &TestEvent{},
	}
	result := e.GetPayload()
	assert.Equal(t, e.Payload, result)
}

func TestGetPayload_ReturnsEvent(t *testing.T) {
	e := &Event[TestEvent]{
		Payload: &TestEvent{},
		Event:   &TestEvent{Name: "fallback", Value: 1},
	}
	result := e.GetPayload()
	assert.Equal(t, e.Event, result)
}

func TestIsZeroValue_NilPointer(t *testing.T) {
	var val *TestEvent
	assert.True(t, isZeroValue(val))
}

func TestIsZeroValue_ZeroValueStruct(t *testing.T) {
	val := &TestEvent{}
	assert.True(t, isZeroValue(val))
}

func TestIsZeroValue_NonZeroValueStruct(t *testing.T) {
	val := &TestEvent{Name: "test", Value: 1}
	assert.False(t, isZeroValue(val))
}
