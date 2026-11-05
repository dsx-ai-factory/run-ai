/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"errors"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("getPodsOfFunctionVersion", func() {
	// A failed list is not "this function version has no pods": the callers
	// report the error, and they used to replace it with a misleading
	// "no pods found for function version" instead.
	It("returns the list error instead of reporting no pods", func() {
		listErr := errors.New("connection refused")
		cachedClient := fake.NewClientBuilder().
			WithScheme(scheme_init.Scheme()).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
					return listErr
				},
			}).
			Build()

		pods, err := getPodsOfFunctionVersion(context.Background(), "function-version-id", "nvcf-ns", cachedClient)

		Expect(err).To(MatchError(listErr))
		Expect(pods).To(BeEmpty())
	})

	It("returns no pods and no error when the function version id is empty", func() {
		pods, err := getPodsOfFunctionVersion(context.Background(), "", "nvcf-ns", nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(BeEmpty())
	})
})
