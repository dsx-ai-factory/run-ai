/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("checkCRD", func() {
	forbiddenErr := apierrors.NewForbidden(schema.GroupResource{Group: runv1.GroupVersion.Group, Resource: "runaijobs"}, "", errors.New("denied"))
	timeoutErr := errors.New("timeout")

	DescribeTable("reports CRD availability without hiding API errors",
		func(listErr error, wantAvailable bool, wantErr error) {
			available, err := checkCRD(testRunaiJobLister{err: listErr})
			Expect(available).To(Equal(wantAvailable))
			if wantErr == nil {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(wantErr))
			}
		},
		Entry("available", nil, true, nil),
		Entry("absent", apierrors.NewNotFound(schema.GroupResource{Group: runv1.GroupVersion.Group, Resource: "runaijobs"}, ""), false, nil),
		Entry("forbidden", forbiddenErr, false, forbiddenErr),
		Entry("other error", timeoutErr, false, timeoutErr),
	)
})

type testRunaiJobLister struct {
	err error
}

func (l testRunaiJobLister) List(context.Context, metav1.ListOptions) (*runv1.RunaiJobList, error) {
	return &runv1.RunaiJobList{}, l.err
}
