/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcile

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State struct {
	pluginIssueMap   map[string][]issue.Issue
	controllerIssues []issue.Issue
	requeueAfter     time.Duration
}

func New() *State {
	return &State{pluginIssueMap: make(map[string][]issue.Issue)}
}

// RequestRequeueAfter retries recoverable plugin work without preventing the
// workload, remaining plugins, and status from being reconciled first.
func (s *State) RequestRequeueAfter(delay time.Duration) {
	if delay > 0 && (s.requeueAfter == 0 || delay < s.requeueAfter) {
		s.requeueAfter = delay
	}
}

func (s *State) RequeueAfter() time.Duration {
	return s.requeueAfter
}

func (s *State) AddControllerIssue(object client.Object, err error, isCritical bool, description string, group issue.Group) {
	s.controllerIssues = append(s.controllerIssues, *issue.New(object, err, isCritical, description, group))
}

func (s *State) AddPluginIssue(ctx context.Context, reporter Reporter, object client.Object, err error, isCritical bool, description string, group issue.Group) error {
	if err == nil {
		logging.FromContext(ctx).Warn("Ignoring nil error for plugin issue")
		return nil
	}

	_, ok := s.pluginIssueMap[reporter.Name()]
	if !ok {
		s.pluginIssueMap[reporter.Name()] = make([]issue.Issue, 0)
	}

	s.pluginIssueMap[reporter.Name()] = append(s.pluginIssueMap[reporter.Name()], *issue.New(object, err, isCritical, description, group))
	return err
}

func (s *State) PluginIssues(reporter Reporter) []issue.Issue {
	return s.pluginIssueMap[reporter.Name()]
}

func (s *State) Issues() []issue.Issue {
	issues := make([]issue.Issue, 0)
	issues = append(issues, s.controllerIssues...)
	for _, plugin := range s.pluginIssueMap {
		issues = append(issues, plugin...)
	}
	return issues
}

func (s *State) PluginsMap() map[string][]issue.Issue {
	return s.pluginIssueMap
}

func (s *State) Controller() []issue.Issue {
	return s.controllerIssues
}

func (s *State) HasCriticalControllerIssues() bool {
	for _, i := range s.controllerIssues {
		if i.IsCritical {
			return true
		}
	}
	return false
}
