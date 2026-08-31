/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestDefaultGlobalErrorHandler(t *testing.T) {
	require.NotPanics(t, func() { defaultErrorCallback(errors.New("")) })
}

func TestPublish(t *testing.T) {
	enablePubsub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdValue)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectValue)
	ctx = context.WithValue(ctx, runaictx.CfGroups, []string{groupsValue1, groupsValue2})
	ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, supplementalGroupsValue)
	pubSuberMock := mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberMock
	defer reset()

	testEventActual := &TestEvent{Name: "test", Value: 1}
	runaiMsg := pevent.New[TestEvent]().WithPayload(testEventActual).WithCtx(ctx)
	jsonPayload, _ := json.Marshal(runaiMsg)

	pubSuberMock.On("Publish", "TestEvent", jsonPayload).Return(nil).Once()
	_ = Publish[TestEvent](ctx, testEventActual)
}

func TestSubscribe(t *testing.T) {
	enablePubsub()
	pubSuberMock := mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberMock
	defer reset()

	pubSuberMock.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	_ = Subscribe[TestEvent](func(c context.Context, event *TestEvent) error {
		return nil
	})
}

func TestSubscribeAsync_WithAdditional(t *testing.T) {
	cfg := &config_models.Pubsub{
		Type:                       "nats",
		Url:                        "test",
		EnableAdditionalSubscriber: true,
		AdditionalSubscriberType:   "redis",
		AdditionalSubscriberUrl:    "local",
	}
	require.NoError(t, config.SetWithDefaults(cfg))
	pubSuberMock := mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberMock
	enablePubsub()

	additionalPubSuberMock := mocks.NewPubSuber(t)
	additionalPubSuberInstance = additionalPubSuberMock
	defer reset()

	done := make(chan bool)
	pubSuberMock.On("SubscribeAsync", mock.Anything, "TestEvent", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Once()
	additionalPubSuberMock.On("SubscribeAsync", mock.Anything, "TestEvent", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Once()

	SubscribeAsync(func(c context.Context, event *TestEvent) error { return nil })
	require.True(t, <-done)
	require.True(t, <-done)
}

func TestSubscribeAsync(t *testing.T) {
	cfg := &config_models.Pubsub{
		Type:                       "nats",
		Url:                        "test",
		EnableAdditionalSubscriber: false,
	}
	require.NoError(t, config.SetWithDefaults(cfg))
	enablePubsub()

	done := make(chan bool)
	pubSuberMock := mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberMock
	defer reset()

	pubSuberMock.On("SubscribeAsync", mock.Anything, "TestEvent", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Once()
	SubscribeAsync[TestEvent](func(c context.Context, event *TestEvent) error {
		return nil
	})

	require.True(t, <-done)
}

func TestClose_ClosesInitAsync(t *testing.T) {
	enablePubsub()
	mountSecretsToDir(t)
	pubSuberMock := mocks.NewPubSuber(t)
	pubSuberInstance = pubSuberMock
	t.Setenv("PUBSUB_TYPE", "kafka")
	t.Setenv("PUBSUB_URL", "localhost:12345")
	t.Setenv("SERVICE_NAME", "test-service")

	defer reset()
	pubSuberMock.On("Close").Return(nil).Once()

	errCh, err := InitAsync(nil)
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		select {
		case err, ok := <-errCh:
			require.True(t, ok)
			require.Contains(t, err.Error(), "failed to connect pubsub instance")
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for error from errCh")
		}
	}
	Close()
	require.Equal(t, "context canceled", closeContext.Err().Error())
}

func TestInit_PubsubDisabled(t *testing.T) {
	mountSecretsToDir(t)
	t.Setenv("PUBSUB_ENABLED", "false")
	t.Setenv("PUBSUB_TYPE", "kafka")
	t.Setenv("PUBSUB_URL", "localhost:12345")
	t.Setenv("SERVICE_NAME", "test-service")
	defer func() {
		Close()
		reset()
	}()

	err := Init(nil)
	require.NoError(t, err)
	require.False(t, IsInitialized())
}

func TestInitAsync_PubsubDisabled(t *testing.T) {
	mountSecretsToDir(t)
	t.Setenv("PUBSUB_ENABLED", "false")
	t.Setenv("PUBSUB_TYPE", "kafka")
	t.Setenv("PUBSUB_URL", "localhost:12345")
	t.Setenv("SERVICE_NAME", "test-service")
	defer func() {
		Close()
		reset()
	}()

	errCh, err := InitAsync(nil)
	require.NoError(t, err)
	require.NotNil(t, errCh)
	select {
	case <-errCh:
		t.Fatal("expected no errors on channel when disabled")
	default:
	}
	require.False(t, IsInitialized())
}

func TestPublish_PubsubDisabled(t *testing.T) {
	disablePubsub()
	defer reset()

	ctx := context.Background()
	event := &TestEvent{Name: "test", Value: 1}
	err := Publish[TestEvent](ctx, event)
	require.NoError(t, err)
}

func TestSubscribe_PubsubDisabled(t *testing.T) {
	disablePubsub()
	defer reset()

	err := Subscribe[TestEvent](func(ctx context.Context, event *TestEvent) error {
		return nil
	})
	require.NoError(t, err)
}

func TestSubscribeAsync_PubsubDisabled(t *testing.T) {
	disablePubsub()
	defer reset()

	errCh := SubscribeAsync[TestEvent](func(ctx context.Context, event *TestEvent) error {
		return nil
	})
	require.NotNil(t, errCh)
	select {
	case <-errCh:
		t.Fatal("expected no errors on channel when disabled")
	default:
	}
}

func TestClose_PubsubDisabled(t *testing.T) {
	disablePubsub()
	defer reset()

	err := Close()
	require.NoError(t, err)
}

func TestIsInitialized(t *testing.T) {
	pubSuberInstance = mocks.NewPubSuber(t)
	defer reset()

	require.True(t, IsInitialized())
}

func TestSetMock(t *testing.T) {
	pubSuberMock := mocks.NewPubSuber(t)
	defer reset()

	SetMock(pubSuberMock)
	require.Equal(t, pubSuberMock, pubSuberInstance)
	require.True(t, pubsubEnabled)
}

func TestIsInitialized_FalseBeforeInitialization(t *testing.T) {
	require.False(t, IsInitialized())
}

func enablePubsub() {
	pubsubEnabled = true
}

func disablePubsub() {
	pubsubEnabled = false
}

func reset() {
	pubsubEnabled = false
	pubSuberInstance = nil
	additionalPubSuberInstance = nil
	config.FlushAll()
}

func mountSecretsToDir(t *testing.T) {
	secretsDir := t.TempDir()
	secretFile := filepath.Join(secretsDir, "PUBSUB_PASSWORD")
	require.NoError(t, os.WriteFile(secretFile, []byte("test-password"), 0o600))
	t.Setenv("SECRETS_MOUNT_PATH", secretsDir)
}
