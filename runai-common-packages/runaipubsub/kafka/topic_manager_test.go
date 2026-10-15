/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// Callers reach CreateIfNotExist through retry.RetryStreamErrors, so a failed
// creation must be reported to them and must not be remembered as done —
// otherwise the retry succeeds against a topic that was never created.
func TestCreateIfNotExistReturnsTheCreationError(t *testing.T) {
	manager := &topicManager{}
	manager.createTopic = func(string) error { return errors.New("broker unavailable") }

	err := manager.CreateIfNotExist("events")

	if err == nil {
		t.Fatal("got nil, want the creation error")
	}
	if !strings.Contains(err.Error(), "failed to create topic events") {
		t.Fatalf("got %q, want it to name the topic", err)
	}
}

func TestCreateIfNotExistRetriesAFailedCreation(t *testing.T) {
	attempts := 0
	manager := &topicManager{}
	manager.createTopic = func(string) error {
		attempts++
		if attempts == 1 {
			return errors.New("broker unavailable")
		}
		return nil
	}

	if err := manager.CreateIfNotExist("events"); err == nil {
		t.Fatal("first call got nil, want the creation error")
	}
	if err := manager.CreateIfNotExist("events"); err != nil {
		t.Fatalf("second call got %v, want the topic to be created on the retry", err)
	}
	if attempts != 2 {
		t.Fatalf("createTopic called %d times, want 2", attempts)
	}
}

func TestCreateIfNotExistCreatesATopicOnce(t *testing.T) {
	attempts := 0
	manager := &topicManager{}
	manager.createTopic = func(string) error {
		attempts++
		return nil
	}

	for call := 1; call <= 3; call++ {
		if err := manager.CreateIfNotExist("events"); err != nil {
			t.Fatalf("call %d got %v, want nil", call, err)
		}
	}
	if attempts != 1 {
		t.Fatalf("createTopic called %d times, want 1", attempts)
	}
}

// Only the caller that happens to run the sync.Once function used to see the
// error; the others were told the topic exists.
func TestCreateIfNotExistReportsAFailureToEveryConcurrentCaller(t *testing.T) {
	const callers = 5

	manager := &topicManager{}
	manager.createTopic = func(string) error { return errors.New("broker unavailable") }

	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = manager.CreateIfNotExist("events")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err == nil {
			t.Errorf("caller %d got nil, want the creation error", i)
		}
	}
}

func TestCreateIfNotExistIsPerTopic(t *testing.T) {
	attempts := map[string]int{}
	manager := &topicManager{}
	manager.createTopic = func(topic string) error {
		attempts[topic]++
		return nil
	}

	for _, topic := range []string{"events", "alerts", "events"} {
		if err := manager.CreateIfNotExist(topic); err != nil {
			t.Fatalf("CreateIfNotExist(%q) got %v, want nil", topic, err)
		}
	}

	if attempts["events"] != 1 || attempts["alerts"] != 1 {
		t.Fatalf("createTopic calls = %v, want one per topic", attempts)
	}
}
