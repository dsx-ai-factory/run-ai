/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package render

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// SchedulingProfile is a named, preloaded scoring configuration.
type SchedulingProfile struct {
	Name    string
	Scorers []Scorer
}

var profileNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func (in *Input) prepareProfiles() error {
	if !in.RuntimeProfiles {
		if len(in.Profiles) != 0 || in.ActiveProfile != "" || in.DefaultProfile != "" {
			return errors.New("scheduling profiles require the llm-d-v0.10 implementation")
		}
		return nil
	}
	if len(in.Scorers) != 0 || len(in.Profiles) < 1 || len(in.Profiles) > 8 {
		return errors.New("runtime scheduling requires 1-8 profiles and no flat scorers list")
	}
	names := make(map[string]bool, len(in.Profiles))
	plugins := make(map[string]bool)
	for _, profile := range in.Profiles {
		if names[profile.Name] || len(profile.Name) > 63 || !profileNamePattern.MatchString(profile.Name) {
			return fmt.Errorf("invalid or duplicate scheduling profile name %q", profile.Name)
		}
		names[profile.Name] = true
		if err := validateProfileScorers(profile); err != nil {
			return err
		}
		for _, scorer := range profile.Scorers {
			plugins[scorer.Plugin] = true
		}
	}
	if !names[in.DefaultProfile] || !names[in.ActiveProfile] {
		return errors.New("defaultProfile and activeProfile must name configured profiles")
	}
	// Canonicalize map-like API lists without mutating caller-owned slices.
	// Reordering an unchanged catalog must not roll the EPP.
	in.Profiles = slices.Clone(in.Profiles)
	for i := range in.Profiles {
		in.Profiles[i].Scorers = slices.Clone(in.Profiles[i].Scorers)
		slices.SortFunc(in.Profiles[i].Scorers, func(a, b Scorer) int { return strings.Compare(a.Plugin, b.Plugin) })
	}
	slices.SortFunc(in.Profiles, func(a, b SchedulingProfile) int { return strings.Compare(a.Name, b.Name) })
	in.ProfilePlugins = make([]string, 0, len(plugins))
	for plugin := range plugins {
		in.ProfilePlugins = append(in.ProfilePlugins, plugin)
	}
	slices.Sort(in.ProfilePlugins)
	return nil
}

func validateProfileScorers(profile SchedulingProfile) error {
	if len(profile.Scorers) < 1 || len(profile.Scorers) > 3 {
		return fmt.Errorf("profile %q requires 1-3 scorers", profile.Name)
	}
	seen := make(map[string]bool, len(profile.Scorers))
	positive := false
	for _, scorer := range profile.Scorers {
		switch scorer.Plugin {
		case "queue-scorer", "kv-cache-utilization-scorer", "prefix-cache-scorer":
		default:
			return fmt.Errorf("profile %q has unsupported scorer %q", profile.Name, scorer.Plugin)
		}
		if seen[scorer.Plugin] || scorer.Weight < 0 || scorer.Weight > 100 {
			return fmt.Errorf("profile %q has duplicate scorer or weight outside 0-100", profile.Name)
		}
		seen[scorer.Plugin] = true
		positive = positive || scorer.Weight > 0
	}
	if !positive {
		return fmt.Errorf("profile %q needs at least one positive scorer weight", profile.Name)
	}
	return nil
}
