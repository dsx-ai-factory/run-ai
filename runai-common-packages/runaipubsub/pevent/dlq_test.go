/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pevent

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	config.Set(&config_models.Pubsub{
		ServiceName: "test-service",
	})
	defer config.FlushAll()

	runaiEvent := NewDlq("Test error", "TestTopic", []byte("Test body"))
	assert.Equal(t, "Test error", runaiEvent.Payload.Error)
	assert.Equal(t, "TestTopic", runaiEvent.Payload.OriginTopic)
	assert.Equal(t, "test-service", runaiEvent.Payload.Consumer)
	assert.Equal(t, []byte("Test body"), runaiEvent.Payload.Body)
	assert.NotNil(t, runaiEvent.Payload.Timestamp)
}

func TestGetTopicName(t *testing.T) {
	config.Set(&config_models.Pubsub{
		ServiceName: "test-service",
	})
	defer config.FlushAll()

	result := GetDlqTopicName()
	assert.Equal(t, "TestServiceDlq", result)
}

func TestToPascalCase_KebabCase(t *testing.T) {
	result := toPascalCase("test-service")
	assert.Equal(t, "TestService", result)
}

func TestToPascalCase_SnakeCase(t *testing.T) {
	result := toPascalCase("test_service")
	assert.Equal(t, "TestService", result)
}

func TestToPascalCase_SpaceSeparated(t *testing.T) {
	result := toPascalCase("test service")
	assert.Equal(t, "TestService", result)
}

func TestToPascalCase_CamelCase(t *testing.T) {
	result := toPascalCase("testService")
	assert.Equal(t, "TestService", result)
}
