/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
)

var _ = Describe("arrayMergeByKey", func() {
	It("returns the target unchanged when src is empty", func() {
		opts := newSetMergeOptions(true, &RuntimeArrayKey{SpecField: "name"})
		want := map[string]any{"name": "a", "v": 1}
		target := reflect.ValueOf([]any{want})
		src := reflect.ValueOf([]any{})
		out, err := arrayMergeByKey(target, src, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Pointer()).To(Equal(target.Pointer()))
		Expect(out.Len()).To(Equal(1))
		Expect(out.Index(0).Interface()).To(Equal(want))
	})

	It("returns src when target is empty", func() {
		opts := newSetMergeOptions(true, &RuntimeArrayKey{SpecField: "name"})
		want := map[string]any{"name": "a", "v": 1}
		target := reflect.ValueOf([]any{})
		src := reflect.ValueOf([]any{want})
		out, err := arrayMergeByKey(target, src, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Len()).To(Equal(1))
		Expect(out.Index(0).Interface()).To(Equal(want))
	})

	It("does not overwrite a matched element when Overwrite is false", func() {
		opts := newSetMergeOptions(false, &RuntimeArrayKey{SpecField: "name"})
		want := map[string]any{"name": "x", "n": 1}
		target := reflect.ValueOf([]any{want})
		src := reflect.ValueOf([]any{map[string]any{"name": "x", "n": 2}})
		out, err := arrayMergeByKey(target, src, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Index(0).Interface()).To(Equal(want))
	})

	It("overwrites a matched element when Overwrite is true", func() {
		opts := newSetMergeOptions(true, &RuntimeArrayKey{SpecField: "name"})
		want := map[string]any{"name": "x", "n": 2}
		target := reflect.ValueOf([]any{map[string]any{"name": "x", "n": 1}})
		src := reflect.ValueOf([]any{want})
		out, err := arrayMergeByKey(target, src, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Index(0).Interface()).To(Equal(want))
	})

	It("replaces target with src when the merge key path is empty", func() {
		opts := newSetMergeOptions(false, &RuntimeArrayKey{})
		want := map[string]any{"k": 2}
		target := reflect.ValueOf([]any{map[string]any{"k": 1}})
		src := reflect.ValueOf([]any{want})
		out, err := arrayMergeByKey(target, src, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Len()).To(Equal(1))
		Expect(out.Index(0).Interface()).To(Equal(want))
	})

	It("returns an error when arrayKey context is missing from options", func() {
		opts := conjungo.NewOptions()
		opts.Overwrite = true
		opts.Context = context.Background()
		target := reflect.ValueOf([]any{map[string]any{"name": "a"}})
		src := reflect.ValueOf([]any{map[string]any{"name": "b"}})
		_, err := arrayMergeByKey(target, src, opts)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("missing arrayKeyCtx"))
	})
})
