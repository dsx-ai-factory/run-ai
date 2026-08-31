/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package issue

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Group categorizes an issue for downstream consumers (cluster-sync, UI).
type Group string

const (
	Internal Group = "Internal"
	Network  Group = "Network"
	Compute  Group = "Compute"
	Storage  Group = "Storage"
)

func (g Group) String() string { return string(g) }

// Issue is the in-memory representation of a controller or plugin problem.
// Converted to v1alpha1.Issue when written to WorkloadRunner.status.
// Issue implements the error interface, so it can be returned and wrapped
// like any other error.
type Issue struct {
	Name        string
	GVK         schema.GroupVersionKind
	IsCritical  bool
	Description string
	Group       Group

	err error
}

// New builds an Issue. The caller decides the Group - no auto-classify.
func New(object client.Object, err error, isCritical bool, desc string, group Group) *Issue {
	return &Issue{
		Name:        object.GetName(),
		GVK:         object.GetObjectKind().GroupVersionKind(),
		IsCritical:  isCritical,
		Description: desc,
		Group:       group,
		err:         err,
	}
}

// Error renders the issue as "[Kind:Name] description: error".
func (i *Issue) Error() string {
	var parts []string
	if i.Description != "" {
		parts = append(parts, i.Description)
	}
	if i.err != nil {
		parts = append(parts, i.err.Error())
	}
	return fmt.Sprintf("[%s:%s] %s", i.GVK.Kind, i.Name, strings.Join(parts, ": "))
}

// Unwrap exposes the underlying error for errors.Is / errors.As.
func (i *Issue) Unwrap() error { return i.err }
