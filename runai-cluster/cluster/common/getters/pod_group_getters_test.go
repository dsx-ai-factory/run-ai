/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package getters

import (
	"context"
	"errors"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// failingReader stands in for a client whose list call fails, e.g. because the
// API server is unreachable or the caller is not allowed to list pods.
type failingReader struct {
	err error
}

func (f failingReader) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return f.err
}

func (f failingReader) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return f.err
}

// A failed list is not an empty result: it has to reach the caller, which
// decides between "no pods" and "could not tell".
func TestGetPodsWithSelectorReturnsTheListError(t *testing.T) {
	listErr := errors.New("connection refused")

	pods, err := getPodsWithSelector(context.Background(), "runai-ns", failingReader{err: listErr},
		&client.ListOptions{Namespace: "runai-ns"})

	if !errors.Is(err, listErr) {
		t.Fatalf("got %v, want the list error %v", err, listErr)
	}
	if len(pods) != 0 {
		t.Fatalf("got %d pods, want none", len(pods))
	}
}

func TestGetPodsWithSelectorRequiresASelector(t *testing.T) {
	if _, err := getPodsWithSelector(context.Background(), "runai-ns", failingReader{}, nil); err == nil {
		t.Fatal("got nil, want an error for a nil selector")
	}
}
