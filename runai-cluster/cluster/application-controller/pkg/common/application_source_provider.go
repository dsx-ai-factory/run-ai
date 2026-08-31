/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationFactory creates Application CRDs from source-specific CRDs.
type ApplicationFactory interface {
	// GetApplicationSource returns the source identifier (e.g., "helm", "dynamo")
	GetApplicationSource() string

	// GetSourceResource returns a new instance of the source object type to watch
	GetSourceResource() ctrlClient.Object

	// FilterSourceResource determines if a source resource event should be processed by the informer
	// This is called by the informer's FilterFunc before invoking event handlers
	FilterSourceResource(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) bool

	// ShouldCreateApplication checks if this source resource should trigger Application creation
	// Returns: shouldCreate, applicationName, error
	ShouldCreateApplication(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) (bool, string, error)

	// CreateApplicationMetadata builds Application metadata from a source resource
	CreateApplicationMetadata(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) (ApplicationMetadata, error)
}

// ApplicationSourceProvider queries and validates application resources at runtime.
type ApplicationSourceProvider interface {
	// GetApplicationSource returns the source identifier (e.g., "helm", "dynamo")
	GetApplicationSource() string

	// GetListOptions returns ListOptions for querying resources belonging to this source
	GetListOptions(namespace string) *ctrlClient.ListOptions

	// BelongsTo checks if a resource belongs to the specified application name
	BelongsTo(resource ctrlClient.Object, applicationName string) bool

	// MatchesSource checks if a resource belongs to this application source (for event filtering)
	MatchesSource(resource ctrlClient.Object) bool

	// GetApplicationName extracts the application name from a resource (for object-to-application mapping)
	// Returns empty string if the application name cannot be extracted
	GetApplicationName(resource ctrlClient.Object) string

	// GetApplicationPhase extracts the current phase/status from a source resource
	GetApplicationPhase(sourceResource ctrlClient.Object) string
}

// ApplicationMetadata holds the metadata needed to create an Application CRD
type ApplicationMetadata struct {
	Name            string
	Namespace       string
	Type            string // Chart name, app type, etc.
	Source          string // "helm", "dynamo", etc.
	Labels          map[string]string
	OwnerReferences []metav1.OwnerReference
}
