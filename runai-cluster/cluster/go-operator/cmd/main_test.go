/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCacheOptionsDoNotRequireEngineConfigCRD(t *testing.T) {
	engineConfigGVK := schema.GroupVersionKind{
		Group:   "engine.run.ai",
		Version: "v1alpha1",
		Kind:    "Config",
	}

	options := cacheOptions()
	if _, err := options.Scheme.New(engineConfigGVK); !runtime.IsNotRegisteredError(err) {
		t.Fatalf("Engine Config must not be registered in the manager scheme; got error %v", err)
	}

	for object := range options.ByObject {
		gvks, _, err := options.Scheme.ObjectKinds(object)
		if err != nil {
			t.Fatalf("failed resolving cached object %T: %v", object, err)
		}
		for _, gvk := range gvks {
			if gvk.Group == engineConfigGVK.Group && gvk.Kind == engineConfigGVK.Kind {
				t.Fatalf("Engine Config must not be registered in the manager cache")
			}
		}
	}
}
