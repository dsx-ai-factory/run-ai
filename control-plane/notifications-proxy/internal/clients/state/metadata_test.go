/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package state

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestFromState_EmptyState(t *testing.T) {
	state := []notification_state.Stater{}
	meta := FromState(state)
	require.Equal(t, map[string]string{}, meta)
}

func TestFromState_EmailEnabledOnly(t *testing.T) {
	state := []notification_state.Stater{&EmailState{Enabled: ptr.To(true)}}
	meta := FromState(state)
	require.Equal(t, map[string]string{
		EmailEnabled: "true",
	}, meta)
}

func TestFromState_SlackEnabledAndWorkspace(t *testing.T) {
	state := []notification_state.Stater{&SlackState{Enabled: ptr.To(true)}}
	meta := FromState(state)
	require.Equal(t, map[string]string{
		SlackEnabled: "true",
	}, meta)
}

func TestToState_EmptyMap_GivesNilPointers(t *testing.T) {
	state := ToState(map[string]string{})
	require.Equal(t, []notification_state.Stater{&EmailState{}, &SlackState{}}, state)
}

func TestToState_EmailTrueOnly(t *testing.T) {
	meta := map[string]string{EmailEnabled: "true"}
	state := ToState(meta)

	require.Equal(t, []notification_state.Stater{&EmailState{Enabled: ptr.To(true)}, &SlackState{}}, state)
}

func TestToState_SlackFalseWithWorkspace(t *testing.T) {
	meta := map[string]string{
		SlackEnabled:       "false",
		SlackWorkspaceName: "workspace",
	}
	state := ToState(meta)

	require.Equal(t, []notification_state.Stater{&EmailState{}, &SlackState{Enabled: ptr.To(false), WorkspaceName: ptr.To("workspace")}}, state)
}

func TestToState_BothChannels(t *testing.T) {
	meta := map[string]string{
		EmailEnabled:       "false",
		SlackEnabled:       "true",
		SlackWorkspaceName: "workspace",
	}
	state := ToState(meta)

	require.EqualValues(t, []notification_state.Stater{&EmailState{Enabled: ptr.To(false)}, &SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")}}, state)
}

func TestToStateWithDetails_SlackFalseWithWorkspace(t *testing.T) {
	meta := map[string]string{
		SlackEnabled:       "false",
		SlackWorkspaceName: "workspace",
	}
	state := ToDetailedState(meta, false)

	require.EqualValues(t, []notification_state.Stater{&EmailState{}, &SlackState{Enabled: ptr.To(false), WorkspaceName: ptr.To("workspace"), AppCredentialsConfigured: ptr.To(false)}}, state)
}

func TestToStateWithDetails_BothChannels(t *testing.T) {
	meta := map[string]string{
		EmailEnabled:       "false",
		SlackEnabled:       "true",
		SlackWorkspaceName: "workspace",
	}
	state := ToDetailedState(meta, true)

	require.EqualValues(t, []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace"), AppCredentialsConfigured: ptr.To(true)}}, state)
}

func TestMergeMetadata_NilCurrent_NonEmptyUpdate(t *testing.T) {
	current := map[string]string(nil)
	update := map[string]string{"a": "1"}

	merged := MergeMetadata(current, update)

	require.Equal(t, update, merged)
}

func TestMergeMetadata_EmptyCurrent_NonEmptyUpdate(t *testing.T) {
	current := map[string]string{}
	update := map[string]string{"a": "1"}

	merged := MergeMetadata(current, update)

	require.Equal(t, update, merged)
}

func TestMergeMetadata_NonEmptyCurrent_EmptyUpdate(t *testing.T) {
	current := map[string]string{"a": "1"}
	update := map[string]string{}

	merged := MergeMetadata(current, update)

	require.Equal(t, current, merged)
}

func TestMergeMetadata_DisjointKeys(t *testing.T) {
	current := map[string]string{"a": "1"}
	update := map[string]string{"b": "2"}

	merged := MergeMetadata(current, update)

	require.Equal(t, map[string]string{"a": "1", "b": "2"}, merged)
}

func TestMergeMetadata_OverwriteExistingKey(t *testing.T) {
	current := map[string]string{"a": "1", "b": "2"}
	update := map[string]string{"b": "99"}

	merged := MergeMetadata(current, update)

	require.Equal(t, map[string]string{"a": "1", "b": "99"}, merged)
}

func TestMergeMetadata_BothNil(t *testing.T) {
	merged := MergeMetadata(nil, nil)

	require.Equal(t, map[string]string{}, merged)
}

func TestGetRouteMetadata_NilRoute(t *testing.T) {
	metadata := getRouteMetadata(nil)
	require.Nil(t, metadata)
}

func TestGetRouteMetadata_NonNilRoute(t *testing.T) {
	m := map[string]string{"k": "v"}
	route := &types.BasicRouteConfig{
		Metadata: m,
	}
	metadata := getRouteMetadata(route)
	require.Equal(t, m, metadata)
}
