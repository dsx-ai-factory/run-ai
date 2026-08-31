//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/stretchr/testify/require"

	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/stretchr/testify/assert"
)

type someTestEvent struct {
	Data string
}

func initAsync() (chan error, error) {
	logging.Init(nil)
	runaipubsub.SetLogger(logging.Actual())
	return runaipubsub.InitAsync(func(e error) {
		fmt.Printf("Error: %v\n", e)
	})
}

func TestAsyncInitAndSubscribe(t *testing.T) {

	eventString := "Sub-->Pub"
	ev := &someTestEvent{
		Data: eventString,
	}
	setEnvs(t)
	defer tearDownIntegrationTest()
	initAsync()

	count := 0
	runaipubsub.SubscribeAsync(func(ctx context.Context, event *someTestEvent) error {
		assert.Equal(t, eventString, event.Data)
		count += 1
		fmt.Println("TestAsyncInitAndSubscribe: time: ", time.Now().String(), " count: ", count)
		return nil
	})

	require.Eventually(t, func() bool {
		err := runaipubsub.Publish(context.Background(), ev)
		return err == nil
	}, 2*time.Second, 100*time.Millisecond)

	assert.Eventually(t, func() bool {
		return count == 1
	}, time.Duration(waitForWantedStateSeconds)*time.Second, 1*time.Second)
	assert.Never(t, func() bool {
		return count > 1
	}, time.Duration(waitForStateStabilizationSeconds)*time.Second, 1*time.Second)
	runaipubsub.Close()
}

func TestAsyncInitAndSubscribe_Fails(t *testing.T) {
	setEnvs(t)
	t.Setenv("PUBSUB_URL", "bad_host:9092")
	defer tearDownIntegrationTest()

	errs, err := initAsync()
	require.NoError(t, err)

	eventString := "Sub-->Pub"
	runaipubsub.SubscribeAsync(func(ctx context.Context, event *someTestEvent) error {
		return nil
	})

	require.Eventually(t, func() bool {
		select {
		case _ = <-errs:
			runaipubsub.Close()
			return true
		}
	}, time.Second, 100*time.Millisecond)

	ev := &someTestEvent{
		Data: eventString,
	}

	err = runaipubsub.Publish(context.Background(), ev)
	assert.Error(t, err)
}

func TestAsyncInitAndSubscribe_Fails_BadConfig(t *testing.T) {
	defer tearDownIntegrationTest()

	_, err := initAsync()
	require.Error(t, err)

	require.False(t, runaipubsub.IsInitialized())
}

func TestIsInitialized_OnAsyncInitEventuallyReturnsTrue(t *testing.T) {
	defer tearDownIntegrationTest()
	setEnvs(t)

	require.False(t, runaipubsub.IsInitialized())

	_, err := initAsync()
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return runaipubsub.IsInitialized()
	}, 2*time.Second, 100*time.Millisecond)
}
