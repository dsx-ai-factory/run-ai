/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	personalcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	personalCredentialsSourceIndex = "run.ai/personal-credential-source"
	personalCredentialsFieldOwner  = "application-personal-credentials"
	personalCredentialsOwnerLabel  = "run.ai/personal-credential-owner"
)

var errPersonalCredentialsUnavailable = errors.New("personal credentials unavailable")

func (c *ApplicationController) materializePersonalCredentials(
	ctx context.Context,
	application *runv1alpha1.Application,
) error {
	entries, err := personalcreds.Get(application)
	if err != nil {
		return fmt.Errorf("parse personal credentials annotation: %w", err)
	}

	key := []byte(os.Getenv(constants.RunaiAdsEncryptionKeyEnvVar))
	if len(entries) > 0 && len(key) == 0 {
		return fmt.Errorf("encryption key is unavailable")
	}

	desired := make(map[string]struct{}, len(entries))
	var materializationErrors []error
	for _, entry := range entries {
		destName := personalcreds.DestName(application.Name, entry.SourceSecret)
		desired[destName] = struct{}{}
		if err := c.materializePersonalCredential(ctx, application, entry.SourceSecret, destName, key); err != nil {
			materializationErrors = append(materializationErrors, err)
		}
	}

	// Replacements are applied above before stale or revoked copies are removed.
	// This avoids temporarily removing a working credential during annotation changes.
	var copies corev1.SecretList
	if err := c.client.List(ctx, &copies,
		client.InNamespace(application.Namespace),
		client.MatchingLabels{
			personalcreds.CopyLabel:       "true",
			personalCredentialsOwnerLabel: application.Name,
		},
	); err != nil {
		return fmt.Errorf("list personal credential copies: %w", err)
	}
	for i := range copies.Items {
		copy := &copies.Items[i]
		if !metav1.IsControlledBy(copy, application) {
			continue
		}
		_, stillDesired := desired[copy.Name]
		if stillDesired && !containsRevocationError(materializationErrors, copy.Name) {
			continue
		}
		if err := c.client.Delete(ctx, copy); err != nil && !apierrors.IsNotFound(err) {
			materializationErrors = append(materializationErrors, fmt.Errorf("delete credential copy %s/%s: %w", copy.Namespace, copy.Name, err))
		}
	}

	if len(materializationErrors) > 0 {
		return errors.Join(materializationErrors...)
	}
	return nil
}

func (c *ApplicationController) materializePersonalCredential(
	ctx context.Context,
	application *runv1alpha1.Application,
	sourceName, destName string,
	key []byte,
) error {
	var source corev1.Secret
	if err := c.client.Get(ctx, types.NamespacedName{
		Name: sourceName, Namespace: c.runaiNamespace,
	}, &source); err != nil {
		return destinationError{
			destination: destName,
			revoke:      apierrors.IsNotFound(err),
			err:         fmt.Errorf("get source secret %s/%s: %w", c.runaiNamespace, sourceName, err),
		}
	}
	if source.Labels[personalcreds.UserCredentialsLabel] != "true" {
		return destinationError{
			destination: destName,
			revoke:      true,
			err:         fmt.Errorf("source secret %s/%s is not a personal credential", source.Namespace, source.Name),
		}
	}

	dest, err := personalcreds.BuildDestination(destName, application.Namespace, &source, key)
	if err != nil {
		return destinationError{destination: destName, err: fmt.Errorf("build destination for source %s/%s: %w", source.Namespace, source.Name, err)}
	}
	dest.Labels[personalCredentialsOwnerLabel] = application.Name

	var existing corev1.Secret
	err = c.client.Get(ctx, client.ObjectKeyFromObject(dest), &existing)
	switch {
	case err == nil && !metav1.IsControlledBy(&existing, application):
		return destinationError{destination: destName, err: fmt.Errorf("destination secret %s/%s is controlled by another object", dest.Namespace, dest.Name)}
	case err != nil && !apierrors.IsNotFound(err):
		return destinationError{destination: destName, err: fmt.Errorf("check destination secret %s/%s: %w", dest.Namespace, dest.Name, err)}
	}

	if err := controllerutil.SetControllerReference(application, dest, c.client.Scheme()); err != nil {
		return destinationError{destination: destName, err: fmt.Errorf("set destination owner: %w", err)}
	}
	dest.SetManagedFields(nil)
	dest.SetResourceVersion("")
	//nolint:staticcheck // Typed Secrets require the Patch API for server-side apply.
	if err := c.client.Patch(ctx, dest, client.Apply, client.ForceOwnership, client.FieldOwner(personalCredentialsFieldOwner)); err != nil {
		return destinationError{destination: destName, err: fmt.Errorf("apply destination secret %s/%s: %w", dest.Namespace, dest.Name, err)}
	}
	return nil
}

type destinationError struct {
	destination string
	revoke      bool
	err         error
}

func (e destinationError) Error() string { return e.err.Error() }
func (e destinationError) Unwrap() error { return e.err }

func containsRevocationError(errs []error, destination string) bool {
	for _, err := range errs {
		var destErr destinationError
		if errors.As(err, &destErr) && destErr.destination == destination && destErr.revoke {
			return true
		}
	}
	return false
}
