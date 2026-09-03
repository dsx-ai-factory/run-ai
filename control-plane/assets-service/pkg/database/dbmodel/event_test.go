/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewEventModelFromSyncApi(t *testing.T) {
	// Setup test data
	now := time.Now()
	clusterId := uuid.New()
	assetId := uuid.New()

	// Create two events for the same PVC in different namespaces
	events := []*api.EventSyncFields{
		{
			Id:        uuid.New().String(),
			ClusterId: clusterId.String(),
			Message:   "waiting for first consumer to be created",
			Reason:    WAIT_FOR_FIRST_CONSUMER,
			CreatedAt: now,
			Type:      "Normal",
			Source:    "source1",
			InvolvedObject: api.InvolvedObject{
				Kind:      PVC_OBJECT_KIND,
				Namespace: "runai",
				Name:      "pvc1",
				Uid:       "uid1",
			},
		},
		{
			Id:        uuid.New().String(),
			ClusterId: clusterId.String(),
			Message:   "waiting for first consumer to be created",
			Reason:    "DifferentReason",
			CreatedAt: now,
			Type:      "Normal",
			Source:    "source1",
			InvolvedObject: api.InvolvedObject{
				Kind:      PVC_OBJECT_KIND,
				Namespace: "runai-team-a",
				Name:      "pvc1",
				Uid:       "uid1",
			},
		},
	}

	// Convert both events and verify their messages
	for _, event := range events {
		got := NewEventModelFromSyncApi(1, event, assetId.String(), "scope1", "type1")

		expectedMessage := event.Message
		if event.Reason == WAIT_FOR_FIRST_CONSUMER {
			expectedMessage = expectedMessage + " (" + event.InvolvedObject.Namespace + ")"
		}

		assert.Equal(t, expectedMessage, got.Message)
		assert.Equal(t, event.InvolvedObject.Namespace, got.ObjectNamespace)
		assert.Equal(t, PVC_OBJECT_KIND, got.ObjectKind)
		assert.Equal(t, "pvc1", got.ObjectName)
	}
}
