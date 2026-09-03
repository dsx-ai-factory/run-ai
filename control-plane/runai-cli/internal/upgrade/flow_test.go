/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"context"
	"errors"
	"os"

	upgrademocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/upgrade"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BasicFlow", func() {
	var (
		mockCtrl       *gomock.Controller
		mockDownloader *upgrademocks.MockDownloader
		mockInstaller  *upgrademocks.MockInstaller
		basicFlow      *BasicFlow
		ctx            context.Context
		execShell      string
	)

	BeforeEach(func() {
		err := config.GetInstance().InitLogger(true)
		Expect(err).NotTo(HaveOccurred())
		mockCtrl = gomock.NewController(GinkgoT())
		mockDownloader = upgrademocks.NewMockDownloader(mockCtrl)
		mockInstaller = upgrademocks.NewMockInstaller(mockCtrl)

		basicFlow = NewBasicFlow(mockDownloader, mockInstaller, "", "fileName")
		ctx = context.Background()
		execShell = "bash"
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should return error when Download fails", func() {
		mockDownloader.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("error"))
		installOutput, err := basicFlow.Upgrade(ctx, execShell)
		Expect(err).To(HaveOccurred())
		Expect(installOutput).To(Equal(""))
	})

	It("should return error when Install fails", func() {
		mockDownloader.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any()).Return("path", nil)
		mockInstaller.EXPECT().Install(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("error"))
		installOutput, err := basicFlow.Upgrade(ctx, execShell)
		Expect(err).To(HaveOccurred())
		Expect(installOutput).To(Equal(""))
	})

	It("should Upgrade successfully without cache file", func() {
		mockDownloader.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any()).Return("path", nil)
		mockInstaller.EXPECT().Install(gomock.Any(), gomock.Any(), gomock.Any()).Return("123", nil)

		installOutput, err := basicFlow.Upgrade(ctx, execShell)
		Expect(err).NotTo(HaveOccurred())
		Expect(installOutput).NotTo(Equal(""))
	})

	It("should Upgrade successfully with cache file", func() {
		mockDownloader.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any()).Return("path", nil)
		mockInstaller.EXPECT().Install(gomock.Any(), gomock.Any(), gomock.Any()).Return("123", nil)

		tmpDir, err := os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
		fileObj, err := os.CreateTemp(tmpDir, "test_cache.json")
		Expect(err).NotTo(HaveOccurred())
		cacheFile := fileObj.Name()
		localBasicFlow := NewBasicFlow(mockDownloader, mockInstaller, cacheFile, "fileName")
		installOutput, err := localBasicFlow.Upgrade(ctx, execShell)
		Expect(err).NotTo(HaveOccurred())
		Expect(installOutput).NotTo(Equal(""))
		_, err = os.Stat(cacheFile)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
