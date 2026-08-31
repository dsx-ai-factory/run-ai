/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"fmt"
	mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestResubscribe_ShouldCallAllAddedSubscribers_NoError(t *testing.T) {
	subscriber1 := mocks.NewSubscriber(t)
	subscriber2 := mocks.NewSubscriber(t)
	handler := ResubscribeHandler{
		lock: &sync.RWMutex{},
		subscribers: map[string]Subscriber{
			"first":  subscriber1,
			"second": subscriber2,
		},
	}
	subscriber1.On("Resubscribe").Return(nil).Once()
	subscriber2.On("Resubscribe").Return(nil).Once()

	err := handler.Resubscribe()
	require.NoError(t, err)
}

func TestResubscribe_ShouldCallAllAddedSubscribers_WithError(t *testing.T) {
	subscriber1 := mocks.NewSubscriber(t)
	subscriber2 := mocks.NewSubscriber(t)
	handler := ResubscribeHandler{
		lock: &sync.RWMutex{},
		subscribers: map[string]Subscriber{
			"first": subscriber1, "second": subscriber2,
		},
	}
	subscriber1.On("Resubscribe").Return(fmt.Errorf("failed in subscriber1")).Once()
	subscriber2.On("Resubscribe").Return(nil).Once()

	err := handler.Resubscribe()
	require.Errorf(t, err, "failed in subscriber1")
}
