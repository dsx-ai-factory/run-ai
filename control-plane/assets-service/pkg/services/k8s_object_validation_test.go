/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Direct function-call tests of the shared k8s-object validators. No HTTP,
// no mocks — the helpers are pure functions and testable at that level.
// The TestServices suite runner in services_suite_test.go picks these up.
//
// End-to-end wiring ("does CreateRct actually call these?") is covered by
// the RCT-specific integration tests in api_rct_service_test.go.

var _ = Describe("validateK8sObjectMeta", func() {
	It("accepts a nil meta (no-op)", func() {
		Expect(validateK8sObjectMeta(nil)).To(BeNil())
	})

	It("accepts an empty meta struct", func() {
		Expect(validateK8sObjectMeta(&api.K8sObjectMeta{})).To(BeNil())
	})

	It("accepts a valid label + annotation payload", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels:      &map[string]string{"app": "my-app", "example.com/tier": "gpu"},
			Annotations: &map[string]string{"example.com/owner": "team-a"},
		})
		Expect(err).To(BeNil())
	})

	It("accepts an empty label value (K8s spec allows it)", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels: &map[string]string{"app": ""},
		})
		Expect(err).To(BeNil())
	})

	It("rejects an invalid label key character", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels: &map[string]string{"my@label": "v"},
		})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("spec.objectMeta.labels"))
		Expect(err.Message).To(ContainSubstring("valid Kubernetes qualified name"))
	})

	It("rejects a label key name longer than DNS1123LabelMaxLength (63) chars", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels: &map[string]string{strings.Repeat("a", validation.DNS1123LabelMaxLength+1): "v"},
		})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("spec.objectMeta.labels"))
	})

	It("rejects a label key prefix longer than DNS1123SubdomainMaxLength (253) chars", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels: &map[string]string{strings.Repeat("a", validation.DNS1123SubdomainMaxLength+1) + "/name": "v"},
		})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("prefix component"))
	})

	It("rejects an invalid label value character", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Labels: &map[string]string{"app": "bad@value"},
		})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("valid Kubernetes label value"))
	})

	It("rejects an invalid annotation key character", func() {
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Annotations: &map[string]string{"my@ann": "v"},
		})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("spec.objectMeta.annotations"))
	})

	It("accepts annotations whose combined size is just under the K8s limit", func() {
		// One annotation of (limit - key_len) bytes exactly hits the ceiling.
		key := "example.com/big"
		value := strings.Repeat("x", validation.TotalAnnotationSizeLimitB-len(key))
		err := validateK8sObjectMeta(&api.K8sObjectMeta{
			Annotations: &map[string]string{key: value},
		})
		Expect(err).To(BeNil())
	})

	It("rejects annotations whose combined UTF-8 byte size exceeds the K8s limit", func() {
		// 65 KiB x 5 = 325 KiB > 256 KiB. Chosen so no single key/value is
		// huge on its own — only the aggregate trips the limit, which is the
		// exact gap the schema's per-key maxLength can't catch.
		anns := map[string]string{}
		for i := 0; i < 5; i++ {
			anns["example.com/chunk-"+string(rune('a'+i))] = strings.Repeat("x", 65*1024)
		}
		err := validateK8sObjectMeta(&api.K8sObjectMeta{Annotations: &anns})
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring("spec.objectMeta.annotations"))
		Expect(err.Message).To(ContainSubstring("exceeds Kubernetes limit"))
	})
})

// The generic validator is exercised against RCT's compiled schema
// (rctObjectSpecSchema) + resourcev1.ResourceClaimSpec — the only pair we
// have wired up today. When a second k8s-object kind lands, add a sibling
// Describe with its own <schema, T> to prove the generic works there too.
var _ = Describe("validateK8sObjectSpec (via RCT schema + ResourceClaimSpec)", func() {
	const (
		pathPrefix = "spec.objectSpec"
		kindLabel  = "resource.k8s.io/v1 ResourceClaimSpec"
	)

	validSpec := func() map[string]any {
		return map[string]any{
			"devices": map[string]any{
				"requests": []any{
					map[string]any{
						"name": "gpu",
						"exactly": map[string]any{
							"deviceClassName": "gpu.nvidia.com",
							"count":           1,
							"allocationMode":  "ExactCount",
						},
					},
				},
			},
		}
	}

	It("accepts a minimal valid ResourceClaimSpec", func() {
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			validSpec(), rctObjectSpecSchema, pathPrefix, kindLabel)
		Expect(err).To(BeNil())
	})

	It("rejects an empty objectSpec (short-circuit before schema)", func() {
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			map[string]any{}, rctObjectSpecSchema, pathPrefix, kindLabel)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(Equal("spec.objectSpec is required"))
	})

	It("returns 503 when the schema is nil (validator not initialized)", func() {
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			validSpec(), nil, pathPrefix, kindLabel)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
		Expect(err.Message).To(ContainSubstring("schema-based validator is not initialized"))
	})

	It("rejects a top-level typo like 'devics' via layer 2 (typed decode)", func() {
		// K8s OpenAPI schemas are permissive (no additionalProperties: false),
		// so layer 1 accepts unknown fields. Layer 2 (DisallowUnknownFields
		// against resourcev1.ResourceClaimSpec) is what catches this — the
		// regression that motivated the two-layer design in the first place.
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			map[string]any{
				"devics":  map[string]any{},
				"devices": map[string]any{"requests": []any{}},
			},
			rctObjectSpecSchema, pathPrefix, kindLabel)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		// Direct function-call assertions don't go through JSON encoding, so
		// quotes stay raw (no `\"…\"` escape trick needed).
		Expect(err.Message).To(ContainSubstring(`unknown field "devics"`))
	})

	It("rejects an unknown nested field inside devices.requests[] via layer 2", func() {
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			map[string]any{
				"devices": map[string]any{
					"requests": []any{
						map[string]any{
							"name":                "gpu",
							"exactly":             map[string]any{"deviceClassName": "gpu.nvidia.com", "count": 1},
							"unexpectedNestedKey": "oops",
						},
					},
				},
			},
			rctObjectSpecSchema, pathPrefix, kindLabel)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring(`unknown field "unexpectedNestedKey"`))
	})

	It("rejects a shape mismatch (devices.requests as string) via layer 1", func() {
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			map[string]any{
				"devices": map[string]any{"requests": "not-an-array"},
			},
			rctObjectSpecSchema, pathPrefix, kindLabel)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusBadRequest))
		Expect(err.Message).To(ContainSubstring(kindLabel))
	})

	It("echoes the caller-supplied pathPrefix and kindLabel in error messages", func() {
		// Sanity check on the parameterisation: a future asset kind using a
		// different pathPrefix / kindLabel should see them in its errors.
		err := validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
			map[string]any{}, rctObjectSpecSchema, "custom.path", "Custom Schema Label")
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(Equal("custom.path is required"))
	})
})
