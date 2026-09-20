/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type fakeClocks struct {
	local  time.Time
	offset time.Duration
}

func (c *fakeClocks) now() time.Time {
	c.local = c.local.Add(10 * time.Millisecond)
	return c.local
}

func (c *fakeClocks) server() time.Time {
	return c.local.Add(c.offset).Truncate(time.Second)
}

func clockClient(clocks *fakeClocks) k8sclient.WithWatch {
	return fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Create: func(ctx context.Context, c k8sclient.WithWatch, obj k8sclient.Object, opts ...k8sclient.CreateOption) error {
			obj.(*corev1.ConfigMap).CreationTimestamp = metav1.NewTime(clocks.server())
			return c.Create(ctx, obj, opts...)
		},
	}).Build()
}

func TestClockOffsetSign(t *testing.T) {
	cases := map[string]time.Duration{
		"cluster ahead":  90 * time.Second,
		"cluster behind": -45 * time.Second,
		"in sync":        0,
	}
	for name, offset := range cases {
		t.Run(name, func(t *testing.T) {
			clocks := &fakeClocks{local: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC), offset: offset}
			kube := clockClient(clocks)

			got, err := clockOffset(context.Background(), kube, "default", clocks.now)
			if err != nil {
				t.Fatal(err)
			}
			if diff := (got - offset).Abs(); diff > time.Second {
				t.Fatalf("offset %s is more than a second away from the injected %s", got, offset)
			}

			var probes corev1.ConfigMapList
			if err := kube.List(context.Background(), &probes, k8sclient.InNamespace("default")); err != nil {
				t.Fatal(err)
			}
			if len(probes.Items) != 0 {
				t.Fatalf("all probes must be deleted, %d left", len(probes.Items))
			}
		})
	}
}

func TestClockOffsetFailsWithoutServerTimestamp(t *testing.T) {
	kube := fake.NewClientBuilder().Build()

	if _, err := clockOffset(context.Background(), kube, "default", time.Now); err == nil {
		t.Fatal("a server that reports no creationTimestamp must be an error, not a zero offset")
	}
}
