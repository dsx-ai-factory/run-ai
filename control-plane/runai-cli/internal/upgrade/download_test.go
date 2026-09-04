/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	upgrademocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/upgrade"
	systems "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/os"
)

var _ = Describe("Download", func() {
	var (
		mockCtrl            *gomock.Controller
		mockInstallerGetter *upgrademocks.MockInstallerGetter
		mockBinaryGetter    *upgrademocks.MockBinaryGetter
		scriptDownloader    *ScriptDownloader
		binaryDownloader    *BinaryDownloader
		ctx                 context.Context
		dir                 string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockInstallerGetter = upgrademocks.NewMockInstallerGetter(mockCtrl)
		mockBinaryGetter = upgrademocks.NewMockBinaryGetter(mockCtrl)
		scriptDownloader = NewScriptDownloader(mockInstallerGetter)
		binaryDownloader = NewBinaryDownloader(mockBinaryGetter, systems.Linux, systems.AMD64)
		ctx = context.Background()
		var err error
		dir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	Context("ScriptDownloader", func() {
		It("should return error when GetInstaller fails", func() {
			mockInstallerGetter.EXPECT().GetInstallerUnix(gomock.Any()).Return("", errors.New("error"))
			_, err := scriptDownloader.Download(ctx, dir, "fileName")
			Expect(err).To(HaveOccurred())
		})

		It("should return fullPath when Download is successful", func() {
			mockInstallerGetter.EXPECT().GetInstallerUnix(gomock.Any()).Return("scriptData", nil)
			fullPath, err := scriptDownloader.Download(ctx, dir, "fileName")
			Expect(err).NotTo(HaveOccurred())
			Expect(fullPath).To(Equal(path.Join(dir, "fileName")))
			fileObj, err := os.Open(fullPath)
			Expect(err).NotTo(HaveOccurred())
			defer fileObj.Close()
			data := make([]byte, 100)
			n, err := fileObj.Read(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(10))
			Expect(string(data[:n])).To(Equal("scriptData"))

		})

	})

	Context("BinaryDownloader", func() {
		It("should return error when GetBinary fails", func() {
			mockBinaryGetter.EXPECT().GetBinary(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			_, err := binaryDownloader.Download(ctx, dir, "fileName")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when binaryData is nil", func() {
			mockBinaryGetter.EXPECT().GetBinary(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			_, err := binaryDownloader.Download(ctx, dir, "fileName")
			Expect(err).To(HaveOccurred())
		})

		It("should return fullPath when Download is successful", func() {

			file, err := os.CreateTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(file.Name())

			mockBinaryGetter.EXPECT().GetBinary(gomock.Any(), gomock.Any(), gomock.Any()).Return(file, nil)
			fullPath, err := binaryDownloader.Download(ctx, dir, "fileName")
			Expect(err).NotTo(HaveOccurred())

			Expect(fullPath).To(Equal(path.Join(dir, "fileName")))
			fileObj, err := os.Open(fullPath)
			Expect(err).NotTo(HaveOccurred())
			defer fileObj.Close()

		})
	})
})

var _ = Describe("Download", func() {
	var (
		filePath string
		body     io.ReadCloser
	)

	BeforeEach(func() {
		filePath = "/tmp/testfile"
		body = io.NopCloser(bytes.NewReader([]byte("test data")))
	})

	AfterEach(func() {
		os.Remove(filePath)
	})

	It("should return error when file creation fails", func() {
		filePath = "/invalid/path/testfile"
		err := writeToFile(filePath, body)
		Expect(err).To(HaveOccurred())
	})

	It("should return error when writing to file fails", func() {
		body = io.NopCloser(&errorReader{})
		err := writeToFile(filePath, body)
		Expect(err).To(HaveOccurred())
	})

	It("should write to file successfully", func() {
		err := writeToFile(filePath, body)
		Expect(err).NotTo(HaveOccurred())
		data, readErr := os.ReadFile(filePath)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("test data"))
	})
})

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced error")
}
