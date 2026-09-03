/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filters

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
)

func TestMatching_Equal(t *testing.T) {
	matcher := PropertyMatcher{
		MatchType: "==",
		Key:       "test",
		Value:     "test",
	}
	match := matching("test", &matcher)
	assert.True(t, match)
}

func TestMatching_NotEqual(t *testing.T) {
	matcher := PropertyMatcher{
		MatchType: "==",
		Key:       "test",
		Value:     "test1",
	}
	match := matching("test", &matcher)
	assert.False(t, match)
}

func TestMatching_RegexMatch(t *testing.T) {
	matcher := PropertyMatcher{
		MatchType: "=~",
		Key:       "test",
		Value:     "^Runai.*$",
	}
	match := matching("RunaiRandomAlert", &matcher)
	assert.True(t, match)
}

func TestMatching_RegexNotMatch(t *testing.T) {
	matcher := PropertyMatcher{
		MatchType: "=~",
		Key:       "test",
		Value:     "^Runai$",
	}
	match := matching("RunaiRandomAlert", &matcher)
	assert.False(t, match)
}

func TestFilter_FilterDrops(t *testing.T) {
	filter := &MatcherFilter{
		config: &MatcherFilterConfig{
			{
				MatchType: "==",
				Key:       "test",
				Value:     "test",
			},
		},
	}
	event := &events.Event{
		Properties: map[string]interface{}{
			"test": "bla",
		},
	}

	dropped, err := filter.Filter(context.Background(), event)
	assert.Nil(t, err)
	assert.True(t, dropped)
}

func TestFilter_FilterPass(t *testing.T) {
	filter := &MatcherFilter{
		config: &MatcherFilterConfig{
			{
				MatchType: "==",
				Key:       "test",
				Value:     "test",
			},
		},
	}
	event := &events.Event{
		Properties: map[string]interface{}{
			"test": "test",
		},
	}

	dropped, err := filter.Filter(context.Background(), event)
	assert.Nil(t, err)
	assert.False(t, dropped)
}
