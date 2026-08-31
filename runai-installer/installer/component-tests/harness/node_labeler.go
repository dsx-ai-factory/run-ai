/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"
	"sync"
)

// NodeLabelCall records one LabelWorkerNodes invocation.
type NodeLabelCall struct {
	Key   string
	Value string
}

// NodeNamesLabelCall records one LabelNodes invocation.
type NodeNamesLabelCall struct {
	Names []string
	Key   string
	Value string
}

// FakeNodeLabeler records node-label invocations and reports a fixed count.
type FakeNodeLabeler struct {
	mu         sync.Mutex
	Calls      []NodeLabelCall
	NodesCalls []NodeNamesLabelCall
	Labeled    int
	LabelErr   error
}

func (f *FakeNodeLabeler) LabelWorkerNodes(_ context.Context, key, value string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, NodeLabelCall{Key: key, Value: value})
	if f.LabelErr != nil {
		return 0, f.LabelErr
	}
	return f.Labeled, nil
}

func (f *FakeNodeLabeler) LabelNodes(_ context.Context, names []string, key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.NodesCalls = append(f.NodesCalls, NodeNamesLabelCall{Names: names, Key: key, Value: value})
	return f.LabelErr
}
