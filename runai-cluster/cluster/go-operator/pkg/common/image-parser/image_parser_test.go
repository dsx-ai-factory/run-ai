/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package image_parser_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	imageparser "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/common/image-parser"
	sdkcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
)

func TestImageParsing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Parse Suite")
}

var _ = Describe("common image parse", func() {
	It("parse image string", func(ctx context.Context) {
		imageObj, err := imageparser.ParseImageString(sdkcommon.GitSyncImageRef)
		Expect(err).To(BeNil())
		Expect(*imageObj.Registry).To(Equal(sdkcommon.GitSyncImageRegistry))
		Expect(*imageObj.Name).To(Equal(sdkcommon.GitSyncImageName))
		Expect(*imageObj.Tag).To(Equal(sdkcommon.GitSyncImageTag))
		Expect(imageObj.PullPolicy).To(Equal(v1.PullAlways))
	})

	It("parse image string 2", func(ctx context.Context) {
		imageStr := "gcr.io/run-ai-lab/init-ca-base:0.0.0-1000157.v217-run-19582-4d6ccce"
		imageObj, err := imageparser.ParseImageString(imageStr)
		Expect(err).To(BeNil())
		Expect(*imageObj.Registry).To(Equal("gcr.io"))
		Expect(*imageObj.Name).To(Equal("run-ai-lab/init-ca-base"))
		Expect(*imageObj.Tag).To(Equal("0.0.0-1000157.v217-run-19582-4d6ccce"))
		Expect(imageObj.PullPolicy).To(Equal(v1.PullAlways))
	})
})
