/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package httputils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
)

type fakeEnum string

const (
	fake1 fakeEnum = "fake1"
)

var (
	fakeEnum1 = fake1

	supportEnum = true
)

type fakeRequest struct {
	fakeEnum   fakeEnum
	justString string
}

var _ = Describe("Request", func() {
	Context("SetOrderFunc", func() {
		It("nil", func() {
			var request *fakeRequest
			httputils.SetOrderFunc(request, sortOrder)
			Expect(request).To(BeNil())
		})

		It("sort is set", func() {
			request := &fakeRequest{}
			httputils.SetOrderFunc(request, sortOrder)
			Expect(request).ToNot(BeNil())
			Expect(request.justString).To(Equal("asc"))
		})
	})

	Context("SetSortByFunc", func() {
		It("nil", func() {
			var request *fakeRequest
			err := httputils.SetSortByFunc(request, sortBy, setter)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("request is nil"))
		})

		It("setter does not familiar with enum value", func() {
			supportEnum = false
			request := fakeRequest{}
			err := httputils.SetSortByFunc(&request, sortBy, setter)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not found"))
		})

		It("setter return new object", func() {
			supportEnum = true
			request := &fakeRequest{justString: "shouldBeRemoved"}
			err := httputils.SetSortByFunc(request, sortBy, setter)
			Expect(err).ToNot(HaveOccurred())
			// new object does not update justString
			Expect(request.justString).To(BeEmpty())
			// new object has the fakeEnum set
			Expect(request.fakeEnum).To(Equal(fakeEnum1))
		})
	})
})

func sortBy(fake fakeEnum) fakeRequest {
	req := fakeRequest{}
	req.fakeEnum = fake
	return req
}

func setter(value string) (*fakeEnum, error) {
	if supportEnum && value == "createdAt" {
		return &fakeEnum1, nil
	}
	return nil, fmt.Errorf("not found")
}

func sortOrder(sortOrder string) fakeRequest {
	return fakeRequest{justString: sortOrder}
}
