/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LegacyRIGVK identifies the legacy ResourceInterface custom resource. It
// lives here (not in the ri/ subpackage) because the scheme builder needs to
// register both the singular and list kinds for unstructured RI ops.
var LegacyRIGVK = schema.GroupVersionKind{
	Group:   "optimization.nvidia.com",
	Version: "v1alpha1",
	Kind:    "ResourceInterface",
}

// NewK8sClient returns a single controller-runtime client.Client that can
// drive every operation the migration needs: typed Karta CRUD, typed CRD
// reads, and unstructured List/Update/Delete on legacy RIs. The scheme
// registers core + apiextensionsv1 + kartav1alpha1, plus an unstructured
// alias for the RI singular + list kinds so client.Client handles them
// natively without a separate dynamic client (per natasharomm PR review).
func NewK8sClient(cfg *rest.Config) (client.Client, error) {
	s, err := buildScheme()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		return nil, fmt.Errorf("build k8s client: %w", err)
	}
	return c, nil
}

func buildScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("add core scheme: %w", err)
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("add apiextensions scheme: %w", err)
	}
	if err := kartav1alpha1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("add karta scheme: %w", err)
	}
	// The legacy RI types are deleted from the engine; register the GVK as
	// unstructured so client.Client.List/Update/Delete can drive RIs.
	s.AddKnownTypeWithName(LegacyRIGVK, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(LegacyRIGVK.GroupVersion().WithKind(LegacyRIGVK.Kind+"List"), &unstructured.UnstructuredList{})
	return s, nil
}

// SchemeForTesting returns the same scheme NewK8sClient uses. Exposed for
// test helpers that need to build a fake client with the same registrations.
func SchemeForTesting() (*runtime.Scheme, error) {
	return buildScheme()
}
