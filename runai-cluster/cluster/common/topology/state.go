/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology

import (
	"context"
)

type State struct {
	enabled   bool
	name      string
	preferred string
	required  string
}

func DisabledState() *State {
	return &State{enabled: false}
}

func EnabledState(name, preferred, required string) *State {
	return &State{enabled: true, name: name, preferred: preferred, required: required}
}

func GetState(ctx context.Context, annotations map[string]string) *State {
	name := annotations[KAITopologyAnnotationKey]
	if name == "" {
		return DisabledState()
	}

	preferred := annotations[KAITopologyAnnotationPreferredKey]
	required := annotations[KAITopologyAnnotationRequiredKey]
	return EnabledState(name, preferred, required)
}

func (s *State) Enabled() bool {
	return s.enabled
}
