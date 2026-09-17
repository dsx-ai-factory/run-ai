/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const clockRounds = 5

// ClockOffset measures how far the cluster clock is from ours: positive means
// the cluster is ahead. Each round creates a tiny ConfigMap in namespace,
// compares the server's creationTimestamp with the midpoint of our send, and
// deletes the probe; the median of five rounds is returned. The server stores
// creationTimestamp in whole seconds, so the result is honest only to about a
// second and is reported as metadata, never mixed into any latency.
func ClockOffset(ctx context.Context, kube k8sclient.Client, namespace string) (time.Duration, error) {
	return clockOffset(ctx, kube, namespace, time.Now)
}

func clockOffset(ctx context.Context, kube k8sclient.Client, namespace string, now func() time.Time) (time.Duration, error) {
	offsets := make([]time.Duration, 0, clockRounds)
	for round := 1; round <= clockRounds; round++ {
		offset, err := probeOffset(ctx, kube, namespace, now)
		if err != nil {
			return 0, fmt.Errorf("clock offset round %d: %w", round, err)
		}
		offsets = append(offsets, offset)
	}
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	return offsets[len(offsets)/2], nil
}

func probeOffset(ctx context.Context, kube k8sclient.Client, namespace string, now func() time.Time) (time.Duration, error) {
	suffix, err := randomSuffix()
	if err != nil {
		return 0, err
	}
	probe := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "flowmark-clock-" + suffix}}

	before := now()
	if err := kube.Create(ctx, probe); err != nil {
		return 0, fmt.Errorf("creating probe %s/%s: %w", namespace, probe.Name, err)
	}
	after := now()
	deleteErr := k8sclient.IgnoreNotFound(kube.Delete(ctx, probe))

	if probe.CreationTimestamp.IsZero() {
		return 0, errors.New("server did not report a creationTimestamp")
	}
	if deleteErr != nil {
		return 0, fmt.Errorf("deleting probe %s/%s: %w", namespace, probe.Name, deleteErr)
	}
	midpoint := before.Add(after.Sub(before) / 2)
	return probe.CreationTimestamp.Time.Sub(midpoint), nil
}

func randomSuffix() (string, error) {
	var raw [4]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generating probe name: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}
